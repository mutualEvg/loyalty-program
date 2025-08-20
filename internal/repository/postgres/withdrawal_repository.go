package postgres

import (
	"context"
	"fmt"
	"time"

	"gofemart/internal/database"
	"gofemart/internal/models"
)

type WithdrawalRepository struct {
	db *database.DB
}

func NewWithdrawalRepository(db *database.DB) *WithdrawalRepository {
	return &WithdrawalRepository{db: db}
}

func (r *WithdrawalRepository) Create(ctx context.Context, userID int, orderNumber string, sum float64) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO withdrawals (user_id, order_number, sum, processed_at) 
		VALUES ($1, $2, $3, $4)`,
		userID, orderNumber, sum, time.Now())
	if err != nil {
		return fmt.Errorf("failed to record withdrawal: %w", err)
	}
	return nil
}

func (r *WithdrawalRepository) GetByUserID(ctx context.Context, userID int) ([]*models.WithdrawalResponse, error) {
	rows, err := r.db.Query(ctx, `
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
