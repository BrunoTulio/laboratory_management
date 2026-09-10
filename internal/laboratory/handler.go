package laboratory

import (
	"encoding/json"
	"net/http"
	"strconv"

	"laboratory-management/internal/middleware"
)

// Handler handles HTTP requests for laboratories
type Handler struct {
	labService *Service
}

// NewHandler creates a new laboratory handler
func NewHandler(labService *Service) *Handler {
	return &Handler{
		labService: labService,
	}
}

// ListHandler handles listing all laboratories
func (h *Handler) ListHandler(w http.ResponseWriter, r *http.Request) {
	labs, err := h.labService.GetAllLaboratories()
	if err != nil {
		http.Error(w, "Error fetching laboratories", http.StatusInternalServerError)
		return
	}

	// Render laboratory list
	h.renderLaboratoryList(w, r, labs)
}

// DetailHandler handles viewing a laboratory detail
func (h *Handler) DetailHandler(w http.ResponseWriter, r *http.Request) {
	labIDStr := r.URL.Query().Get("id")
	labID, err := strconv.Atoi(labIDStr)
	if err != nil {
		http.Error(w, "Invalid laboratory ID", http.StatusBadRequest)
		return
	}

	lab, err := h.labService.GetLaboratoryByID(labID)
	if err != nil {
		http.Error(w, "Laboratory not found", http.StatusNotFound)
		return
	}

	// Render laboratory detail
	h.renderLaboratoryDetail(w, r, lab)
}

// AdminLabListHandler handles listing all laboratories (admin)
func (h *Handler) AdminLabListHandler(w http.ResponseWriter, r *http.Request) {
	labs, err := h.labService.GetAllLaboratories()
	if err != nil {
		http.Error(w, "Error fetching laboratories", http.StatusInternalServerError)
		return
	}

	// Render admin laboratory list
	h.renderAdminLabList(w, r, labs)
}

// AdminLabCreateHandler handles creating a laboratory (admin)
func (h *Handler) AdminLabCreateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		h.renderLabCreateForm(w, r)
		return
	}

	if r.Method == http.MethodPost {
		var req CreateLaboratoryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		creatorID := middleware.GetUserIDFromContext(r)
		lab, err := h.labService.CreateLaboratory(&req, creatorID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(lab)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// AdminLabEditHandler handles editing a laboratory (admin)
func (h *Handler) AdminLabEditHandler(w http.ResponseWriter, r *http.Request) {
	labIDStr := r.URL.Query().Get("id")
	labID, err := strconv.Atoi(labIDStr)
	if err != nil {
		http.Error(w, "Invalid laboratory ID", http.StatusBadRequest)
		return
	}

	if r.Method == http.MethodGet {
		lab, err := h.labService.GetLaboratoryByID(labID)
		if err != nil {
			http.Error(w, "Laboratory not found", http.StatusNotFound)
			return
		}
		h.renderLabEditForm(w, r, lab)
		return
	}

	if r.Method == http.MethodPost {
		var req UpdateLaboratoryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		updaterID := middleware.GetUserIDFromContext(r)
		lab, err := h.labService.UpdateLaboratory(labID, &req, updaterID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(lab)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// renderLaboratoryList renders the laboratory list page
func (h *Handler) renderLaboratoryList(w http.ResponseWriter, r *http.Request, labs []*Laboratory) {
	// This will be implemented with HTML templates
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(labs)
}

// renderLaboratoryDetail renders the laboratory detail page
func (h *Handler) renderLaboratoryDetail(w http.ResponseWriter, r *http.Request, lab *Laboratory) {
	// This will be implemented with HTML templates
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lab)
}

// renderAdminLabList renders the admin laboratory list page
func (h *Handler) renderAdminLabList(w http.ResponseWriter, r *http.Request, labs []*Laboratory) {
	// This will be implemented with HTML templates
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(labs)
}

// renderLabCreateForm renders the laboratory creation form
func (h *Handler) renderLabCreateForm(w http.ResponseWriter, r *http.Request) {
	// This will be implemented with HTML templates
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Laboratory creation form - to be implemented with HTML templates",
	})
}

// renderLabEditForm renders the laboratory edit form
func (h *Handler) renderLabEditForm(w http.ResponseWriter, r *http.Request, lab *Laboratory) {
	// This will be implemented with HTML templates
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lab)
}