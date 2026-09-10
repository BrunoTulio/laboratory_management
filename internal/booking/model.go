package booking

import (
	"time"
	"laboratory-management/internal/class"
	"laboratory-management/internal/laboratory"
	"laboratory-management/internal/user"
)

// Booking represents a laboratory booking
type Booking struct {
	ID               int        `json:"id"`
	UserID           int        `json:"user_id"`
	User             *user.User `json:"user,omitempty"`
	LaboratoryID     int        `json:"laboratory_id"`
	Laboratory       *laboratory.Laboratory `json:"laboratory,omitempty"`
	ClassID          *int       `json:"class_id,omitempty"`
	Class            *class.Class `json:"class,omitempty"`
	BookingDate      time.Time  `json:"booking_date"`
	StartTime        time.Time  `json:"start_time"`
	EndTime          time.Time  `json:"end_time"`
	Shift            string     `json:"shift"`
	Reason           string     `json:"reason"`
	Observation      string     `json:"observation"`
	Status           string     `json:"status"`
	ApprovedBy       *int       `json:"approved_by,omitempty"`
	ApprovedAt       *time.Time `json:"approved_at,omitempty"`
	RejectedBy       *int       `json:"rejected_by,omitempty"`
	RejectedAt       *time.Time `json:"rejected_at,omitempty"`
	RejectionReason  string     `json:"rejection_reason,omitempty"`
	CancelledBy      *int       `json:"cancelled_by,omitempty"`
	CancelledAt      *time.Time `json:"cancelled_at,omitempty"`
	CancellationReason string   `json:"cancellation_reason,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// BookingStatus constants
const (
	StatusPending   = "PENDING"
	StatusApproved  = "APPROVED"
	StatusRejected  = "REJECTED"
	StatusCancelled = "CANCELLED"
	StatusCompleted = "CONCLUIDO"
)

// CreateBookingRequest represents the request to create a booking
type CreateBookingRequest struct {
	LaboratoryID int       `json:"laboratory_id"`
	ClassID      *int      `json:"class_id,omitempty"`
	BookingDate  time.Time `json:"booking_date"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	Shift        string    `json:"shift"`
	Reason       string    `json:"reason"`
	Observation  string    `json:"observation"`
}

// ApproveBookingRequest represents the request to approve a booking
type ApproveBookingRequest struct {
	BookingID int `json:"booking_id"`
}

// RejectBookingRequest represents the request to reject a booking
type RejectBookingRequest struct {
	BookingID      int    `json:"booking_id"`
	RejectionReason string `json:"rejection_reason"`
}

// CancelBookingRequest represents the request to cancel a booking
type CancelBookingRequest struct {
	BookingID           int    `json:"booking_id"`
	CancellationReason  string `json:"cancellation_reason"`
}

// TimeRange represents a time range for conflict detection
type TimeRange struct {
	Start time.Time
	End   time.Time
}

// HasConflict checks if two time ranges overlap
func (tr *TimeRange) HasConflict(other TimeRange) bool {
	// No conflict if one ends before the other starts
	if tr.End.Before(other.Start) || other.End.Before(tr.Start) {
		return false
	}
	// No conflict if one starts exactly when the other ends
	if tr.End.Equal(other.Start) || other.End.Equal(tr.Start) {
		return false
	}
	// Otherwise, there's a conflict
	return true
}