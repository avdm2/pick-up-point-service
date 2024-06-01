package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"homework-1/internal/models"
	"os"
	"reflect"
	"time"
)

var (
	errFileCreation    = errors.New("can not create a file")
	errAlreadyExists   = errors.New("this order is already exists")
	errWrongExpiration = errors.New("wrong expiration date")
	errDelete          = errors.New("can not delete this order. this order might be already received or expiration date is not passed")
	errRefund          = errors.New("can not refund this order. make sure it is yours and refund time (2 weeks) has not passed")
	errPagination      = errors.New("page is out of range")
	errReceive         = errors.New("can not receive other orders. one of them probably has not belong to customer or expiration time has passed")
)

type Storage struct {
	fileName string
}

func NewStorage(fileName string) (Storage, error) {
	if _, err := os.Stat(fileName); err == nil {
		return Storage{fileName: fileName}, nil
	}

	if err := createFile(fileName); err != nil {
		return Storage{}, fmt.Errorf("storage.NewStorage error: %w", errFileCreation)
	}

	return Storage{fileName: fileName}, nil
}

func (s Storage) AddOrder(modelOrder models.Order) error {
	orders, errJson := readJson(s.fileName)
	if errJson != nil {
		return fmt.Errorf("storage.AddOrder error: %w", errJson)
	}

	order := transform(modelOrder)

	if order.ExpirationTime.Before(time.Now()) {
		return errWrongExpiration
	}

	for _, v := range orders {
		if reflect.DeepEqual(v, order) {
			return fmt.Errorf("storage.AddOrder error: %w", errAlreadyExists)
		}
	}

	orders = append(orders, order)
	return s.writeJson(orders)
}

func (s Storage) ReturnOrder(orderId models.ID) error {
	orders, errJson := readJson(s.fileName)
	if errJson != nil {
		return fmt.Errorf("storage.ReturnOrder error: %w", errJson)
	}

	for i, v := range orders {
		if v.OrderID == id(orderId) {
			if v.ReceivedByCustomer || v.ExpirationTime.Before(time.Now()) {
				// Удаление (добавление всех заказов до данного индекса и после)
				orders = append(orders[:i], orders[i+1:]...)
				break
			} else {
				return fmt.Errorf("storage.ReturnOrder error: %w", errDelete)
			}
		}
	}

	return s.writeJson(orders)
}

func (s Storage) ReceiveOrders(ordersId []models.ID) ([]models.Order, error) {
	orders, errJson := readJson(s.fileName)
	if errJson != nil {
		return nil, fmt.Errorf("storage.ReceiveOrders error: %w", errJson)
	}

	ordersMap := make(map[id]orderRecord, len(orders))
	for _, v := range orders {
		ordersMap[v.OrderID] = v
	}

	// Предположим, что первый заказ из списка принадлежит какому-то клиенту. С этим клиентом будет сравнение
	// следующих заказов.
	customerId := ordersMap[id(ordersId[0])].CustomerID
	var result []models.Order
	for _, orderId := range ordersId {
		if order, ok := ordersMap[id(orderId)]; ok && order.ExpirationTime.After(time.Now()) && order.CustomerID == customerId {
			// upd orderRecord
			order.ReceivedTime = time.Now()
			order.ReceivedByCustomer = true
			ordersMap[id(orderId)] = order
			result = append(result, order.toDomain())
		} else {
			return nil, fmt.Errorf("storage.ReceiveOrders error: %w", errReceive)
		}
	}

	var changedOrders []orderRecord
	for _, order := range ordersMap {
		changedOrders = append(changedOrders, order)
	}

	return result, s.writeJson(changedOrders)
}

func (s Storage) GetOrders(customerId models.ID, n int) ([]models.Order, error) {
	orders, errJson := readJson(s.fileName)
	if errJson != nil {
		return nil, fmt.Errorf("storage.GetOrders error: %w", errJson)
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
		return fmt.Errorf("storage.CreateRefund error: %w", errJson)
	}

	ordersMap := make(map[id]orderRecord, len(orders))
	for _, v := range orders {
		ordersMap[v.OrderID] = v
	}

	if toRefund, ok := ordersMap[id(orderId)]; ok && toRefund.CustomerID == id(customerId) &&
		toRefund.ReceivedTime.Add(time.Hour*24*2).Before(time.Now()) {

		toRefund.Refunded = true
		ordersMap[id(orderId)] = toRefund
	} else {
		return fmt.Errorf("storage.CreateRefund error: %w", errRefund)
	}

	var changedOrders []orderRecord
	for _, order := range ordersMap {
		changedOrders = append(changedOrders, order)
	}

	return s.writeJson(changedOrders)
}

func (s Storage) GetRefunds(page int, limit int) ([]models.Order, error) {
	orders, errJson := readJson(s.fileName)
	if errJson != nil {
		return nil, fmt.Errorf("storage.GetRefunds error: %w", errJson)
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
		return nil, fmt.Errorf("storage.GetRefunds error: %w", errPagination)
	}

	if end > len(refunds) {
		end = len(refunds)
	}

	return refunds[start:end], nil
}

func createFile(fileName string) error {
	f, errCreate := os.Create(fileName)
	if errCreate != nil {
		return fmt.Errorf("storage.createFile error: %w", errCreate)
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
		return nil, fmt.Errorf("storage.readJson error: %w", errReadFile)
	}

	if len(b) == 0 {
		return nil, nil
	}

	var orders []orderRecord
	if errUnmarshal := json.Unmarshal(b, &orders); errUnmarshal != nil {
		return nil, fmt.Errorf("storage.readJson error: %w", errUnmarshal)
	}

	return orders, nil
}

func (s Storage) writeJson(orders []orderRecord) error {
	bWrite, errMarshal := json.MarshalIndent(orders, "  ", "  ")
	if errMarshal != nil {
		return fmt.Errorf("storage.writeJson error: %w", errMarshal)
	}

	errWriting := os.WriteFile(s.fileName, bWrite, 0666)
	if errWriting != nil {
		return fmt.Errorf("storage.writeJson error: %w", errWriting)
	}

	return nil
}
