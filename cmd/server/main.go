package main

import (
	"context"
	"fmt"
	"homework-1/pkg/api/proto/orders_grpc/v1/orders_grpc/v1"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	service "homework-1/internal/api"
	"homework-1/internal/config"
	"homework-1/internal/module"
	"homework-1/internal/storage"
)

const (
	grpcPort = 50051

	cfgPath = "config/config.yaml"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	cfg := getConfig()
	s := initDB(cfg)

	ordersModule := module.NewModule(module.Deps{
		Storage: s,
	})

	ordersService := &service.OrderService{
		Module: ordersModule,
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	orders_grpc.RegisterOrdersServiceServer(grpcServer, ordersService)

	if err = grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
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
