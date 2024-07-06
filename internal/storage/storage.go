//go:generate mockgen -source ./storage.go -destination=./mocks/storage_mock.go -package=storage_mock

package storage

import "homework-1/internal/models"

type Storage interface {
	AddOrder(order models.Order) error
	GetOrder(orderId models.ID) (models.Order, error)
	GetCustomersOrders(customerId models.ID) ([]models.Order, error)
	GetRefunds() ([]models.Order, error)
	ChangeOrder(order models.Order) error
	ReceiveOrder(orderId models.ID) (models.Order, error)
	ReturnOrder(orderId models.ID) error
}
