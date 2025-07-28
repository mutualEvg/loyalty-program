package services

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gofemart/internal/database"
	"gofemart/internal/models"
)

type WithdrawalService struct {
	db *database.DB
}

func NewWithdrawalService(db *database.DB) *WithdrawalService {
	return &WithdrawalService{db: db}
}

// WithdrawPoints processes a withdrawal request
func (s *WithdrawalService) WithdrawPoints(userID int, req *models.WithdrawalRequest) error {
	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Get current balance
	var currentBalance float64
	err = tx.QueryRow(`
		SELECT current_balance FROM user_balances WHERE user_id = $1`,
		userID).Scan(&currentBalance)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("user balance not found")
		}
		return fmt.Errorf("failed to get current balance: %w", err)
	}

	// Check if user has sufficient balance
	if currentBalance < req.Sum {
		return errors.New("insufficient funds")
	}

	// Update balance
	_, err = tx.Exec(`
		UPDATE user_balances 
		SET current_balance = current_balance - $1, 
		    total_withdrawn = total_withdrawn + $1,
		    updated_at = $2 
		WHERE user_id = $3`,
		req.Sum, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	// Record withdrawal
	_, err = tx.Exec(`
		INSERT INTO withdrawals (user_id, order_number, sum, processed_at) 
		VALUES ($1, $2, $3, $4)`,
		userID, req.OrderNumber, req.Sum, time.Now())
	if err != nil {
		return fmt.Errorf("failed to record withdrawal: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetUserWithdrawals retrieves all withdrawals for a user
func (s *WithdrawalService) GetUserWithdrawals(userID int) ([]*models.WithdrawalResponse, error) {
	rows, err := s.db.Query(`
		SELECT order_number, sum, processed_at 
		FROM withdrawals 
		WHERE user_id = $1 
		ORDER BY processed_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user withdrawals: %w", err)
	}
	defer rows.Close()

	var withdrawals []*models.WithdrawalResponse
	for rows.Next() {
		var withdrawal models.WithdrawalResponse
		err := rows.Scan(&withdrawal.OrderNumber, &withdrawal.Sum, &withdrawal.ProcessedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan withdrawal: %w", err)
		}
		withdrawals = append(withdrawals, &withdrawal)
	}

	// Check for errors during iteration
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate over withdrawals: %w", err)
	}

	return withdrawals, nil
}
