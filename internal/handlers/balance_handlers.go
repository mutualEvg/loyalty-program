package handlers

import (
	"encoding/json"
	"net/http"

	appErrors "gofemart/internal/errors"
	"gofemart/internal/middleware"
	"gofemart/internal/models"
	"gofemart/internal/services"
)

type BalanceHandlers struct {
	userService       *services.UserService
	withdrawalService *services.WithdrawalService
}

func NewBalanceHandlers(userService *services.UserService, withdrawalService *services.WithdrawalService) *BalanceHandlers {
	return &BalanceHandlers{
		userService:       userService,
		withdrawalService: withdrawalService,
	}
}

// GetBalance handles user balance retrieval
func (h *BalanceHandlers) GetBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from context
	claims, ok := middleware.GetUserFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user balance
	balance, err := h.userService.GetBalance(r.Context(), claims.UserID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(balance)
}

// WithdrawPoints handles points withdrawal request
func (h *BalanceHandlers) WithdrawPoints(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from context
	claims, ok := middleware.GetUserFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.WithdrawalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Basic validation
	if req.OrderNumber == "" || req.Sum <= 0 {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Process withdrawal
	err := h.withdrawalService.WithdrawPoints(r.Context(), claims.UserID, &req)
	if err != nil {
		if err == appErrors.ErrInsufficientFunds {
			http.Error(w, "Insufficient funds", http.StatusPaymentRequired)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetWithdrawals handles retrieval of user withdrawals
func (h *BalanceHandlers) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from context
	claims, ok := middleware.GetUserFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user withdrawals
	withdrawals, err := h.withdrawalService.GetUserWithdrawals(r.Context(), claims.UserID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return empty array if no withdrawals
	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(withdrawals)
}
