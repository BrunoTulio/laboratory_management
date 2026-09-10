package notification

import (
	"encoding/json"
	"net/http"
	"strconv"

	"laboratory-management/internal/middleware"
)

// Handler handles HTTP requests for notifications
type Handler struct {
	notificationService *Service
}

// NewHandler creates a new notification handler
func NewHandler(notificationService *Service) *Handler {
	return &Handler{
		notificationService: notificationService,
	}
}

// ListHandler handles listing user's notifications
func (h *Handler) ListHandler(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r)
	if userID == 0 {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	notifications, err := h.notificationService.GetNotificationsByUserID(userID)
	if err != nil {
		http.Error(w, "Error fetching notifications", http.StatusInternalServerError)
		return
	}

	// Render notification list
	h.renderNotificationList(w, r, notifications)
}

// UnreadCountHandler handles getting unread notification count
func (h *Handler) UnreadCountHandler(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	count, err := h.notificationService.GetUnreadCount(userID)
	if err != nil {
		http.Error(w, "Error fetching unread count", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"unread_count": count,
	})
}

// MarkAsReadHandler handles marking a notification as read
func (h *Handler) MarkAsReadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	notificationIDStr := r.URL.Query().Get("id")
	notificationID, err := strconv.Atoi(notificationIDStr)
	if err != nil {
		http.Error(w, "Invalid notification ID", http.StatusBadRequest)
		return
	}

	err = h.notificationService.MarkAsRead(notificationID)
	if err != nil {
		http.Error(w, "Error marking notification as read", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

// MarkAllAsReadHandler handles marking all notifications as read
func (h *Handler) MarkAllAsReadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetUserIDFromContext(r)
	err := h.notificationService.MarkAllAsRead(userID)
	if err != nil {
		http.Error(w, "Error marking all notifications as read", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

// renderNotificationList renders the notification list page
func (h *Handler) renderNotificationList(w http.ResponseWriter, r *http.Request, notifications []*Notification) {
	// This will be implemented with HTML templates
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notifications)
}