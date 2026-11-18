package audit

import (
	"database/sql"
	"fmt"
)

// Repository handles audit log data operations
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new audit repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Create creates a new audit log entry
func (r *Repository) Create(auditLog *AuditLog) error {
	query := `
		INSERT INTO audit_logs (user_id, action, entity_type, entity_id, details, ip_address, user_agent)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		RETURNING id, created_at
	`
	
	err := r.db.QueryRow(
		query,
		auditLog.UserID,
		auditLog.Action,
		auditLog.EntityType,
		auditLog.EntityID,
		auditLog.Details,
		auditLog.IPAddress,
		auditLog.UserAgent,
	).Scan(&auditLog.ID, &auditLog.CreatedAt)
	
	if err != nil {
		return fmt.Errorf("error creating audit log: %w", err)
	}
	
	return nil
}

// GetByID retrieves an audit log by ID
func (r *Repository) GetByID(id int) (*AuditLog, error) {
	query := `
		SELECT id, user_id, action, entity_type, entity_id, details, ip_address, user_agent, created_at
		FROM audit_logs
		WHERE id = ?
	`
	
	auditLog := &AuditLog{}
	err := r.db.QueryRow(query, id).Scan(
		&auditLog.ID,
		&auditLog.UserID,
		&auditLog.Action,
		&auditLog.EntityType,
		&auditLog.EntityID,
		&auditLog.Details,
		&auditLog.IPAddress,
		&auditLog.UserAgent,
		&auditLog.CreatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("audit log not found")
		}
		return nil, fmt.Errorf("error getting audit log: %w", err)
	}
	
	return auditLog, nil
}

// GetByUserID retrieves all audit logs for a specific user
func (r *Repository) GetByUserID(userID int) ([]*AuditLog, error) {
	query := `
		SELECT id, user_id, action, entity_type, entity_id, details, ip_address, user_agent, created_at
		FROM audit_logs
		WHERE user_id = ?
		ORDER BY created_at DESC
	`
	
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting audit logs: %w", err)
	}
	defer rows.Close()
	
	var auditLogs []*AuditLog
	for rows.Next() {
		auditLog := &AuditLog{}
		
		err := rows.Scan(
			&auditLog.ID,
			&auditLog.UserID,
			&auditLog.Action,
			&auditLog.EntityType,
			&auditLog.EntityID,
			&auditLog.Details,
			&auditLog.IPAddress,
			&auditLog.UserAgent,
			&auditLog.CreatedAt,
		)
		
		if err != nil {
			return nil, fmt.Errorf("error scanning audit log: %w", err)
		}
		
		auditLogs = append(auditLogs, auditLog)
	}
	
	return auditLogs, nil
}

// GetByAction retrieves all audit logs for a specific action
func (r *Repository) GetByAction(action string) ([]*AuditLog, error) {
	query := `
		SELECT id, user_id, action, entity_type, entity_id, details, ip_address, user_agent, created_at
		FROM audit_logs
		WHERE action = ?
		ORDER BY created_at DESC
	`
	
	rows, err := r.db.Query(query, action)
	if err != nil {
		return nil, fmt.Errorf("error getting audit logs: %w", err)
	}
	defer rows.Close()
	
	var auditLogs []*AuditLog
	for rows.Next() {
		auditLog := &AuditLog{}
		
		err := rows.Scan(
			&auditLog.ID,
			&auditLog.UserID,
			&auditLog.Action,
			&auditLog.EntityType,
			&auditLog.EntityID,
			&auditLog.Details,
			&auditLog.IPAddress,
			&auditLog.UserAgent,
			&auditLog.CreatedAt,
		)
		
		if err != nil {
			return nil, fmt.Errorf("error scanning audit log: %w", err)
		}
		
		auditLogs = append(auditLogs, auditLog)
	}
	
	return auditLogs, nil
}

// GetByEntity retrieves all audit logs for a specific entity
func (r *Repository) GetByEntity(entityType string, entityID int) ([]*AuditLog, error) {
	query := `
		SELECT id, user_id, action, entity_type, entity_id, details, ip_address, user_agent, created_at
		FROM audit_logs
		WHERE entity_type = ? AND entity_id = ?
		ORDER BY created_at DESC
	`
	
	rows, err := r.db.Query(query, entityType, entityID)
	if err != nil {
		return nil, fmt.Errorf("error getting audit logs: %w", err)
	}
	defer rows.Close()
	
	var auditLogs []*AuditLog
	for rows.Next() {
		auditLog := &AuditLog{}
		
		err := rows.Scan(
			&auditLog.ID,
			&auditLog.UserID,
			&auditLog.Action,
			&auditLog.EntityType,
			&auditLog.EntityID,
			&auditLog.Details,
			&auditLog.IPAddress,
			&auditLog.UserAgent,
			&auditLog.CreatedAt,
		)
		
		if err != nil {
			return nil, fmt.Errorf("error scanning audit log: %w", err)
		}
		
		auditLogs = append(auditLogs, auditLog)
	}
	
	return auditLogs, nil
}

// GetAll retrieves all audit logs (for admin)
func (r *Repository) GetAll() ([]*AuditLog, error) {
	query := `
		SELECT id, user_id, action, entity_type, entity_id, details, ip_address, user_agent, created_at
		FROM audit_logs
		ORDER BY created_at DESC
		LIMIT 1000
	`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error getting audit logs: %w", err)
	}
	defer rows.Close()
	
	var auditLogs []*AuditLog
	for rows.Next() {
		auditLog := &AuditLog{}
		
		err := rows.Scan(
			&auditLog.ID,
			&auditLog.UserID,
			&auditLog.Action,
			&auditLog.EntityType,
			&auditLog.EntityID,
			&auditLog.Details,
			&auditLog.IPAddress,
			&auditLog.UserAgent,
			&auditLog.CreatedAt,
		)
		
		if err != nil {
			return nil, fmt.Errorf("error scanning audit log: %w", err)
		}
		
		auditLogs = append(auditLogs, auditLog)
	}
	
	return auditLogs, nil
}