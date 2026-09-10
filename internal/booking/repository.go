package booking

import (
	"database/sql"
	"fmt"
	"time"
)

// Repository handles booking data operations
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new booking repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Create creates a new booking
func (r *Repository) Create(booking *Booking) error {
	query := `
		INSERT INTO bookings (user_id, laboratory_id, class_id, booking_date, start_time, end_time, shift, reason, observation, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at
	`
	
	err := r.db.QueryRow(
		query,
		booking.UserID,
		booking.LaboratoryID,
		booking.ClassID,
		booking.BookingDate,
		booking.StartTime,
		booking.EndTime,
		booking.Shift,
		booking.Reason,
		booking.Observation,
		booking.Status,
	).Scan(&booking.ID, &booking.CreatedAt, &booking.UpdatedAt)
	
	if err != nil {
		return fmt.Errorf("error creating booking: %w", err)
	}
	
	return nil
}

// GetByID retrieves a booking by ID
func (r *Repository) GetByID(id int) (*Booking, error) {
	query := `
		SELECT b.id, b.user_id, b.laboratory_id, b.class_id, b.booking_date, b.start_time, b.end_time,
		       b.shift, b.reason, b.observation, b.status, b.approved_by, b.approved_at,
		       b.rejected_by, b.rejected_at, b.rejection_reason, b.cancelled_by, b.cancelled_at,
		       b.cancellation_reason, b.created_at, b.updated_at
		FROM bookings b
		WHERE b.id = $1
	`
	
	booking := &Booking{}
	err := r.db.QueryRow(query, id).Scan(
		&booking.ID,
		&booking.UserID,
		&booking.LaboratoryID,
		&booking.ClassID,
		&booking.BookingDate,
		&booking.StartTime,
		&booking.EndTime,
		&booking.Shift,
		&booking.Reason,
		&booking.Observation,
		&booking.Status,
		&booking.ApprovedBy,
		&booking.ApprovedAt,
		&booking.RejectedBy,
		&booking.RejectedAt,
		&booking.RejectionReason,
		&booking.CancelledBy,
		&booking.CancelledAt,
		&booking.CancellationReason,
		&booking.CreatedAt,
		&booking.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("booking not found")
		}
		return nil, fmt.Errorf("error getting booking: %w", err)
	}
	
	return booking, nil
}

// GetByUserID retrieves all bookings for a specific user
func (r *Repository) GetByUserID(userID int) ([]*Booking, error) {
	query := `
		SELECT b.id, b.user_id, b.laboratory_id, b.class_id, b.booking_date, b.start_time, b.end_time,
		       b.shift, b.reason, b.observation, b.status, b.approved_by, b.approved_at,
		       b.rejected_by, b.rejected_at, b.rejection_reason, b.cancelled_by, b.cancelled_at,
		       b.cancellation_reason, b.created_at, b.updated_at
		FROM bookings b
		WHERE b.user_id = $1
		ORDER BY b.booking_date DESC, b.start_time DESC
	`
	
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting bookings: %w", err)
	}
	defer rows.Close()
	
	var bookings []*Booking
	for rows.Next() {
		booking := &Booking{}
		
		err := rows.Scan(
			&booking.ID,
			&booking.UserID,
			&booking.LaboratoryID,
			&booking.ClassID,
			&booking.BookingDate,
			&booking.StartTime,
			&booking.EndTime,
			&booking.Shift,
			&booking.Reason,
			&booking.Observation,
			&booking.Status,
			&booking.ApprovedBy,
			&booking.ApprovedAt,
			&booking.RejectedBy,
			&booking.RejectedAt,
			&booking.RejectionReason,
			&booking.CancelledBy,
			&booking.CancelledAt,
			&booking.CancellationReason,
			&booking.CreatedAt,
			&booking.UpdatedAt,
		)
		
		if err != nil {
			return nil, fmt.Errorf("error scanning booking: %w", err)
		}
		
		bookings = append(bookings, booking)
	}
	
	return bookings, nil
}

