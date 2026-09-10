package class

import "time"

// Class represents a class/turma in the system
type Class struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Course    string    `json:"course"`
	GradeYear string    `json:"grade_year"`
	Shift     string    `json:"shift"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ClassStatus constants
const (
	ClassStatusActive   = "ACTIVE"
	ClassStatusInactive = "INACTIVE"
)

// Shift constants
const (
	ShiftMorning   = "MATUTINO"
	ShiftAfternoon = "VESPERTINO"
	ShiftNight     = "NOTURNO"
)

// CreateClassRequest represents the request to create a class
type CreateClassRequest struct {
	Name      string `json:"name"`
	Course    string `json:"course"`
	GradeYear string `json:"grade_year"`
	Shift     string `json:"shift"`
}

// UpdateClassRequest represents the request to update a class
type UpdateClassRequest struct {
	Name      string `json:"name"`
	Course    string `json:"course"`
	GradeYear string `json:"grade_year"`
	Shift     string `json:"shift"`
	Status    string `json:"status"`
}