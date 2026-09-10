package user

import (
	"database/sql"
	"fmt"
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
		VALUES ($1, $2, $3, $4, $5, $6, $7)
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
		       u.phone, u.status, u.created_at, u.updated_at,
		       r.id, r.name, r.description, r.created_at, r.updated_at
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		WHERE u.id = $1
	`
	
	user := &User{}
	var role Role
	
	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.RoleID,
		&user.RegistrationNumber, &user.Phone, &user.Status, &user.CreatedAt, &user.UpdatedAt,
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
		       u.phone, u.status, u.created_at, u.updated_at,
		       r.id, r.name, r.description, r.created_at, r.updated_at
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		WHERE u.email = $1
	`
	
	user := &User{}
	var role Role
	
	err := r.db.QueryRow(query, email).Scan(
		&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.RoleID,
		&user.RegistrationNumber, &user.Phone, &user.Status, &user.CreatedAt, &user.UpdatedAt,
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
		       u.phone, u.status, u.created_at, u.updated_at,
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
			&user.RegistrationNumber, &user.Phone, &user.Status, &user.CreatedAt, &user.UpdatedAt,
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
		SET name = $1, email = $2, phone = $3, status = $4, updated_at = $5
		WHERE id = $6
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

// Delete deletes a user
func (r *Repository) Delete(id int) error {
	query := `DELETE FROM users WHERE id = $1`
	
	result, err := r.db.Exec(query, id)
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
	
	return nil
}

// GetRoleByName retrieves a role by name
func (r *Repository) GetRoleByName(name string) (*Role, error) {
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM roles
		WHERE name = $1
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
		WHERE id = $1
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