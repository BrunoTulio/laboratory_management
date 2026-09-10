package user

import (
	_ "database/sql"
	"encoding/json"
	"net/http"
)

type Handler struct {
	userService *Service
}

// AdminUserListHandler handles listing all users (admin only)
func (h *Handler) AdminUserListHandler(w http.ResponseWriter, r *http.Request) {
	users, err := h.userService.GetAllUsers()
	if err != nil {
		http.Error(w, "Error fetching users", http.StatusInternalServerError)
		return
	}

	// Render user list page
	h.renderUserList(w, r, users)
}

// renderUserList renders the user list page
func (h *Handler) renderUserList(w http.ResponseWriter, r *http.Request, users []*User) {
	// This will be implemented with HTML templates
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func NewHandler(userService *Service) *Handler {
	return &Handler{
		userService: userService,
	}
}
