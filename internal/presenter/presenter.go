package presenter

import (
	"time"

	"github.com/MarkelovSergey/gofermart/internal/models"
)

type OrderResponse struct {
	Number     string             `json:"number"`
	Status     models.OrderStatus `json:"status"`
	Accrual    *float64           `json:"accrual,omitempty"`
	UploadedAt string             `json:"uploaded_at"`
}

type WithdrawalResponse struct {
	Order       string  `json:"order"`
	Sum         float64 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}

type Presenter struct{}

func NewPresenter() *Presenter {
	return &Presenter{}
}

func (p *Presenter) FormatOrder(order models.Order) OrderResponse {
	return OrderResponse{
		Number:     order.Number,
		Status:     order.Status,
		Accrual:    order.Accrual,
		UploadedAt: order.UploadedAt.Format(time.RFC3339),
	}
}

func (p *Presenter) FormatOrders(orders []models.Order) []OrderResponse {
	if len(orders) == 0 {
		return []OrderResponse{}
	}

	result := make([]OrderResponse, len(orders))
	for i, order := range orders {
		result[i] = p.FormatOrder(order)
	}

	return result
}

func (p *Presenter) FormatWithdrawal(withdrawal models.Withdrawal) WithdrawalResponse {
	return WithdrawalResponse{
		Order:       withdrawal.Order,
		Sum:         withdrawal.Sum,
		ProcessedAt: withdrawal.ProcessedAt.Format(time.RFC3339),
	}
}

func (p *Presenter) FormatWithdrawals(withdrawals []models.Withdrawal) []WithdrawalResponse {
	if len(withdrawals) == 0 {
		return []WithdrawalResponse{}
	}

	result := make([]WithdrawalResponse, len(withdrawals))
	for i, withdrawal := range withdrawals {
		result[i] = p.FormatWithdrawal(withdrawal)
	}

	return result
}
