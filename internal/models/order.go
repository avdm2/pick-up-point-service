package models

import (
	"fmt"
	"time"
)

type ID int64

type Order struct {
	OrderID    ID
	CustomerID ID

	ExpirationTime time.Time
	ReceivedTime   time.Time

	ReceivedByCustomer bool
	Refunded           bool
}

func (o Order) String() string {
	return fmt.Sprintf(
		"OrderID: %d; CustomerID: %d; ExpirationTime: %s; ReceivedTime: %s; "+
			"ReceivedByCustomer: %t; Refunded: %t;",
		o.OrderID, o.CustomerID, o.ExpirationTime, o.ReceivedTime, o.ReceivedByCustomer, o.Refunded)
}
