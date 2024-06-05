package module

import (
	"errors"
	"fmt"
	"homework-1/internal/models"
	"homework-1/internal/storage"
	"time"
)

var (
	errWrongExpiration = errors.New("wrong expiration date")
	errAlreadyExists   = errors.New("order with this id is already exists")
	errDelete          = errors.New("can not delete this order. this order might be already received or expiration date is not passed")
	errRefund          = errors.New("can not refund this order. make sure it is yours and refund time (2 weeks) has not passed")
	errPagination      = errors.New("page is out of range")
	errReceive         = errors.New("can not receive other orders. one of them probably has not belong to customer or expiration time has passed")
)

type Deps struct {
	Storage storage.Storage
}

type Module struct {
	Deps
}

func NewModule(d Deps) Module {
	return Module{Deps: d}
}

func (m Module) AddOrder(order models.Order) error {
	orders, errJson := m.Storage.ReadJson()
	if errJson != nil {
		return fmt.Errorf("module.AddOrder error: %w", errJson)
	}

	if order.ExpirationTime.Before(time.Now()) {
		return errWrongExpiration
	}

	for _, v := range orders {
		if v.OrderID == order.CustomerID {
			return fmt.Errorf("module.AddOrder error: %w", errAlreadyExists)
		}
	}

	orders = append(orders, order)
	return m.Storage.WriteJson(orders)

}

func (m Module) ReturnOrder(id models.ID) error {
	orders, errJson := m.Storage.ReadJson()
	if errJson != nil {
		return fmt.Errorf("module.Return error: %w", errJson)
	}

	for i, v := range orders {
		if v.OrderID == id {
			if v.ReceivedByCustomer || v.ExpirationTime.Before(time.Now()) {
				// Удаление (добавление всех заказов до данного индекса и после)
				orders = append(orders[:i], orders[i+1:]...)
				break
			} else {
				return fmt.Errorf("storage.ReturnOrder error: %w", errDelete)
			}
		}
	}

	return m.Storage.WriteJson(orders)
}

func (m Module) ReceiveOrders(ordersId []models.ID) ([]models.Order, error) {
	orders, errJson := m.Storage.ReadJson()
	if errJson != nil {
		return nil, fmt.Errorf("module.ReceiveOrders error: %w", errJson)
	}

	ordersMap := make(map[models.ID]models.Order, len(orders))
	for _, v := range orders {
		ordersMap[v.OrderID] = v
	}

	// Предположим, что первый заказ из списка принадлежит какому-то клиенту. С этим клиентом будет сравнение
	// следующих заказов.
	customerId := ordersMap[ordersId[0]].CustomerID
	var result []models.Order
	for _, orderId := range ordersId {
		if order, ok := ordersMap[orderId]; ok && order.ExpirationTime.After(time.Now()) && order.CustomerID == customerId {
			// upd orderRecord
			order.ReceivedTime = time.Now()
			order.ReceivedByCustomer = true
			ordersMap[orderId] = order
			result = append(result, order)
		} else {
			return nil, fmt.Errorf("storage.ReceiveOrders error: %w", errReceive)
		}
	}

	var changedOrders []models.Order
	for _, order := range ordersMap {
		changedOrders = append(changedOrders, order)
	}

	return result, m.Storage.WriteJson(changedOrders)

}

func (m Module) GetOrders(customerId models.ID, n int) ([]models.Order, error) {
	orders, errJson := m.Storage.ReadJson()
	if errJson != nil {
		return nil, fmt.Errorf("storage.GetOrders error: %w", errJson)
	}

	var result []models.Order
	for _, order := range orders {
		if order.CustomerID == customerId {
			result = append(result, order)
			if n > 0 && len(result) >= n {
				break
			}
		}
	}

	return result, nil

}

func (m Module) RefundOrder(customerId models.ID, orderId models.ID) error {
	orders, errJson := m.Storage.ReadJson()
	if errJson != nil {
		return fmt.Errorf("storage.CreateRefund error: %w", errJson)
	}

	ordersMap := make(map[models.ID]models.Order, len(orders))
	for _, v := range orders {
		ordersMap[v.OrderID] = v
	}

	if toRefund, ok := ordersMap[orderId]; ok && toRefund.CustomerID == customerId &&
		toRefund.ReceivedTime.Add(time.Hour*24*2).Before(time.Now()) {

		toRefund.Refunded = true
		ordersMap[orderId] = toRefund
	} else {
		return fmt.Errorf("storage.CreateRefund error: %w", errRefund)
	}

	var changedOrders []models.Order
	for _, order := range ordersMap {
		changedOrders = append(changedOrders, order)
	}

	return m.Storage.WriteJson(changedOrders)

}

func (m Module) GetRefunds(page int, limit int) ([]models.Order, error) {
	orders, errJson := m.Storage.ReadJson()
	if errJson != nil {
		return nil, fmt.Errorf("storage.GetRefunds error: %w", errJson)
	}

	var refunds []models.Order
	for _, v := range orders {
		if v.Refunded {
			refunds = append(refunds, v)
		}
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
