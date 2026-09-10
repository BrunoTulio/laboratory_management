package audit

import "time"

// AuditLog represents an audit log entry
type AuditLog struct {
	ID         int       `json:"id"`
	UserID     *int      `json:"user_id,omitempty"`
	Action     string    `json:"action"`
	EntityType string    `json:"entity_type"`
	EntityID   *int      `json:"entity_id,omitempty"`
	Details    string    `json:"details"`
	IPAddress  string    `json:"ip_address"`
	UserAgent  string    `json:"user_agent"`
	CreatedAt  time.Time `json:"created_at"`
}

// AuditAction constants
const (
	ActionLogin           = "LOGIN"
	ActionLogout          = "LOGOUT"
	ActionCreateUser      = "CREATE_USER"
	ActionUpdateUser      = "UPDATE_USER"
	ActionDeleteUser      = "DELETE_USER"
	ActionCreateBooking   = "CREATE_BOOKING"
	ActionApproveBooking  = "APPROVE_BOOKING"
	ActionRejectBooking   = "REJECT_BOOKING"
	ActionCancelBooking   = "CANCEL_BOOKING"
	ActionChangeRole      = "CHANGE_ROLE"
	ActionViewAudit       = "VIEW_AUDIT"
)

// EntityType constants
const (
	EntityUser       = "USER"
	EntityBooking    = "BOOKING"
	EntityLaboratory = "LABORATORY"
	EntityClass      = "CLASS"
	EntityNotification = "NOTIFICATION"
)

// CreateAuditLogRequest represents the request to create an audit log
type CreateAuditLogRequest struct {
	UserID     *int   `json:"user_id,omitempty"`
	Action     string `json:"action"`
	EntityType string `json:"entity_type"`
	EntityID   *int   `json:"entity_id,omitempty"`
	Details    string `json:"details"`
	IPAddress  string `json:"ip_address"`
	UserAgent  string `json:"user_agent"`
}