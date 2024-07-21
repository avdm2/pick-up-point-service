//go:build integration
// +build integration

package api

import (
	"context"
	"fmt"
	"homework-1/internal/module"
	"homework-1/internal/storage"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	mockcache "homework-1/internal/cache/mocks"
	"homework-1/internal/models"
	mockmodule "homework-1/internal/module/mocks"
	"homework-1/pkg/api/proto/orders_grpc/v1/orders_grpc/v1"
)

func TestOrderService_AddOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockModule := mockmodule.NewMockModuleInterface(ctrl)
	mockCache := mockcache.NewMockCacheInterface(ctrl)
	orderService := &OrderService{Module: mockModule, Redis: mockCache}

	t.Run("Успешное добавление заказа", func(t *testing.T) {
		request := &orders_grpc.AddOrderRequest{
			OrderId:        100,
			CustomerId:     100,
			ExpirationTime: "10-10-2024",
			PackageType:    "box",
			Weight:         1,
			Cost:           1,
		}

		expirationDate, _ := time.Parse(dateLayout, request.ExpirationTime)

		mockModule.EXPECT().AddOrder(models.ID(100), models.ID(100), expirationDate, models.PackageType("box"), models.Kilo(1), models.Rub(1)).Return(nil)
		mockCache.EXPECT().Delete(gomock.Any(), fmt.Sprintf("getOrders_%d", request.CustomerId)).Return(nil)

		_, err := orderService.AddOrder(context.Background(), request)
		require.NoError(t, err)
	})

	t.Run("Попытка добавить заказ с существующим ID", func(t *testing.T) {
		request := &orders_grpc.AddOrderRequest{
			OrderId:        1,
			CustomerId:     1,
			ExpirationTime: "10-10-2024",
			PackageType:    "box",
			Weight:         1,
			Cost:           1,
		}

		expirationDate, _ := time.Parse(dateLayout, request.ExpirationTime)
		mockModule.EXPECT().AddOrder(models.ID(1), models.ID(1), expirationDate, models.PackageType("box"), models.Kilo(1), models.Rub(1)).Return(storage.ErrOrderExists)

		_, err := orderService.AddOrder(context.Background(), request)
		require.Error(t, err)
		assert.Contains(t, err.Error(), storage.ErrOrderExists.Error())
	})
}

func TestOrderService_ReturnOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockModule := mockmodule.NewMockModuleInterface(ctrl)
	mockCache := mockcache.NewMockCacheInterface(ctrl)
	orderService := &OrderService{Module: mockModule, Redis: mockCache}

	t.Run("Успешный возврат товара курьеру", func(t *testing.T) {
		request := &orders_grpc.ReturnOrderRequest{
			OrderId: 1,
		}

		order := models.Order{CustomerID: models.ID(1)}

		mockModule.EXPECT().ReturnOrder(models.ID(1)).Return(order, nil)
		mockCache.EXPECT().Delete(gomock.Any(), fmt.Sprintf("getOrders_%d", order.CustomerID)).Return(nil)

		_, err := orderService.ReturnOrder(context.Background(), request)
		require.NoError(t, err)
	})

	t.Run("Попытка вернуть курьеру ранее забранный заказ", func(t *testing.T) {
		request := &orders_grpc.ReturnOrderRequest{
			OrderId: 1,
		}

		mockModule.EXPECT().ReturnOrder(models.ID(1)).Return(models.Order{}, module.ErrReturn)

		_, err := orderService.ReturnOrder(context.Background(), request)
		require.Error(t, err)
		assert.Contains(t, err.Error(), module.ErrReturn.Error())
	})
}

func TestOrderService_ReceiveOrders(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockModule := mockmodule.NewMockModuleInterface(ctrl)
	mockCache := mockcache.NewMockCacheInterface(ctrl)
	orderService := &OrderService{Module: mockModule, Redis: mockCache}

	t.Run("Успешное получение заказа", func(t *testing.T) {
		request := &orders_grpc.ReceiveOrdersRequest{
			OrderIds: []int64{100},
		}

		customerID := models.ID(100)
		expirationDate := time.Now().Add(time.Hour)
		packageType := models.PackageType("box")
		weight := models.Kilo(10)
		cost := models.Rub(1000)
		order := models.Order{
			OrderID:            models.ID(100),
			CustomerID:         customerID,
			ExpirationTime:     expirationDate,
			ReceivedByCustomer: false,
			Refunded:           false,
			Package:            packageType,
			Weight:             weight,
			Cost:               cost,
			PackageCost:        100,
		}

		mockModule.EXPECT().ReceiveOrders([]models.ID{models.ID(100)}).Return([]models.Order{order}, nil)
		mockCache.EXPECT().Delete(gomock.Any(), fmt.Sprintf("getOrders_%d", order.CustomerID)).Return(nil)

		response, err := orderService.ReceiveOrders(context.Background(), request)
		require.NoError(t, err)
		assert.Len(t, response.Orders, 1)
		assert.Equal(t, response.Orders[0].OrderId, int64(order.OrderID))
	})
}

