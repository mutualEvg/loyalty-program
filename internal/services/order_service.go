package services

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"gofemart/internal/database"
	"gofemart/internal/models"
	"gofemart/internal/utils"
)

type OrderService struct {
	db                *database.DB
	accrualSystemAddr string
}

type AccrualResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual,omitempty"`
}

func NewOrderService(db *database.DB) *OrderService {
	return &OrderService{db: db}
}

// SetAccrualSystemAddress configures the external accrual system address
func (s *OrderService) SetAccrualSystemAddress(addr string) {
	s.accrualSystemAddr = addr
}

// SubmitOrder submits a new order for processing
func (s *OrderService) SubmitOrder(userID int, orderNumber string) error {
	// Validate order number using Luhn algorithm
	if !utils.IsValidLuhn(orderNumber) {
		return errors.New("invalid order number format")
	}

	// Check if order already exists
	var existingUserID int
	err := s.db.QueryRow("SELECT user_id FROM orders WHERE number = $1", orderNumber).Scan(&existingUserID)
	if err == nil {
		if existingUserID == userID {
			return nil // Order already uploaded by this user (200 response)
		}
		return errors.New("order already uploaded by another user")
	} else if err != sql.ErrNoRows {
		return fmt.Errorf("failed to check existing order: %w", err)
	}

	// Create new order
	_, err = s.db.Exec(`
		INSERT INTO orders (user_id, number, status, uploaded_at) 
		VALUES ($1, $2, $3, $4)`,
		userID, orderNumber, models.OrderStatusNew, time.Now())
	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}

	// Start processing order asynchronously
	go s.processOrderAsync(orderNumber)

	return nil
}

// GetUserOrders retrieves all orders for a user
func (s *OrderService) GetUserOrders(userID int) ([]*models.OrderResponse, error) {
	rows, err := s.db.Query(`
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

// processOrderAsync handles order processing with external accrual system or mock
func (s *OrderService) processOrderAsync(orderNumber string) {
	// Update status to PROCESSING
	s.db.Exec("UPDATE orders SET status = $1, updated_at = $2 WHERE number = $3",
		models.OrderStatusProcessing, time.Now(), orderNumber)

	// Simulate processing time
	time.Sleep(2 * time.Second)

	var accrual float64
	var status models.OrderStatus

	if s.accrualSystemAddr != "" {
		// Use real accrual system
		accrual, status = s.queryAccrualSystem(orderNumber)
	} else {
		// Use mock implementation
		accrual, status = s.simulateLoyaltyCalculation(orderNumber)
	}

	// Update order with result
	if accrual > 0 {
		// Update order
		s.db.Exec(`
			UPDATE orders 
			SET status = $1, accrual = $2, updated_at = $3 
			WHERE number = $4`,
			status, accrual, time.Now(), orderNumber)

		// Update user balance
		s.db.Exec(`
			UPDATE user_balances 
			SET current_balance = current_balance + $1, updated_at = $2 
			WHERE user_id = (SELECT user_id FROM orders WHERE number = $3)`,
			accrual, time.Now(), orderNumber)
	} else {
		s.db.Exec(`
			UPDATE orders 
			SET status = $1, updated_at = $2 
			WHERE number = $3`,
			status, time.Now(), orderNumber)
	}
}

// queryAccrualSystem queries the external accrual system
func (s *OrderService) queryAccrualSystem(orderNumber string) (float64, models.OrderStatus) {
	url := fmt.Sprintf("%s/api/orders/%s", s.accrualSystemAddr, orderNumber)

	resp, err := http.Get(url)
	if err != nil {
		// If accrual system is unavailable, mark as processing
		return 0, models.OrderStatusProcessing
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var accrualResp AccrualResponse
		if err := json.NewDecoder(resp.Body).Decode(&accrualResp); err != nil {
			return 0, models.OrderStatusProcessing
		}

		// Map accrual system statuses to our statuses
		switch accrualResp.Status {
		case "REGISTERED":
			return 0, models.OrderStatusNew
		case "PROCESSING":
			return 0, models.OrderStatusProcessing
		case "INVALID":
			return 0, models.OrderStatusInvalid
		case "PROCESSED":
			return accrualResp.Accrual, models.OrderStatusProcessed
		default:
			return 0, models.OrderStatusProcessing
		}

	case http.StatusNoContent:
		// Order not registered in accrual system, mark as invalid
		return 0, models.OrderStatusInvalid

	case http.StatusTooManyRequests:
		// Rate limited, retry later
		time.Sleep(60 * time.Second)
		return s.queryAccrualSystem(orderNumber)

	default:
		// Unknown error, keep processing
		return 0, models.OrderStatusProcessing
	}
}

// simulateLoyaltyCalculation mocks the external loyalty points calculation system
func (s *OrderService) simulateLoyaltyCalculation(orderNumber string) (float64, models.OrderStatus) {
	// Simple mock logic - in real system this would call external service
	// For simulation purposes:
	// - Orders ending in 0-3: get points (order value * 0.1)
	// - Orders ending in 4-6: no points but valid
	// - Orders ending in 7-9: invalid

	lastDigit := orderNumber[len(orderNumber)-1]

	switch {
	case lastDigit >= '0' && lastDigit <= '3':
		// Simulate points calculation based on order value
		// Mock order value calculation (could be fetched from external system)
		mockOrderValue := 100.0 + float64((lastDigit-'0')*50) // $100-$250
		points := mockOrderValue * 0.1                        // 10% cashback
		return points, models.OrderStatusProcessed

	case lastDigit >= '4' && lastDigit <= '6':
		// Valid order but no points
		return 0, models.OrderStatusProcessed

	default:
		// Invalid order
		return 0, models.OrderStatusInvalid
	}
}
