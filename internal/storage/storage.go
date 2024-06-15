package storage

import (
	"context"
	"errors"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"homework-1/internal/models"
	"homework-1/internal/storage/schema"
	"homework-1/internal/storage/transactor"
	"sync"
	"time"
)

var (
	errOrderNotFound = errors.New("order not found")
	ErrOrderExists   = errors.New("order already exists")
)

var (
	orderColumns = []string{"order_id", "customer_id", "expiration_time", "received_time", "received_by_customer", "refunded"}
	orderTable   = "orders"
)

type Storage struct {
	db *pgxpool.Pool
	tr *transactor.Transactor
	mu sync.Mutex
}

func NewStorage(connUrl string) (*Storage, error) {
	db, err := pgxpool.New(context.Background(), connUrl)
	if err != nil {
		return nil, fmt.Errorf("storage.NewStorage error: %w", err)

	}

	return &Storage{
		db: db,
		tr: &transactor.Transactor{Db: db},
	}, nil
}

func (s *Storage) AddOrder(order models.Order) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	ordRecord := schema.Transform(order)

	sql, args, errSql := sq.
		Insert(orderTable).
		Columns(orderColumns...).
		Values(ordRecord.OrderID, ordRecord.CustomerID, ordRecord.ExpirationTime, ordRecord.ReceivedTime, ordRecord.ReceivedByCustomer, ordRecord.Refunded).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if errSql != nil {
		return fmt.Errorf("storage.AddOrder error: %w", errSql)
	}

	_, errExec := s.db.Exec(context.Background(), sql, args...)
	if errExec != nil {
		if errors.Is(errExec, pgx.ErrNoRows) {
			return fmt.Errorf("storage.AddOrder error: %w", ErrOrderExists)
		}
		return fmt.Errorf("storage.AddOrder error: %w", errExec)
	}

	return nil
}

