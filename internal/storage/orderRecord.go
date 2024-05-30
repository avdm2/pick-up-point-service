package storage

import (
	"homework-1/internal/models"
	"time"
)

type id int64

type orderRecord struct {
	OrderID    id `json:"order_id"`
	CustomerID id `json:"customer_id"`

	ExpirationTime time.Time `json:"expiration_time"`
	ReceivedTime   time.Time `json:"received_time"`

	ReceivedByCustomer bool `json:"received_by_customer"`
	Refunded           bool `json:"refunded"`
}

func (o orderRecord) toDomain() models.Order {
	return models.Order{
		OrderID:            models.ID(o.OrderID),
		CustomerID:         models.ID(o.CustomerID),
		ExpirationTime:     o.ExpirationTime,
		ReceivedTime:       o.ReceivedTime,
		ReceivedByCustomer: o.ReceivedByCustomer,
		Refunded:           o.Refunded,
	}
}

func transform(orderModel models.Order) orderRecord {
	return orderRecord{
		OrderID:            id(orderModel.OrderID),
		CustomerID:         id(orderModel.CustomerID),
		ExpirationTime:     orderModel.ExpirationTime,
		ReceivedTime:       orderModel.ReceivedTime,
		ReceivedByCustomer: orderModel.ReceivedByCustomer,
		Refunded:           orderModel.Refunded,
	}
}