// GetByLaboratoryIDAndDate retrieves bookings for a specific laboratory and date
func (r *Repository) GetByLaboratoryIDAndDate(laboratoryID int, date time.Time) ([]*Booking, error) {
	query := `
		SELECT b.id, b.user_id, b.laboratory_id, b.class_id, b.booking_date, b.start_time, b.end_time,
		       b.shift, b.reason, b.observation, b.status, b.approved_by, b.approved_at,
		       b.rejected_by, b.rejected_at, b.rejection_reason, b.cancelled_by, b.cancelled_at,
		       b.cancellation_reason, b.created_at, b.updated_at
		FROM bookings b
		WHERE b.laboratory_id = $1 AND b.booking_date = $2 AND b.status IN ('APPROVED', 'PENDING')
		ORDER BY b.start_time
	`
	
	rows, err := r.db.Query(query, laboratoryID, date)
	if err != nil {
		return nil, fmt.Errorf("error getting bookings: %w", err)
	}
	defer rows.Close()
	
	var bookings []*Booking
	for rows.Next() {
		booking := &Booking{}
		
		err := rows.Scan(
			&booking.ID,
			&booking.UserID,
			&booking.LaboratoryID,
			&booking.ClassID,
			&booking.BookingDate,
			&booking.StartTime,
			&booking.EndTime,
			&booking.Shift,
			&booking.Reason,
			&booking.Observation,
			&booking.Status,
			&booking.ApprovedBy,
			&booking.ApprovedAt,
			&booking.RejectedBy,
			&booking.RejectedAt,
			&booking.RejectionReason,
			&booking.CancelledBy,
			&booking.CancelledAt,
			&booking.CancellationReason,
			&booking.CreatedAt,
			&booking.UpdatedAt,
		)
		
		if err != nil {
			return nil, fmt.Errorf("error scanning booking: %w", err)
		}
		
		bookings = append(bookings, booking)
	}
	
	return bookings, nil
}

// GetPending retrieves all pending bookings
func (r *Repository) GetPending() ([]*Booking, error) {
	query := `
		SELECT b.id, b.user_id, b.laboratory_id, b.class_id, b.booking_date, b.start_time, b.end_time,
		       b.shift, b.reason, b.observation, b.status, b.approved_by, b.approved_at,
		       b.rejected_by, b.rejected_at, b.rejection_reason, b.cancelled_by, b.cancelled_at,
		       b.cancellation_reason, b.created_at, b.updated_at
		FROM bookings b
		WHERE b.status = 'PENDING'
		ORDER BY b.booking_date ASC, b.start_time ASC
	`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error getting pending bookings: %w", err)
	}
	defer rows.Close()
	
	var bookings []*Booking
	for rows.Next() {
		booking := &Booking{}
		
		err := rows.Scan(
			&booking.ID,
			&booking.UserID,
			&booking.LaboratoryID,
			&booking.ClassID,
			&booking.BookingDate,
			&booking.StartTime,
			&booking.EndTime,
			&booking.Shift,
			&booking.Reason,
			&booking.Observation,
			&booking.Status,
			&booking.ApprovedBy,
			&booking.ApprovedAt,
			&booking.RejectedBy,
			&booking.RejectedAt,
			&booking.RejectionReason,
			&booking.CancelledBy,
			&booking.CancelledAt,
			&booking.CancellationReason,
			&booking.CreatedAt,
			&booking.UpdatedAt,
		)
		
		if err != nil {
			return nil, fmt.Errorf("error scanning booking: %w", err)
		}
		
		bookings = append(bookings, booking)
	}
	
	return bookings, nil
}

