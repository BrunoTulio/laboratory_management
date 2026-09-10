package audit

// Service handles audit log business logic
type Service struct {
	auditRepo *Repository
}

// NewService creates a new audit service
func NewService(auditRepo *Repository) *Service {
	return &Service{
		auditRepo: auditRepo,
	}
}

// GetAuditLogByID retrieves an audit log by ID
func (s *Service) GetAuditLogByID(id int) (*AuditLog, error) {
	return s.auditRepo.GetByID(id)
}

// GetAuditLogsByUserID retrieves all audit logs for a user
func (s *Service) GetAuditLogsByUserID(userID int) ([]*AuditLog, error) {
	return s.auditRepo.GetByUserID(userID)
}

// GetAuditLogsByAction retrieves all audit logs for a specific action
func (s *Service) GetAuditLogsByAction(action string) ([]*AuditLog, error) {
	return s.auditRepo.GetByAction(action)
}

// GetAuditLogsByEntity retrieves all audit logs for a specific entity
func (s *Service) GetAuditLogsByEntity(entityType string, entityID int) ([]*AuditLog, error) {
	return s.auditRepo.GetByEntity(entityType, entityID)
}

// GetAllAuditLogs retrieves all audit logs (admin)
func (s *Service) GetAllAuditLogs() ([]*AuditLog, error) {
	return s.auditRepo.GetAll()
}

// CreateAuditLog creates a new audit log entry
func (s *Service) CreateAuditLog(req *CreateAuditLogRequest) (*AuditLog, error) {
	auditLog := &AuditLog{
		UserID:     req.UserID,
		Action:     req.Action,
		EntityType: req.EntityType,
		EntityID:   req.EntityID,
		Details:    req.Details,
		IPAddress:  req.IPAddress,
		UserAgent:  req.UserAgent,
	}

	err := s.auditRepo.Create(auditLog)
	if err != nil {
		return nil, err
	}

	return auditLog, nil
}