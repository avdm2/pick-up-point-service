package module

import (
	"homework-1/internal/services/packaging"
	"homework-1/internal/storage"
	mockstorage "homework-1/internal/storage/mocks"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"homework-1/internal/models"
)

func TestModule_AddOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mockstorage.NewMockStorage(ctrl)
	module := NewModule(Deps{Storage: mockStorage})

	t.Run("[OK] module.AddOrder", func(t *testing.T) {
		t.Parallel()

		orderID := models.ID(100)
		customerID := models.ID(100)
		expirationDate := time.Now().Add(time.Hour)
		pack := models.PackageType("box")
		weight := models.Kilo(10)
		cost := models.Rub(100)

		mockStorage.EXPECT().GetOrder(orderID).Return(models.Order{}, nil)
		mockStorage.EXPECT().AddOrder(gomock.Any()).Return(nil)

		err := module.AddOrder(orderID, customerID, expirationDate, pack, weight, cost)
		require.NoError(t, err)
	})

	t.Run("[X] module.AddOrder", func(t *testing.T) {
		t.Parallel()

		orderID := models.ID(1)
		customerID := models.ID(1)
		expirationDate := time.Now().Add(time.Hour)
		pack := models.PackageType("box")
		weight := models.Kilo(10)
		cost := models.Rub(100)

		existingOrder := models.Order{
			OrderID:    orderID,
			CustomerID: customerID,
		}

		mockStorage.EXPECT().GetOrder(orderID).Return(existingOrder, nil)

		err := module.AddOrder(orderID, customerID, expirationDate, pack, weight, cost)
		require.Error(t, err)
		assert.Contains(t, err.Error(), storage.ErrOrderExists.Error())
	})
}

func TestModule_ReturnOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mockstorage.NewMockStorage(ctrl)
	module := NewModule(Deps{Storage: mockStorage})

	t.Run("[OK] module.ReturnOrder", func(t *testing.T) {
		t.Parallel()

		orderID := models.ID(1)
		order := models.Order{
			OrderID:            orderID,
			ReceivedByCustomer: true,
			ExpirationTime:     time.Now().Add(-24 * time.Hour),
		}

		mockStorage.EXPECT().GetOrder(orderID).Return(order, nil)
		mockStorage.EXPECT().ReturnOrder(orderID).Return(nil)

		err := module.ReturnOrder(orderID)
		require.NoError(t, err)
	})

	t.Run("[X] module.ReturnOrder", func(t *testing.T) {
		t.Parallel()

		orderID := models.ID(1)
		order := models.Order{
			OrderID:            orderID,
			ReceivedByCustomer: false,
			ExpirationTime:     time.Now().Add(-24 * time.Hour),
		}

		mockStorage.EXPECT().GetOrder(orderID).Return(order, nil)

		err := module.ReturnOrder(orderID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), ErrReturn.Error())
	})
}

func TestModule_ReceiveOrders(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mockstorage.NewMockStorage(ctrl)
	module := NewModule(Deps{Storage: mockStorage})

	t.Run("[OK] module.ReceiveOrders", func(t *testing.T) {
		t.Parallel()

		orderID := models.ID(100)
		customerID := models.ID(100)
		expirationDate := time.Now().Add(24 * time.Hour)
		pack := models.PackageType("box")
		weight := models.Kilo(10)
		cost := models.Rub(100)
		p, _ := packaging.ParsePackage(pack)

		order := models.Order{
			OrderID:            orderID,
			CustomerID:         customerID,
			ExpirationTime:     expirationDate,
			ReceivedTime:       time.Time{},
			ReceivedByCustomer: false,
			Refunded:           false,
			Package:            pack,
			Weight:             weight,
			Cost:               cost,
			PackageCost:        p.GetCost(),
		}

		mockStorage.EXPECT().GetOrder(orderID).Return(order, nil)
		mockStorage.EXPECT().GetOrder(orderID).Return(order, nil)
		mockStorage.EXPECT().ReceiveOrder(orderID).Return(order, nil)

		receivedOrders, err := module.ReceiveOrders([]models.ID{orderID})
		require.NoError(t, err)
		assert.Equal(t, 1, len(receivedOrders))
		assert.Equal(t, orderID, receivedOrders[0].OrderID)
	})
}

func TestModule_GetOrders(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mockstorage.NewMockStorage(ctrl)
	module := NewModule(Deps{Storage: mockStorage})

	t.Run("[OK] module.GetOrders", func(t *testing.T) {
		t.Parallel()

		customerID := models.ID(1)
		orders := []models.Order{
			{OrderID: models.ID(1)},
			{OrderID: models.ID(2)},
		}

		mockStorage.EXPECT().GetCustomersOrders(customerID).Return(orders, nil)

		result, err := module.GetOrders(customerID, 2)
		require.NoError(t, err)
		assert.Equal(t, 2, len(result))
	})
}

func TestModule_RefundOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mockstorage.NewMockStorage(ctrl)
	module := NewModule(Deps{Storage: mockStorage})

	t.Run("[OK] module.RefundOrder", func(t *testing.T) {
		t.Parallel()

		orderID := models.ID(1)
		customerID := models.ID(1)
		order := models.Order{
			OrderID:            orderID,
			CustomerID:         customerID,
			ReceivedByCustomer: true,
			Refunded:           false,
			ReceivedTime:       time.Now().Add(-time.Hour),
		}

		mockStorage.EXPECT().GetOrder(orderID).Return(order, nil)
		mockStorage.EXPECT().ChangeOrder(gomock.Any()).Return(nil)

		err := module.RefundOrder(customerID, orderID)
		require.NoError(t, err)
	})

	t.Run("[X] module.RefundOrder", func(t *testing.T) {
		t.Parallel()

		orderID := models.ID(1)
		customerID := models.ID(1)
		order := models.Order{
			OrderID:            orderID,
			CustomerID:         models.ID(2),
			ReceivedByCustomer: true,
			Refunded:           false,
			ReceivedTime:       time.Now().Add(-time.Hour),
		}

		mockStorage.EXPECT().GetOrder(orderID).Return(order, nil)

		err := module.RefundOrder(customerID, orderID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), ErrRefund.Error())
	})
}

func TestModule_GetRefunds(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mockstorage.NewMockStorage(ctrl)
	module := NewModule(Deps{Storage: mockStorage})

	t.Run("[OK] module.GetRefunds", func(t *testing.T) {
		t.Parallel()

		page := 0
		limit := 2
		refunds := []models.Order{
			{OrderID: models.ID(1)},
			{OrderID: models.ID(2)},
		}

		mockStorage.EXPECT().GetRefunds().Return(refunds, nil)

		result, err := module.GetRefunds(page, limit)
		require.NoError(t, err)
		assert.Equal(t, 2, len(result))
	})
}
