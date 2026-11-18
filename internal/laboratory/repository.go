package laboratory

import (
	"database/sql"
	"fmt"
	"time"
)

// Repository handles laboratory data operations
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new laboratory repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Create creates a new laboratory
func (r *Repository) Create(lab *Laboratory) error {
	query := `
		INSERT INTO laboratories (name, description, location, capacity, status)
		VALUES (?, ?, ?, ?, ?)
		RETURNING id, created_at, updated_at
	`
	
	err := r.db.QueryRow(
		query,
		lab.Name,
		lab.Description,
		lab.Location,
		lab.Capacity,
		lab.Status,
	).Scan(&lab.ID, &lab.CreatedAt, &lab.UpdatedAt)
	
	if err != nil {
		return fmt.Errorf("error creating laboratory: %w", err)
	}
	
	return nil
}

// GetByID retrieves a laboratory by ID
func (r *Repository) GetByID(id int) (*Laboratory, error) {
	query := `
		SELECT id, name, description, location, capacity, status, created_at, updated_at
		FROM laboratories
		WHERE id = ?
	`
	
	lab := &Laboratory{}
	err := r.db.QueryRow(query, id).Scan(
		&lab.ID,
		&lab.Name,
		&lab.Description,
		&lab.Location,
		&lab.Capacity,
		&lab.Status,
		&lab.CreatedAt,
		&lab.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("laboratory not found")
		}
		return nil, fmt.Errorf("error getting laboratory: %w", err)
	}
	
	return lab, nil
}

// GetAll retrieves all laboratories
func (r *Repository) GetAll() ([]*Laboratory, error) {
	query := `
		SELECT id, name, description, location, capacity, status, created_at, updated_at
		FROM laboratories
		ORDER BY name
	`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error getting laboratories: %w", err)
	}
	defer rows.Close()
	
	var laboratories []*Laboratory
	for rows.Next() {
		lab := &Laboratory{}
		
		err := rows.Scan(
			&lab.ID,
			&lab.Name,
			&lab.Description,
			&lab.Location,
			&lab.Capacity,
			&lab.Status,
			&lab.CreatedAt,
			&lab.UpdatedAt,
		)
		
		if err != nil {
			return nil, fmt.Errorf("error scanning laboratory: %w", err)
		}
		
		laboratories = append(laboratories, lab)
	}
	
	return laboratories, nil
}

// Update updates a laboratory
func (r *Repository) Update(lab *Laboratory) error {
	query := `
		UPDATE laboratories 
		SET name = ?, description = ?, location = ?, capacity = ?, status = ?, updated_at = ?
		WHERE id = ?
	`
	
	lab.UpdatedAt = time.Now()
	
	result, err := r.db.Exec(
		query,
		lab.Name,
		lab.Description,
		lab.Location,
		lab.Capacity,
		lab.Status,
		lab.UpdatedAt,
		lab.ID,
	)
	
	if err != nil {
		return fmt.Errorf("error updating laboratory: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("laboratory not found")
	}
	
	return nil
}

// Delete deletes a laboratory
func (r *Repository) Delete(id int) error {
	query := `DELETE FROM laboratories WHERE id = ?`
	
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error deleting laboratory: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("laboratory not found")
	}
	
	return nil
}