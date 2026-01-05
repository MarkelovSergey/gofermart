package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestUserCredentials_JSON(t *testing.T) {
	creds := UserCredentials{
		Login:    "testuser",
		Password: "testpass",
	}

	data, err := json.Marshal(creds)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded UserCredentials
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if decoded.Login != creds.Login {
		t.Errorf("Login = %v, want %v", decoded.Login, creds.Login)
	}
	if decoded.Password != creds.Password {
		t.Errorf("Password = %v, want %v", decoded.Password, creds.Password)
	}
}

func TestOrder_JSON(t *testing.T) {
	accrual := 500.0
	order := Order{
		Number:     "12345678903",
		Status:     OrderStatusProcessed,
		Accrual:    &accrual,
		UploadedAt: time.Date(2020, 12, 10, 15, 15, 45, 0, time.UTC),
	}

	data, err := json.Marshal(order)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded Order
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if decoded.Number != order.Number {
		t.Errorf("Number = %v, want %v", decoded.Number, order.Number)
	}
	if decoded.Status != order.Status {
		t.Errorf("Status = %v, want %v", decoded.Status, order.Status)
	}
	if decoded.Accrual == nil || *decoded.Accrual != *order.Accrual {
		t.Errorf("Accrual = %v, want %v", decoded.Accrual, order.Accrual)
	}
}

func TestOrder_JSON_WithoutAccrual(t *testing.T) {
	order := Order{
		Number:     "12345678903",
		Status:     OrderStatusProcessing,
		UploadedAt: time.Date(2020, 12, 10, 15, 15, 45, 0, time.UTC),
	}

	data, err := json.Marshal(order)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if _, exists := m["accrual"]; exists {
		t.Error("accrual should be omitted when nil")
	}
}

func TestBalance_JSON(t *testing.T) {
	balance := Balance{
		Current:   500.5,
		Withdrawn: 42.0,
	}

	data, err := json.Marshal(balance)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded Balance
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if decoded.Current != balance.Current {
		t.Errorf("Current = %v, want %v", decoded.Current, balance.Current)
	}
	if decoded.Withdrawn != balance.Withdrawn {
		t.Errorf("Withdrawn = %v, want %v", decoded.Withdrawn, balance.Withdrawn)
	}
}

func TestWithdrawal_JSON(t *testing.T) {
	withdrawal := Withdrawal{
		Order:       "2377225624",
		Sum:         500.0,
		ProcessedAt: time.Date(2020, 12, 9, 16, 9, 57, 0, time.UTC),
	}

	data, err := json.Marshal(withdrawal)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded Withdrawal
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if decoded.Order != withdrawal.Order {
		t.Errorf("Order = %v, want %v", decoded.Order, withdrawal.Order)
	}
	if decoded.Sum != withdrawal.Sum {
		t.Errorf("Sum = %v, want %v", decoded.Sum, withdrawal.Sum)
	}
}

func TestWithdrawRequest_JSON(t *testing.T) {
	req := WithdrawRequest{
		Order: "2377225624",
		Sum:   751.0,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded WithdrawRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if decoded.Order != req.Order {
		t.Errorf("Order = %v, want %v", decoded.Order, req.Order)
	}
	if decoded.Sum != req.Sum {
		t.Errorf("Sum = %v, want %v", decoded.Sum, req.Sum)
	}
}

func TestOrderStatus_Values(t *testing.T) {
	tests := []struct {
		status   OrderStatus
		expected string
	}{
		{OrderStatusNew, "NEW"},
		{OrderStatusProcessing, "PROCESSING"},
		{OrderStatusInvalid, "INVALID"},
		{OrderStatusProcessed, "PROCESSED"},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if string(tt.status) != tt.expected {
				t.Errorf("OrderStatus = %v, want %v", tt.status, tt.expected)
			}
		})
	}
}

func TestAccrualResponse_JSON(t *testing.T) {
	accrual := 500.0
	resp := AccrualResponse{
		Order:   "12345678903",
		Status:  "PROCESSED",
		Accrual: &accrual,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded AccrualResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if decoded.Order != resp.Order {
		t.Errorf("Order = %v, want %v", decoded.Order, resp.Order)
	}
	if decoded.Status != resp.Status {
		t.Errorf("Status = %v, want %v", decoded.Status, resp.Status)
	}
	if decoded.Accrual == nil || *decoded.Accrual != *resp.Accrual {
		t.Errorf("Accrual = %v, want %v", decoded.Accrual, resp.Accrual)
	}
}
