package service

import (
	"context"

	"github.com/MarkelovSergey/gofermart/internal/luhn"
	"github.com/MarkelovSergey/gofermart/internal/models"
	"github.com/MarkelovSergey/gofermart/internal/storage"
)

type WithdrawalService interface {
	CreateWithdrawal(ctx context.Context, userID int, order string, sum float64) error
	GetWithdrawalsByUserID(ctx context.Context, userID int) ([]models.Withdrawal, error)
}

type withdrawalService struct {
	storage storage.Storage
}

func NewWithdrawalService(storage storage.Storage) *withdrawalService {
	return &withdrawalService{storage}
}

func (s *withdrawalService) CreateWithdrawal(ctx context.Context, userID int, order string, sum float64) error {
	if !luhn.IsValidOrderNumber(order) {
		return ErrInvalidOrderNumber
	}

	return s.storage.CreateWithdrawal(ctx, userID, order, sum)
}

func (s *withdrawalService) GetWithdrawalsByUserID(ctx context.Context, userID int) ([]models.Withdrawal, error) {
	return s.storage.GetWithdrawalsByUserID(ctx, userID)
}
