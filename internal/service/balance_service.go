package service

import (
	"context"

	"github.com/MarkelovSergey/gofermart/internal/models"
	"github.com/MarkelovSergey/gofermart/internal/storage"
)

type BalanceService interface {
	GetBalance(ctx context.Context, userID int) (*models.Balance, error)
}

type balanceService struct {
	storage storage.Storage
}

func NewBalanceService(storage storage.Storage) *balanceService {
	return &balanceService{storage}
}

func (s *balanceService) GetBalance(ctx context.Context, userID int) (*models.Balance, error) {
	return s.storage.GetBalance(ctx, userID)
}
