package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/MarkelovSergey/gofermart/internal/accrual"
	"github.com/MarkelovSergey/gofermart/internal/auth"
	"github.com/MarkelovSergey/gofermart/internal/config"
	"github.com/MarkelovSergey/gofermart/internal/handlers"
	"github.com/MarkelovSergey/gofermart/internal/middleware"
	"github.com/MarkelovSergey/gofermart/internal/service"
	"github.com/MarkelovSergey/gofermart/internal/storage"
	"github.com/MarkelovSergey/gofermart/internal/worker"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

type App struct {
	server  *http.Server
	storage storage.Storage
	worker  *worker.Worker
}

func New(cfg *config.Config) *App {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	store, err := storage.NewPostgresStorage(ctx, cfg.DatabaseURI)
	if err != nil {
		log.Fatalf("failed to initialize storage: %v", err)
	}

	jwtManager := auth.NewJWTManager("gophermart-secret-key", 24*time.Hour)

	// Создаём сервисы
	userService := service.NewUserService(store, jwtManager)
	orderService := service.NewOrderService(store)
	balanceService := service.NewBalanceService(store)
	withdrawalService := service.NewWithdrawalService(store)

	h := handlers.NewHandler(userService, orderService, balanceService, withdrawalService)

	// Настраиваем роутер
	r := chi.NewRouter()
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.GzipMiddleware)

	// Публичные эндпоинты
	r.Post("/api/user/register", h.Register)
	r.Post("/api/user/login", h.Login)

	// Защищённые эндпоинты
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(jwtManager))
		r.Post("/api/user/orders", h.CreateOrder)
		r.Get("/api/user/orders", h.GetOrders)
		r.Get("/api/user/balance", h.GetBalance)
		r.Post("/api/user/balance/withdraw", h.Withdraw)
		r.Get("/api/user/withdrawals", h.GetWithdrawals)
	})

	// Инициализируем фоновый воркер для обработки заказов
	var w *worker.Worker
	if cfg.AccrualSystemAddress != "" {
		accrualClient := accrual.NewClient(cfg.AccrualSystemAddress)
		w = worker.NewWorker(store, accrualClient, 5*time.Second)
	}

	// Создаём HTTP сервер
	server := &http.Server{
		Addr:    cfg.RunAddress,
		Handler: r,
	}

	return &App{
		server:  server,
		storage: store,
		worker:  w,
	}
}

func (a *App) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Запускаем фоновый воркер
	if a.worker != nil {
		go a.worker.Start(context.Background())
	}

	// Запускаем HTTP сервер
	go func() {
		log.Printf("Starting server on %s", a.server.Addr)
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server failed to start: %v", err)
		}
	}()

	<-ctx.Done()

	log.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	// Останавливаем воркер
	if a.worker != nil {
		a.worker.Stop()
	}

	// Закрываем хранилище
	if a.storage != nil {
		a.storage.Close()
	}

	log.Println("Server exited gracefully")

	return nil
}
