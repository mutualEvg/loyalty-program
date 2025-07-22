package services

import (
	"database/sql"
	"errors"
	"fmt"

	"gofemart/internal/database"
	"gofemart/internal/models"
	"gofemart/internal/utils"
)

type UserService struct {
	db *database.DB
}

func NewUserService(db *database.DB) *UserService {
	return &UserService{db: db}
}

// Register creates a new user account
func (s *UserService) Register(req *models.RegisterRequest) (*models.User, string, error) {
	// Check if user already exists
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE login = $1)", req.Login).Scan(&exists)
	if err != nil {
		return nil, "", fmt.Errorf("failed to check if user exists: %w", err)
	}
	if exists {
		return nil, "", errors.New("login already exists")
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, "", fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	var user models.User
	err = s.db.QueryRow(`
		INSERT INTO users (login, password_hash) 
		VALUES ($1, $2) 
		RETURNING id, login, created_at, updated_at`,
		req.Login, hashedPassword).Scan(&user.ID, &user.Login, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create user: %w", err)
	}

	// Initialize user balance
	_, err = s.db.Exec(`
		INSERT INTO user_balances (user_id, current_balance, total_withdrawn) 
		VALUES ($1, 0.00, 0.00)`, user.ID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to initialize user balance: %w", err)
	}

	// Generate token
	token, err := utils.GenerateToken(user.ID, user.Login)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	return &user, token, nil
}

// Login authenticates a user
func (s *UserService) Login(req *models.LoginRequest) (*models.User, string, error) {
	var user models.User
	err := s.db.QueryRow(`
		SELECT id, login, password_hash, created_at, updated_at 
		FROM users WHERE login = $1`, req.Login).Scan(
		&user.ID, &user.Login, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, "", errors.New("invalid credentials")
		}
		return nil, "", fmt.Errorf("failed to get user: %w", err)
	}

	// Check password
	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		return nil, "", errors.New("invalid credentials")
	}

	// Generate token
	token, err := utils.GenerateToken(user.ID, user.Login)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	return &user, token, nil
}

// GetBalance retrieves user's balance information
func (s *UserService) GetBalance(userID int) (*models.UserBalance, error) {
	var balance models.UserBalance
	err := s.db.QueryRow(`
		SELECT user_id, current_balance, total_withdrawn 
		FROM user_balances WHERE user_id = $1`, userID).Scan(
		&balance.UserID, &balance.Current, &balance.Withdrawn)
	if err != nil {
		if err == sql.ErrNoRows {
			return &models.UserBalance{UserID: userID, Current: 0, Withdrawn: 0}, nil
		}
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	return &balance, nil
}
