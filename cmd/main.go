package main

import (
	"fmt"
	"homework-1/internal/cli"
	"homework-1/internal/module"
	"homework-1/internal/storage"
	"os"
)

const (
	storagePath = "orders.json"
)

func main() {
	s, errStorage := storage.NewStorage(storagePath)
	if errStorage != nil {
		fmt.Printf("Storage error. %s\n", errStorage)
		os.Exit(1)
	}

	ordersModule := module.NewModule(module.Deps{
		Storage: s,
	})

	c := cli.NewCLI(cli.Deps{Module: ordersModule})
	if errCLI := c.Run(); errCLI != nil {
		fmt.Printf("CLI error. %s\n", errCLI)
		os.Exit(1)
	}

	fmt.Printf("end")
}
