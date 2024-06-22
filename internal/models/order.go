package models

import (
	"fmt"
	"time"
)

type ID int64
type Rub int64
type Kilo float32
type PackageType string

type Order struct {
	OrderID            ID
	CustomerID         ID
	ExpirationTime     time.Time
	ReceivedTime       time.Time
	ReceivedByCustomer bool
	Refunded           bool
	Package            PackageType
	Weight             Kilo
	Cost               Rub
}

func (o Order) String() string {
	return fmt.Sprintf(
		"OrderID: %d; CustomerID: %d; ExpirationTime: %s; ReceivedTime: %s; "+
			"ReceivedByCustomer: %t; Refunded: %t; Package: %s; Weight: %f; Cost: %d",
		o.OrderID, o.CustomerID, o.ExpirationTime, o.ReceivedTime, o.ReceivedByCustomer, o.Refunded, o.Package, o.Weight, o.Cost)
}