func TestOrderService_GetOrders(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockModule := mockmodule.NewMockModuleInterface(ctrl)
	mockCache := mockcache.NewMockCacheInterface(ctrl)
	orderService := &OrderService{Module: mockModule, Redis: mockCache}

	t.Run("Успешное получение списка заказов", func(t *testing.T) {
		request := &orders_grpc.GetOrdersRequest{
			CustomerId: 1,
			N:          2,
		}

		orders := []models.Order{
			{OrderID: models.ID(1)},
			{OrderID: models.ID(2)},
		}

		mockCache.EXPECT().Get(gomock.Any(), fmt.Sprintf("getOrders_%d", request.CustomerId)).Return(nil, false)
		mockModule.EXPECT().GetOrders(models.ID(1), 2).Return(orders, nil)
		mockCache.EXPECT().Set(gomock.Any(), fmt.Sprintf("getOrders_%d", request.CustomerId), orders, gomock.Any()).Return(nil)

		response, err := orderService.GetOrders(context.Background(), request)
		require.NoError(t, err)
		assert.Len(t, response.Orders, 2)
	})
}

func TestOrderService_CreateRefund(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockModule := mockmodule.NewMockModuleInterface(ctrl)
	mockCache := mockcache.NewMockCacheInterface(ctrl)
	orderService := &OrderService{Module: mockModule, Redis: mockCache}

	t.Run("Успешный возврат товара", func(t *testing.T) {
		request := &orders_grpc.CreateRefundRequest{
			OrderId:    1,
			CustomerId: 1,
		}

		mockModule.EXPECT().RefundOrder(models.ID(1), models.ID(1)).Return(nil)
		mockCache.EXPECT().Delete(gomock.Any(), fmt.Sprintf("getOrders_%d", request.CustomerId)).Return(nil)

		_, err := orderService.CreateRefund(context.Background(), request)
		require.NoError(t, err)
	})

	t.Run("Попытка вернуть ранее возвращенный товар", func(t *testing.T) {
		request := &orders_grpc.CreateRefundRequest{
			OrderId:    1,
			CustomerId: 1,
		}

		mockModule.EXPECT().RefundOrder(models.ID(1), models.ID(1)).Return(module.ErrRefund)

		_, err := orderService.CreateRefund(context.Background(), request)
		require.Error(t, err)
		assert.Contains(t, err.Error(), module.ErrRefund.Error())
	})
}

func TestOrderService_GetRefunds(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockModule := mockmodule.NewMockModuleInterface(ctrl)
	mockCache := mockcache.NewMockCacheInterface(ctrl)
	orderService := &OrderService{Module: mockModule, Redis: mockCache}

	t.Run("Успешное получение списка возвратов", func(t *testing.T) {
		request := &orders_grpc.GetRefundsRequest{
			Page:  0,
			Limit: 2,
		}

		refunds := []models.Order{
			{OrderID: models.ID(1)},
			{OrderID: models.ID(2)},
		}

		mockCache.EXPECT().Get(gomock.Any(), fmt.Sprintf("getRefunds_p%d_l%d", request.Page, request.Limit)).Return(nil, false)
		mockModule.EXPECT().GetRefunds(int(request.Page), int(request.Limit)).Return(refunds, nil)
		mockCache.EXPECT().Set(gomock.Any(), fmt.Sprintf("getRefunds_p%d_l%d", request.Page, request.Limit), refunds, gomock.Any()).Return(nil)

		response, err := orderService.GetRefunds(context.Background(), request)
		require.NoError(t, err)
		assert.Len(t, response.Refunds, 2)
	})

	t.Run("Ошибка получения списка возвратов из кеша", func(t *testing.T) {
		request := &orders_grpc.GetRefundsRequest{
			Page:  0,
			Limit: 2,
		}

		mockCache.EXPECT().Get(gomock.Any(), fmt.Sprintf("getRefunds_p%d_l%d", request.Page, request.Limit)).Return(nil, false)
		mockModule.EXPECT().GetRefunds(int(request.Page), int(request.Limit)).Return(nil, fmt.Errorf("err"))

		_, err := orderService.GetRefunds(context.Background(), request)
		require.Error(t, err)
	})
}
