package models

import (
	"time"
)

type Withdrawal struct {
	ID          int       `json:"-" db:"id"`
	UserID      int       `json:"-" db:"user_id"`
	OrderNumber string    `json:"order" db:"order_number"`
	Sum         float64   `json:"sum" db:"sum"`
	ProcessedAt time.Time `json:"processed_at" db:"processed_at"`
}

type WithdrawalRequest struct {
	OrderNumber string  `json:"order"`
	Sum         float64 `json:"sum"`
}

type WithdrawalResponse struct {
	OrderNumber string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}
