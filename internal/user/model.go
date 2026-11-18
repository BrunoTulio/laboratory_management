package user

import (
	"strings"
	"time"
)

// Role represents a user role in the system
type Role struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// User represents a system user
type User struct {
	ID                int       `json:"id"`
	Name              string    `json:"name"`
	Email             string    `json:"email"`
	PasswordHash      string    `json:"-"` // Never expose in JSON
	RoleID            int       `json:"role_id"`
	Role              *Role     `json:"role,omitempty"`
	RegistrationNumber string   `json:"registration_number"`
	Phone             string    `json:"phone"`
	Status            string    `json:"status"`
	ClassesCSV        string    `json:"classes_csv,omitempty"`
	SubjectsCSV       string    `json:"subjects_csv,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// ClassesList returns the user's configured classes as a trimmed slice.
func (u *User) ClassesList() []string {
	return splitCSV(u.ClassesCSV)
}

// SubjectsList returns the user's configured subjects as a trimmed slice.
func (u *User) SubjectsList() []string {
	return splitCSV(u.SubjectsCSV)
}

func splitCSV(raw string) []string {
	var out []string
	for _, part := range strings.Split(raw, ",") {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

// UserStatus constants
const (
	StatusActive   = "ACTIVE"
	StatusPending  = "PENDING"
	StatusInactive = "INACTIVE"
	StatusBlocked  = "BLOCKED"
)

// RoleName constants
const (
	RoleProfessor = "PROFESSOR"
	RoleStudent   = "ALUNO"
	RoleAdmin     = "DIRECAO"
)

// CreateUserRequest represents the request to create a new user
type CreateUserRequest struct {
	Name              string `json:"name"`
	Email             string `json:"email"`
	Password          string `json:"password"`
	RoleID            int    `json:"role_id"`
	RegistrationNumber string `json:"registration_number"`
	Phone             string `json:"phone"`
}

// LoginRequest represents the login request
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// UpdateUserRequest represents the request to update a user
type UpdateUserRequest struct {
	Name              string `json:"name"`
	Email             string `json:"email"`
	Phone             string `json:"phone"`
	Status            string `json:"status"`
}