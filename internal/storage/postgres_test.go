package storage

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"homework-1/internal/config"
	"homework-1/internal/models"
	"testing"
	"time"
)

const (
	path = "../../config/config.yaml"
)

func getConnUrl(path string) (string, error) {
	cfg, errCfg := config.LoadConfig(path)
	if errCfg != nil {
		return "", fmt.Errorf("storage.getConnUrl error: %w", errCfg)
	}

	connUrl := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.DatabaseConfig.User, cfg.DatabaseConfig.Password,
		cfg.DatabaseConfig.Host, cfg.DatabaseConfig.Port,
		cfg.DatabaseConfig.Name)

	return connUrl, nil
}

func clearDB(connURL string) error {
	db, err := pgxpool.New(context.Background(), connURL)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(context.Background(), "TRUNCATE TABLE orders CASCADE;")
	return err
}

func setupDB(t *testing.T, connURL string) {
	err := clearDB(connURL)
	require.NoError(t, err)

	db, err := NewStorage(connURL)
	require.NoError(t, err)

	// Add initial data for tests that need it
	initialOrder := models.Order{
		OrderID:        models.ID(1),
		CustomerID:     models.ID(1),
		ExpirationTime: time.Now().Add(time.Hour),
		Package:        "box",
		Weight:         10,
		Cost:           100,
		PackageCost:    10,
	}
	err = db.AddOrder(initialOrder)
	require.NoError(t, err)
}

func TestPostgresDB_AddOrder(t *testing.T) {
	t.Run("[OK] AddOrder", func(t *testing.T) {
		t.Parallel()

		connURL, err := getConnUrl(path)
		require.NoError(t, err)
		err = clearDB(connURL)
		require.NoError(t, err)
		db, err := NewStorage(connURL)
		require.NoError(t, err)

		order := models.Order{
			OrderID:        models.ID(2),
			CustomerID:     models.ID(2),
			ExpirationTime: time.Now().Add(time.Hour),
			Package:        "box",
			Weight:         10,
			Cost:           100,
			PackageCost:    10,
		}

		err = db.AddOrder(order)
		assert.NoError(t, err)
	})
}

func TestPostgresDB_GetOrder(t *testing.T) {
	t.Run("[OK] GetOrder", func(t *testing.T) {
		t.Parallel()

		connURL, err := getConnUrl(path)
		require.NoError(t, err)
		setupDB(t, connURL)
		db, err := NewStorage(connURL)
		require.NoError(t, err)

		orderID := models.ID(1)

		order, err := db.GetOrder(orderID)
		assert.NoError(t, err)
		assert.Equal(t, orderID, order.OrderID)
	})
}

func TestPostgresDB_GetCustomersOrders(t *testing.T) {
	t.Run("[OK] GetCustomersOrders", func(t *testing.T) {
		t.Parallel()

		connURL, err := getConnUrl(path)
		require.NoError(t, err)
		setupDB(t, connURL)
		db, err := NewStorage(connURL)
		require.NoError(t, err)

		customerID := models.ID(1)

		orders, err := db.GetCustomersOrders(customerID)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(orders))
	})
}

func TestPostgresDB_GetRefunds(t *testing.T) {
	t.Run("[OK] GetRefunds", func(t *testing.T) {
		t.Parallel()

		connURL, err := getConnUrl(path)
		require.NoError(t, err)
		setupDB(t, connURL)
		db, err := NewStorage(connURL)
		require.NoError(t, err)

		order, _ := db.GetOrder(models.ID(1))
		order.Refunded = true
		err = db.ChangeOrder(order)

		refunds, err := db.GetRefunds()
		assert.NoError(t, err)
		assert.Equal(t, 1, len(refunds))
	})
}

func TestPostgresDB_ChangeOrder(t *testing.T) {
	t.Run("[OK] ChangeOrder", func(t *testing.T) {
		t.Parallel()

		connURL, err := getConnUrl(path)
		require.NoError(t, err)
		setupDB(t, connURL)
		db, err := NewStorage(connURL)
		require.NoError(t, err)

		order := models.Order{
			OrderID:            models.ID(1),
			CustomerID:         models.ID(1),
			ReceivedByCustomer: true,
			Refunded:           true,
			ReceivedTime:       time.Now().Add(-time.Hour),
		}
		err = db.ChangeOrder(order)
		assert.NoError(t, err)

		order, _ = db.GetOrder(models.ID(1))
		assert.Equal(t, true, order.ReceivedByCustomer)
		assert.Equal(t, true, order.Refunded)
	})
}

func TestPostgresDB_ReceiveOrder(t *testing.T) {
	t.Run("[OK] ReceiveOrder", func(t *testing.T) {
		t.Parallel()

		connURL, err := getConnUrl(path)
		require.NoError(t, err)

		setupDB(t, connURL)

		db, err := NewStorage(connURL)
		require.NoError(t, err)

		orderID := models.ID(1)

		order, err := db.ReceiveOrder(orderID)
		assert.NoError(t, err)

		order, _ = db.GetOrder(orderID)
		assert.Equal(t, true, order.ReceivedByCustomer)
	})
}

func TestPostgresDB_ReturnOrder(t *testing.T) {
	t.Run("[OK] ReturnOrder", func(t *testing.T) {
		t.Parallel()

		connURL, err := getConnUrl(path)
		require.NoError(t, err)
		setupDB(t, connURL)
		db, err := NewStorage(connURL)
		require.NoError(t, err)

		orderID := models.ID(1)
		err = db.ReturnOrder(orderID)
		assert.NoError(t, err)

		order, _ := db.GetOrder(orderID)
		assert.Equal(t, models.Order{}, order)
	})
}
