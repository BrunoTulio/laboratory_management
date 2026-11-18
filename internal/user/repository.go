package user

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// Repository handles user data operations
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new user repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Create creates a new user
func (r *Repository) Create(user *User) error {
	query := `
		INSERT INTO users (name, email, password_hash, role_id, registration_number, phone, status)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		RETURNING id, created_at, updated_at
	`
	
	err := r.db.QueryRow(
		query,
		user.Name,
		user.Email,
		user.PasswordHash,
		user.RoleID,
		user.RegistrationNumber,
		user.Phone,
		user.Status,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	
	if err != nil {
		return fmt.Errorf("error creating user: %w", err)
	}
	
	return nil
}

// GetByID retrieves a user by ID
func (r *Repository) GetByID(id int) (*User, error) {
	query := `
		SELECT u.id, u.name, u.email, u.password_hash, u.role_id, u.registration_number, 
		       u.phone, u.status, COALESCE(u.classes_csv, ''), COALESCE(u.subjects_csv, ''), u.created_at, u.updated_at,
		       r.id, r.name, r.description, r.created_at, r.updated_at
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		WHERE u.id = ?
	`
	
	user := &User{}
	var role Role
	
	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.RoleID,
		&user.RegistrationNumber, &user.Phone, &user.Status, &user.ClassesCSV, &user.SubjectsCSV, &user.CreatedAt, &user.UpdatedAt,
		&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("error getting user: %w", err)
	}
	
	if role.ID > 0 {
		user.Role = &role
	}
	
	return user, nil
}

// GetByEmail retrieves a user by email
func (r *Repository) GetByEmail(email string) (*User, error) {
	query := `
		SELECT u.id, u.name, u.email, u.password_hash, u.role_id, u.registration_number, 
		       u.phone, u.status, COALESCE(u.classes_csv, ''), COALESCE(u.subjects_csv, ''), u.created_at, u.updated_at,
		       r.id, r.name, r.description, r.created_at, r.updated_at
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		WHERE u.email = ?
	`
	
	user := &User{}
	var role Role
	
	err := r.db.QueryRow(query, email).Scan(
		&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.RoleID,
		&user.RegistrationNumber, &user.Phone, &user.Status, &user.ClassesCSV, &user.SubjectsCSV, &user.CreatedAt, &user.UpdatedAt,
		&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("error getting user: %w", err)
	}
	
	if role.ID > 0 {
		user.Role = &role
	}
	
	return user, nil
}

// GetAll retrieves all users
func (r *Repository) GetAll() ([]*User, error) {
	query := `
		SELECT u.id, u.name, u.email, u.password_hash, u.role_id, u.registration_number, 
		       u.phone, u.status, COALESCE(u.classes_csv, ''), COALESCE(u.subjects_csv, ''), u.created_at, u.updated_at,
		       r.id, r.name, r.description, r.created_at, r.updated_at
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		ORDER BY u.created_at DESC
	`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error getting users: %w", err)
	}
	defer rows.Close()
	
	var users []*User
	for rows.Next() {
		user := &User{}
		var role Role
		
		err := rows.Scan(
			&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.RoleID,
			&user.RegistrationNumber, &user.Phone, &user.Status, &user.ClassesCSV, &user.SubjectsCSV, &user.CreatedAt, &user.UpdatedAt,
			&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt,
		)
		
		if err != nil {
			return nil, fmt.Errorf("error scanning user: %w", err)
		}
		
		if role.ID > 0 {
			user.Role = &role
		}
		
		users = append(users, user)
	}
	
	return users, nil
}

