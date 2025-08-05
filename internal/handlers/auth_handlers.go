package handlers

import (
	"encoding/json"
	"net/http"

	appErrors "gofemart/internal/errors"
	"gofemart/internal/models"
	"gofemart/internal/services"
)

type AuthHandlers struct {
	userService *services.UserService
}

func NewAuthHandlers(userService *services.UserService) *AuthHandlers {
	return &AuthHandlers{userService: userService}
}

// Register handles user registration
func (h *AuthHandlers) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Basic validation
	if req.Login == "" || req.Password == "" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	user, token, err := h.userService.Register(&req)
	if err != nil {
		if err == appErrors.ErrLoginAlreadyExists {
			http.Error(w, "Login already taken", http.StatusConflict)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Set authentication cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400, // 24 hours
	})

	// Set Authorization header
	w.Header().Set("Authorization", "Bearer "+token)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

// Login handles user authentication
func (h *AuthHandlers) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Basic validation
	if req.Login == "" || req.Password == "" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	user, token, err := h.userService.Login(&req)
	if err != nil {
		if err == appErrors.ErrInvalidCredentials {
			http.Error(w, "Invalid login/password", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Set authentication cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400, // 24 hours
	})

	// Set Authorization header
	w.Header().Set("Authorization", "Bearer "+token)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}
