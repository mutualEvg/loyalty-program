package postgres

import (
	"context"
	"fmt"
	"time"

	"gofemart/internal/database"
	"gofemart/internal/models"

	"github.com/jackc/pgx/v5"
)

type OrderRepository struct {
	db *database.DB
}

func NewOrderRepository(db *database.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(ctx context.Context, userID int, orderNumber string, status models.OrderStatus) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO orders (user_id, number, status, uploaded_at) 
		VALUES ($1, $2, $3, $4)`,
		userID, orderNumber, status, time.Now())
	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}
	return nil
}

func (r *OrderRepository) GetByNumber(ctx context.Context, orderNumber string) (*models.Order, error) {
	var order models.Order
	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, number, status, accrual, uploaded_at, updated_at 
		FROM orders WHERE number = $1`, orderNumber).Scan(
		&order.ID, &order.UserID, &order.Number, &order.Status,
		&order.Accrual, &order.UploadedAt, &order.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}
	return &order, nil
}

func (r *OrderRepository) GetByUserID(ctx context.Context, userID int) ([]*models.OrderResponse, error) {
	rows, err := r.db.Query(ctx, `
		SELECT number, status, accrual, uploaded_at 
		FROM orders 
		WHERE user_id = $1 
		ORDER BY uploaded_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user orders: %w", err)
	}
	defer rows.Close()

	var orders []*models.OrderResponse
	for rows.Next() {
		var order models.OrderResponse
		err := rows.Scan(&order.Number, &order.Status, &order.Accrual, &order.UploadedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, &order)
	}

	// Check for errors during iteration
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate over orders: %w", err)
	}

	return orders, nil
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, orderNumber string, status models.OrderStatus, accrual *float64) error {
	if accrual != nil {
		_, err := r.db.Exec(ctx, `
			UPDATE orders 
			SET status = $1, accrual = $2, updated_at = $3 
			WHERE number = $4`,
			status, *accrual, time.Now(), orderNumber)
		return err
	} else {
		_, err := r.db.Exec(ctx, `
			UPDATE orders 
			SET status = $1, updated_at = $2 
			WHERE number = $3`,
			status, time.Now(), orderNumber)
		return err
	}
}

func (r *OrderRepository) ExistsByNumber(ctx context.Context, orderNumber string) (int, error) {
	var userID int
	err := r.db.QueryRow(ctx, "SELECT user_id FROM orders WHERE number = $1", orderNumber).Scan(&userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, nil // Order doesn't exist
		}
		return 0, fmt.Errorf("failed to check existing order: %w", err)
	}
	return userID, nil
}
