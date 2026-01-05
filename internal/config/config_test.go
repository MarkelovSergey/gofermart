package config

import (
	"os"
	"testing"
)

func TestNew(t *testing.T) {
	// Сохраняем текущие значения переменных окружения
	oldRunAddr := os.Getenv("RUN_ADDRESS")
	oldDBURI := os.Getenv("DATABASE_URI")
	oldAccrualAddr := os.Getenv("ACCRUAL_SYSTEM_ADDRESS")
	defer func() {
		os.Setenv("RUN_ADDRESS", oldRunAddr)
		os.Setenv("DATABASE_URI", oldDBURI)
		os.Setenv("ACCRUAL_SYSTEM_ADDRESS", oldAccrualAddr)
	}()

	t.Run("default values", func(t *testing.T) {
		os.Unsetenv("RUN_ADDRESS")
		os.Unsetenv("DATABASE_URI")
		os.Unsetenv("ACCRUAL_SYSTEM_ADDRESS")

		cfg := New()
		if cfg.RunAddress == "" {
			t.Error("RunAddress should have default value")
		}
	})

	t.Run("environment variables override", func(t *testing.T) {
		os.Setenv("RUN_ADDRESS", "localhost:9090")
		os.Setenv("DATABASE_URI", "postgres://test")
		os.Setenv("ACCRUAL_SYSTEM_ADDRESS", "http://localhost:8081")

		cfg := New()

		if cfg.RunAddress != "localhost:9090" {
			t.Errorf("expected RunAddress to be localhost:9090, got %s", cfg.RunAddress)
		}
		if cfg.DatabaseURI != "postgres://test" {
			t.Errorf("expected DatabaseURI to be postgres://test, got %s", cfg.DatabaseURI)
		}
		if cfg.AccrualSystemAddress != "http://localhost:8081" {
			t.Errorf("expected AccrualSystemAddress to be http://localhost:8081, got %s", cfg.AccrualSystemAddress)
		}
	})
}
