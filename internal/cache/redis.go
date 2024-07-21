package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/redis/go-redis/v9"
	"homework-1/internal/models"
	"log"
	"time"
)

func MustNew(ctx context.Context, url string, pwd string, db int, ttl time.Duration) *Redis {
	client := redis.NewClient(&redis.Options{
		Addr:     url,
		Password: pwd,
		DB:       db,
	})

	status := client.Ping(ctx)
	if status.Err() != nil {
		log.Fatalf("failed to connect to redis: %v", status.Err())
	}

	return &Redis{
		ttl:    ttl,
		client: client,
	}
}

type Redis struct {
	ttl    time.Duration
	client *redis.Client
}

func (r *Redis) Get(ctx context.Context, key string) ([]models.Order, bool) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "cache.Redis.Get")
	defer span.Finish()

	val, errGet := r.client.Get(ctx, key).Result()
	if errGet != nil {
		if errors.Is(errGet, redis.Nil) {
			return nil, false
		}

		log.Printf("failed to fetch key %s: %v", key, errGet)
		return nil, false
	}

	var order []models.Order
	if errUnmarshall := json.Unmarshal([]byte(val), &order); errUnmarshall != nil {
		log.Printf("failed to unmarshal order: %v", errUnmarshall)
		return nil, false
	}

	return order, true
}

func (r *Redis) Set(ctx context.Context, key string, orders []models.Order, now time.Time) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "cache.Redis.Set")
	defer span.Finish()

	data, err := json.Marshal(orders)
	if err != nil {
		return fmt.Errorf("cache.Redis.Set error: %w", err)
	}

	if errSet := r.client.Set(ctx, key, data, r.ttl).Err(); errSet != nil {
		return fmt.Errorf("cache.Redis.Set error: %w", errSet)
	}

	return nil
}

func (r *Redis) Delete(ctx context.Context, key string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "cache.Redis.Delete")
	defer span.Finish()

	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("cache.Redis.Delete error: %w", err)
	}

	return nil
}
