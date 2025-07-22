package models

import (
	"time"
)

type User struct {
	ID           int       `json:"id" db:"id"`
	Login        string    `json:"login" db:"login"`
	PasswordHash string    `json:"-" db:"password_hash"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type UserBalance struct {
	UserID    int     `json:"-" db:"user_id"`
	Current   float64 `json:"current" db:"current_balance"`
	Withdrawn float64 `json:"withdrawn" db:"total_withdrawn"`
}

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
