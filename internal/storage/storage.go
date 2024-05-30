package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"homework-1/internal/models"
	"os"
	"time"
)

var (
	errFileCreation    = errors.New("can not create a file")
	errAlreadyExists   = errors.New("this order is already exists")
	errWrongExpiration = errors.New("wrong expiration date")
	errDelete          = errors.New("can not delete this order")
	errRefund          = errors.New("can not refund this order")
	errPagination      = errors.New("page is out of range")
)

type Storage struct {
	fileName string
}

func NewStorage(fileName string) (Storage, error) {
	if _, err := os.Stat(fileName); err == nil {
		return Storage{fileName: fileName}, nil
	}

	if err := createFile(fileName); err != nil {
		return Storage{}, errFileCreation
	}

	return Storage{fileName: fileName}, nil
}

func (s Storage) AddOrder(modelOrder models.Order) error {
	orders, errJson := readJson(s.fileName)
	if errJson != nil {
		return errJson
	}

	order := transform(modelOrder)

	if order.ExpirationTime.Before(time.Now()) {
		return errWrongExpiration
	}

	for _, v := range orders {
		if v == order {
			return errAlreadyExists
		}
	}

	orders = append(orders, order)
	bWrite, errMarshal := json.MarshalIndent(orders, "  ", "  ")
	if errMarshal != nil {
		return errMarshal
	}

	return os.WriteFile(s.fileName, bWrite, 0666)
}

func (s Storage) DeleteOrder(orderId models.ID) error {
	orders, errJson := readJson(s.fileName)
	if errJson != nil {
		return errJson
	}

	for i, v := range orders {
		if v.OrderID == id(orderId) {
			if v.ReceivedByCustomer || v.ExpirationTime.Before(time.Now()) {
				// Удаление (добавление всех заказов до данного индекса и после)
				orders = append(orders[:i], orders[i+1:]...)
				break
			} else {
				return errDelete
			}
		}
	}

	bWrite, errMarshal := json.MarshalIndent(orders, "  ", "  ")
	if errMarshal != nil {
		return errMarshal
	}

	return os.WriteFile(s.fileName, bWrite, 0666)
}

// ReceiveOrders По условию требуется принимать только один параметр - список айди заказов. С другой стороны
// в условии написано "Все ID заказов должны принадлежать только одному клиенту". Значит ли это, что параметров на самом
// деле два (список заказов и айди клиента, который будет получать)? Или это просто ограничение на параметр айди заказа,
// что у двух разных людей не может быть заказа с одним ID?
func (s Storage) ReceiveOrders(ordersId []models.ID) ([]models.Order, error) {
	orders, errJson := readJson(s.fileName)
	if errJson != nil {
		return nil, errJson
	}

	ordersMap := make(map[id]orderRecord, len(orders))
	for _, v := range orders {
		ordersMap[v.OrderID] = v
	}

	var result []models.Order
	for _, orderId := range ordersId {
		if order, ok := ordersMap[id(orderId)]; ok && order.ExpirationTime.After(time.Now()) {
			// upd orderRecord
			order.ReceivedTime = time.Now()
			order.ReceivedByCustomer = true
			ordersMap[id(orderId)] = order
			result = append(result, order.toDomain())
		}
	}

	var changedOrders []orderRecord
	for _, order := range ordersMap {
		changedOrders = append(changedOrders, order)
	}

	bWrite, errMarshal := json.MarshalIndent(changedOrders, "  ", "  ")
	if errMarshal != nil {
		return nil, errMarshal
	}

	return result, os.WriteFile(s.fileName, bWrite, 0666)
}

func (s Storage) GetOrders(customerId models.ID, n int) ([]models.Order, error) {
	orders, errJson := readJson(s.fileName)
	if errJson != nil {
		return nil, errJson
	}

	var result []models.Order
	for _, order := range orders {
		if order.CustomerID == id(customerId) {
			result = append(result, order.toDomain())
			if n > 0 && len(result) >= n {
				break
			}
		}
	}

	return result, nil
}

func (s Storage) CreateRefund(customerId models.ID, orderId models.ID) error {
	orders, errJson := readJson(s.fileName)
	if errJson != nil {
		return errJson
	}

	ordersMap := make(map[id]orderRecord, len(orders))
	for _, v := range orders {
		ordersMap[v.OrderID] = v
	}

	if toRefund, ok := ordersMap[id(orderId)]; ok && toRefund.CustomerID == id(customerId) {
		if toRefund.ReceivedTime.Add(time.Hour * 24 * 2).Before(time.Now()) {
			return errRefund

		}
		toRefund.Refunded = true
		ordersMap[id(orderId)] = toRefund
	}

	var changedOrders []orderRecord
	for _, order := range ordersMap {
		changedOrders = append(changedOrders, order)
	}

	bWrite, errMarshal := json.MarshalIndent(changedOrders, "  ", "  ")
	if errMarshal != nil {
		return errMarshal
	}

	return os.WriteFile(s.fileName, bWrite, 0666)
}

func (s Storage) GetRefunds(page int, limit int) ([]models.Order, error) {
	orders, errJson := readJson(s.fileName)
	if errJson != nil {
		return nil, errJson
	}

	var refunds []models.Order
	for _, v := range orders {
		if v.Refunded {
			refunds = append(refunds, v.toDomain())
		}
	}

	if limit <= 0 {
		return refunds, nil
	}

	start := page * limit
	end := start + limit

	if start > len(refunds) {
		return nil, errPagination
	}

	if end > len(refunds) {
		end = len(refunds)
	}

	return refunds[start:end], nil
}

func createFile(fileName string) error {
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}

	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			fmt.Printf("Can not close file %s: %s\n", fileName, err)
			return
		}
	}(f)
	return nil
}

func readJson(fileName string) ([]orderRecord, error) {
	b, errReadFile := os.ReadFile(fileName)
	if errReadFile != nil {
		return nil, errReadFile
	}

	if len(b) == 0 {
		return nil, nil
	}

	var orders []orderRecord
	if errUnmarshal := json.Unmarshal(b, &orders); errUnmarshal != nil {
		return nil, errUnmarshal
	}

	return orders, nil
}
