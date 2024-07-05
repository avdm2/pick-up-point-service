package main

import (
	"fmt"
	"homework-1/internal/cli"
	"homework-1/internal/config"
	"homework-1/internal/module"
	"homework-1/internal/storage"
	"os"
)

const (
	cfgPath = "config/config.yaml"
)

func main() {
	cfg, errCfg := config.LoadConfig(cfgPath)
	if errCfg != nil {
		fmt.Printf("Config error. %s\n", errCfg)
		os.Exit(1)
	}

	connUrl := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.DatabaseConfig.User, cfg.DatabaseConfig.Password,
		cfg.DatabaseConfig.Host, cfg.DatabaseConfig.Port,
		cfg.DatabaseConfig.Name)

	s, errStorage := storage.NewStorage(connUrl)
	if errStorage != nil {
		fmt.Printf("storageImpl error. %s\n", errStorage)
		os.Exit(1)
	}

	ordersModule := module.NewModule(module.Deps{
		Storage: s,
	})

	c := cli.NewCLI(cli.Deps{
		Module: ordersModule,
	})
	if errCLI := c.Run(); errCLI != nil {
		fmt.Printf("CLI error. %s\n", errCLI)
		os.Exit(1)
	}
}
