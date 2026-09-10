package laboratory

import (
	"fmt"
	"laboratory-management/internal/audit"
)

// Service handles laboratory business logic
type Service struct {
	labRepo   *Repository
	auditRepo *audit.Repository
}

// NewService creates a new laboratory service
func NewService(labRepo *Repository, auditRepo *audit.Repository) *Service {
	return &Service{
		labRepo:   labRepo,
		auditRepo: auditRepo,
	}
}

// GetLaboratoryByID retrieves a laboratory by ID
func (s *Service) GetLaboratoryByID(id int) (*Laboratory, error) {
	return s.labRepo.GetByID(id)
}

// GetAllLaboratories retrieves all laboratories
func (s *Service) GetAllLaboratories() ([]*Laboratory, error) {
	return s.labRepo.GetAll()
}

// CreateLaboratory creates a new laboratory
func (s *Service) CreateLaboratory(req *CreateLaboratoryRequest, creatorID int) (*Laboratory, error) {
	// Validate input
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	newLab := &Laboratory{
		Name:        req.Name,
		Description: req.Description,
		Location:    req.Location,
		Capacity:    req.Capacity,
		Status:      LabStatusActive,
	}

	err := s.labRepo.Create(newLab)
	if err != nil {
		return nil, fmt.Errorf("error creating laboratory: %w", err)
	}

	// Log audit
	auditLog := &audit.AuditLog{
		UserID:     &creatorID,
		Action:     audit.ActionCreateUser, // Reuse action or create new one
		EntityType: audit.EntityLaboratory,
		EntityID:   &newLab.ID,
		Details:    fmt.Sprintf("Created laboratory: %s", newLab.Name),
	}
	s.auditRepo.Create(auditLog)

	return newLab, nil
}

// UpdateLaboratory updates a laboratory
func (s *Service) UpdateLaboratory(id int, req *UpdateLaboratoryRequest, updaterID int) (*Laboratory, error) {
	// Get existing laboratory
	existingLab, err := s.labRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("laboratory not found")
	}

	// Update fields
	existingLab.Name = req.Name
	existingLab.Description = req.Description
	existingLab.Location = req.Location
	existingLab.Capacity = req.Capacity
	existingLab.Status = req.Status

	err = s.labRepo.Update(existingLab)
	if err != nil {
		return nil, fmt.Errorf("error updating laboratory: %w", err)
	}

	// Log audit
	auditLog := &audit.AuditLog{
		UserID:     &updaterID,
		Action:     audit.ActionUpdateUser, // Reuse action or create new one
		EntityType: audit.EntityLaboratory,
		EntityID:   &id,
		Details:    fmt.Sprintf("Updated laboratory: %s", existingLab.Name),
	}
	s.auditRepo.Create(auditLog)

	return existingLab, nil
}

// DeleteLaboratory deletes a laboratory
func (s *Service) DeleteLaboratory(id int, deleterID int) error {
	// Get laboratory for audit log
	lab, err := s.labRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("laboratory not found")
	}

	err = s.labRepo.Delete(id)
	if err != nil {
		return fmt.Errorf("error deleting laboratory: %w", err)
	}

	// Log audit
	auditLog := &audit.AuditLog{
		UserID:     &deleterID,
		Action:     audit.ActionDeleteUser, // Reuse action or create new one
		EntityType: audit.EntityLaboratory,
		EntityID:   &id,
		Details:    fmt.Sprintf("Deleted laboratory: %s", lab.Name),
	}
	s.auditRepo.Create(auditLog)

	return nil
}