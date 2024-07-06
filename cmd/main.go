package main

import (
	"fmt"
	"homework-1/internal/cli"
	"homework-1/internal/config"
	"homework-1/internal/infrastructure/kafka"
	"homework-1/internal/infrastructure/messaging"
	"homework-1/internal/module"
	"homework-1/internal/storage"
	"os"
)

const (
	cfgPath = "config/config.yaml"
)

func main() {
	cfg := getConfig()

	s := initDB(cfg)

	ordersModule := module.NewModule(module.Deps{
		Storage: s,
	})

	sender := initSender(cfg)

	receiver := initReceiver(cfg)

	runCLI(ordersModule, sender, receiver)
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

func initSender(cfg *config.Config) *kafka.KafkaSender {
	kafkaProducer, errKafka := messaging.NewKafkaProducer(cfg.Brokers)
	if errKafka != nil {
		fmt.Printf("error while initializing messaging producer: %s\n", errKafka)
		os.Exit(1)
	}

	return kafka.NewKafkaSender(kafkaProducer, cfg.Topic)
}

func initReceiver(cfg *config.Config) *kafka.KafkaReceiver {
	if !cfg.ConsolePrinting {
		return nil
	}

	kafkaConsumer, errKafka := messaging.NewKafkaConsumer(cfg.Brokers)
	if errKafka != nil {
		fmt.Printf("error while initializing messaging consumer: %s\n", errKafka)
		os.Exit(1)
	}

	return kafka.NewKafkaReceiver(kafkaConsumer, cfg.Topic)
}

func runCLI(ordersModule *module.Module, sender *kafka.KafkaSender, receiver *kafka.KafkaReceiver) {
	c := cli.NewCLI(cli.Deps{
		Module:   ordersModule,
		Sender:   sender,
		Receiver: receiver,
	})

	if errCLI := c.Run(); errCLI != nil {
		fmt.Printf("error while running cli: %s\n", errCLI)
		return
	}
}
