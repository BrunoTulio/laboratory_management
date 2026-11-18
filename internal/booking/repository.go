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
		INSERT INTO bookings (user_id, laboratory_id, class_id, booking_date, start_time, end_time, shift, teacher_name, class_name, subject_name, reason, observation, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id, created_at, updated_at
	`
	
	err := r.db.QueryRow(
		query,
		booking.UserID,
		booking.LaboratoryID,
		booking.ClassID,
		booking.BookingDate.Format("2006-01-02"),
		booking.StartTime,
		booking.EndTime,
		booking.Shift,
		booking.TeacherName,
		booking.ClassName,
		booking.SubjectName,
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
		WHERE b.id = ?
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
		WHERE b.user_id = ?
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
		WHERE b.laboratory_id = ? AND b.booking_date = ? AND b.status IN ('APPROVED', 'PENDING')
		ORDER BY b.start_time
	`
	
	rows, err := r.db.Query(query, laboratoryID, date.Format("2006-01-02"))
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
		WHERE laboratory_id = ?
		  AND booking_date = ?
		  AND status IN ('APPROVED', 'PENDING')
		  AND (? = 0 OR id != ?)
		  AND start_time < ?
		  AND end_time > ?
	`

	var count int
	err := r.db.QueryRow(query, laboratoryID, date.Format("2006-01-02"), excludeID, excludeID, endTime, startTime).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("error checking conflict: %w", err)
	}

	return count > 0, nil
}

// Approve approves a booking
func (r *Repository) Approve(bookingID, approvedBy int) error {
	query := `
		UPDATE bookings 
		SET status = 'APPROVED', approved_by = ?, approved_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
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
		SET status = 'REJECTED', rejected_by = ?, rejected_at = CURRENT_TIMESTAMP, rejection_reason = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
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
		SET status = 'CANCELLED', cancelled_by = ?, cancelled_at = CURRENT_TIMESTAMP, cancellation_reason = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
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

// GetByDate retrieves active (APPROVED/PENDING) bookings for a specific date with class, teacher, and subject metadata.
func (r *Repository) GetByDate(date time.Time) ([]*Booking, error) {
	query := `
		SELECT b.id, b.user_id, b.laboratory_id, b.class_id, b.booking_date, b.start_time, b.end_time,
		       COALESCE(b.shift, ''), COALESCE(b.teacher_name, ''), COALESCE(b.subject_name, ''),
		       COALESCE(NULLIF(b.class_name, ''), COALESCE(c.name, '')), b.reason, COALESCE(b.observation, ''), b.status,
		       b.created_at, b.updated_at
		FROM bookings b
		LEFT JOIN classes c ON b.class_id = c.id
		WHERE b.booking_date = ? AND b.status IN ('APPROVED', 'PENDING')
		ORDER BY b.start_time ASC
	`
	return r.scanSummaryRows(query, date.Format("2006-01-02"))
}

// GetUpcoming retrieves upcoming active bookings starting from fromDate.
func (r *Repository) GetUpcoming(fromDate time.Time, limit int) ([]*Booking, error) {
	query := `
		SELECT b.id, b.user_id, b.laboratory_id, b.class_id, b.booking_date, b.start_time, b.end_time,
		       COALESCE(b.shift, ''), COALESCE(b.teacher_name, ''), COALESCE(b.subject_name, ''),
		       COALESCE(NULLIF(b.class_name, ''), COALESCE(c.name, '')), b.reason, COALESCE(b.observation, ''), b.status,
		       b.created_at, b.updated_at
		FROM bookings b
		LEFT JOIN classes c ON b.class_id = c.id
		WHERE b.booking_date >= ? AND b.status IN ('APPROVED', 'PENDING')
		ORDER BY b.booking_date ASC, b.start_time ASC
		LIMIT ?
	`
	return r.scanSummaryRows(query, fromDate.Format("2006-01-02"), limit)
}

func (r *Repository) scanSummaryRows(query string, args ...interface{}) ([]*Booking, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error querying bookings: %w", err)
	}
	defer rows.Close()

	var bookings []*Booking
	for rows.Next() {
		b := &Booking{}
		if err := rows.Scan(
			&b.ID, &b.UserID, &b.LaboratoryID, &b.ClassID, &b.BookingDate, &b.StartTime, &b.EndTime,
			&b.Shift, &b.TeacherName, &b.SubjectName, &b.ClassName,
			&b.Reason, &b.Observation, &b.Status, &b.CreatedAt, &b.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("error scanning booking summary: %w", err)
		}
		bookings = append(bookings, b)
	}
	return bookings, nil
}