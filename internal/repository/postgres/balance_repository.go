package postgres

import (
	"context"
	"fmt"
	"time"

	"gofemart/internal/database"
	appErrors "gofemart/internal/errors"
	"gofemart/internal/models"

	"github.com/jackc/pgx/v5"
)

type BalanceRepository struct {
	db *database.DB
}

func NewBalanceRepository(db *database.DB) *BalanceRepository {
	return &BalanceRepository{db: db}
}

func (r *BalanceRepository) GetByUserID(ctx context.Context, userID int) (*models.UserBalance, error) {
	var balance models.UserBalance
	err := r.db.QueryRow(ctx, `
		SELECT user_id, current_balance, total_withdrawn 
		FROM user_balances WHERE user_id = $1`, userID).Scan(
		&balance.UserID, &balance.Current, &balance.Withdrawn)
	if err != nil {
		if err == pgx.ErrNoRows {
			return &models.UserBalance{UserID: userID, Current: 0, Withdrawn: 0}, nil
		}
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}
	return &balance, nil
}

func (r *BalanceRepository) Create(ctx context.Context, userID int) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO user_balances (user_id, current_balance, total_withdrawn) 
		VALUES ($1, 0.00, 0.00)`, userID)
	if err != nil {
		return fmt.Errorf("failed to create user balance: %w", err)
	}
	return nil
}

func (r *BalanceRepository) UpdateBalance(ctx context.Context, userID int, amount float64) error {
	_, err := r.db.Exec(ctx, `
		UPDATE user_balances 
		SET current_balance = current_balance + $1, updated_at = $2 
		WHERE user_id = $3`,
		amount, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}
	return nil
}

func (r *BalanceRepository) Withdraw(ctx context.Context, userID int, amount float64) error {

	// Start transaction
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Get current balance
	var currentBalance float64
	err = tx.QueryRow(ctx, `
		SELECT current_balance FROM user_balances WHERE user_id = $1`,
		userID).Scan(&currentBalance)
	if err != nil {
		if err == pgx.ErrNoRows {
			return appErrors.ErrUserBalanceNotFound
		}
		return fmt.Errorf("failed to get current balance: %w", err)
	}

	// Check if user has sufficient balance
	if currentBalance < amount {
		return appErrors.ErrInsufficientFunds
	}

	// Update balance
	_, err = tx.Exec(ctx, `
		UPDATE user_balances 
		SET current_balance = current_balance - $1, 
		    total_withdrawn = total_withdrawn + $1,
		    updated_at = $2 
		WHERE user_id = $3`,
		amount, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
