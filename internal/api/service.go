package api

import (
	"context"
	"errors"
	"fmt"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"homework-1/internal/models"
	"homework-1/internal/module"
	"homework-1/pkg/api/proto/orders_grpc/v1/orders_grpc/v1"
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
}

func (o *OrderService) AddOrder(ctx context.Context, request *orders_grpc.AddOrderRequest) (*emptypb.Empty, error) {
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

	return &emptypb.Empty{}, nil
}

func (o *OrderService) ReturnOrder(ctx context.Context, request *orders_grpc.ReturnOrderRequest) (*emptypb.Empty, error) {
	orderId := models.ID(request.GetOrderId())
	if orderId <= 0 {
		return nil, fmt.Errorf("OrderService.ReturnOrder error: %w", errIncorrectId)
	}

	if errReturn := o.Module.ReturnOrder(orderId); errReturn != nil {
		return nil, fmt.Errorf("OrderService.ReturnOrder error: %w", errReturn)
	}

	return &emptypb.Empty{}, nil
}

func (o *OrderService) ReceiveOrders(ctx context.Context, request *orders_grpc.ReceiveOrdersRequest) (*orders_grpc.ReceiveOrdersResponse, error) {
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

	response := &orders_grpc.ReceiveOrdersResponse{}
	for _, order := range orders {
		response.Orders = append(response.Orders, &orders_grpc.Order{
			OrderId:        int64(order.OrderID),
			CustomerId:     int64(order.CustomerID),
			ExpirationDate: timestamppb.New(order.ExpirationTime),
			PackageType:    string(order.Package),
			Weight:         float64(order.Weight),
			Cost:           float64(order.Cost),
		})
	}

	return response, nil
}

func (o *OrderService) GetOrders(ctx context.Context, request *orders_grpc.GetOrdersRequest) (*orders_grpc.GetOrdersResponse, error) {
	orders, err := o.Module.GetOrders(models.ID(request.GetCustomerId()), int(request.GetN()))
	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}

	response := &orders_grpc.GetOrdersResponse{}
	for _, order := range orders {
		response.Orders = append(response.Orders, &orders_grpc.Order{
			OrderId:        int64(order.OrderID),
			CustomerId:     int64(order.CustomerID),
			ExpirationDate: timestamppb.New(order.ExpirationTime),
			PackageType:    string(order.Package),
			Weight:         float64(order.Weight),
			Cost:           float64(order.Cost),
		})
	}

	return response, nil
}

func (o *OrderService) CreateRefund(ctx context.Context, request *orders_grpc.CreateRefundRequest) (*emptypb.Empty, error) {
	orderId := models.ID(request.GetOrderId())
	customerId := models.ID(request.GetCustomerId())
	if orderId <= 0 || customerId <= 0 {
		return nil, fmt.Errorf("OrderService.CreateRefund error: %w", errIncorrectId)
	}

	if errRefund := o.Module.RefundOrder(customerId, orderId); errRefund != nil {
		return nil, fmt.Errorf("OrderService.CreateRefund error: %w", errRefund)
	}

	return &emptypb.Empty{}, nil
}

func (o *OrderService) GetRefunds(ctx context.Context, request *orders_grpc.GetRefundsRequest) (*orders_grpc.GetRefundsResponse, error) {
	refunds, err := o.Module.GetRefunds(int(request.GetPage()), int(request.GetLimit()))
	if err != nil {
		return nil, fmt.Errorf("OrderService.GetRefunds error: %w", err)
	}

	response := &orders_grpc.GetRefundsResponse{}
	for _, refund := range refunds {
		response.Refunds = append(response.Refunds, &orders_grpc.Order{
			OrderId:        int64(refund.OrderID),
			CustomerId:     int64(refund.CustomerID),
			ExpirationDate: timestamppb.New(refund.ExpirationTime),
			PackageType:    string(refund.Package),
			Weight:         float64(refund.Weight),
			Cost:           float64(refund.Cost),
		})
	}

	return response, nil
}
