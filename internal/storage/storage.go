package storage

import (
	"context"
	"errors"
	"time"

	"github.com/MarkelovSergey/gofermart/internal/migration"
	"github.com/MarkelovSergey/gofermart/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUserExists           = errors.New("user already exists")                   // Возвращается при попытке создать пользователя с существующим логином.
	ErrUserNotFound         = errors.New("user not found")                        // Возвращается когда пользователь не найден.
	ErrOrderExistsSameUser  = errors.New("order already exists for this user")    // Возвращается когда заказ уже загружен этим же пользователем.
	ErrOrderExistsOtherUser = errors.New("order already exists for another user") // Возвращается когда заказ уже загружен другим пользователем.
	ErrInsufficientFunds    = errors.New("insufficient funds")                    // Возвращается при недостаточном балансе для списания.
)

type Storage interface {
	CreateUser(ctx context.Context, login, passwordHash string) (*models.User, error)                             // Создаёт нового пользователя.
	GetUserByLogin(ctx context.Context, login string) (*models.User, error)                                       // Возвращает пользователя по логину.
	CreateOrder(ctx context.Context, userID int, orderNumber string) error                                        // Создаёт новый заказ.
	GetOrdersByUserID(ctx context.Context, userID int) ([]models.Order, error)                                    // Возвращает список заказов пользователя.
	GetBalance(ctx context.Context, userID int) (*models.Balance, error)                                          // Возвращает баланс пользователя.
	CreateWithdrawal(ctx context.Context, userID int, order string, sum float64) error                            // Создаёт запись о списании баллов.
	GetWithdrawalsByUserID(ctx context.Context, userID int) ([]models.Withdrawal, error)                          // Возвращает список списаний пользователя.
	GetPendingOrders(ctx context.Context) ([]models.Order, error)                                                 // Возвращает заказы, ожидающие обработки.
	UpdateOrderStatus(ctx context.Context, orderNumber string, status models.OrderStatus, accrual *float64) error // Обновляет статус и начисление заказа.
	Close() error                                                                                                 // Закрывает соединение с хранилищем.
}

type PostgresStorage struct {
	pool *pgxpool.Pool
}

func NewPostgresStorage(ctx context.Context, dsn string) (Storage, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	storage := &PostgresStorage{pool}

	if err := migration.RunMigrations(dsn); err != nil {
		return nil, err
	}

	return storage, nil
}

func (s *PostgresStorage) CreateUser(ctx context.Context, login, passwordHash string) (*models.User, error) {
	var user models.User
	err := s.pool.QueryRow(ctx,
		`INSERT INTO users (login, password_hash) VALUES ($1, $2) 
		 RETURNING id, login, password_hash, created_at`,
		login, passwordHash,
	).Scan(&user.ID, &user.Login, &user.PasswordHash, &user.CreatedAt)

	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrUserExists
		}
		return nil, err
	}

	return &user, nil
}

func (s *PostgresStorage) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	var user models.User
	err := s.pool.QueryRow(ctx,
		`SELECT id, login, password_hash, created_at FROM users WHERE login = $1`,
		login,
	).Scan(&user.ID, &user.Login, &user.PasswordHash, &user.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (s *PostgresStorage) CreateOrder(ctx context.Context, userID int, orderNumber string) error {
	var existingUserID int
	err := s.pool.QueryRow(ctx,
		`SELECT user_id FROM orders WHERE number = $1`,
		orderNumber,
	).Scan(&existingUserID)

	if err == nil {
		if existingUserID == userID {
			return ErrOrderExistsSameUser
		}

		return ErrOrderExistsOtherUser
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	_, err = s.pool.Exec(ctx,
		`INSERT INTO orders (user_id, number, status) VALUES ($1, $2, $3)`,
		userID, orderNumber, models.OrderStatusNew,
	)

	return err
}

func (s *PostgresStorage) GetOrdersByUserID(ctx context.Context, userID int) ([]models.Order, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, user_id, number, status, accrual, uploaded_at 
		 FROM orders WHERE user_id = $1 
		 ORDER BY uploaded_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		var accrual *float64
		err := rows.Scan(&order.ID, &order.UserID, &order.Number, &order.Status, &accrual, &order.UploadedAt)
		if err != nil {
			return nil, err
		}
		order.Accrual = accrual
		orders = append(orders, order)
	}

	return orders, rows.Err()
}

func (s *PostgresStorage) GetBalance(ctx context.Context, userID int) (*models.Balance, error) {
	var (
		balance      models.Balance
		totalAccrual float64
	)

	err := s.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(accrual), 0) FROM orders WHERE user_id = $1 AND status = $2`,
		userID, models.OrderStatusProcessed,
	).Scan(&totalAccrual)
	if err != nil {
		return nil, err
	}

	var totalWithdrawn float64
	err = s.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(sum), 0) FROM withdrawals WHERE user_id = $1`,
		userID,
	).Scan(&totalWithdrawn)
	if err != nil {
		return nil, err
	}

	balance.Withdrawn = totalWithdrawn
	balance.Current = totalAccrual - totalWithdrawn

	return &balance, nil
}

