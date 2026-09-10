package class

import (
	"fmt"
	"laboratory-management/internal/audit"
)

// Service handles class business logic
type Service struct {
	classRepo  *Repository
	auditRepo  *audit.Repository
}

// NewService creates a new class service
func NewService(classRepo *Repository, auditRepo *audit.Repository) *Service {
	return &Service{
		classRepo: classRepo,
		auditRepo: auditRepo,
	}
}

// GetClassByID retrieves a class by ID
func (s *Service) GetClassByID(id int) (*Class, error) {
	return s.classRepo.GetByID(id)
}

// GetAllClasses retrieves all classes
func (s *Service) GetAllClasses() ([]*Class, error) {
	return s.classRepo.GetAll()
}

// CreateClass creates a new class
func (s *Service) CreateClass(req *CreateClassRequest, creatorID int) (*Class, error) {
	// Validate input
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	newClass := &Class{
		Name:      req.Name,
		Course:    req.Course,
		GradeYear: req.GradeYear,
		Shift:     req.Shift,
		Status:    ClassStatusActive,
	}

	err := s.classRepo.Create(newClass)
	if err != nil {
		return nil, fmt.Errorf("error creating class: %w", err)
	}

	// Log audit
	auditLog := &audit.AuditLog{
		UserID:     &creatorID,
		Action:     audit.ActionCreateUser, // Reuse action or create new one
		EntityType: audit.EntityClass,
		EntityID:   &newClass.ID,
		Details:    fmt.Sprintf("Created class: %s", newClass.Name),
	}
	s.auditRepo.Create(auditLog)

	return newClass, nil
}

// UpdateClass updates a class
func (s *Service) UpdateClass(id int, req *UpdateClassRequest, updaterID int) (*Class, error) {
	// Get existing class
	existingClass, err := s.classRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("class not found")
	}

	// Update fields
	existingClass.Name = req.Name
	existingClass.Course = req.Course
	existingClass.GradeYear = req.GradeYear
	existingClass.Shift = req.Shift
	existingClass.Status = req.Status

	err = s.classRepo.Update(existingClass)
	if err != nil {
		return nil, fmt.Errorf("error updating class: %w", err)
	}

	// Log audit
	auditLog := &audit.AuditLog{
		UserID:     &updaterID,
		Action:     audit.ActionUpdateUser, // Reuse action or create new one
		EntityType: audit.EntityClass,
		EntityID:   &id,
		Details:    fmt.Sprintf("Updated class: %s", existingClass.Name),
	}
	s.auditRepo.Create(auditLog)

	return existingClass, nil
}

// DeleteClass deletes a class
func (s *Service) DeleteClass(id int, deleterID int) error {
	// Get class for audit log
	class, err := s.classRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("class not found")
	}

	err = s.classRepo.Delete(id)
	if err != nil {
		return fmt.Errorf("error deleting class: %w", err)
	}

	// Log audit
	auditLog := &audit.AuditLog{
		UserID:     &deleterID,
		Action:     audit.ActionDeleteUser, // Reuse action or create new one
		EntityType: audit.EntityClass,
		EntityID:   &id,
		Details:    fmt.Sprintf("Deleted class: %s", class.Name),
	}
	s.auditRepo.Create(auditLog)

	return nil
}