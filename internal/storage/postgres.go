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
	"time"
)

var (
	ErrOrderNotFound = errors.New("order not found")
	ErrOrderExists   = errors.New("order already exists")
)

var (
	orderColumns = []string{
		"order_id", "customer_id",
		"expiration_time", "received_time",
		"received_by_customer", "refunded",
		"package", "weight", "cost", "package_cost"}
	orderTable = "orders"
)

type PostgresDB struct {
	db *pgxpool.Pool
	tr *transactor.Transactor
}

func NewStorage(connUrl string) (*PostgresDB, error) {
	db, err := pgxpool.New(context.Background(), connUrl)
	if err != nil {
		return nil, fmt.Errorf("storage.NewStorage error: %w", err)

	}

	return &PostgresDB{
		db: db,
		tr: transactor.NewTransactor(db),
	}, nil
}

func (s *PostgresDB) AddOrder(order models.Order) error {
	ordRecord := schema.Transform(order)

	sql, args, errSql := sq.
		Insert(orderTable).
		Columns(orderColumns...).
		Values(ordRecord.OrderID, ordRecord.CustomerID,
			ordRecord.ExpirationTime, ordRecord.ReceivedTime,
			ordRecord.ReceivedByCustomer, ordRecord.Refunded,
			ordRecord.Package, ordRecord.Weight, ordRecord.Cost, ordRecord.PackageCost).
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

func (s *PostgresDB) GetOrder(orderId models.ID) (models.Order, error) {
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
			return models.Order{}, fmt.Errorf("storage.GetOrder error: %w", ErrOrderNotFound)
		}
		return models.Order{}, fmt.Errorf("storage.GetOrder error: %w", errQuery)
	}
	defer rows.Close()

	var ordRecord schema.OrderRecord
	for rows.Next() {
		if errScan := rows.Scan(&ordRecord.OrderID, &ordRecord.CustomerID,
			&ordRecord.ExpirationTime, &ordRecord.ReceivedTime,
			&ordRecord.ReceivedByCustomer, &ordRecord.Refunded,
			&ordRecord.Package, &ordRecord.Weight, &ordRecord.Cost, &ordRecord.PackageCost); errScan != nil {
			return models.Order{}, fmt.Errorf("storage.GetOrder error: %w", errScan)
		}
	}

	return ordRecord.ToDomain(), nil
}

func (s *PostgresDB) GetCustomersOrders(customerId models.ID) ([]models.Order, error) {
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
			return nil, fmt.Errorf("storage.GetCustomersOrders error: %w", ErrOrderNotFound)
		}
		return nil, fmt.Errorf("storage.GetCustomersOrders error: %w", errQuery)
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var ordRecord schema.OrderRecord
		if errScan := rows.Scan(&ordRecord.OrderID, &ordRecord.CustomerID,
			&ordRecord.ExpirationTime, &ordRecord.ReceivedTime,
			&ordRecord.ReceivedByCustomer, &ordRecord.Refunded,
			&ordRecord.Package, &ordRecord.Weight, &ordRecord.Cost, &ordRecord.PackageCost); errScan != nil {
			return nil, fmt.Errorf("storage.GetCustomersOrders error: %w", errScan)
		}
		orders = append(orders, ordRecord.ToDomain())
	}

	return orders, nil
}

func (s *PostgresDB) GetRefunds() ([]models.Order, error) {
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
			return nil, fmt.Errorf("storage.GetRefunds error: %w", ErrOrderNotFound)
		}
		return nil, fmt.Errorf("storage.GetRefunds error: %w", errQuery)
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var ordRecord schema.OrderRecord
		if errScan := rows.Scan(&ordRecord.OrderID, &ordRecord.CustomerID,
			&ordRecord.ExpirationTime, &ordRecord.ReceivedTime,
			&ordRecord.ReceivedByCustomer, &ordRecord.Refunded,
			&ordRecord.Package, &ordRecord.Weight, &ordRecord.Cost, &ordRecord.PackageCost); errScan != nil {
			return nil, fmt.Errorf("storage.GetRefunds error: %w", errScan)
		}
		orders = append(orders, ordRecord.ToDomain())
	}

	return orders, nil
}

