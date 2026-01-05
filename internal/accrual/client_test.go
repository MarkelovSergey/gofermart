package accrual

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MarkelovSergey/gofermart/internal/models"
)

func TestClient_GetOrderAccrual(t *testing.T) {
	t.Run("successful response", func(t *testing.T) {
		accrual := 500.0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Errorf("Method = %v, want %v", r.Method, http.MethodGet)
			}
			if r.URL.Path != "/api/orders/12345678903" {
				t.Errorf("Path = %v, want /api/orders/12345678903", r.URL.Path)
			}

			resp := models.AccrualResponse{
				Order:   "12345678903",
				Status:  "PROCESSED",
				Accrual: &accrual,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		result, err := client.GetOrderAccrual(context.Background(), "12345678903")

		if err != nil {
			t.Fatalf("GetOrderAccrual() error = %v", err)
		}
		if result.Order != "12345678903" {
			t.Errorf("Order = %v, want 12345678903", result.Order)
		}
		if result.Status != "PROCESSED" {
			t.Errorf("Status = %v, want PROCESSED", result.Status)
		}
		if result.Accrual == nil || *result.Accrual != 500.0 {
			t.Errorf("Accrual = %v, want 500.0", result.Accrual)
		}
	})

	t.Run("order not registered (204)", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		_, err := client.GetOrderAccrual(context.Background(), "12345678903")

		if err != ErrOrderNotRegistered {
			t.Errorf("GetOrderAccrual() error = %v, want %v", err, ErrOrderNotRegistered)
		}
	})

	t.Run("too many requests (429)", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Retry-After", "60")
			w.WriteHeader(http.StatusTooManyRequests)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		_, err := client.GetOrderAccrual(context.Background(), "12345678903")

		if err != ErrTooManyRequests {
			t.Errorf("GetOrderAccrual() error = %v, want %v", err, ErrTooManyRequests)
		}

		retryAfter := client.GetRetryAfter()
		if retryAfter != 60*time.Second {
			t.Errorf("GetRetryAfter() = %v, want 60s", retryAfter)
		}
	})

	t.Run("server error (500)", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client := NewClient(server.URL)
		_, err := client.GetOrderAccrual(context.Background(), "12345678903")

		if err != ErrServerError {
			t.Errorf("GetOrderAccrual() error = %v, want %v", err, ErrServerError)
		}
	})
}

func TestClient_GetRetryAfter(t *testing.T) {
	client := NewClient("http://localhost")

	// По умолчанию должен возвращать 1 секунду
	retryAfter := client.GetRetryAfter()
	if retryAfter != time.Second {
		t.Errorf("GetRetryAfter() = %v, want 1s", retryAfter)
	}

	// После установки значения
	client.retryAfter = 30 * time.Second
	retryAfter = client.GetRetryAfter()
	if retryAfter != 30*time.Second {
		t.Errorf("GetRetryAfter() = %v, want 30s", retryAfter)
	}
}

func TestClient_ResetRetryAfter(t *testing.T) {
	client := NewClient("http://localhost")
	client.retryAfter = 30 * time.Second

	client.ResetRetryAfter()

	if client.retryAfter != 0 {
		t.Errorf("retryAfter = %v, want 0", client.retryAfter)
	}
}
