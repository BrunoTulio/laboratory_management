package notification

// Service handles notification business logic
type Service struct {
	notificationRepo *Repository
}

// NewService creates a new notification service
func NewService(notificationRepo *Repository) *Service {
	return &Service{
		notificationRepo: notificationRepo,
	}
}

// GetNotificationByID retrieves a notification by ID
func (s *Service) GetNotificationByID(id int) (*Notification, error) {
	return s.notificationRepo.GetByID(id)
}

// GetNotificationsByUserID retrieves all notifications for a user
func (s *Service) GetNotificationsByUserID(userID int) ([]*Notification, error) {
	return s.notificationRepo.GetByUserID(userID)
}

// GetUnreadCount retrieves the count of unread notifications for a user
func (s *Service) GetUnreadCount(userID int) (int, error) {
	return s.notificationRepo.GetUnreadCount(userID)
}

// CreateNotification creates a new notification
func (s *Service) CreateNotification(req *CreateNotificationRequest) (*Notification, error) {
	notification := &Notification{
		UserID:    req.UserID,
		BookingID: req.BookingID,
		Type:      req.Type,
		Title:     req.Title,
		Message:   req.Message,
		IsRead:    false,
	}

	err := s.notificationRepo.Create(notification)
	if err != nil {
		return nil, err
	}

	return notification, nil
}

// MarkAsRead marks a notification as read
func (s *Service) MarkAsRead(notificationID int) error {
	return s.notificationRepo.MarkAsRead(notificationID)
}

// MarkAllAsRead marks all notifications for a user as read
func (s *Service) MarkAllAsRead(userID int) error {
	return s.notificationRepo.MarkAllAsRead(userID)
}

// DeleteNotification deletes a notification
func (s *Service) DeleteNotification(notificationID int) error {
	return s.notificationRepo.Delete(notificationID)
}