package user

import (
	"fmt"
	"laboratory-management/internal/audit"
	"strings"
)

// Service handles user business logic
type Service struct {
	userRepo  *Repository
	auditRepo *audit.Repository
}

// NewService creates a new user service
func NewService(userRepo *Repository, auditRepo *audit.Repository) *Service {
	return &Service{
		userRepo:  userRepo,
		auditRepo: auditRepo,
	}
}

// GetUserByID retrieves a user by ID
func (s *Service) GetUserByID(id int) (*User, error) {
	return s.userRepo.GetByID(id)
}

// GetUserByEmail retrieves a user by email
func (s *Service) GetUserByEmail(email string) (*User, error) {
	return s.userRepo.GetByEmail(email)
}

// GetAllUsers retrieves all users
func (s *Service) GetAllUsers() ([]*User, error) {
	return s.userRepo.GetAll()
}

// CreateUser creates a new user
func (s *Service) CreateUser(req *CreateUserRequest, creatorID int) (*User, error) {
	// Validate input
	if req.Name == "" || req.Email == "" || req.Password == "" {
		return nil, fmt.Errorf("name, email, and password are required")
	}

	// Check if email already exists
	_, err := s.userRepo.GetByEmail(req.Email)
	if err == nil {
		return nil, fmt.Errorf("email already registered")
	}

	// Prevent users from creating admin accounts (unless they are admin)
	role, err := s.userRepo.GetRoleByID(req.RoleID)
	if err != nil {
		return nil, fmt.Errorf("invalid role")
	}

	if role.Name == RoleAdmin {
		// In production, check if creator is admin
		return nil, fmt.Errorf("cannot create admin accounts")
	}

	// Hash password should be done in auth service
	// For now, we'll expect the password to already be hashed
	// This is a simplification - in production, use auth service
	
	newUser := &User{
		Name:              req.Name,
		Email:             req.Email,
		PasswordHash:      req.Password, // Should be hashed
		RoleID:            req.RoleID,
		RegistrationNumber: req.RegistrationNumber,
		Phone:             req.Phone,
		Status:            StatusActive,
	}

	err = s.userRepo.Create(newUser)
	if err != nil {
		return nil, fmt.Errorf("error creating user: %w", err)
	}

	// Log audit
	auditLog := &audit.AuditLog{
		UserID:     &creatorID,
		Action:     audit.ActionCreateUser,
		EntityType: audit.EntityUser,
		EntityID:   &newUser.ID,
		Details:    fmt.Sprintf("Created user: %s (%s)", newUser.Name, newUser.Email),
	}
	s.auditRepo.Create(auditLog)

	// Remove password hash from response
	newUser.PasswordHash = ""

	return newUser, nil
}

// UpdateUser updates a user
func (s *Service) UpdateUser(id int, req *UpdateUserRequest, updaterID int) (*User, error) {
	// Get existing user
	existingUser, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Update fields
	existingUser.Name = req.Name
	existingUser.Email = req.Email
	existingUser.Phone = req.Phone
	existingUser.Status = req.Status

	err = s.userRepo.Update(existingUser)
	if err != nil {
		return nil, fmt.Errorf("error updating user: %w", err)
	}

	// Log audit
	auditLog := &audit.AuditLog{
		UserID:     &updaterID,
		Action:     audit.ActionUpdateUser,
		EntityType: audit.EntityUser,
		EntityID:   &id,
		Details:    fmt.Sprintf("Updated user: %s (%s)", existingUser.Name, existingUser.Email),
	}
	s.auditRepo.Create(auditLog)

	// Remove password hash from response
	existingUser.PasswordHash = ""

	return existingUser, nil
}

// SetUserStatus activates or deactivates a user, protecting the primary admin and the acting user.
func (s *Service) SetUserStatus(id int, newStatus string, actorID int, protectedAdminEmail string) error {
	target, err := s.userRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("usuário não encontrado")
	}
	if target.ID == actorID || strings.EqualFold(target.Email, protectedAdminEmail) {
		return fmt.Errorf("não é permitido alterar o status do usuário administrativo principal")
	}
	if newStatus != StatusActive && newStatus != StatusInactive {
		return fmt.Errorf("status inválido")
	}
	if err := s.userRepo.SetStatus(id, newStatus); err != nil {
		return err
	}
	if s.auditRepo != nil {
		s.auditRepo.Create(&audit.AuditLog{
			UserID:     &actorID,
			Action:     audit.ActionUpdateUser,
			EntityType: audit.EntityUser,
			EntityID:   &id,
			Details:    fmt.Sprintf("Changed status of %s (%s) to %s", target.Name, target.Email, newStatus),
		})
	}
	return nil
}

// DeleteUserSafe deletes a non-primary user along with their bookings and notifications.
func (s *Service) DeleteUserSafe(id int, deleterID int, protectedAdminEmail string) error {
	target, err := s.userRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("usuário não encontrado")
	}
	if target.ID == deleterID || strings.EqualFold(target.Email, protectedAdminEmail) {
		return fmt.Errorf("não é permitido excluir o usuário administrativo principal")
	}
	if err := s.userRepo.Delete(id); err != nil {
		return fmt.Errorf("erro ao excluir usuário: %w", err)
	}
	if s.auditRepo != nil {
		s.auditRepo.Create(&audit.AuditLog{
			UserID:     &deleterID,
			Action:     audit.ActionDeleteUser,
			EntityType: audit.EntityUser,
			EntityID:   &id,
			Details:    fmt.Sprintf("Deleted user and bookings: %s (%s)", target.Name, target.Email),
		})
	}
	return nil
}

// DeleteUser deletes a user
func (s *Service) DeleteUser(id int, deleterID int) error {
	// Get user for audit log
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	err = s.userRepo.Delete(id)
	if err != nil {
		return fmt.Errorf("error deleting user: %w", err)
	}

	// Log audit
	auditLog := &audit.AuditLog{
		UserID:     &deleterID,
		Action:     audit.ActionDeleteUser,
		EntityType: audit.EntityUser,
		EntityID:   &id,
		Details:    fmt.Sprintf("Deleted user: %s (%s)", user.Name, user.Email),
	}
	s.auditRepo.Create(auditLog)

	return nil
}

// GetRoleByID retrieves a role by ID
func (s *Service) GetRoleByID(id int) (*Role, error) {
	return s.userRepo.GetRoleByID(id)
}