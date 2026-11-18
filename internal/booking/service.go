package booking

import (
	"fmt"
	"time"
	"laboratory-management/internal/audit"
	"laboratory-management/internal/notification"
)

// Service handles booking business logic
type Service struct {
	bookingRepo     *Repository
	auditRepo       *audit.Repository
	notificationRepo *notification.Repository
}

// NewService creates a new booking service
func NewService(bookingRepo *Repository, auditRepo *audit.Repository, notificationRepo *notification.Repository) *Service {
	return &Service{
		bookingRepo:     bookingRepo,
		auditRepo:       auditRepo,
		notificationRepo: notificationRepo,
	}
}

// GetBookingByID retrieves a booking by ID
func (s *Service) GetBookingByID(id int) (*Booking, error) {
	return s.bookingRepo.GetByID(id)
}

// GetBookingsByUserID retrieves all bookings for a user
func (s *Service) GetBookingsByUserID(userID int) ([]*Booking, error) {
	return s.bookingRepo.GetByUserID(userID)
}

// GetAllBookings retrieves all bookings (admin)
func (s *Service) GetAllBookings() ([]*Booking, error) {
	return s.bookingRepo.GetAll()
}

// GetPendingBookings retrieves all pending bookings
func (s *Service) GetPendingBookings() ([]*Booking, error) {
	return s.bookingRepo.GetPending()
}

// GetBookingsByDate retrieves active bookings for a specific date.
func (s *Service) GetBookingsByDate(date time.Time) ([]*Booking, error) {
	return s.bookingRepo.GetByDate(date)
}

// GetUpcomingBookings retrieves upcoming active bookings starting from fromDate.
func (s *Service) GetUpcomingBookings(fromDate time.Time, limit int) ([]*Booking, error) {
	return s.bookingRepo.GetUpcoming(fromDate, limit)
}

// CreateBooking creates a new booking with conflict detection
func (s *Service) CreateBooking(req *CreateBookingRequest, userID int) (*Booking, error) {
	// Validate input
	if req.LaboratoryID == 0 {
		return nil, fmt.Errorf("laboratory ID is required")
	}
	if req.BookingDate.IsZero() {
		return nil, fmt.Errorf("booking date is required")
	}
	if req.StartTime.IsZero() || req.EndTime.IsZero() {
		return nil, fmt.Errorf("start and end times are required")
	}
	if req.StartTime.After(req.EndTime) || req.StartTime.Equal(req.EndTime) {
		return nil, fmt.Errorf("end time must be after start time")
	}
	if req.Reason == "" {
		return nil, fmt.Errorf("reason is required")
	}

	// Check for time conflicts
	hasConflict, err := s.bookingRepo.CheckConflict(
		req.LaboratoryID,
		req.BookingDate,
		req.StartTime,
		req.EndTime,
		0, // 0 means we're checking for new booking (no ID to exclude)
	)
	if err != nil {
		return nil, fmt.Errorf("error checking for conflicts: %w", err)
	}
	if hasConflict {
		return nil, fmt.Errorf("time conflict: laboratory is already booked for this time slot")
	}

	// Create booking
	newBooking := &Booking{
		UserID:       userID,
		LaboratoryID: req.LaboratoryID,
		ClassID:      req.ClassID,
		BookingDate:  req.BookingDate,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
		Shift:        req.Shift,
		TeacherName:  req.TeacherName,
		ClassName:    req.ClassName,
		SubjectName:  req.SubjectName,
		Reason:       req.Reason,
		Observation:  req.Observation,
		Status:       StatusPending,
	}

	err = s.bookingRepo.Create(newBooking)
	if err != nil {
		return nil, fmt.Errorf("error creating booking: %w", err)
	}

	// Log audit
	auditLog := &audit.AuditLog{
		UserID:     &userID,
		Action:     audit.ActionCreateBooking,
		EntityType: audit.EntityBooking,
		EntityID:   &newBooking.ID,
		Details:    fmt.Sprintf("Created booking for laboratory %d on %s", req.LaboratoryID, req.BookingDate.Format("2006-01-02")),
	}
	s.auditRepo.Create(auditLog)

	// Create notification for the user
	notification := &notification.Notification{
		UserID:    userID,
		BookingID: &newBooking.ID,
		Type:      notification.TypeBookingCreated,
		Title:     "Solicitação de Agendamento Criada",
		Message:   fmt.Sprintf("Sua solicitação de agendamento para o laboratório foi criada e está pendente de aprovação."),
		IsRead:    false,
	}
	s.notificationRepo.Create(notification)

	return newBooking, nil
}

