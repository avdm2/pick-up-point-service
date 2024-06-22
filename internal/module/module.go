package module

import (
	"errors"
	"fmt"
	"homework-1/internal/models"
	"homework-1/internal/models/packaging"
	"homework-1/internal/storage"
	"time"
)

var (
	errWrongExpiration = errors.New("wrong expiration date")
	errReturn          = errors.New("can not delete this order. this order might be already received or expiration date is not passed")
	errRefund          = errors.New("can not refund this order. make sure it is yours, you received it and refund time (2 days) has not passed")
	errPagination      = errors.New("page is out of range")
	errReceive         = errors.New("can not receive other orders. one of them probably has not belong to customer or already received or expiration time has passed")
)

type Deps struct {
	Storage *storage.Storage
}

type Module struct {
	Deps
}

func NewModule(d Deps) *Module {
	return &Module{Deps: d}
}

func (m *Module) AddOrder(orderId models.ID, customerId models.ID, expirationDate time.Time, pack models.PackageType, weight models.Kilo, cost models.Rub) error {
	if expirationDate.Before(time.Now()) {
		return errWrongExpiration
	}

	p, errParse := packaging.ParsePackage(pack)
	if errParse != nil {
		return fmt.Errorf("module.AddOrder error: %w", errParse)
	}

	if errWeight := p.ValidateWeight(weight); errWeight != nil {
		return fmt.Errorf("module.AddOrder error: %w", errWeight)
	}

	fromDb, errGetOrder := m.Storage.GetOrder(orderId)
	if errGetOrder != nil {
		return fmt.Errorf("module.AddOrder error: %w", errGetOrder)
	}

	if fromDb.OrderID == orderId {
		return fmt.Errorf("module.AddOrder error: %w", storage.ErrOrderExists)
	}

	order := models.Order{
		OrderID:            orderId,
		CustomerID:         customerId,
		ExpirationTime:     expirationDate,
		ReceivedTime:       time.Time{},
		ReceivedByCustomer: false,
		Refunded:           false,
		Package:            pack,
		Weight:             weight,
		Cost:               cost + p.GetCost(),
	}
	return m.Storage.AddOrder(order)
}

func (m *Module) ReturnOrder(id models.ID) error {
	order, errGet := m.Storage.GetOrder(id)
	if errGet != nil {
		return fmt.Errorf("storage.ReturnOrder error: %w", errGet)
	}

	if order.ReceivedByCustomer && order.ExpirationTime.Before(time.Now()) {
		return m.Storage.ReturnOrder(id)
	}

	return fmt.Errorf("storage.ReturnOrder error: %w", errReturn)
}

func (m *Module) ReceiveOrders(ordersId []models.ID) ([]models.Order, error) {
	order, errGetOrder := m.Storage.GetOrder(ordersId[0])
	if errGetOrder != nil {
		return nil, fmt.Errorf("storage.ReceiveOrders error: %w", errGetOrder)
	}

	customerId := order.CustomerID
	var received []models.Order
	for _, orderId := range ordersId {
		if toReceive, errGet := m.Storage.GetOrder(orderId); errGet != nil || toReceive.ExpirationTime.Before(time.Now()) ||
			toReceive.ReceivedByCustomer || toReceive.CustomerID != customerId {
			return nil, fmt.Errorf("storage.ReceiveOrders error: %w", errReceive)
		}

		receivedOrder, errRec := m.Storage.ReceiveOrder(orderId)
		if errRec != nil {
			return nil, fmt.Errorf("storage.ReceiveOrders error: %w", errRec)
		}

		received = append(received, receivedOrder)
	}

	return received, nil
}

func (m *Module) GetOrders(customerId models.ID, n int) ([]models.Order, error) {
	orders, errGet := m.Storage.GetCustomersOrders(customerId)
	if errGet != nil {
		return nil, fmt.Errorf("storage.GetOrders error: %w", errGet)
	}

	if n <= 0 {
		return orders, nil
	}

	if n > len(orders) {
		n = len(orders)
	}

	return orders[:n], nil

}

func (m *Module) RefundOrder(customerId models.ID, orderId models.ID) error {
	order, errGet := m.Storage.GetOrder(orderId)
	if errGet != nil {
		return fmt.Errorf("storage.ReturnOrder error: %w", errGet)
	}

	if order.CustomerID != customerId || !order.ReceivedByCustomer || order.Refunded || order.ReceivedTime.Add(time.Hour*24*2).Before(time.Now()) {
		return fmt.Errorf("storage.CreateRefund error: %w", errRefund)
	}

	order.Refunded = true
	return m.Storage.ChangeOrder(order)
}

func (m *Module) GetRefunds(page int, limit int) ([]models.Order, error) {
	refunds, errGet := m.Storage.GetRefunds()
	if errGet != nil {
		return nil, fmt.Errorf("storage.GetRefunds error: %w", errGet)
	}

	if limit <= 0 {
		return refunds, nil
	}

	start := page * limit
	end := start + limit

	if start > len(refunds) {
		return nil, fmt.Errorf("storage.GetRefunds error: %w", errPagination)
	}

	if end > len(refunds) {
		end = len(refunds)
	}

	return refunds[start:end], nil
}
