package repository

import (
	"gofemart/internal/models"
)

// UserRepository defines interface for user data operations
type UserRepository interface {
	Create(user *models.User, passwordHash string) error
	GetByLogin(login string) (*models.User, error)
	ExistsByLogin(login string) (bool, error)
}

// OrderRepository defines interface for order data operations
type OrderRepository interface {
	Create(userID int, orderNumber string, status models.OrderStatus) error
	GetByNumber(orderNumber string) (*models.Order, error)
	GetByUserID(userID int) ([]*models.OrderResponse, error)
	UpdateStatus(orderNumber string, status models.OrderStatus, accrual *float64) error
	ExistsByNumber(orderNumber string) (int, error) // returns userID if exists
}

// BalanceRepository defines interface for balance operations
type BalanceRepository interface {
	GetByUserID(userID int) (*models.UserBalance, error)
	Create(userID int) error
	UpdateBalance(userID int, amount float64) error
	Withdraw(userID int, amount float64) error
}

// WithdrawalRepository defines interface for withdrawal operations
type WithdrawalRepository interface {
	Create(userID int, orderNumber string, sum float64) error
	GetByUserID(userID int) ([]*models.WithdrawalResponse, error)
}
