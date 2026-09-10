package laboratory

import "time"

// Laboratory represents a laboratory in the system
type Laboratory struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	Capacity    int       `json:"capacity"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// LaboratoryStatus constants
const (
	LabStatusActive   = "ACTIVE"
	LabStatusInactive = "INACTIVE"
	LabStatusMaintenance = "MAINTENANCE"
)

// CreateLaboratoryRequest represents the request to create a laboratory
type CreateLaboratoryRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Location    string `json:"location"`
	Capacity    int    `json:"capacity"`
}

// UpdateLaboratoryRequest represents the request to update a laboratory
type UpdateLaboratoryRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Location    string `json:"location"`
	Capacity    int    `json:"capacity"`
	Status      string `json:"status"`
}