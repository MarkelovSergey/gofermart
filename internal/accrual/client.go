package accrual

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/MarkelovSergey/gofermart/internal/models"
)

// Ошибки клиента системы начислений.
var (
	ErrOrderNotRegistered = errors.New("order not registered in accrual system") // Возвращается когда заказ не зарегистрирован в системе.
	ErrTooManyRequests    = errors.New("too many requests")                      // Возвращается при превышении лимита запросов.
	ErrServerError        = errors.New("accrual server error")                   // Возвращается при ошибке сервера.
)

// Client клиент для взаимодействия с системой расчёта начислений.
type Client struct {
	baseURL    string
	httpClient *http.Client
	retryAfter time.Duration
}

// NewClient создаёт новый клиент системы расчёта начислений.
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetOrderAccrual получает информацию о начислении для заказа.
func (c *Client) GetOrderAccrual(ctx context.Context, orderNumber string) (*models.AccrualResponse, error) {
	url := fmt.Sprintf("%s/api/orders/%s", c.baseURL, orderNumber)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var result models.AccrualResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, err
		}
		return &result, nil

	case http.StatusNoContent:
		return nil, ErrOrderNotRegistered

	case http.StatusTooManyRequests:
		// Читаем заголовок Retry-After
		retryAfterStr := resp.Header.Get("Retry-After")
		if retryAfterStr != "" {
			if seconds, err := strconv.Atoi(retryAfterStr); err == nil {
				c.retryAfter = time.Duration(seconds) * time.Second
			}
		}
		return nil, ErrTooManyRequests

	default:
		return nil, ErrServerError
	}
}

// GetRetryAfter возвращает время ожидания после ошибки 429.
func (c *Client) GetRetryAfter() time.Duration {
	if c.retryAfter > 0 {
		return c.retryAfter
	}
	
	return time.Second
}

// ResetRetryAfter сбрасывает время ожидания.
func (c *Client) ResetRetryAfter() {
	c.retryAfter = 0
}
