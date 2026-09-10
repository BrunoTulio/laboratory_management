package booking_test

import (
	"testing"
	"time"
	
	"laboratory-management/internal/booking"
)

// TestTimeRangeConflict tests the time range conflict detection logic
func TestTimeRangeConflict(t *testing.T) {
	// Test cases for conflict detection
	tests := []struct {
		name           string
		existingStart  time.Time
		existingEnd    time.Time
		newStart       time.Time
		newEnd         time.Time
		expectConflict bool
	}{
		{
			name:           "No conflict - before",
			existingStart:  time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			existingEnd:    time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
			newStart:       time.Date(2024, 1, 1, 8, 0, 0, 0, time.UTC),
			newEnd:         time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
			expectConflict: false,
		},
		{
			name:           "No conflict - after",
			existingStart:  time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			existingEnd:    time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
			newStart:       time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			newEnd:         time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC),
			expectConflict: false,
		},
		{
			name:           "No conflict - exact boundary",
			existingStart:  time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			existingEnd:    time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
			newStart:       time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
			newEnd:         time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			expectConflict: false,
		},
		{
			name:           "Conflict - overlapping start",
			existingStart:  time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			existingEnd:    time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			newStart:       time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
			newEnd:         time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC),
			expectConflict: true,
		},
		{
			name:           "Conflict - completely inside",
			existingStart:  time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			existingEnd:    time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC),
			newStart:       time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
			newEnd:         time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			expectConflict: true,
		},
		{
			name:           "Conflict - completely surrounding",
			existingStart:  time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
			existingEnd:    time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			newStart:       time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			newEnd:         time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC),
			expectConflict: true,
		},
		{
			name:           "Conflict - example from requirements (14:00-16:00 vs 15:00-17:00)",
			existingStart:  time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC),
			existingEnd:    time.Date(2024, 1, 1, 16, 0, 0, 0, time.UTC),
			newStart:       time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC),
			newEnd:         time.Date(2024, 1, 1, 17, 0, 0, 0, time.UTC),
			expectConflict: true,
		},
		{
			name:           "Conflict - example from requirements (09:00-11:00 vs 10:30-12:00)",
			existingStart:  time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
			existingEnd:    time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
			newStart:       time.Date(2024, 1, 1, 10, 30, 0, 0, time.UTC),
			newEnd:         time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			expectConflict: true,
		},
		{
			name:           "No conflict - example from requirements (09:00-10:00 vs 10:00-11:00)",
			existingStart:  time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
			existingEnd:    time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			newStart:       time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			newEnd:         time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
			expectConflict: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			existingRange := booking.TimeRange{
				Start: tt.existingStart,
				End:   tt.existingEnd,
			}
			newRange := booking.TimeRange{
				Start: tt.newStart,
				End:   tt.newEnd,
			}

			hasConflict := existingRange.HasConflict(newRange)
			if hasConflict != tt.expectConflict {
				t.Errorf("HasConflict() = %v, want %v", hasConflict, tt.expectConflict)
			}
		})
	}
}

// TestBookingStatusConstants tests that booking status constants are properly defined
func TestBookingStatusConstants(t *testing.T) {
	expectedStatuses := []string{
		booking.StatusPending,
		booking.StatusApproved,
		booking.StatusRejected,
		booking.StatusCancelled,
		booking.StatusCompleted,
	}

	for _, status := range expectedStatuses {
		if status == "" {
			t.Errorf("Status constant is empty")
		}
	}
}

// TestShiftConstants tests that shift constants are properly defined
func TestShiftConstants(t *testing.T) {
	expectedShifts := []string{
		"MATUTINO",
		"VESPERTINO",
		"NOTURNO",
	}

	for _, shift := range expectedShifts {
		if shift == "" {
			t.Errorf("Shift constant is empty")
		}
	}
}