func (s *Storage) GetOrder(orderId models.ID) (models.Order, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sql, args, errSql := sq.
		Select(orderColumns...).
		From(orderTable).
		Where(sq.Eq{"order_id": orderId}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if errSql != nil {
		return models.Order{}, fmt.Errorf("storage.GetOrder error: %w", errSql)
	}

	rows, errQuery := s.db.Query(context.Background(), sql, args...)
	if errQuery != nil {
		if errors.Is(errQuery, pgx.ErrNoRows) {
			return models.Order{}, fmt.Errorf("storage.GetOrder error: %w", errOrderNotFound)
		}
		return models.Order{}, fmt.Errorf("storage.GetOrder error: %w", errQuery)
	}
	defer rows.Close()

	var ordRecord schema.OrderRecord
	for rows.Next() {
		if errScan := rows.Scan(&ordRecord.OrderID, &ordRecord.CustomerID,
			&ordRecord.ExpirationTime, &ordRecord.ReceivedTime,
			&ordRecord.ReceivedByCustomer, &ordRecord.Refunded); errScan != nil {
			return models.Order{}, fmt.Errorf("storage.GetOrder error: %w", errScan)
		}
	}

	return ordRecord.ToDomain(), nil
}

func (s *Storage) GetCustomersOrders(customerId models.ID) ([]models.Order, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sql, args, errSql := sq.
		Select(orderColumns...).
		From(orderTable).
		Where(sq.Eq{"customer_id": customerId}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if errSql != nil {
		return nil, fmt.Errorf("storage.GetCustomersOrders error: %w", errSql)
	}

	rows, errQuery := s.db.Query(context.Background(), sql, args...)
	if errQuery != nil {
		if errors.Is(errQuery, pgx.ErrNoRows) {
			return nil, fmt.Errorf("storage.GetCustomersOrders error: %w", errOrderNotFound)
		}
		return nil, fmt.Errorf("storage.GetCustomersOrders error: %w", errQuery)
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var ordRecord schema.OrderRecord
		if errScan := rows.Scan(&ordRecord.OrderID, &ordRecord.CustomerID,
			&ordRecord.ExpirationTime, &ordRecord.ReceivedTime,
			&ordRecord.ReceivedByCustomer, &ordRecord.Refunded); errScan != nil {
			return nil, fmt.Errorf("storage.GetCustomersOrders error: %w", errScan)
		}
		orders = append(orders, ordRecord.ToDomain())
	}

	return orders, nil
}

func (s *Storage) GetRefunds() ([]models.Order, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sql, args, errSql := sq.
		Select(orderColumns...).
		From(orderTable).
		Where(sq.Eq{"refunded": true}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if errSql != nil {
		return nil, fmt.Errorf("storage.GetRefunds error: %w", errSql)
	}

	rows, errQuery := s.db.Query(context.Background(), sql, args...)
	if errQuery != nil {
		if errors.Is(errQuery, pgx.ErrNoRows) {
			return nil, fmt.Errorf("storage.GetRefunds error: %w", errOrderNotFound)
		}
		return nil, fmt.Errorf("storage.GetRefunds error: %w", errQuery)
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var ordRecord schema.OrderRecord
		if errScan := rows.Scan(&ordRecord.OrderID, &ordRecord.CustomerID,
			&ordRecord.ExpirationTime, &ordRecord.ReceivedTime,
			&ordRecord.ReceivedByCustomer, &ordRecord.Refunded); errScan != nil {
			return nil, fmt.Errorf("storage.GetRefunds error: %w", errScan)
		}
		orders = append(orders, ordRecord.ToDomain())
	}

	return orders, nil
}

func (s *Storage) ChangeOrder(order models.Order) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	ordRecord := schema.Transform(order)

	f := func(ctxTX context.Context) error {

		sql, args, errSql := sq.
			Update(orderTable).
			Set("customer_id", ordRecord.CustomerID).
			Set("expiration_time", ordRecord.ExpirationTime).
			Set("received_time", ordRecord.ReceivedTime).
			Set("received_by_customer", ordRecord.ReceivedByCustomer).
			Set("refunded", ordRecord.Refunded).
			Where(sq.Eq{"order_id": ordRecord.OrderID}).
			PlaceholderFormat(sq.Dollar).
			ToSql()
		if errSql != nil {
			return fmt.Errorf("storage.ChangeOrder error: %w", errSql)
		}

		_, errExec := s.db.Exec(context.Background(), sql, args...)
		if errExec != nil {
			if errors.Is(errExec, pgx.ErrNoRows) {
				return fmt.Errorf("storage.ChangeOrder error: %w", errOrderNotFound)
			}
			return fmt.Errorf("storage.ChangeOrder error: %w", errExec)
		}

		return nil

	}

	if err := s.tr.RunRepeatableRead(context.Background(), f); err != nil {
		return fmt.Errorf("storage.ChangeOrder error: %w", err)
	}

	return nil
}

func (s *Storage) ReceiveOrder(orderId models.ID) (models.Order, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sql, args, errSql := sq.
		Update(orderTable).
		Set("received_time", time.Now()).
		Set("received_by_customer", true).
		Where(sq.Eq{"order_id": orderId}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if errSql != nil {
		return models.Order{}, fmt.Errorf("storage.ReceiveOrder error: %w", errSql)
	}

	rows, errQuery := s.db.Query(context.Background(), sql, args...)
	if errQuery != nil {
		if errors.Is(errQuery, pgx.ErrNoRows) {
			return models.Order{}, fmt.Errorf("storage.ReceiveOrder error: %w", errOrderNotFound)
		}
		return models.Order{}, fmt.Errorf("storage.ReceiveOrder error: %w", errQuery)
	}
	defer rows.Close()

	var order models.Order
	for rows.Next() {
		var ordRecord schema.OrderRecord
		if errScan := rows.Scan(&ordRecord.OrderID, &ordRecord.CustomerID,
			&ordRecord.ExpirationTime, &ordRecord.ReceivedTime,
			&ordRecord.ReceivedByCustomer, &ordRecord.Refunded); errScan != nil {
			return models.Order{}, fmt.Errorf("storage.ReceiveOrder error: %w", errScan)
		}
		order = ordRecord.ToDomain()
	}

	return order, nil
}

func (s *Storage) ReturnOrder(orderId models.ID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sql, args, errSql := sq.
		Delete(orderTable).
		Where(sq.Eq{"order_id": orderId}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if errSql != nil {
		return fmt.Errorf("storage.ReturnOrder error: %w", errSql)
	}

	_, errExec := s.db.Exec(context.Background(), sql, args...)
	if errExec != nil {
		if errors.Is(errExec, pgx.ErrNoRows) {
			return fmt.Errorf("storage.ReceiveOrder error: %w", errOrderNotFound)
		}
		return fmt.Errorf("storage.ReceiveOrder error: %w", errExec)
	}

	return nil
}
