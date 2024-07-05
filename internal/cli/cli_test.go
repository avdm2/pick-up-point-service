package cli

import (
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"homework-1/internal/models"
	"homework-1/internal/module"
	mockmodule "homework-1/internal/module/mocks"
	"homework-1/internal/services/packaging"
	"homework-1/internal/storage"
	"strconv"
	"testing"
	"time"
)

func TestCLI_AddOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockModule := mockmodule.NewMockModuleInterface(ctrl)
	c := NewCLI(Deps{Module: mockModule})

	t.Run("[OK] cli.AddOrder", func(t *testing.T) {
		t.Parallel()

		orderID := "100"
		customerID := "100"
		expirationDate := "10-10-2024"
		pack := "box"
		weight := "1"
		cost := "1"

		mockModule.EXPECT().AddOrder(models.ID(100), models.ID(100), gomock.Any(), models.PackageType(pack), models.Kilo(1), models.Rub(1)).Return(nil)

		err := c.addOrder([]string{orderID, customerID, expirationDate, pack, weight, cost})
		require.NoError(t, err)
	})

	t.Run("[X] cli.AddOrder", func(t *testing.T) {
		t.Parallel()

		orderID := "1"
		customerID := "1"
		expirationDate := "10-10-2024"
		pack := "box"
		weight := "1"
		cost := "1"

		mockModule.EXPECT().AddOrder(models.ID(1), models.ID(1), gomock.Any(), models.PackageType(pack), models.Kilo(1), models.Rub(1)).Return(storage.ErrOrderExists)

		err := c.addOrder([]string{orderID, customerID, expirationDate, pack, weight, cost})
		require.Error(t, err)
		assert.Contains(t, err.Error(), storage.ErrOrderExists.Error())
	})
}

func TestCLI_ReturnOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockModule := mockmodule.NewMockModuleInterface(ctrl)
	c := CLI{Deps: Deps{Module: mockModule}}

	t.Run("[OK] cli.ReturnOrder", func(t *testing.T) {
		t.Parallel()

		orderID := "1"

		mockModule.EXPECT().ReturnOrder(models.ID(1)).Return(nil)

		err := c.returnOrder([]string{orderID})
		require.NoError(t, err)
	})

	t.Run("[X] cli.ReturnOrder", func(t *testing.T) {
		t.Parallel()

		orderID := "1"
		errReturn := fmt.Errorf("cli.returnOrder error: %w", module.ErrReturn)

		mockModule.EXPECT().ReturnOrder(models.ID(1)).Return(errReturn)

		err := c.returnOrder([]string{orderID})
		require.Error(t, err)
		assert.Contains(t, err.Error(), errReturn.Error())
	})
}

func TestCLI_ReceiveOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockModule := mockmodule.NewMockModuleInterface(ctrl)
	c := CLI{Deps: Deps{Module: mockModule}}

	t.Run("[OK] cli.ReceiveOrder", func(t *testing.T) {
		t.Parallel()

		orderID := "100"
		customerID := models.ID(100)
		expirationDate := time.Now().Add(time.Hour)
		pack := models.PackageType("box")
		weight := models.Kilo(10)
		cost := models.Rub(1000)
		p, _ := packaging.ParsePackage(pack)

		order := models.Order{
			OrderID:            models.ID(100),
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

		mockModule.EXPECT().ReceiveOrders([]models.ID{models.ID(100)}).Return([]models.Order{order}, nil)

		err := c.receiveOrder([]string{orderID})
		require.NoError(t, err)
	})
}

func TestCLI_GetOrders(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockModule := mockmodule.NewMockModuleInterface(ctrl)
	c := CLI{Deps: Deps{Module: mockModule}}

	t.Run("[OK] cli.GetOrders", func(t *testing.T) {
		t.Parallel()

		customerID := "1"
		orders := []models.Order{
			{OrderID: models.ID(1)},
			{OrderID: models.ID(2)},
		}

		mockModule.EXPECT().GetOrders(models.ID(1), 2).Return(orders, nil)

		err := c.getOrders([]string{customerID, "2"})
		require.NoError(t, err)
	})
}

func TestCLI_CreateRefund(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockModule := mockmodule.NewMockModuleInterface(ctrl)
	c := CLI{Deps: Deps{Module: mockModule}}

	t.Run("[OK] cli.CreateRefund", func(t *testing.T) {
		t.Parallel()

		orderID := "1"
		customerID := "1"

		mockModule.EXPECT().RefundOrder(models.ID(1), models.ID(1)).Return(nil)

		err := c.createRefund([]string{customerID, orderID})
		require.NoError(t, err)
	})

	t.Run("[X] cli.CreateRefund", func(t *testing.T) {
		t.Parallel()

		orderID := "1"
		customerID := "1"
		errRefund := fmt.Errorf("cli.createRefund error: %w", module.ErrRefund)

		mockModule.EXPECT().RefundOrder(models.ID(1), models.ID(1)).Return(errRefund)

		err := c.createRefund([]string{customerID, orderID})
		require.Error(t, err)
		assert.Contains(t, err.Error(), errRefund.Error())
	})
}

func TestCLI_GetRefunds(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockModule := mockmodule.NewMockModuleInterface(ctrl)
	c := CLI{Deps: Deps{Module: mockModule}}

	t.Run("[OK] cli.GetRefunds", func(t *testing.T) {
		t.Parallel()

		page := "0"
		limit := "2"
		refunds := []models.Order{
			{OrderID: models.ID(1)},
			{OrderID: models.ID(2)},
		}
		pageInt, _ := strconv.Atoi(page)
		limitInt, _ := strconv.Atoi(limit)

		mockModule.EXPECT().GetRefunds(pageInt, limitInt).Return(refunds, nil)

		err := c.getRefunds([]string{page, limit})
		require.NoError(t, err)
	})
}
