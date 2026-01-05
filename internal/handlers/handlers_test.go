package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MarkelovSergey/gofermart/internal/middleware"
	"github.com/MarkelovSergey/gofermart/internal/models"
	"github.com/MarkelovSergey/gofermart/internal/service"
	"github.com/MarkelovSergey/gofermart/internal/storage"
)

type mockUserService struct {
	registerFunc func(ctx context.Context, login, password string) (string, error)
	loginFunc    func(ctx context.Context, login, password string) (string, error)
}

func (m *mockUserService) RegisterUser(ctx context.Context, login, password string) (string, error) {
	if m.registerFunc != nil {
		return m.registerFunc(ctx, login, password)
	}
	return "", errors.New("not implemented")
}

func (m *mockUserService) LoginUser(ctx context.Context, login, password string) (string, error) {
	if m.loginFunc != nil {
		return m.loginFunc(ctx, login, password)
	}
	return "", errors.New("not implemented")
}

type mockOrderService struct {
	createOrderFunc func(ctx context.Context, userID int, orderNumber string) error
	getOrdersFunc   func(ctx context.Context, userID int) ([]models.Order, error)
}

func (m *mockOrderService) CreateOrder(ctx context.Context, userID int, orderNumber string) error {
	if m.createOrderFunc != nil {
		return m.createOrderFunc(ctx, userID, orderNumber)
	}
	return errors.New("not implemented")
}

func (m *mockOrderService) GetOrdersByUserID(ctx context.Context, userID int) ([]models.Order, error) {
	if m.getOrdersFunc != nil {
		return m.getOrdersFunc(ctx, userID)
	}
	return nil, errors.New("not implemented")
}

type mockBalanceService struct {
	getBalanceFunc func(ctx context.Context, userID int) (*models.Balance, error)
}

func (m *mockBalanceService) GetBalance(ctx context.Context, userID int) (*models.Balance, error) {
	if m.getBalanceFunc != nil {
		return m.getBalanceFunc(ctx, userID)
	}
	return nil, errors.New("not implemented")
}

type mockWithdrawalService struct {
	createWithdrawalFunc func(ctx context.Context, userID int, order string, sum float64) error
	getWithdrawalsFunc   func(ctx context.Context, userID int) ([]models.Withdrawal, error)
}

func (m *mockWithdrawalService) CreateWithdrawal(ctx context.Context, userID int, order string, sum float64) error {
	if m.createWithdrawalFunc != nil {
		return m.createWithdrawalFunc(ctx, userID, order, sum)
	}
	return errors.New("not implemented")
}

func (m *mockWithdrawalService) GetWithdrawalsByUserID(ctx context.Context, userID int) ([]models.Withdrawal, error) {
	if m.getWithdrawalsFunc != nil {
		return m.getWithdrawalsFunc(ctx, userID)
	}
	return nil, errors.New("not implemented")
}

