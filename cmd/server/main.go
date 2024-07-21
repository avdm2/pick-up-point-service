package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	service "homework-1/internal/api"
	"homework-1/internal/cache"
	"homework-1/internal/config"
	"homework-1/internal/http"
	"homework-1/internal/module"
	"homework-1/internal/storage"
	"homework-1/internal/tracing"
	"homework-1/pkg/api/proto/orders_grpc/v1/orders_grpc/v1"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

const (
	grpcPort = 50051

	cfgPath = "config/config.yaml"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := getConfig()
	s := initDB(cfg)

	ordersModule := module.NewModule(module.Deps{
		Storage: s,
	})

	redis := cache.MustNew(ctx, cfg.RedisConfig.Url, cfg.RedisConfig.Password, cfg.RedisConfig.DB, time.Duration(cfg.RedisConfig.TTL)*time.Second)
	ordersService := &service.OrderService{
		Module: ordersModule,
		Redis:  redis,
	}

	tracing.MustSetup(ctx, "orders-service")

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		http.MustRun(ctx, 5*time.Second, fmt.Sprintf(":%d", cfg.HttpConfig.Port))
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}

		grpcServer := grpc.NewServer()
		orders_grpc.RegisterOrdersServiceServer(grpcServer, ordersService)

		if err = grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	wg.Wait()
}

func getConfig() *config.Config {
	cfg, errCfg := config.LoadConfig(cfgPath)
	if errCfg != nil {
		fmt.Printf("error while reading config: %s\n", errCfg)
		os.Exit(1)
	}
	return cfg
}

func initDB(cfg *config.Config) *storage.PostgresDB {
	connUrl := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.DatabaseConfig.User, cfg.DatabaseConfig.Password,
		cfg.DatabaseConfig.Host, cfg.DatabaseConfig.Port,
		cfg.DatabaseConfig.Name)

	s, errStorage := storage.NewStorage(connUrl)
	if errStorage != nil {
		fmt.Printf("error while initializing storage: %s\n", errStorage)
		os.Exit(1)
	}
	return s
}
