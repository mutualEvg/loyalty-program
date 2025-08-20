package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	appErrors "gofemart/internal/errors"
	"gofemart/internal/middleware"
	"gofemart/internal/services"
)

type OrderHandlers struct {
	orderService *services.OrderService
}

func NewOrderHandlers(orderService *services.OrderService) *OrderHandlers {
	return &OrderHandlers{orderService: orderService}
}

// SubmitOrder handles order submission
func (h *OrderHandlers) SubmitOrder(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Read order number from request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	orderNumber := strings.TrimSpace(string(body))
	if orderNumber == "" {
		http.Error(w, "Order number is required", http.StatusBadRequest)
		return
	}

	err = h.orderService.SubmitOrder(r.Context(), claims.UserID, orderNumber)
	if err != nil {
		switch err {
		case appErrors.ErrInvalidOrderNumberFormat:
			http.Error(w, "Invalid order number format", http.StatusUnprocessableEntity)
			return
		case appErrors.ErrOrderAlreadyUploadedByUser:
			w.WriteHeader(http.StatusOK) // 200 - already uploaded by this user
			return
		case appErrors.ErrOrderAlreadyUploadedByAnotherUser:
			http.Error(w, "Order already uploaded by another user", http.StatusConflict)
			return
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusAccepted) // 202 - accepted for processing
}

// GetUserOrders handles retrieving user's orders
func (h *OrderHandlers) GetUserOrders(w http.ResponseWriter, r *http.Request) {
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

	// Get user orders
	orders, err := h.orderService.GetUserOrders(r.Context(), claims.UserID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return empty array if no orders
	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(orders)
}