func (s *PostgresStorage) CreateWithdrawal(ctx context.Context, userID int, order string, sum float64) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var totalAccrual, totalWithdrawn float64
	err = tx.QueryRow(ctx,
		`SELECT COALESCE(SUM(accrual), 0) FROM orders WHERE user_id = $1 AND status = $2`,
		userID, models.OrderStatusProcessed,
	).Scan(&totalAccrual)
	if err != nil {
		return err
	}

	err = tx.QueryRow(ctx,
		`SELECT COALESCE(SUM(sum), 0) FROM withdrawals WHERE user_id = $1`,
		userID,
	).Scan(&totalWithdrawn)
	if err != nil {
		return err
	}

	currentBalance := totalAccrual - totalWithdrawn
	if currentBalance < sum {
		return ErrInsufficientFunds
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO withdrawals (user_id, order_number, sum, processed_at) VALUES ($1, $2, $3, $4)`,
		userID, order, sum, time.Now(),
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *PostgresStorage) GetWithdrawalsByUserID(ctx context.Context, userID int) ([]models.Withdrawal, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, user_id, order_number, sum, processed_at 
		 FROM withdrawals WHERE user_id = $1 
		 ORDER BY processed_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var withdrawals []models.Withdrawal
	for rows.Next() {
		var w models.Withdrawal
		err := rows.Scan(&w.ID, &w.UserID, &w.Order, &w.Sum, &w.ProcessedAt)
		if err != nil {
			return nil, err
		}
		withdrawals = append(withdrawals, w)
	}

	return withdrawals, rows.Err()
}

func (s *PostgresStorage) GetPendingOrders(ctx context.Context) ([]models.Order, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, user_id, number, status, accrual, uploaded_at 
		 FROM orders WHERE status IN ($1, $2)
		 ORDER BY uploaded_at ASC`,
		models.OrderStatusNew, models.OrderStatusProcessing,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var (
			order   models.Order
			accrual *float64
		)

		err := rows.Scan(&order.ID, &order.UserID, &order.Number, &order.Status, &accrual, &order.UploadedAt)
		if err != nil {
			return nil, err
		}
		order.Accrual = accrual
		orders = append(orders, order)
	}

	return orders, rows.Err()
}

func (s *PostgresStorage) UpdateOrderStatus(ctx context.Context, orderNumber string, status models.OrderStatus, accrual *float64) error {
	var err error
	if accrual != nil {
		_, err = s.pool.Exec(ctx,
			`UPDATE orders SET status = $1, accrual = $2 WHERE number = $3`,
			status, *accrual, orderNumber,
		)
	} else {
		_, err = s.pool.Exec(ctx,
			`UPDATE orders SET status = $1 WHERE number = $2`,
			status, orderNumber,
		)
	}

	return err
}

func (s *PostgresStorage) Close() error {
	s.pool.Close()

	return nil
}

func isUniqueViolation(err error) bool {
	return err != nil && (contains(err.Error(), "unique") || contains(err.Error(), "duplicate") || contains(err.Error(), "23505"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}

	return false
}
