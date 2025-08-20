package errors

import "errors"

// User-related errors
var (
	ErrLoginAlreadyExists  = errors.New("login already exists")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrUserBalanceNotFound = errors.New("user balance not found")
	ErrInsufficientFunds   = errors.New("insufficient funds")
)

// Order-related errors
var (
	ErrInvalidOrderNumberFormat          = errors.New("invalid order number format")
	ErrOrderAlreadyUploadedByUser        = errors.New("order already uploaded by this user")
	ErrOrderAlreadyUploadedByAnotherUser = errors.New("order already uploaded by another user")
)
