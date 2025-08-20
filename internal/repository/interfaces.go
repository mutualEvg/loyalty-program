package repository

import (
	"context"

	"gofemart/internal/models"
)

// UserRepository defines interface for user data operations
type UserRepository interface {
	Create(ctx context.Context, user *models.User, passwordHash string) error
	GetByLogin(ctx context.Context, login string) (*models.User, error)
	ExistsByLogin(ctx context.Context, login string) (bool, error)
}

// OrderRepository defines interface for order data operations
type OrderRepository interface {
	Create(ctx context.Context, userID int, orderNumber string, status models.OrderStatus) error
	GetByNumber(ctx context.Context, orderNumber string) (*models.Order, error)
	GetByUserID(ctx context.Context, userID int) ([]*models.OrderResponse, error)
	UpdateStatus(ctx context.Context, orderNumber string, status models.OrderStatus, accrual *float64) error
	ExistsByNumber(ctx context.Context, orderNumber string) (int, error) // returns userID if exists
}

// BalanceRepository defines interface for balance operations
type BalanceRepository interface {
	GetByUserID(ctx context.Context, userID int) (*models.UserBalance, error)
	Create(ctx context.Context, userID int) error
	UpdateBalance(ctx context.Context, userID int, amount float64) error
	Withdraw(ctx context.Context, userID int, amount float64) error
}

// WithdrawalRepository defines interface for withdrawal operations
type WithdrawalRepository interface {
	Create(ctx context.Context, userID int, orderNumber string, sum float64) error
	GetByUserID(ctx context.Context, userID int) ([]*models.WithdrawalResponse, error)
}
