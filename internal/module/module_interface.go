//go:generate mockgen -source ./module_interface.go -destination=./mocks/module_mock.go -package=module_mock

package module

import (
	"homework-1/internal/models"
	"time"
)

type ModuleInterface interface {
	AddOrder(orderId models.ID, customerId models.ID, expirationTime time.Time, pack models.PackageType, weight models.Kilo, cost models.Rub) error
	ReturnOrder(id models.ID) (models.Order, error)
	ReceiveOrders(ordersId []models.ID) ([]models.Order, error)
	GetOrders(customerId models.ID, n int) ([]models.Order, error)
	RefundOrder(customerId models.ID, orderId models.ID) error
	GetRefunds(page int, limit int) ([]models.Order, error)
}