func (s *PostgresDB) ChangeOrder(order models.Order) error {
	ordRecord := schema.Transform(order)

	f := func(ctxTX context.Context) error {
		queryEngine := s.tr.GetQueryEngine(ctxTX)

		sql, args, errSql := sq.
			Update(orderTable).
			Set("customer_id", ordRecord.CustomerID).
			Set("expiration_time", ordRecord.ExpirationTime).
			Set("received_time", ordRecord.ReceivedTime).
			Set("received_by_customer", ordRecord.ReceivedByCustomer).
			Set("refunded", ordRecord.Refunded).
			Set("package", ordRecord.Package).
			Set("weight", ordRecord.Weight).
			Set("cost", ordRecord.Cost).
			Set("package_cost", ordRecord.PackageCost).
			Where(sq.Eq{"order_id": ordRecord.OrderID}).
			PlaceholderFormat(sq.Dollar).
			ToSql()
		if errSql != nil {
			return fmt.Errorf("storage.ChangeOrder error: %w", errSql)
		}

		_, errExec := queryEngine.Exec(context.Background(), sql, args...)
		if errExec != nil {
			if errors.Is(errExec, pgx.ErrNoRows) {
				return fmt.Errorf("storage.ChangeOrder error: %w", ErrOrderNotFound)
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

func (s *PostgresDB) ReceiveOrder(orderId models.ID) (models.Order, error) {
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
			return models.Order{}, fmt.Errorf("storage.ReceiveOrder error: %w", ErrOrderNotFound)
		}
		return models.Order{}, fmt.Errorf("storage.ReceiveOrder error: %w", errQuery)
	}
	defer rows.Close()

	var order models.Order
	for rows.Next() {
		var ordRecord schema.OrderRecord
		if errScan := rows.Scan(&ordRecord.OrderID, &ordRecord.CustomerID,
			&ordRecord.ExpirationTime, &ordRecord.ReceivedTime,
			&ordRecord.ReceivedByCustomer, &ordRecord.Refunded,
			&ordRecord.Package, &ordRecord.Weight, &ordRecord.Cost, &ordRecord.PackageCost); errScan != nil {
			return models.Order{}, fmt.Errorf("storage.ReceiveOrder error: %w", errScan)
		}
		order = ordRecord.ToDomain()
	}

	return order, nil
}

func (s *PostgresDB) ReturnOrder(orderId models.ID) (models.Order, error) {
	sql, args, errSql := sq.
		Delete(orderTable).
		Where(sq.Eq{"order_id": orderId}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if errSql != nil {
		return models.Order{}, fmt.Errorf("storage.ReturnOrder error: %w", errSql)
	}

	rows, errQuery := s.db.Query(context.Background(), sql, args...)
	if errQuery != nil {
		if errors.Is(errQuery, pgx.ErrNoRows) {
			return models.Order{}, fmt.Errorf("storage.ReceiveOrder error: %w", ErrOrderNotFound)
		}
		return models.Order{}, fmt.Errorf("storage.ReceiveOrder error: %w", errQuery)
	}
	defer rows.Close()

	var order models.Order
	for rows.Next() {
		var ordRecord schema.OrderRecord
		if errScan := rows.Scan(&ordRecord.OrderID, &ordRecord.CustomerID,
			&ordRecord.ExpirationTime, &ordRecord.ReceivedTime,
			&ordRecord.ReceivedByCustomer, &ordRecord.Refunded,
			&ordRecord.Package, &ordRecord.Weight, &ordRecord.Cost, &ordRecord.PackageCost); errScan != nil {
			return models.Order{}, fmt.Errorf("storage.ReceiveOrder error: %w", errScan)
		}
		order = ordRecord.ToDomain()
	}

	return order, nil
}