func TestHandler_Register(t *testing.T) {
	t.Run("successful registration", func(t *testing.T) {
		userService := &mockUserService{
			registerFunc: func(ctx context.Context, login, password string) (string, error) {
				return "test-token", nil
			},
		}
		h := NewHandler(userService, nil, nil, nil)

		body := `{"login": "testuser", "password": "testpass"}`
		req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		h.Register(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("status = %v, want %v", rr.Code, http.StatusOK)
		}

		authHeader := rr.Header().Get("Authorization")
		if authHeader == "" {
			t.Error("Authorization header should be set")
		}
	})

	t.Run("duplicate login", func(t *testing.T) {
		userService := &mockUserService{
			registerFunc: func(ctx context.Context, login, password string) (string, error) {
				return "", storage.ErrUserExists
			},
		}
		h := NewHandler(userService, nil, nil, nil)

		body := `{"login": "testuser", "password": "testpass"}`
		req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		h.Register(rr, req)

		if rr.Code != http.StatusConflict {
			t.Errorf("status = %v, want %v", rr.Code, http.StatusConflict)
		}
	})

	t.Run("invalid body", func(t *testing.T) {
		h := NewHandler(nil, nil, nil, nil)

		req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewBufferString("invalid json"))
		rr := httptest.NewRecorder()

		h.Register(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("status = %v, want %v", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("empty credentials", func(t *testing.T) {
		h := NewHandler(nil, nil, nil, nil)

		body := `{"login": "", "password": ""}`
		req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		h.Register(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("status = %v, want %v", rr.Code, http.StatusBadRequest)
		}
	})
}

func TestHandler_Login(t *testing.T) {
	t.Run("successful login", func(t *testing.T) {
		userService := &mockUserService{
			loginFunc: func(ctx context.Context, login, password string) (string, error) {
				return "test-token", nil
			},
		}
		h := NewHandler(userService, nil, nil, nil)

		body := `{"login": "testuser", "password": "testpass"}`
		req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		h.Login(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("status = %v, want %v", rr.Code, http.StatusOK)
		}

		authHeader := rr.Header().Get("Authorization")
		if authHeader == "" {
			t.Error("Authorization header should be set")
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		userService := &mockUserService{
			loginFunc: func(ctx context.Context, login, password string) (string, error) {
				return "", errors.New("invalid password")
			},
		}
		h := NewHandler(userService, nil, nil, nil)

		body := `{"login": "testuser", "password": "wrongpass"}`
		req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewBufferString(body))
		rr := httptest.NewRecorder()

		h.Login(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Errorf("status = %v, want %v", rr.Code, http.StatusInternalServerError)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		userService := &mockUserService{
			loginFunc: func(ctx context.Context, login, password string) (string, error) {
				return "", storage.ErrUserNotFound
			},
		}
		h := NewHandler(userService, nil, nil, nil)

		body := `{"login": "unknown", "password": "testpass"}`
		req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewBufferString(body))
		rr := httptest.NewRecorder()

		h.Login(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("status = %v, want %v", rr.Code, http.StatusUnauthorized)
		}
	})
}

func TestHandler_CreateOrder(t *testing.T) {
	t.Run("successful order creation", func(t *testing.T) {
		orderService := &mockOrderService{
			createOrderFunc: func(ctx context.Context, userID int, orderNumber string) error {
				return nil
			},
		}
		h := NewHandler(nil, orderService, nil, nil)

		req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewBufferString("79927398713"))
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, 1)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		h.CreateOrder(rr, req)

		if rr.Code != http.StatusAccepted {
			t.Errorf("status = %v, want %v", rr.Code, http.StatusAccepted)
		}
	})

	t.Run("duplicate order same user", func(t *testing.T) {
		orderService := &mockOrderService{
			createOrderFunc: func(ctx context.Context, userID int, orderNumber string) error {
				return storage.ErrOrderExistsSameUser
			},
		}
		h := NewHandler(nil, orderService, nil, nil)

		req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewBufferString("79927398713"))
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, 1)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		h.CreateOrder(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("status = %v, want %v", rr.Code, http.StatusOK)
		}
	})

	t.Run("duplicate order different user", func(t *testing.T) {
		orderService := &mockOrderService{
			createOrderFunc: func(ctx context.Context, userID int, orderNumber string) error {
				return storage.ErrOrderExistsOtherUser
			},
		}
		h := NewHandler(nil, orderService, nil, nil)

		req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewBufferString("79927398713"))
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, int(2))
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		h.CreateOrder(rr, req)

		if rr.Code != http.StatusConflict {
			t.Errorf("status = %v, want %v", rr.Code, http.StatusConflict)
		}
	})

	t.Run("invalid order number (Luhn check)", func(t *testing.T) {
		orderService := &mockOrderService{
			createOrderFunc: func(ctx context.Context, userID int, orderNumber string) error {
				return service.ErrInvalidOrderNumber
			},
		}
		h := NewHandler(nil, orderService, nil, nil)

		req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewBufferString("12345678904"))
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, 1)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		h.CreateOrder(rr, req)

		if rr.Code != http.StatusUnprocessableEntity {
			t.Errorf("status = %v, want %v", rr.Code, http.StatusUnprocessableEntity)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		h := NewHandler(nil, &mockOrderService{}, nil, nil)

		req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewBufferString("79927398713"))
		rr := httptest.NewRecorder()

		h.CreateOrder(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("status = %v, want %v", rr.Code, http.StatusUnauthorized)
		}
	})
}

func TestHandler_GetOrders(t *testing.T) {
	t.Run("no orders", func(t *testing.T) {
		orderService := &mockOrderService{
			getOrdersFunc: func(ctx context.Context, userID int) ([]models.Order, error) {
				return []models.Order{}, nil
			},
		}
		h := NewHandler(nil, orderService, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, 1)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		h.GetOrders(rr, req)

		if rr.Code != http.StatusNoContent {
			t.Errorf("status = %v, want %v", rr.Code, http.StatusNoContent)
		}
	})

	t.Run("with orders", func(t *testing.T) {
		orderService := &mockOrderService{
			getOrdersFunc: func(ctx context.Context, userID int) ([]models.Order, error) {
				return []models.Order{
					{Number: "79927398713", Status: models.OrderStatusNew, UploadedAt: time.Now()},
				}, nil
			},
		}
		h := NewHandler(nil, orderService, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, 1)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		h.GetOrders(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("status = %v, want %v", rr.Code, http.StatusOK)
		}

		var orders []models.Order
		json.NewDecoder(rr.Body).Decode(&orders)
		if len(orders) != 1 {
			t.Errorf("orders count = %v, want 1", len(orders))
		}
	})
}

func TestHandler_GetBalance(t *testing.T) {
	balanceService := &mockBalanceService{
		getBalanceFunc: func(ctx context.Context, userID int) (*models.Balance, error) {
			return &models.Balance{Current: 500.5, Withdrawn: 42.0}, nil
		},
	}
	h := NewHandler(nil, nil, balanceService, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, 1)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	h.GetBalance(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %v, want %v", rr.Code, http.StatusOK)
	}

	var balance models.Balance
	json.NewDecoder(rr.Body).Decode(&balance)
	if balance.Current != 500.5 {
		t.Errorf("Current = %v, want 500.5", balance.Current)
	}
	if balance.Withdrawn != 42.0 {
		t.Errorf("Withdrawn = %v, want 42.0", balance.Withdrawn)
	}
}

