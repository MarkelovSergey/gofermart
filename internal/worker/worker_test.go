package worker

import (
	"context"
	"testing"
	"time"

	"github.com/MarkelovSergey/gofermart/internal/accrual"
	"github.com/MarkelovSergey/gofermart/internal/models"
	"github.com/MarkelovSergey/gofermart/internal/storage"
)

// mockStorage реализация хранилища для тестов worker.
type mockStorage struct {
	pendingOrders []models.Order
	updatedOrders map[string]models.OrderStatus
}

func newMockStorage() *mockStorage {
	return &mockStorage{
		pendingOrders: make([]models.Order, 0),
		updatedOrders: make(map[string]models.OrderStatus),
	}
}

func (m *mockStorage) CreateUser(ctx context.Context, login, passwordHash string) (*models.User, error) {
	return nil, nil
}

func (m *mockStorage) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	return nil, storage.ErrUserNotFound
}

func (m *mockStorage) CreateOrder(ctx context.Context, userID int, orderNumber string) error {
	return nil
}

func (m *mockStorage) GetOrdersByUserID(ctx context.Context, userID int) ([]models.Order, error) {
	return nil, nil
}

func (m *mockStorage) GetBalance(ctx context.Context, userID int) (*models.Balance, error) {
	return &models.Balance{}, nil
}

func (m *mockStorage) CreateWithdrawal(ctx context.Context, userID int, order string, sum float64) error {
	return nil
}

func (m *mockStorage) GetWithdrawalsByUserID(ctx context.Context, userID int) ([]models.Withdrawal, error) {
	return nil, nil
}

func (m *mockStorage) GetPendingOrders(ctx context.Context) ([]models.Order, error) {
	return m.pendingOrders, nil
}

func (m *mockStorage) UpdateOrderStatus(ctx context.Context, orderNumber string, status models.OrderStatus, accrualVal *float64) error {
	m.updatedOrders[orderNumber] = status
	return nil
}

func (m *mockStorage) Close() error {
	return nil
}

func TestNewWorker(t *testing.T) {
	store := newMockStorage()
	client := accrual.NewClient("http://localhost:8081")
	interval := 5 * time.Second

	w := NewWorker(store, client, interval)

	if w == nil {
		t.Fatal("NewWorker() returned nil")
	}
	if w.storage != store {
		t.Error("worker storage not set correctly")
	}
	if w.accrualClient != client {
		t.Error("worker accrualClient not set correctly")
	}
	if w.interval != interval {
		t.Errorf("worker interval = %v, want %v", w.interval, interval)
	}
}

func TestWorker_Stop(t *testing.T) {
	store := newMockStorage()
	client := accrual.NewClient("http://localhost:8081")
	w := NewWorker(store, client, time.Second)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Запускаем worker в горутине
	done := make(chan struct{})
	go func() {
		w.Start(ctx)
		close(done)
	}()

	// Даём воркеру время запуститься
	time.Sleep(100 * time.Millisecond)

	// Останавливаем воркер
	w.Stop()

	// Ждём завершения
	select {
	case <-done:
		// Успешно остановлен
	case <-time.After(2 * time.Second):
		t.Error("worker did not stop in time")
	}
}

func TestWorker_StartWithContextCancel(t *testing.T) {
	store := newMockStorage()
	client := accrual.NewClient("http://localhost:8081")
	w := NewWorker(store, client, time.Second)

	ctx, cancel := context.WithCancel(context.Background())

	// Запускаем worker в горутине
	done := make(chan struct{})
	go func() {
		w.Start(ctx)
		close(done)
	}()

	// Даём воркеру время запуститься
	time.Sleep(100 * time.Millisecond)

	// Отменяем контекст
	cancel()

	// Ждём завершения
	select {
	case <-done:
		// Успешно остановлен
	case <-time.After(2 * time.Second):
		t.Error("worker did not stop in time after context cancel")
	}
}