// ApproveBooking approves a booking with conflict re-check
func (s *Service) ApproveBooking(bookingID, approverID int) (*Booking, error) {
	// Get booking
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return nil, fmt.Errorf("booking not found")
	}

	// Check if booking can be approved
	if booking.Status != StatusPending {
		return nil, fmt.Errorf("booking is not in pending status")
	}

	// Check for conflicts again (to handle race conditions)
	hasConflict, err := s.bookingRepo.CheckConflict(
		booking.LaboratoryID,
		booking.BookingDate,
		booking.StartTime,
		booking.EndTime,
		bookingID, // Exclude current booking from conflict check
	)
	if err != nil {
		return nil, fmt.Errorf("error checking for conflicts: %w", err)
	}
	if hasConflict {
		return nil, fmt.Errorf("time conflict detected during approval: laboratory is now booked for this time slot")
	}

	// Approve booking
	err = s.bookingRepo.Approve(bookingID, approverID)
	if err != nil {
		return nil, fmt.Errorf("error approving booking: %w", err)
	}

	// Refresh booking data
	booking, err = s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return nil, fmt.Errorf("error refreshing booking data: %w", err)
	}

	// Log audit
	auditLog := &audit.AuditLog{
		UserID:     &approverID,
		Action:     audit.ActionApproveBooking,
		EntityType: audit.EntityBooking,
		EntityID:   &bookingID,
		Details:    fmt.Sprintf("Approved booking %d", bookingID),
	}
	s.auditRepo.Create(auditLog)

	// Create notification for the requester
	notification := &notification.Notification{
		UserID:    booking.UserID,
		BookingID: &bookingID,
		Type:      notification.TypeBookingApproved,
		Title:     "Agendamento Aprovado",
		Message:   fmt.Sprintf("Seu agendamento para o laboratório foi aprovado."),
		IsRead:    false,
	}
	s.notificationRepo.Create(notification)

	return booking, nil
}

// RejectBooking rejects a booking
func (s *Service) RejectBooking(bookingID, rejecterID int, rejectionReason string) (*Booking, error) {
	// Get booking
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return nil, fmt.Errorf("booking not found")
	}

	// Check if booking can be rejected
	if booking.Status != StatusPending {
		return nil, fmt.Errorf("booking is not in pending status")
	}

	// Reject booking
	err = s.bookingRepo.Reject(bookingID, rejecterID, rejectionReason)
	if err != nil {
		return nil, fmt.Errorf("error rejecting booking: %w", err)
	}

	// Refresh booking data
	booking, err = s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return nil, fmt.Errorf("error refreshing booking data: %w", err)
	}

	// Log audit
	auditLog := &audit.AuditLog{
		UserID:     &rejecterID,
		Action:     audit.ActionRejectBooking,
		EntityType: audit.EntityBooking,
		EntityID:   &bookingID,
		Details:    fmt.Sprintf("Rejected booking %d. Reason: %s", bookingID, rejectionReason),
	}
	s.auditRepo.Create(auditLog)

	// Create notification for the requester
	notification := &notification.Notification{
		UserID:    booking.UserID,
		BookingID: &bookingID,
		Type:      notification.TypeBookingRejected,
		Title:     "Agendamento Rejeitado",
		Message:   fmt.Sprintf("Seu agendamento foi rejeitado. Motivo: %s", rejectionReason),
		IsRead:    false,
	}
	s.notificationRepo.Create(notification)

	return booking, nil
}

// CancelBooking cancels a booking
func (s *Service) CancelBooking(bookingID, cancellerID int, cancellationReason string) (*Booking, error) {
	// Get booking
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return nil, fmt.Errorf("booking not found")
	}

	// Check if booking can be cancelled
	if booking.Status == StatusCancelled || booking.Status == StatusCompleted {
		return nil, fmt.Errorf("booking cannot be cancelled")
	}

	// Cancel booking
	err = s.bookingRepo.Cancel(bookingID, cancellerID, cancellationReason)
	if err != nil {
		return nil, fmt.Errorf("error cancelling booking: %w", err)
	}

	// Refresh booking data
	booking, err = s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return nil, fmt.Errorf("error refreshing booking data: %w", err)
	}

	// Log audit
	auditLog := &audit.AuditLog{
		UserID:     &cancellerID,
		Action:     audit.ActionCancelBooking,
		EntityType: audit.EntityBooking,
		EntityID:   &bookingID,
		Details:    fmt.Sprintf("Cancelled booking %d. Reason: %s", bookingID, cancellationReason),
	}
	s.auditRepo.Create(auditLog)

	// Create notification for the requester (if not the same person)
	if booking.UserID != cancellerID {
		notification := &notification.Notification{
			UserID:    booking.UserID,
			BookingID: &bookingID,
			Type:      notification.TypeBookingCancelled,
			Title:     "Agendamento Cancelado",
			Message:   fmt.Sprintf("Seu agendamento foi cancelado. Motivo: %s", cancellationReason),
			IsRead:    false,
		}
		s.notificationRepo.Create(notification)
	}

	return booking, nil
}

// CheckAvailability checks availability for a specific laboratory and time
func (s *Service) CheckAvailability(laboratoryID int, date time.Time, startTime, endTime time.Time) (bool, error) {
	hasConflict, err := s.bookingRepo.CheckConflict(laboratoryID, date, startTime, endTime, 0)
	if err != nil {
		return false, fmt.Errorf("error checking availability: %w", err)
	}
	return !hasConflict, nil
}

// GetBookingsByLaboratoryAndDate retrieves bookings for a laboratory on a specific date
func (s *Service) GetBookingsByLaboratoryAndDate(laboratoryID int, date time.Time) ([]*Booking, error) {
	return s.bookingRepo.GetByLaboratoryIDAndDate(laboratoryID, date)
}