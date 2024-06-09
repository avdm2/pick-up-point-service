package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"homework-1/internal/models"
	"os"
	"sync"
)

var (
	errFileCreation = errors.New("can not create a file")
	errWrongOrder   = errors.New("wrong order id")
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

func (s *Storage) GetOrder(orderId models.ID) (models.Order, error) {
	orders, err := s.ReadJson()
	if err != nil {
		return models.Order{}, fmt.Errorf("storage.GetOrder error: %w", err)
	}

	for _, order := range orders {
		if order.OrderID == orderId {
			return order, nil
		}
	}

	return models.Order{}, fmt.Errorf("storage.GetOrder error: %w", errWrongOrder)
}

func (s *Storage) AddOrder(order models.Order) error {
	orders, err := s.ReadJson()
	if err != nil {
		return fmt.Errorf("storage.AddOrder error: %w", err)
	}

	orders = append(orders, order)
	return s.WriteJson(orders)
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
