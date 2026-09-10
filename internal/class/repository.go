package class

import (
	"database/sql"
	"fmt"
	"time"
)

// Repository handles class data operations
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new class repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Create creates a new class
func (r *Repository) Create(cls *Class) error {
	query := `
		INSERT INTO classes (name, course, grade_year, shift, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`
	
	err := r.db.QueryRow(
		query,
		cls.Name,
		cls.Course,
		cls.GradeYear,
		cls.Shift,
		cls.Status,
	).Scan(&cls.ID, &cls.CreatedAt, &cls.UpdatedAt)
	
	if err != nil {
		return fmt.Errorf("error creating class: %w", err)
	}
	
	return nil
}

// GetByID retrieves a class by ID
func (r *Repository) GetByID(id int) (*Class, error) {
	query := `
		SELECT id, name, course, grade_year, shift, status, created_at, updated_at
		FROM classes
		WHERE id = $1
	`
	
	cls := &Class{}
	err := r.db.QueryRow(query, id).Scan(
		&cls.ID,
		&cls.Name,
		&cls.Course,
		&cls.GradeYear,
		&cls.Shift,
		&cls.Status,
		&cls.CreatedAt,
		&cls.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("class not found")
		}
		return nil, fmt.Errorf("error getting class: %w", err)
	}
	
	return cls, nil
}

// GetAll retrieves all classes
func (r *Repository) GetAll() ([]*Class, error) {
	query := `
		SELECT id, name, course, grade_year, shift, status, created_at, updated_at
		FROM classes
		ORDER BY name
	`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error getting classes: %w", err)
	}
	defer rows.Close()
	
	var classes []*Class
	for rows.Next() {
		cls := &Class{}
		
		err := rows.Scan(
			&cls.ID,
			&cls.Name,
			&cls.Course,
			&cls.GradeYear,
			&cls.Shift,
			&cls.Status,
			&cls.CreatedAt,
			&cls.UpdatedAt,
		)
		
		if err != nil {
			return nil, fmt.Errorf("error scanning class: %w", err)
		}
		
		classes = append(classes, cls)
	}
	
	return classes, nil
}

// Update updates a class
func (r *Repository) Update(cls *Class) error {
	query := `
		UPDATE classes 
		SET name = $1, course = $2, grade_year = $3, shift = $4, status = $5, updated_at = $6
		WHERE id = $7
	`
	
	cls.UpdatedAt = time.Now()
	
	result, err := r.db.Exec(
		query,
		cls.Name,
		cls.Course,
		cls.GradeYear,
		cls.Shift,
		cls.Status,
		cls.UpdatedAt,
		cls.ID,
	)
	
	if err != nil {
		return fmt.Errorf("error updating class: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("class not found")
	}
	
	return nil
}

// Delete deletes a class
func (r *Repository) Delete(id int) error {
	query := `DELETE FROM classes WHERE id = $1`
	
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error deleting class: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("class not found")
	}
	
	return nil
}