// CheckConflict checks if there's a time conflict for a booking
func (r *Repository) CheckConflict(laboratoryID int, date time.Time, startTime, endTime time.Time, excludeID int) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM bookings
		WHERE laboratory_id = $1 
		  AND booking_date = $2 
		  AND status IN ('APPROVED', 'PENDING')
		  AND id != $3
		  AND (
			  (start_time < $4 AND end_time > $5) OR
			  (start_time >= $4 AND start_time < $5) OR
			  (end_time > $4 AND end_time <= $5)
		  )
	`
	
	var count int
	var excludeIDParam interface{}
	if excludeID > 0 {
		excludeIDParam = excludeID
	} else {
		excludeIDParam = nil
	}
	
	err := r.db.QueryRow(query, laboratoryID, date, excludeIDParam, endTime, startTime).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("error checking conflict: %w", err)
	}
	
	return count > 0, nil
}

// Approve approves a booking
func (r *Repository) Approve(bookingID, approvedBy int) error {
	query := `
		UPDATE bookings 
		SET status = 'APPROVED', approved_by = $1, approved_at = NOW(), updated_at = NOW()
		WHERE id = $2
	`
	
	result, err := r.db.Exec(query, approvedBy, bookingID)
	if err != nil {
		return fmt.Errorf("error approving booking: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("booking not found")
	}
	
	return nil
}

// Reject rejects a booking
func (r *Repository) Reject(bookingID, rejectedBy int, rejectionReason string) error {
	query := `
		UPDATE bookings 
		SET status = 'REJECTED', rejected_by = $1, rejected_at = NOW(), rejection_reason = $2, updated_at = NOW()
		WHERE id = $3
	`
	
	result, err := r.db.Exec(query, rejectedBy, rejectionReason, bookingID)
	if err != nil {
		return fmt.Errorf("error rejecting booking: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("booking not found")
	}
	
	return nil
}

// Cancel cancels a booking
func (r *Repository) Cancel(bookingID, cancelledBy int, cancellationReason string) error {
	query := `
		UPDATE bookings 
		SET status = 'CANCELLED', cancelled_by = $1, cancelled_at = NOW(), cancellation_reason = $2, updated_at = NOW()
		WHERE id = $3
	`
	
	result, err := r.db.Exec(query, cancelledBy, cancellationReason, bookingID)
	if err != nil {
		return fmt.Errorf("error cancelling booking: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("booking not found")
	}
	
	return nil
}

// GetAll retrieves all bookings (for admin)
func (r *Repository) GetAll() ([]*Booking, error) {
	query := `
		SELECT b.id, b.user_id, b.laboratory_id, b.class_id, b.booking_date, b.start_time, b.end_time,
		       b.shift, b.reason, b.observation, b.status, b.approved_by, b.approved_at,
		       b.rejected_by, b.rejected_at, b.rejection_reason, b.cancelled_by, b.cancelled_at,
		       b.cancellation_reason, b.created_at, b.updated_at
		FROM bookings b
		ORDER BY b.booking_date DESC, b.start_time DESC
	`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error getting bookings: %w", err)
	}
	defer rows.Close()
	
	var bookings []*Booking
	for rows.Next() {
		booking := &Booking{}
		
		err := rows.Scan(
			&booking.ID,
			&booking.UserID,
			&booking.LaboratoryID,
			&booking.ClassID,
			&booking.BookingDate,
			&booking.StartTime,
			&booking.EndTime,
			&booking.Shift,
			&booking.Reason,
			&booking.Observation,
			&booking.Status,
			&booking.ApprovedBy,
			&booking.ApprovedAt,
			&booking.RejectedBy,
			&booking.RejectedAt,
			&booking.RejectionReason,
			&booking.CancelledBy,
			&booking.CancelledAt,
			&booking.CancellationReason,
			&booking.CreatedAt,
			&booking.UpdatedAt,
		)
		
		if err != nil {
			return nil, fmt.Errorf("error scanning booking: %w", err)
		}
		
		bookings = append(bookings, booking)
	}
	
	return bookings, nil
}