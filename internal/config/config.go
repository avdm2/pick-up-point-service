package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
)

type Config struct {
	DatabaseConfig `yaml:"database"`
	KafkaConfig    `yaml:"kafka"`
	RedisConfig    `yaml:"redis"`
	HttpConfig     `yaml:"http"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host" env-default:"localhost"`
	Port     int    `yaml:"port" env-default:"5432"`
	User     string `yaml:"user" env-default:"postgres"`
	Password string `yaml:"password" env-default:"admin"`
	Name     string `yaml:"name" env-default:"ozon_hw3"`
}

type KafkaConfig struct {
	Brokers         []string `yaml:"brokers" env-default:"localhost:9091"`
	Topic           string   `yaml:"topic" env-default:"orders"`
	ConsolePrinting bool     `yaml:"console-printing" env-default:"false"`
}

type RedisConfig struct {
	Url      string `yaml:"url" env-default:"localhost:6379"`
	Password string `yaml:"password" env-default:"admin"`
	TTL      int    `yaml:"ttl-seconds" env-default:"60"`
	DB       int    `yaml:"db" env-default:"0"`
}

type HttpConfig struct {
	Port int `yaml:"port" env-default:"8080"`
}

func LoadConfig(path string) (*Config, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("config.LoadConfig error: %w", err)
	}

	var cfg Config
	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		return nil, fmt.Errorf("config.LoadConfig error: %w", err)
	}

	return &cfg, nil
}
