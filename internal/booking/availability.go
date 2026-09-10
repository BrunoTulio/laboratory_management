package booking

import (
	"time"
)

// AvailabilityService handles availability calculations
type AvailabilityService struct {
	bookingRepo *Repository
}

// NewAvailabilityService creates a new availability service
func NewAvailabilityService(bookingRepo *Repository) *AvailabilityService {
	return &AvailabilityService{
		bookingRepo: bookingRepo,
	}
}

// TimeSlot represents a time slot with its availability status
type TimeSlot struct {
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Status      string    `json:"status"` // AVAILABLE, OCCUPIED, PENDING
	BookingID   *int      `json:"booking_id,omitempty"`
	BookingInfo *Booking  `json:"booking_info,omitempty"`
}

// DailyAvailability represents availability for a full day
type DailyAvailability struct {
	Date       time.Time   `json:"date"`
	LaboratoryID int        `json:"laboratory_id"`
	TimeSlots  []TimeSlot  `json:"time_slots"`
}

// GetDailyAvailability gets availability for a laboratory on a specific date
func (s *AvailabilityService) GetDailyAvailability(laboratoryID int, date time.Time) (*DailyAvailability, error) {
	// Get all bookings for the laboratory on the date
	bookings, err := s.bookingRepo.GetByLaboratoryIDAndDate(laboratoryID, date)
	if err != nil {
		return nil, err
	}

	// Generate time slots for the day
	timeSlots := s.generateTimeSlots(date, bookings)

	return &DailyAvailability{
		Date:        date,
		LaboratoryID: laboratoryID,
		TimeSlots:   timeSlots,
	}, nil
}

// generateTimeSlots generates time slots for a day based on bookings
func (s *AvailabilityService) generateTimeSlots(date time.Time, bookings []*Booking) []TimeSlot {
	// Define standard time slots (e.g., hourly from 8:00 to 18:00)
	// This can be configured based on laboratory hours
	baseSlots := []struct {
		start string
		end   string
	}{
		{"08:00", "09:00"},
		{"09:00", "10:00"},
		{"10:00", "11:00"},
		{"11:00", "12:00"},
		{"13:00", "14:00"},
		{"14:00", "15:00"},
		{"15:00", "16:00"},
		{"16:00", "17:00"},
		{"17:00", "18:00"},
	}

	var timeSlots []TimeSlot

	for _, slot := range baseSlots {
		startTime, _ := time.Parse("15:04", slot.start)
		endTime, _ := time.Parse("15:04", slot.end)

		// Check if this slot is occupied
		var occupiedBooking *Booking
		status := "AVAILABLE"

		for _, booking := range bookings {
			// Check if booking overlaps with this time slot
			bookingStart := booking.StartTime
			bookingEnd := booking.EndTime

			// Simple overlap check
			if (startTime.Before(bookingEnd) || startTime.Equal(bookingEnd)) &&
			   (endTime.After(bookingStart) || endTime.Equal(bookingStart)) {
				occupiedBooking = booking
				if booking.Status == "APPROVED" {
					status = "OCCUPIED"
				} else if booking.Status == "PENDING" {
					status = "PENDING"
				}
				break
			}
		}

		timeSlot := TimeSlot{
			StartTime: startTime,
			EndTime:   endTime,
			Status:    status,
		}

		if occupiedBooking != nil {
			timeSlot.BookingID = &occupiedBooking.ID
			timeSlot.BookingInfo = occupiedBooking
		}

		timeSlots = append(timeSlots, timeSlot)
	}

	return timeSlots
}

// CheckTimeSlotAvailability checks if a specific time slot is available
func (s *AvailabilityService) CheckTimeSlotAvailability(laboratoryID int, date time.Time, startTime, endTime time.Time) (bool, string, error) {
	hasConflict, err := s.bookingRepo.CheckConflict(laboratoryID, date, startTime, endTime, 0)
	if err != nil {
		return false, "", err
	}

	if hasConflict {
		return false, "CONFLICT", nil
	}

	return true, "AVAILABLE", nil
}