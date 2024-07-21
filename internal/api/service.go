package api

import (
	"context"
	"errors"
	"fmt"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"homework-1/internal/cache"
	"homework-1/internal/metrics"
	"homework-1/internal/models"
	"homework-1/internal/module"
	"homework-1/pkg/api/proto/orders_grpc/v1/orders_grpc/v1"
	"log"
	"time"
)

const (
	dateLayout = "02-01-2006"
)

var (
	errIncorrectId    = errors.New("empty or non-positive order or customer id")
	errNegativeWeight = errors.New("weight can not be negative")
	errNegativeCost   = errors.New("cost can not be negative")
)

type OrderService struct {
	Module module.ModuleInterface
	orders_grpc.UnimplementedOrdersServiceServer
	Redis cache.CacheInterface
}

// AddOrder Инвалидация кеша происходит на этапе успешного добавления заказа.
// В этом случае из кеша удаляется ключ, содержащий информацию о всех заказах пользователя, для которого был создан новый заказ.
// Это сделано, потому что после добавления нового заказа информация о заказах пользователя в кеше устаревает (добавялется новый заказ).
func (o *OrderService) AddOrder(ctx context.Context, request *orders_grpc.AddOrderRequest) (*emptypb.Empty, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.OrderService.AddOrder")
	defer span.Finish()

	expirationTime, errDate := time.Parse(dateLayout, request.GetExpirationTime())
	if errDate != nil {
		return nil, fmt.Errorf("OrderService.AddOrder error: %w", errDate)
	}

	orderId := models.ID(request.GetOrderId())
	customerId := models.ID(request.GetCustomerId())
	if orderId <= 0 || customerId <= 0 {
		return nil, fmt.Errorf("OrderService.AddOrder error: %w", errIncorrectId)
	}

	weight := models.Kilo(request.GetWeight())
	if weight < 0 {
		return nil, fmt.Errorf("OrderService.AddOrder error: %w", errNegativeWeight)
	}

	cost := models.Rub(request.GetCost())
	if cost < 0 {
		return nil, fmt.Errorf("OrderService.AddOrder error: %w", errNegativeCost)
	}

	packageType := models.PackageType(request.GetPackageType())
	if errAdd := o.Module.AddOrder(orderId, customerId, expirationTime, packageType, weight, cost); errAdd != nil {
		return nil, fmt.Errorf("OrderService.AddOrder error: %w", errAdd)
	}

	metrics.IncAddedOrders(1)

	if errCache := o.Redis.Delete(ctx, fmt.Sprintf("getOrders_%d", customerId)); errCache != nil {
		return nil, fmt.Errorf("OrderService.AddOrder error clearing cache: %w", errCache)
	}

	return &emptypb.Empty{}, nil
}

// ReturnOrder Инвалидация кеша происходит на этапе успешного возврата заказа курьеру (или его удаление из базы).
// В этом случае из кеша удаляется ключ, содержащий информацию о всех заказах пользователя, для которого был возвращен заказ.
// Это сделано, потому что после добавления возврата заказа информация о заказах пользователя в кеше устаревает (заказ удаляется).
func (o *OrderService) ReturnOrder(ctx context.Context, request *orders_grpc.ReturnOrderRequest) (*emptypb.Empty, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.OrderService.ReturnOrder")
	defer span.Finish()

	orderId := models.ID(request.GetOrderId())
	if orderId <= 0 {
		return nil, fmt.Errorf("OrderService.ReturnOrder error: %w", errIncorrectId)
	}

	order, errReturn := o.Module.ReturnOrder(orderId)
	if errReturn != nil {
		return nil, fmt.Errorf("OrderService.ReturnOrder error: %w", errReturn)
	}

	if errCache := o.Redis.Delete(ctx, fmt.Sprintf("getOrders_%d", order.CustomerID)); errCache != nil {
		return nil, fmt.Errorf("OrderService.ReturnOrder error: %w", errCache)
	}

	return &emptypb.Empty{}, nil
}

// ReceiveOrders Инвалидация кеша происходит на этапе успешного получения заказа.
// В этом случае из кеша удаляется ключ, содержащий информацию о всех заказах пользователя, для которого был получен заказ.
// Это сделано, потому что после получения заказа информация о заказах пользователя в кеше устаревает (изменяется поле Received).
func (o *OrderService) ReceiveOrders(ctx context.Context, request *orders_grpc.ReceiveOrdersRequest) (*orders_grpc.ReceiveOrdersResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.OrderService.ReceiveOrders")
	defer span.Finish()

	ids := make([]models.ID, len(request.GetOrderIds()))
	for i, id := range request.GetOrderIds() {
		if id <= 0 {
			return nil, fmt.Errorf("OrderService.ReceiveOrders error: %w", errIncorrectId)
		}

		ids[i] = models.ID(id)
	}

	orders, err := o.Module.ReceiveOrders(ids)
	if err != nil {
		return nil, fmt.Errorf("OrderService.ReceiveOrders error: %w", err)
	}

	if errCache := o.Redis.Delete(ctx, fmt.Sprintf("getOrders_%d", orders[0].CustomerID)); errCache != nil {
		return nil, fmt.Errorf("OrderService.ReceiveOrders error: %w", errCache)
	}

	response := &orders_grpc.ReceiveOrdersResponse{}
	for _, order := range orders {
		response.Orders = append(response.Orders, &orders_grpc.Order{
			OrderId:        int64(order.OrderID),
			CustomerId:     int64(order.CustomerID),
			ExpirationTime: timestamppb.New(order.ExpirationTime),
			Received:       order.ReceivedByCustomer,
			Refunded:       order.Refunded,
			PackageType:    string(order.Package),
			Weight:         float64(order.Weight),
			Cost:           float64(order.Cost),
			PackCost:       float64(order.PackageCost),
		})
	}

	metrics.IncReceivedOrders(len(orders))

	return response, nil
}

