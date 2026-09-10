package class

import (
	"encoding/json"
	"net/http"
	"strconv"

	"laboratory-management/internal/middleware"
)

// Handler handles HTTP requests for classes
type Handler struct {
	classService *Service
}

// NewHandler creates a new class handler
func NewHandler(classService *Service) *Handler {
	return &Handler{
		classService: classService,
	}
}

// ListHandler handles listing all classes
func (h *Handler) ListHandler(w http.ResponseWriter, r *http.Request) {
	classes, err := h.classService.GetAllClasses()
	if err != nil {
		http.Error(w, "Error fetching classes", http.StatusInternalServerError)
		return
	}

	// Render class list
	h.renderClassList(w, r, classes)
}

// DetailHandler handles viewing a class detail
func (h *Handler) DetailHandler(w http.ResponseWriter, r *http.Request) {
	classIDStr := r.URL.Query().Get("id")
	classID, err := strconv.Atoi(classIDStr)
	if err != nil {
		http.Error(w, "Invalid class ID", http.StatusBadRequest)
		return
	}

	class, err := h.classService.GetClassByID(classID)
	if err != nil {
		http.Error(w, "Class not found", http.StatusNotFound)
		return
	}

	// Render class detail
	h.renderClassDetail(w, r, class)
}

// AdminClassListHandler handles listing all classes (admin)
func (h *Handler) AdminClassListHandler(w http.ResponseWriter, r *http.Request) {
	classes, err := h.classService.GetAllClasses()
	if err != nil {
		http.Error(w, "Error fetching classes", http.StatusInternalServerError)
		return
	}

	// Render admin class list
	h.renderAdminClassList(w, r, classes)
}

// AdminClassCreateHandler handles creating a class (admin)
func (h *Handler) AdminClassCreateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		h.renderClassCreateForm(w, r)
		return
	}

	if r.Method == http.MethodPost {
		var req CreateClassRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		creatorID := middleware.GetUserIDFromContext(r)
		class, err := h.classService.CreateClass(&req, creatorID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(class)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// AdminClassEditHandler handles editing a class (admin)
func (h *Handler) AdminClassEditHandler(w http.ResponseWriter, r *http.Request) {
	classIDStr := r.URL.Query().Get("id")
	classID, err := strconv.Atoi(classIDStr)
	if err != nil {
		http.Error(w, "Invalid class ID", http.StatusBadRequest)
		return
	}

	if r.Method == http.MethodGet {
		class, err := h.classService.GetClassByID(classID)
		if err != nil {
			http.Error(w, "Class not found", http.StatusNotFound)
			return
		}
		h.renderClassEditForm(w, r, class)
		return
	}

	if r.Method == http.MethodPost {
		var req UpdateClassRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		updaterID := middleware.GetUserIDFromContext(r)
		class, err := h.classService.UpdateClass(classID, &req, updaterID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(class)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// renderClassList renders the class list page
func (h *Handler) renderClassList(w http.ResponseWriter, r *http.Request, classes []*Class) {
	// This will be implemented with HTML templates
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(classes)
}

// renderClassDetail renders the class detail page
func (h *Handler) renderClassDetail(w http.ResponseWriter, r *http.Request, class *Class) {
	// This will be implemented with HTML templates
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(class)
}

// renderAdminClassList renders the admin class list page
func (h *Handler) renderAdminClassList(w http.ResponseWriter, r *http.Request, classes []*Class) {
	// This will be implemented with HTML templates
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(classes)
}

// renderClassCreateForm renders the class creation form
func (h *Handler) renderClassCreateForm(w http.ResponseWriter, r *http.Request) {
	// This will be implemented with HTML templates
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Class creation form - to be implemented with HTML templates",
	})
}

// renderClassEditForm renders the class edit form
func (h *Handler) renderClassEditForm(w http.ResponseWriter, r *http.Request, class *Class) {
	// This will be implemented with HTML templates
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(class)
}