// Update updates a user
func (r *Repository) Update(user *User) error {
	query := `
		UPDATE users 
		SET name = ?, email = ?, phone = ?, status = ?, updated_at = ?
		WHERE id = ?
	`
	
	user.UpdatedAt = time.Now()
	
	result, err := r.db.Exec(
		query,
		user.Name,
		user.Email,
		user.Phone,
		user.Status,
		user.UpdatedAt,
		user.ID,
	)
	
	if err != nil {
		return fmt.Errorf("error updating user: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	
	return nil
}

// SetStatus updates a user's status (ACTIVE, INACTIVE, etc.).
func (r *Repository) SetStatus(id int, status string) error {
	result, err := r.db.Exec(
		`UPDATE users SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		status, id,
	)
	if err != nil {
		return fmt.Errorf("error updating user status: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

// Delete deletes a user and removes their associated bookings and notifications inside a transaction.
func (r *Repository) Delete(id int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("error starting delete transaction: %w", err)
	}
	defer tx.Rollback()

	cleanupQueries := []string{
		`DELETE FROM notifications WHERE user_id = ?`,
		`DELETE FROM notifications WHERE booking_id IN (SELECT id FROM bookings WHERE user_id = ?)`,
		`UPDATE bookings SET approved_by = NULL WHERE approved_by = ?`,
		`UPDATE bookings SET rejected_by = NULL WHERE rejected_by = ?`,
		`UPDATE bookings SET cancelled_by = NULL WHERE cancelled_by = ?`,
		`DELETE FROM bookings WHERE user_id = ?`,
		`UPDATE audit_logs SET user_id = NULL WHERE user_id = ?`,
	}
	for _, q := range cleanupQueries {
		if _, err := tx.Exec(q, id); err != nil {
			return fmt.Errorf("error cleaning user dependencies: %w", err)
		}
	}

	result, err := tx.Exec(`DELETE FROM users WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("error deleting user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return tx.Commit()
}

// GetRoleByName retrieves a role by name
func (r *Repository) GetRoleByName(name string) (*Role, error) {
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM roles
		WHERE name = ?
	`
	
	role := &Role{}
	err := r.db.QueryRow(query, name).Scan(
		&role.ID,
		&role.Name,
		&role.Description,
		&role.CreatedAt,
		&role.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("role not found")
		}
		return nil, fmt.Errorf("error getting role: %w", err)
	}
	
	return role, nil
}

// GetRoleByID retrieves a role by ID
func (r *Repository) GetRoleByID(id int) (*Role, error) {
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM roles
		WHERE id = ?
	`
	
	role := &Role{}
	err := r.db.QueryRow(query, id).Scan(
		&role.ID,
		&role.Name,
		&role.Description,
		&role.CreatedAt,
		&role.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("role not found")
		}
		return nil, fmt.Errorf("error getting role: %w", err)
	}
	
	return role, nil
}

// CreateInvited creates a pending user with an activation token.
func (r *Repository) CreateInvited(u *User, token string, expiresAt time.Time) error {
	query := `
		INSERT INTO users (name, email, password_hash, role_id, registration_number, phone, status, activation_token, activation_expires_at)
		VALUES (?, ?, '', ?, ?, ?, ?, ?, ?)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(
		query,
		u.Name,
		u.Email,
		u.RoleID,
		u.RegistrationNumber,
		u.Phone,
		StatusPending,
		token,
		expiresAt.UTC().Format(time.RFC3339),
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
}

// GetByActivationToken retrieves a pending user by activation token.
func (r *Repository) GetByActivationToken(token string) (*User, time.Time, error) {
	query := `
		SELECT u.id, u.name, u.email, u.role_id, u.status, COALESCE(CAST(u.activation_expires_at AS TEXT), ''),
		       COALESCE(r.id, 0), COALESCE(r.name, '')
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		WHERE u.activation_token = ? AND u.activation_token != ''
	`
	u := &User{}
	var role Role
	var expiresRaw string
	err := r.db.QueryRow(query, token).Scan(&u.ID, &u.Name, &u.Email, &u.RoleID, &u.Status, &expiresRaw, &role.ID, &role.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, time.Time{}, fmt.Errorf("invalid activation token")
		}
		return nil, time.Time{}, fmt.Errorf("error querying activation token: %w", err)
	}
	if role.ID > 0 {
		u.Role = &role
	}

	expiresAt, err := parseFlexibleTime(expiresRaw)
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("activation token has invalid expiration: %w", err)
	}
	return u, expiresAt, nil
}

func parseFlexibleTime(raw string) (time.Time, error) {
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05.999999999 -0700 MST",
		"2006-01-02 15:04:05 -0700 MST",
		"2006-01-02 15:04:05",
	}
	// Strip monotonic clock suffix (" m=+...") if present from earlier Go time.String() writes
	if idx := strings.Index(raw, " m="); idx > 0 {
		raw = raw[:idx]
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, raw); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognized time format: %s", raw)
}

// ActivateWithPassword sets the user's password, optional professor classes/subjects, marks it ACTIVE, and clears the activation token.
func (r *Repository) ActivateWithPassword(userID int, passwordHash, classesCSV, subjectsCSV string) error {
	query := `
		UPDATE users
		SET password_hash = ?, classes_csv = ?, subjects_csv = ?, status = ?, activation_token = '', activation_expires_at = NULL, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err := r.db.Exec(query, passwordHash, classesCSV, subjectsCSV, StatusActive, userID)
	return err
}