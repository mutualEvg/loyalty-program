package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"gofemart/internal/models"
	"gofemart/internal/repository"
	"gofemart/internal/utils"
)

type OrderService struct {
	orderRepo         repository.OrderRepository
	balanceRepo       repository.BalanceRepository
	accrualSystemAddr string
}

type AccrualResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual,omitempty"`
}

func NewOrderService(orderRepo repository.OrderRepository, balanceRepo repository.BalanceRepository) *OrderService {
	return &OrderService{
		orderRepo:   orderRepo,
		balanceRepo: balanceRepo,
	}
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
	existingUserID, err := s.orderRepo.ExistsByNumber(orderNumber)
	if err != nil {
		return err
	}
	if existingUserID != 0 {
		if existingUserID == userID {
			return errors.New("order already uploaded by this user")
		}
		return errors.New("order already uploaded by another user")
	}

	// Create new order
	if err := s.orderRepo.Create(userID, orderNumber, models.OrderStatusNew); err != nil {
		return err
	}

	// Start processing order asynchronously
	go s.processOrderAsync(orderNumber)

	return nil
}

// GetUserOrders retrieves all orders for a user
func (s *OrderService) GetUserOrders(userID int) ([]*models.OrderResponse, error) {
	return s.orderRepo.GetByUserID(userID)
}

// processOrderAsync handles order processing with external accrual system or mock
func (s *OrderService) processOrderAsync(orderNumber string) {
	// Update status to PROCESSING
	s.orderRepo.UpdateStatus(orderNumber, models.OrderStatusProcessing, nil)

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
		s.orderRepo.UpdateStatus(orderNumber, status, &accrual)

		// Get order to find user ID
		order, err := s.orderRepo.GetByNumber(orderNumber)
		if err == nil && order != nil {
			// Update user balance
			s.balanceRepo.UpdateBalance(order.UserID, accrual)
		}
	} else {
		s.orderRepo.UpdateStatus(orderNumber, status, nil)
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
