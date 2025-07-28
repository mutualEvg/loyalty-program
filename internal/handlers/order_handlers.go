package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"gofemart/internal/middleware"
	"gofemart/internal/services"
)

type OrderHandlers struct {
	orderService *services.OrderService
}

func NewOrderHandlers(orderService *services.OrderService) *OrderHandlers {
	return &OrderHandlers{orderService: orderService}
}

// SubmitOrder handles order number submission
func (h *OrderHandlers) SubmitOrder(w http.ResponseWriter, r *http.Request) {
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

	// Check content type
	if r.Header.Get("Content-Type") != "text/plain" {
		http.Error(w, "Bad request", http.StatusBadRequest)
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
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Submit order
	err = h.orderService.SubmitOrder(claims.UserID, orderNumber)
	if err != nil {
		switch err.Error() {
		case "invalid order number format":
			http.Error(w, "Invalid order number format", http.StatusUnprocessableEntity)
			return
		case "order already uploaded by this user":
			w.WriteHeader(http.StatusOK) // 200 - already uploaded by this user
			return
		case "order already uploaded by another user":
			http.Error(w, "Order already uploaded by another user", http.StatusConflict)
			return
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	// New order accepted
	w.WriteHeader(http.StatusAccepted)
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
	orders, err := h.orderService.GetUserOrders(claims.UserID)
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
