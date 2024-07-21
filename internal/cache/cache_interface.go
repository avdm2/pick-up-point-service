//go:generate mockgen -source ./cache_interface.go -destination=./mocks/cache_mock.go -package=cache_mock
package cache

import (
	"context"
	"homework-1/internal/models"
	"time"
)

type CacheInterface interface {
	Get(ctx context.Context, key string) ([]models.Order, bool)
	Set(ctx context.Context, key string, orders []models.Order, now time.Time) error
	Delete(ctx context.Context, key string) error
}
