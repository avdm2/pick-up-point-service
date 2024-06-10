package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"homework-1/internal/models"
	"os"
	"sync"
	"time"
)

var (
	errFileCreation  = errors.New("can not create a file")
	errOrderNotFound = errors.New("order not found")
)

type Storage struct {
	fileName string
	mu       sync.Mutex
}

func NewStorage(fileName string) (*Storage, error) {
	if _, err := os.Stat(fileName); err == nil {
		return &Storage{fileName: fileName}, nil
	}

	if err := createFile(fileName); err != nil {
		return &Storage{}, fmt.Errorf("storage.NewStorage error: %w", errFileCreation)
	}

	return &Storage{fileName: fileName}, nil
}

func (s *Storage) AddOrder(order models.Order) error {
	orders, err := s.ReadJson()
	if err != nil {
		return fmt.Errorf("storage.AddOrder error: %w", err)
	}

	orders = append(orders, order)
	return s.WriteJson(orders)
}

func (s *Storage) GetOrder(orderId models.ID) (models.Order, error) {
	orders, errJson := s.ReadJson()
	if errJson != nil {
		return models.Order{}, fmt.Errorf("storage.GetOrder error: %w", errJson)
	}

	for _, order := range orders {
		if order.OrderID == orderId {
			return order, nil
		}
	}

	return models.Order{}, fmt.Errorf("storage.GetOrder error: %w", errOrderNotFound)
}

func (s *Storage) GetCustomersOrders(customerId models.ID) ([]models.Order, error) {
	orders, err := s.ReadJson()
	if err != nil {
		return nil, fmt.Errorf("storage.GetCustomersOrders error: %w", err)
	}

	var customersOrders []models.Order
	for _, order := range orders {
		if order.CustomerID == customerId {
			customersOrders = append(customersOrders, order)
		}
	}

	return customersOrders, nil

}

func (s *Storage) GetRefunds() ([]models.Order, error) {
	orders, err := s.ReadJson()
	if err != nil {
		return nil, fmt.Errorf("storage.GetRefunds error: %w", err)
	}

	var refunds []models.Order
	for _, order := range orders {
		if order.Refunded {
			refunds = append(refunds, order)
		}
	}

	return refunds, nil
}

func (s *Storage) ChangeOrder(order models.Order) error {
	orders, err := s.ReadJson()
	if err != nil {
		return fmt.Errorf("storage.ChangeOrder error: %w", err)
	}

	for i, v := range orders {
		if v.OrderID == order.OrderID {
			orders[i] = order
			return s.WriteJson(orders)
		}
	}

	return fmt.Errorf("storage.ChangeOrder error: %w", errOrderNotFound)
}

func (s *Storage) ReceiveOrder(orderId models.ID) (models.Order, error) {
	orders, err := s.ReadJson()
	if err != nil {
		return models.Order{}, fmt.Errorf("storage.ReceiveOrder error: %w", err)
	}

	for i, v := range orders {
		if v.OrderID == orderId {
			orders[i].ReceivedTime = time.Now()
			orders[i].ReceivedByCustomer = true
			return orders[i], s.WriteJson(orders)
		}
	}

	return models.Order{}, fmt.Errorf("storage.ReceiveOrder error: %w", errOrderNotFound)
}

func (s *Storage) ReturnOrder(orderId models.ID) error {
	orders, err := s.ReadJson()
	if err != nil {
		return fmt.Errorf("storage.ReturnOrder error: %w", err)
	}

	for i, v := range orders {
		if v.OrderID == orderId {
			orders = append(orders[:i], orders[i+1:]...)
			return s.WriteJson(orders)
		}
	}

	return fmt.Errorf("storage.ReturnOrder error: %w", errOrderNotFound)

}

func (s *Storage) ReadJson() ([]models.Order, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	b, errReadFile := os.ReadFile(s.fileName)
	if errReadFile != nil {
		return nil, fmt.Errorf("storage.readJson error: %w", errReadFile)
	}

	if len(b) == 0 {
		return nil, nil
	}

	var orderRecords []orderRecord
	if errUnmarshal := json.Unmarshal(b, &orderRecords); errUnmarshal != nil {
		return nil, fmt.Errorf("storage.readJson error: %w", errUnmarshal)
	}

	var orders []models.Order
	for _, orderRec := range orderRecords {
		orders = append(orders, orderRec.toDomain())
	}

	return orders, nil
}

func (s *Storage) WriteJson(orders []models.Order) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var orderRecords []orderRecord
	for _, order := range orders {
		orderRecords = append(orderRecords, transform(order))
	}

	bWrite, errMarshal := json.MarshalIndent(orderRecords, "  ", "  ")
	if errMarshal != nil {
		return fmt.Errorf("storage.writeJson error: %w", errMarshal)
	}

	errWriting := os.WriteFile(s.fileName, bWrite, 0666)
	if errWriting != nil {
		return fmt.Errorf("storage.writeJson error: %w", errWriting)
	}

	return nil
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
