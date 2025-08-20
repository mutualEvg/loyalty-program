package services

import (
	"context"

	appErrors "gofemart/internal/errors"
	"gofemart/internal/models"
	"gofemart/internal/repository"
)

type WithdrawalService struct {
	withdrawalRepo repository.WithdrawalRepository
	balanceRepo    repository.BalanceRepository
}

func NewWithdrawalService(withdrawalRepo repository.WithdrawalRepository, balanceRepo repository.BalanceRepository) *WithdrawalService {
	return &WithdrawalService{
		withdrawalRepo: withdrawalRepo,
		balanceRepo:    balanceRepo,
	}
}

// WithdrawPoints processes a withdrawal request
func (s *WithdrawalService) WithdrawPoints(ctx context.Context, userID int, req *models.WithdrawalRequest) error {
	// Try to withdraw from balance (includes transaction logic)
	err := s.balanceRepo.Withdraw(ctx, userID, req.Sum)
	if err != nil {
		if err.Error() == "insufficient funds" {
			return appErrors.ErrInsufficientFunds
		}
		return err
	}

	// Record withdrawal
	return s.withdrawalRepo.Create(ctx, userID, req.OrderNumber, req.Sum)
}

// GetUserWithdrawals retrieves all withdrawals for a user
func (s *WithdrawalService) GetUserWithdrawals(ctx context.Context, userID int) ([]*models.WithdrawalResponse, error) {
	return s.withdrawalRepo.GetByUserID(ctx, userID)
}
