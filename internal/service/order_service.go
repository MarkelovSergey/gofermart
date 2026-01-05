package service

import (
	"context"
	"errors"

	"github.com/MarkelovSergey/gofermart/internal/luhn"
	"github.com/MarkelovSergey/gofermart/internal/models"
	"github.com/MarkelovSergey/gofermart/internal/storage"
)

var (
	ErrInvalidOrderNumber = errors.New("invalid order number format")
)

type OrderService interface {
	CreateOrder(ctx context.Context, userID int, orderNumber string) error
	GetOrdersByUserID(ctx context.Context, userID int) ([]models.Order, error)
}

type orderService struct {
	storage storage.Storage
}

func NewOrderService(storage storage.Storage) *orderService {
	return &orderService{storage}
}

func (s *orderService) CreateOrder(ctx context.Context, userID int, orderNumber string) error {
	if orderNumber == "" {
		return errors.New("order number required")
	}

	if !luhn.IsValidOrderNumber(orderNumber) {
		return ErrInvalidOrderNumber
	}

	return s.storage.CreateOrder(ctx, userID, orderNumber)
}

func (s *orderService) GetOrdersByUserID(ctx context.Context, userID int) ([]models.Order, error) {
	return s.storage.GetOrdersByUserID(ctx, userID)
}