func TestHandler_Withdraw(t *testing.T) {
	t.Run("successful withdrawal", func(t *testing.T) {
		withdrawalService := &mockWithdrawalService{
			createWithdrawalFunc: func(ctx context.Context, userID int, order string, sum float64) error {
				return nil
			},
		}
		h := NewHandler(nil, nil, nil, withdrawalService)

		body := `{"order": "79927398713", "sum": 100}`
		req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewBufferString(body))
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, 1)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		h.Withdraw(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("status = %v, want %v", rr.Code, http.StatusOK)
		}
	})

	t.Run("insufficient funds", func(t *testing.T) {
		withdrawalService := &mockWithdrawalService{
			createWithdrawalFunc: func(ctx context.Context, userID int, order string, sum float64) error {
				return storage.ErrInsufficientFunds
			},
		}
		h := NewHandler(nil, nil, nil, withdrawalService)

		body := `{"order": "12345678903", "sum": 100}`
		req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewBufferString(body))
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, int(2))
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		h.Withdraw(rr, req)

		if rr.Code != http.StatusPaymentRequired {
			t.Errorf("status = %v, want %v", rr.Code, http.StatusPaymentRequired)
		}
	})

	t.Run("invalid order number", func(t *testing.T) {
		withdrawalService := &mockWithdrawalService{
			createWithdrawalFunc: func(ctx context.Context, userID int, order string, sum float64) error {
				return service.ErrInvalidOrderNumber
			},
		}
		h := NewHandler(nil, nil, nil, withdrawalService)

		body := `{"order": "invalid", "sum": 100}`
		req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewBufferString(body))
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, 1)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		h.Withdraw(rr, req)

		if rr.Code != http.StatusUnprocessableEntity {
			t.Errorf("status = %v, want %v", rr.Code, http.StatusUnprocessableEntity)
		}
	})
}

func TestHandler_GetWithdrawals(t *testing.T) {
	t.Run("no withdrawals", func(t *testing.T) {
		withdrawalService := &mockWithdrawalService{
			getWithdrawalsFunc: func(ctx context.Context, userID int) ([]models.Withdrawal, error) {
				return []models.Withdrawal{}, nil
			},
		}
		h := NewHandler(nil, nil, nil, withdrawalService)

		req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, 1)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		h.GetWithdrawals(rr, req)

		if rr.Code != http.StatusNoContent {
			t.Errorf("status = %v, want %v", rr.Code, http.StatusNoContent)
		}
	})

	t.Run("with withdrawals", func(t *testing.T) {
		withdrawalService := &mockWithdrawalService{
			getWithdrawalsFunc: func(ctx context.Context, userID int) ([]models.Withdrawal, error) {
				return []models.Withdrawal{
					{Order: "79927398713", Sum: 100.0, ProcessedAt: time.Now()},
				}, nil
			},
		}
		h := NewHandler(nil, nil, nil, withdrawalService)

		req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, 1)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		h.GetWithdrawals(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("status = %v, want %v", rr.Code, http.StatusOK)
		}

		var withdrawals []models.Withdrawal
		json.NewDecoder(rr.Body).Decode(&withdrawals)
		if len(withdrawals) != 1 {
			t.Errorf("withdrawals count = %v, want 1", len(withdrawals))
		}
	})
}

func TestHandler_ErrorHandling(t *testing.T) {
	t.Run("GetOrders error", func(t *testing.T) {
		orderService := &mockOrderService{
			getOrdersFunc: func(ctx context.Context, userID int) ([]models.Order, error) {
				return nil, errors.New("database error")
			},
		}
		h := NewHandler(nil, orderService, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, 1)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		h.GetOrders(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Errorf("status = %v, want %v", rr.Code, http.StatusInternalServerError)
		}
	})

	t.Run("GetBalance error", func(t *testing.T) {
		balanceService := &mockBalanceService{
			getBalanceFunc: func(ctx context.Context, userID int) (*models.Balance, error) {
				return nil, errors.New("database error")
			},
		}
		h := NewHandler(nil, nil, balanceService, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, 1)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		h.GetBalance(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Errorf("status = %v, want %v", rr.Code, http.StatusInternalServerError)
		}
	})

	t.Run("GetWithdrawals error", func(t *testing.T) {
		withdrawalService := &mockWithdrawalService{
			getWithdrawalsFunc: func(ctx context.Context, userID int) ([]models.Withdrawal, error) {
				return nil, errors.New("database error")
			},
		}
		h := NewHandler(nil, nil, nil, withdrawalService)

		req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, 1)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		h.GetWithdrawals(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Errorf("status = %v, want %v", rr.Code, http.StatusInternalServerError)
		}
	})
}
