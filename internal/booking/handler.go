package booking

import (
	"encoding/json"
	"laboratory-management/internal/middleware"
	"net/http"
	"strconv"
	"time"
)

// Handler handles HTTP requests for bookings
type Handler struct {
	bookingService *Service
}

// NewHandler creates a new booking handler
func NewHandler(bookingService *Service) *Handler {
	return &Handler{
		bookingService: bookingService,
	}
}

// ListHandler handles listing user's bookings
func (h *Handler) ListHandler(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r)
	if userID == 0 {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	bookings, err := h.bookingService.GetBookingsByUserID(userID)
	if err != nil {
		http.Error(w, "Error fetching bookings", http.StatusInternalServerError)
		return
	}

	// Render booking list
	h.renderBookingList(w, r, bookings)
}

// DetailHandler handles viewing a booking detail
func (h *Handler) DetailHandler(w http.ResponseWriter, r *http.Request) {
	bookingIDStr := r.URL.Query().Get("id")
	bookingID, err := strconv.Atoi(bookingIDStr)
	if err != nil {
		http.Error(w, "Invalid booking ID", http.StatusBadRequest)
		return
	}

	booking, err := h.bookingService.GetBookingByID(bookingID)
	if err != nil {
		http.Error(w, "Booking not found", http.StatusNotFound)
		return
	}

	// Check if user has permission to view this booking
	userID := middleware.GetUserIDFromContext(r)
	userRole := middleware.GetUserRoleFromContext(r)
	if booking.UserID != userID && userRole != "DIRECAO" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Render booking detail
	h.renderBookingDetail(w, r, booking)
}

// CreateHandler handles creating a new booking
func (h *Handler) CreateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		h.renderBookingCreateForm(w, r)
		return
	}

	if r.Method == http.MethodPost {
		var req CreateBookingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		userID := middleware.GetUserIDFromContext(r)
		booking, err := h.bookingService.CreateBooking(&req, userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(booking)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// AvailabilityHandler handles checking availability
func (h *Handler) AvailabilityHandler(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	labIDStr := r.URL.Query().Get("laboratory_id")
	dateStr := r.URL.Query().Get("date")
	startTimeStr := r.URL.Query().Get("start_time")
	endTimeStr := r.URL.Query().Get("end_time")

	labID, err := strconv.Atoi(labIDStr)
	if err != nil {
		http.Error(w, "Invalid laboratory ID", http.StatusBadRequest)
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		http.Error(w, "Invalid date format", http.StatusBadRequest)
		return
	}

	startTime, err := time.Parse("15:04", startTimeStr)
	if err != nil {
		http.Error(w, "Invalid start time format", http.StatusBadRequest)
		return
	}

	endTime, err := time.Parse("15:04", endTimeStr)
	if err != nil {
		http.Error(w, "Invalid end time format", http.StatusBadRequest)
		return
	}

	available, err := h.bookingService.CheckAvailability(labID, date, startTime, endTime)
	if err != nil {
		http.Error(w, "Error checking availability", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"available": available,
	})
}

// AdminBookingListHandler handles listing all bookings (admin)
func (h *Handler) AdminBookingListHandler(w http.ResponseWriter, r *http.Request) {
	bookings, err := h.bookingService.GetAllBookings()
	if err != nil {
		http.Error(w, "Error fetching bookings", http.StatusInternalServerError)
		return
	}

	// Render admin booking list
	h.renderAdminBookingList(w, r, bookings)
}

// AdminPendingListHandler handles listing pending bookings (admin)
func (h *Handler) AdminPendingListHandler(w http.ResponseWriter, r *http.Request) {
	bookings, err := h.bookingService.GetPendingBookings()
	if err != nil {
		http.Error(w, "Error fetching pending bookings", http.StatusInternalServerError)
		return
	}

	// Render pending booking list
	h.renderPendingBookingList(w, r, bookings)
}

// AdminApproveHandler handles approving a booking (admin)
func (h *Handler) AdminApproveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ApproveBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	approverID := middleware.GetUserIDFromContext(r)
	booking, err := h.bookingService.ApproveBooking(req.BookingID, approverID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(booking)
}

// AdminRejectHandler handles rejecting a booking (admin)
func (h *Handler) AdminRejectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RejectBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	rejecterID := middleware.GetUserIDFromContext(r)
	booking, err := h.bookingService.RejectBooking(req.BookingID, rejecterID, req.RejectionReason)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(booking)
}

// AdminCancelHandler handles cancelling a booking (admin)
func (h *Handler) AdminCancelHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CancelBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	cancellerID := middleware.GetUserIDFromContext(r)
	booking, err := h.bookingService.CancelBooking(req.BookingID, cancellerID, req.CancellationReason)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(booking)
}

// renderBookingList renders the booking list page
func (h *Handler) renderBookingList(w http.ResponseWriter, r *http.Request, bookings []*Booking) {
	// This will be implemented with HTML templates
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bookings)
}

// renderBookingDetail renders the booking detail page
func (h *Handler) renderBookingDetail(w http.ResponseWriter, r *http.Request, booking *Booking) {
	// This will be implemented with HTML templates
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(booking)
}

// renderBookingCreateForm renders the booking creation form
func (h *Handler) renderBookingCreateForm(w http.ResponseWriter, r *http.Request) {
	// This will be implemented with HTML templates
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Booking creation form - to be implemented with HTML templates",
	})
}

// renderAdminBookingList renders the admin booking list page
func (h *Handler) renderAdminBookingList(w http.ResponseWriter, r *http.Request, bookings []*Booking) {
	// This will be implemented with HTML templates
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bookings)
}

// renderPendingBookingList renders the pending booking list page
func (h *Handler) renderPendingBookingList(w http.ResponseWriter, r *http.Request, bookings []*Booking) {
	// This will be implemented with HTML templates
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bookings)
}