func (o *OrderService) GetOrders(ctx context.Context, request *orders_grpc.GetOrdersRequest) (resp *orders_grpc.GetOrdersResponse, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.OrderService.GetOrders")
	defer span.Finish()

	cachedKey := fmt.Sprintf("getOrders_%d", request.GetCustomerId())

	orders, ok := o.Redis.Get(ctx, cachedKey)
	if !ok {
		log.Println("cache is empty for key", cachedKey)
		orders, err = o.Module.GetOrders(models.ID(request.GetCustomerId()), int(request.GetN()))
		if err != nil {
			return nil, fmt.Errorf("service.OrderService error: %w", err)
		}

		if len(orders) > 0 {
			if err = o.Redis.Set(ctx, cachedKey, orders, time.Now()); err != nil {
				return nil, fmt.Errorf("service.OrderService error: %w", err)
			}
		}
	}

	if ok {
		log.Println("cache is not empty for key", cachedKey)
	}

	resp = &orders_grpc.GetOrdersResponse{}
	for _, order := range orders {
		resp.Orders = append(resp.Orders, &orders_grpc.Order{
			OrderId:        int64(order.OrderID),
			CustomerId:     int64(order.CustomerID),
			ExpirationTime: timestamppb.New(order.ExpirationTime),
			Received:       order.ReceivedByCustomer,
			Refunded:       order.Refunded,
			PackageType:    string(order.Package),
			Weight:         float64(order.Weight),
			Cost:           float64(order.Cost),
			PackCost:       float64(order.PackageCost),
		})
	}

	return resp, nil
}

// CreateRefund Инвалидация кеша происходит на этапе успешного создания заявки на возврат заказа.
// В этом случае из кеша удаляется ключ, содержащий информацию о всех заказах пользователя, для которого был возвращен заказ.
// Это сделано, потому что после возврата заказа информация о заказах пользователя в кеше устаревает (изменяется поле Refunded).
func (o *OrderService) CreateRefund(ctx context.Context, request *orders_grpc.CreateRefundRequest) (*emptypb.Empty, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.OrderService.CreateRefund")
	defer span.Finish()

	orderId := models.ID(request.GetOrderId())
	customerId := models.ID(request.GetCustomerId())
	if orderId <= 0 || customerId <= 0 {
		return nil, fmt.Errorf("OrderService.CreateRefund error: %w", errIncorrectId)
	}

	if errRefund := o.Module.RefundOrder(customerId, orderId); errRefund != nil {
		return nil, fmt.Errorf("OrderService.CreateRefund error: %w", errRefund)
	}

	if errCache := o.Redis.Delete(ctx, fmt.Sprintf("getOrders_%d", customerId)); errCache != nil {
		return nil, fmt.Errorf("OrderService.CreateRefund error: %w", errCache)
	}

	metrics.IncRefundedOrders(1)

	return &emptypb.Empty{}, nil
}

func (o *OrderService) GetRefunds(ctx context.Context, request *orders_grpc.GetRefundsRequest) (resp *orders_grpc.GetRefundsResponse, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.OrderService.GetRefunds")
	defer span.Finish()

	cachedKey := fmt.Sprintf("getRefunds_p%d_l%d", request.GetPage(), request.GetLimit())

	refunds, ok := o.Redis.Get(ctx, cachedKey)
	if !ok {
		log.Println("cache is empty for key", cachedKey)
		refunds, err = o.Module.GetRefunds(int(request.GetPage()), int(request.GetLimit()))
		if err != nil {
			return nil, fmt.Errorf("OrderService.GetRefunds error: %w", err)
		}

		if len(refunds) > 0 {
			if err = o.Redis.Set(ctx, cachedKey, refunds, time.Now()); err != nil {
				return nil, fmt.Errorf("OrderService.GetRefunds error: %w", err)
			}
		}
	}

	resp = &orders_grpc.GetRefundsResponse{}
	for _, refund := range refunds {
		resp.Refunds = append(resp.Refunds, &orders_grpc.Order{
			OrderId:        int64(refund.OrderID),
			CustomerId:     int64(refund.CustomerID),
			ExpirationTime: timestamppb.New(refund.ExpirationTime),
			Received:       refund.ReceivedByCustomer,
			Refunded:       refund.Refunded,
			PackageType:    string(refund.Package),
			Weight:         float64(refund.Weight),
			Cost:           float64(refund.Cost),
			PackCost:       float64(refund.PackageCost),
		})
	}

	return resp, nil
}
