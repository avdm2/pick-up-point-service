package module

import (
	"homework-1/internal/models"
)

type Storage interface {
	AddOrder(order models.Order) error
	DeleteOrder(id models.ID) error
	ReceiveOrders(ordersId []models.ID) ([]models.Order, error)
	GetOrders(customerId models.ID, n int) ([]models.Order, error)
	CreateRefund(customerId models.ID, orderId models.ID) error
	GetRefunds(page int, limit int) ([]models.Order, error)
}

type Deps struct {
	Storage Storage
}

type Module struct {
	Deps
}

func NewModule(d Deps) Module {
	return Module{Deps: d}
}

func (m Module) Add(order models.Order) error {
	return m.Storage.AddOrder(order)
}

func (m Module) Delete(id models.ID) error {
	return m.Storage.DeleteOrder(id)
}

func (m Module) Receive(ordersId []models.ID) ([]models.Order, error) {
	return m.Storage.ReceiveOrders(ordersId)
}

func (m Module) Orders(customerId models.ID, n int) ([]models.Order, error) {
	return m.Storage.GetOrders(customerId, n)
}

func (m Module) Refund(customerId models.ID, orderId models.ID) error {
	return m.Storage.CreateRefund(customerId, orderId)
}

func (m Module) Refunds(page int, limit int) ([]models.Order, error) {
	return m.Storage.GetRefunds(page, limit)
}
