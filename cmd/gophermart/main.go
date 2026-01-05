package main

import (
	"log"

	"github.com/MarkelovSergey/gofermart/internal/app"
	"github.com/MarkelovSergey/gofermart/internal/config"
)

func main() {
	cfg := config.New()

	if cfg.DatabaseURI == "" {
		log.Fatal("database URI is required")
	}

	application := app.New(cfg)

	if err := application.Run(); err != nil {
		log.Fatal("Application error:", err)
	}
}
