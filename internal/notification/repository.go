package notification

import (
	"database/sql"
	"fmt"
)

// Repository handles notification data operations
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new notification repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Create creates a new notification
func (r *Repository) Create(notification *Notification) error {
	query := `
		INSERT INTO notifications (user_id, booking_id, type, title, message, is_read)
		VALUES (?, ?, ?, ?, ?, ?)
		RETURNING id, created_at
	`
	
	err := r.db.QueryRow(
		query,
		notification.UserID,
		notification.BookingID,
		notification.Type,
		notification.Title,
		notification.Message,
		notification.IsRead,
	).Scan(&notification.ID, &notification.CreatedAt)
	
	if err != nil {
		return fmt.Errorf("error creating notification: %w", err)
	}
	
	return nil
}

// GetByID retrieves a notification by ID
func (r *Repository) GetByID(id int) (*Notification, error) {
	query := `
		SELECT id, user_id, booking_id, type, title, message, is_read, created_at
		FROM notifications
		WHERE id = ?
	`
	
	notification := &Notification{}
	err := r.db.QueryRow(query, id).Scan(
		&notification.ID,
		&notification.UserID,
		&notification.BookingID,
		&notification.Type,
		&notification.Title,
		&notification.Message,
		&notification.IsRead,
		&notification.CreatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("notification not found")
		}
		return nil, fmt.Errorf("error getting notification: %w", err)
	}
	
	return notification, nil
}

// GetByUserID retrieves all notifications for a specific user
func (r *Repository) GetByUserID(userID int) ([]*Notification, error) {
	query := `
		SELECT id, user_id, booking_id, type, title, message, is_read, created_at
		FROM notifications
		WHERE user_id = ?
		ORDER BY created_at DESC
	`
	
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting notifications: %w", err)
	}
	defer rows.Close()
	
	var notifications []*Notification
	for rows.Next() {
		notification := &Notification{}
		
		err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.BookingID,
			&notification.Type,
			&notification.Title,
			&notification.Message,
			&notification.IsRead,
			&notification.CreatedAt,
		)
		
		if err != nil {
			return nil, fmt.Errorf("error scanning notification: %w", err)
		}
		
		notifications = append(notifications, notification)
	}
	
	return notifications, nil
}

// GetUnreadCount retrieves the count of unread notifications for a user
func (r *Repository) GetUnreadCount(userID int) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM notifications
		WHERE user_id = ? AND is_read = FALSE
	`
	
	var count int
	err := r.db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error getting unread count: %w", err)
	}
	
	return count, nil
}

// MarkAsRead marks a notification as read
func (r *Repository) MarkAsRead(notificationID int) error {
	query := `UPDATE notifications SET is_read = TRUE WHERE id = ?`
	
	result, err := r.db.Exec(query, notificationID)
	if err != nil {
		return fmt.Errorf("error marking notification as read: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("notification not found")
	}
	
	return nil
}

// MarkAllAsRead marks all notifications for a user as read
func (r *Repository) MarkAllAsRead(userID int) error {
	query := `UPDATE notifications SET is_read = TRUE WHERE user_id = ?`
	
	_, err := r.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("error marking all notifications as read: %w", err)
	}
	
	return nil
}

// Delete deletes a notification
func (r *Repository) Delete(id int) error {
	query := `DELETE FROM notifications WHERE id = ?`
	
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error deleting notification: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("notification not found")
	}
	
	return nil
}