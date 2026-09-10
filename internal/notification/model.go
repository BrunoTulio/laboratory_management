package notification

import "time"

// Notification represents a system notification
type Notification struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	BookingID *int      `json:"booking_id,omitempty"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}

// NotificationType constants
const (
	TypeBookingCreated    = "BOOKING_CREATED"
	TypeBookingApproved   = "BOOKING_APPROVED"
	TypeBookingRejected   = "BOOKING_REJECTED"
	TypeBookingCancelled  = "BOOKING_CANCELLED"
	TypeBookingReminder   = "BOOKING_REMINDER"
)

// CreateNotificationRequest represents the request to create a notification
type CreateNotificationRequest struct {
	UserID    int    `json:"user_id"`
	BookingID *int   `json:"booking_id,omitempty"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	Message   string `json:"message"`
}

// MarkAsReadRequest represents the request to mark notifications as read
type MarkAsReadRequest struct {
	NotificationID int `json:"notification_id"`
}