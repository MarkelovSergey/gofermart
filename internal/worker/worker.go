package worker

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/MarkelovSergey/gofermart/internal/accrual"
	"github.com/MarkelovSergey/gofermart/internal/models"
	"github.com/MarkelovSergey/gofermart/internal/storage"
)

type Worker struct {
	storage       storage.Storage
	accrualClient *accrual.Client
	interval      time.Duration
	done          chan struct{}
}

func NewWorker(storage storage.Storage, accrualClient *accrual.Client, interval time.Duration) *Worker {
	return &Worker{
		storage:       storage,
		accrualClient: accrualClient,
		interval:      interval,
		done:          make(chan struct{}),
	}
}

// Start запускает фоновую обработку заказов.
func (w *Worker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.done:
			return
		case <-ticker.C:
			w.processOrders(ctx)
		}
	}
}

// Stop останавливает фоновую обработку.
func (w *Worker) Stop() {
	close(w.done)
}

// processOrders обрабатывает заказы, ожидающие проверки.
func (w *Worker) processOrders(ctx context.Context) {
	orders, err := w.storage.GetPendingOrders(ctx)
	if err != nil {
		log.Printf("worker: error getting pending orders: %v", err)
		return
	}

	for _, order := range orders {
		select {
		case <-ctx.Done():
			return
		default:
		}

		result, err := w.accrualClient.GetOrderAccrual(ctx, order.Number)
		if err != nil {
			if errors.Is(err, accrual.ErrTooManyRequests) {
				// Ждём перед следующим запросом
				retryAfter := w.accrualClient.GetRetryAfter()
				log.Printf("worker: rate limited, waiting %v", retryAfter)
				time.Sleep(retryAfter)
				continue
			}
			if errors.Is(err, accrual.ErrOrderNotRegistered) {
				// Заказ не зарегистрирован, пропускаем
				continue
			}
			log.Printf("worker: error getting accrual for order %s: %v", order.Number, err)
			continue
		}

		// Обновляем статус заказа
		var newStatus models.OrderStatus
		switch result.Status {
		case "REGISTERED":
			newStatus = models.OrderStatusNew
		case "PROCESSING":
			newStatus = models.OrderStatusProcessing
		case "INVALID":
			newStatus = models.OrderStatusInvalid
		case "PROCESSED":
			newStatus = models.OrderStatusProcessed
		default:
			continue
		}

		err = w.storage.UpdateOrderStatus(ctx, order.Number, newStatus, result.Accrual)
		if err != nil {
			log.Printf("worker: error updating order %s: %v", order.Number, err)
		}

		// Небольшая пауза между запросами
		time.Sleep(100 * time.Millisecond)
	}
}
