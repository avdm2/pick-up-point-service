//go:generate mockgen -source ./moduleInterface.go -destination=./mocks/mock_module.go -package=mock_module

package module

import (
	"homework-1/internal/models"
	"time"
)

type ModuleInterface interface {
	AddOrder(orderId models.ID, customerId models.ID, expirationDate time.Time, pack models.PackageType, weight models.Kilo, cost models.Rub) error
	ReturnOrder(id models.ID) error
	ReceiveOrders(ordersId []models.ID) ([]models.Order, error)
	GetOrders(customerId models.ID, n int) ([]models.Order, error)
	RefundOrder(customerId models.ID, orderId models.ID) error
	GetRefunds(page int, limit int) ([]models.Order, error)
}
