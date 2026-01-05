package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/MarkelovSergey/gofermart/internal/middleware"
	"github.com/MarkelovSergey/gofermart/internal/models"
	"github.com/MarkelovSergey/gofermart/internal/presenter"
	"github.com/MarkelovSergey/gofermart/internal/service"
	"github.com/MarkelovSergey/gofermart/internal/storage"
)

type Handler struct {
	userService       service.UserService
	orderService      service.OrderService
	balanceService    service.BalanceService
	withdrawalService service.WithdrawalService
	presenter         *presenter.Presenter
}

func NewHandler(
	userService service.UserService,
	orderService service.OrderService,
	balanceService service.BalanceService,
	withdrawalService service.WithdrawalService,
) *Handler {
	return &Handler{
		userService,
		orderService,
		balanceService,
		withdrawalService,
		presenter.NewPresenter(),
	}
}

// Register обрабатывает регистрацию пользователя.
// POST /api/user/register
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var creds models.UserCredentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if creds.Login == "" || creds.Password == "" {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	token, err := h.userService.RegisterUser(r.Context(), creds.Login, creds.Password)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			http.Error(w, "login already taken", http.StatusConflict)
			return
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	h.setAuthToken(w, token)
	w.WriteHeader(http.StatusOK)
}

// Login обрабатывает аутентификацию пользователя.
// POST /api/user/login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var creds models.UserCredentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if creds.Login == "" || creds.Password == "" {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	token, err := h.userService.LoginUser(r.Context(), creds.Login, creds.Password)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	h.setAuthToken(w, token)
	w.WriteHeader(http.StatusOK)
}

// CreateOrder обрабатывает загрузку номера заказа.
// POST /api/user/orders
func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	orderNumber := string(body)

	err = h.orderService.CreateOrder(r.Context(), userID, orderNumber)
	if err != nil {
		if errors.Is(err, service.ErrInvalidOrderNumber) {
			http.Error(w, "invalid order number format", http.StatusUnprocessableEntity)
			return
		}
		if errors.Is(err, storage.ErrOrderExistsSameUser) {
			w.WriteHeader(http.StatusOK)
			return
		}
		if errors.Is(err, storage.ErrOrderExistsOtherUser) {
			http.Error(w, "order already submitted by another user", http.StatusConflict)
			return
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// GetOrders возвращает список заказов пользователя.
// GET /api/user/orders
func (h *Handler) GetOrders(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	orders, err := h.orderService.GetOrdersByUserID(r.Context(), userID)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.presenter.FormatOrders(orders))
}

// GetBalance возвращает баланс пользователя.
// GET /api/user/balance
func (h *Handler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	balance, err := h.balanceService.GetBalance(r.Context(), userID)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(balance)
}

// Withdraw обрабатывает запрос на списание баллов.
// POST /api/user/balance/withdraw
func (h *Handler) Withdraw(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.WithdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	err := h.withdrawalService.CreateWithdrawal(r.Context(), userID, req.Order, req.Sum)
	if err != nil {
		if errors.Is(err, service.ErrInvalidOrderNumber) {
			http.Error(w, "invalid order number format", http.StatusUnprocessableEntity)
			return
		}
		if errors.Is(err, storage.ErrInsufficientFunds) {
			http.Error(w, "insufficient funds", http.StatusPaymentRequired)
			return
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetWithdrawals возвращает историю списаний пользователя.
// GET /api/user/withdrawals
func (h *Handler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	withdrawals, err := h.withdrawalService.GetWithdrawalsByUserID(r.Context(), userID)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.presenter.FormatWithdrawals(withdrawals))
}

func (h *Handler) setAuthToken(w http.ResponseWriter, token string) {
	w.Header().Set("Authorization", "Bearer "+token)
}
