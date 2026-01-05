package config

import (
	"flag"
	"os"
	"sync"
)

type Config struct {
	RunAddress           string // Адрес и порт запуска HTTP-сервера.
	DatabaseURI          string // Строка подключения к базе данных PostgreSQL.
	AccrualSystemAddress string // Адрес системы расчёта начислений баллов лояльности.
}

var (
	flagsRegistered bool
	flagsMu         sync.Mutex
)

func New() *Config {
	cfg := &Config{
		RunAddress: "localhost:8080",
	}

	flagsMu.Lock()
	if !flagsRegistered {
		flag.StringVar(&cfg.RunAddress, "a", "localhost:8080", "адрес и порт запуска сервиса")
		flag.StringVar(&cfg.DatabaseURI, "d", "", "адрес подключения к базе данных")
		flag.StringVar(&cfg.AccrualSystemAddress, "r", "", "адрес системы расчёта начислений")
		flagsRegistered = true
	}
	flagsMu.Unlock()

	if !flag.Parsed() {
		flag.Parse()
	}

	if envRunAddr := os.Getenv("RUN_ADDRESS"); envRunAddr != "" {
		cfg.RunAddress = envRunAddr
	}
	if envDatabaseURI := os.Getenv("DATABASE_URI"); envDatabaseURI != "" {
		cfg.DatabaseURI = envDatabaseURI
	}
	if envAccrualAddr := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); envAccrualAddr != "" {
		cfg.AccrualSystemAddress = envAccrualAddr
	}

	return cfg
}
