package services

import (
	"context"

	appErrors "gofemart/internal/errors"
	"gofemart/internal/models"
	"gofemart/internal/repository"
	"gofemart/internal/utils"
)

type UserService struct {
	userRepo    repository.UserRepository
	balanceRepo repository.BalanceRepository
}

func NewUserService(userRepo repository.UserRepository, balanceRepo repository.BalanceRepository) *UserService {
	return &UserService{
		userRepo:    userRepo,
		balanceRepo: balanceRepo,
	}
}

// Register creates a new user account
func (s *UserService) Register(ctx context.Context, req *models.RegisterRequest) (*models.User, string, error) {
	// Check if user already exists
	exists, err := s.userRepo.ExistsByLogin(ctx, req.Login)
	if err != nil {
		return nil, "", err
	}
	if exists {
		return nil, "", appErrors.ErrLoginAlreadyExists
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, "", err
	}

	// Create user
	user := &models.User{Login: req.Login}
	if err := s.userRepo.Create(ctx, user, hashedPassword); err != nil {
		return nil, "", err
	}

	// Initialize user balance
	if err := s.balanceRepo.Create(ctx, user.ID); err != nil {
		return nil, "", err
	}

	// Generate token
	token, err := utils.GenerateToken(user.ID, user.Login)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

// Login authenticates a user
func (s *UserService) Login(ctx context.Context, req *models.LoginRequest) (*models.User, string, error) {
	user, err := s.userRepo.GetByLogin(ctx, req.Login)
	if err != nil {
		return nil, "", err
	}
	if user == nil {
		return nil, "", appErrors.ErrInvalidCredentials
	}

	// Check password
	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		return nil, "", appErrors.ErrInvalidCredentials
	}

	// Generate token
	token, err := utils.GenerateToken(user.ID, user.Login)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

// GetBalance retrieves user's balance information
func (s *UserService) GetBalance(ctx context.Context, userID int) (*models.UserBalance, error) {
	return s.balanceRepo.GetByUserID(ctx, userID)
}
