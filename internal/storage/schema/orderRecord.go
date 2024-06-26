package schema

import (
	"homework-1/internal/models"
	"time"
)

type id int64
type rub int64
type kilo float32
type packageType string

type OrderRecord struct {
	OrderID            id          `db:"order_id"`
	CustomerID         id          `db:"customer_id"`
	ExpirationTime     time.Time   `db:"expiration_time"`
	ReceivedTime       time.Time   `db:"received_time"`
	ReceivedByCustomer bool        `db:"received_by_customer"`
	Refunded           bool        `db:"refunded"`
	Package            packageType `db:"package"`
	Weight             kilo        `db:"weight"`
	Cost               rub         `db:"cost"`
	PackageCost        rub         `db:"package_cost"`
}

func (o OrderRecord) ToDomain() models.Order {
	return models.Order{
		OrderID:            models.ID(o.OrderID),
		CustomerID:         models.ID(o.CustomerID),
		ExpirationTime:     o.ExpirationTime,
		ReceivedTime:       o.ReceivedTime,
		ReceivedByCustomer: o.ReceivedByCustomer,
		Refunded:           o.Refunded,
		Package:            models.PackageType(o.Package),
		Weight:             models.Kilo(o.Weight),
		Cost:               models.Rub(o.Cost),
		PackageCost:        models.Rub(o.PackageCost),
	}
}

func Transform(orderModel models.Order) OrderRecord {
	return OrderRecord{
		OrderID:            id(orderModel.OrderID),
		CustomerID:         id(orderModel.CustomerID),
		ExpirationTime:     orderModel.ExpirationTime,
		ReceivedTime:       orderModel.ReceivedTime,
		ReceivedByCustomer: orderModel.ReceivedByCustomer,
		Refunded:           orderModel.Refunded,
		Package:            packageType(orderModel.Package),
		Weight:             kilo(orderModel.Weight),
		Cost:               rub(orderModel.Cost),
		PackageCost:        rub(orderModel.PackageCost),
	}
}
