package auth

import (
	"encoding/json"
	"laboratory-management/internal/middleware"
	"laboratory-management/internal/user"
	"net/http"
	"strconv"
)

// Handler handles HTTP requests for users
type Handler struct {
	userService *user.Service
	authService *Service
}

// NewHandler creates a new user handler
func NewHandler(userService *user.Service, authService *Service) *Handler {
	return &Handler{
		userService: userService,
		authService: authService,
	}
}

// RegisterHandler handles user registration
func (h *Handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Render registration form
		h.renderRegisterForm(w, r)
		return
	}

	if r.Method == http.MethodPost {
		// Process registration
		var req user.CreateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Hash password
		hashedPassword, err := HashPassword(req.Password)
		if err != nil {
			http.Error(w, "Error processing password", http.StatusInternalServerError)
			return
		}
		req.Password = hashedPassword

		// Create user
		user, err := h.userService.CreateUser(&req, 0) // 0 for system-created
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Return success
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// LoginHandler handles user login
func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Render login form
		h.renderLoginForm(w, r)
		return
	}

	if r.Method == http.MethodPost {
		// Process login
		var req user.LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Authenticate
		user, err := h.authService.Login(req.Email, req.Password)
		if err != nil {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		// Create session
		session, err := h.authService.CreateSession(user)
		if err != nil {
			http.Error(w, "Error creating session", http.StatusInternalServerError)
			return
		}

		// Set session cookie
		middleware.SetSessionCookie(w, session.Token)

		// Return success
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"user":    user,
		})
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// LogoutHandler handles user logout
func (h *Handler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Clear session cookie
	middleware.ClearSessionCookie(w)

	// Redirect to login
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// ProfileHandler handles user profile
func (h *Handler) ProfileHandler(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r)
	if userID == 0 {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Render profile page
	h.renderProfile(w, r, user)
}

// AdminUserEditHandler handles editing a user (admin only)
func (h *Handler) AdminUserEditHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	if r.Method == http.MethodGet {
		user, err := h.userService.GetUserByID(userID)
		if err != nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		h.renderUserEdit(w, r, user)
		return
	}

	if r.Method == http.MethodPost {
		var req user.UpdateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		updaterID := middleware.GetUserIDFromContext(r)
		user, err := h.userService.UpdateUser(userID, &req, updaterID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// renderUserEdit renders the user edit page
func (h *Handler) renderUserEdit(w http.ResponseWriter, r *http.Request, user *user.User) {
	// This will be implemented with HTML templates
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// renderProfile renders the profile page
func (h *Handler) renderProfile(w http.ResponseWriter, r *http.Request, user *user.User) {
	// This will be implemented with HTML templates
	// For now, return a simple JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// renderRegisterForm renders the registration form
func (h *Handler) renderRegisterForm(w http.ResponseWriter, r *http.Request) {
	// This will be implemented with HTML templates
	// For now, return a simple JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Registration form - to be implemented with HTML templates",
	})
}

// renderLoginForm renders the login form
func (h *Handler) renderLoginForm(w http.ResponseWriter, r *http.Request) {
	// This will be implemented with HTML templates
	// For now, return a simple JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Login form - to be implemented with HTML templates",
	})
}
