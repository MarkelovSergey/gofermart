package models

import (
	"time"
)

// OrderStatus представляет статус обработки заказа в системе.
type OrderStatus string

// Доступные статусы обработки заказов.
const (
	OrderStatusNew        OrderStatus = "NEW"        // Загружен в систему, но не попал в обработку.
	OrderStatusProcessing OrderStatus = "PROCESSING" // Вознаграждение за заказ рассчитывается.
	OrderStatusInvalid    OrderStatus = "INVALID"    // Система расчёта вознаграждений отказала в расчёте.
	OrderStatusProcessed  OrderStatus = "PROCESSED"  // Данные по заказу проверены и информация о расчёте успешно получена.
)

// User представляет пользователя системы лояльности.
type User struct {
	ID           int       `json:"-"`     // Уникальный идентификатор пользователя.
	Login        string    `json:"login"` // Уникальный логин пользователя.
	PasswordHash string    `json:"-"`     // Хеш пароля пользователя.
	CreatedAt    time.Time `json:"-"`     // Время создания пользователя.
}

// UserCredentials содержит данные для аутентификации пользователя.
type UserCredentials struct {
	Login    string `json:"login"`    // Логин пользователя.
	Password string `json:"password"` // Пароль пользователя.
}

// Order представляет заказ пользователя в системе лояльности.
type Order struct {
	ID         int         `json:"-"`                 // Уникальный идентификатор записи.
	UserID     int         `json:"-"`                 // Идентификатор пользователя, загрузившего заказ.
	Number     string      `json:"number"`            // Номер заказа.
	Status     OrderStatus `json:"status"`            // Статус обработки заказа.
	Accrual    *float64    `json:"accrual,omitempty"` // Начисленные баллы (может отсутствовать).
	UploadedAt time.Time   `json:"uploaded_at"`       // Время загрузки заказа.
}

// Balance представляет баланс пользователя.
type Balance struct {
	Current   float64 `json:"current"`   // Текущий баланс баллов лояльности.
	Withdrawn float64 `json:"withdrawn"` // Сумма списанных баллов за весь период.
}

// Withdrawal представляет информацию о списании баллов.
type Withdrawal struct {
	ID          int       `json:"-"`            // Уникальный идентификатор записи.
	UserID      int       `json:"-"`            // Идентификатор пользователя.
	Order       string    `json:"order"`        // Номер заказа, в счёт которого списаны баллы.
	Sum         float64   `json:"sum"`          // Сумма списанных баллов.
	ProcessedAt time.Time `json:"processed_at"` // Время списания.
}

// WithdrawRequest представляет запрос на списание баллов.
type WithdrawRequest struct {
	Order string  `json:"order"` // Номер заказа для списания.
	Sum   float64 `json:"sum"`   // Сумма баллов для списания.
}

// AccrualResponse представляет ответ от системы расчёта начислений.
type AccrualResponse struct {
	Order   string   `json:"order"`             // Номер заказа.
	Status  string   `json:"status"`            // Статус расчёта начисления.
	Accrual *float64 `json:"accrual,omitempty"` // Рассчитанные баллы к начислению.
}
