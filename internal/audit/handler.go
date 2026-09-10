package audit

import (
	"encoding/json"
	"net/http"
	"strconv"
)

// Handler handles HTTP requests for audit logs
type Handler struct {
	auditService *Service
}

// NewHandler creates a new audit handler
func NewHandler(auditService *Service) *Handler {
	return &Handler{
		auditService: auditService,
	}
}

// ListHandler handles listing all audit logs (admin only)
func (h *Handler) ListHandler(w http.ResponseWriter, r *http.Request) {
	auditLogs, err := h.auditService.GetAllAuditLogs()
	if err != nil {
		http.Error(w, "Error fetching audit logs", http.StatusInternalServerError)
		return
	}

	// Render audit log list
	h.renderAuditLogList(w, r, auditLogs)
}

// UserAuditHandler handles listing audit logs for a specific user
func (h *Handler) UserAuditHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	auditLogs, err := h.auditService.GetAuditLogsByUserID(userID)
	if err != nil {
		http.Error(w, "Error fetching audit logs", http.StatusInternalServerError)
		return
	}

	// Render user audit log list
	h.renderUserAuditLogList(w, r, auditLogs)
}

// EntityAuditHandler handles listing audit logs for a specific entity
func (h *Handler) EntityAuditHandler(w http.ResponseWriter, r *http.Request) {
	entityType := r.URL.Query().Get("entity_type")
	entityIDStr := r.URL.Query().Get("entity_id")
	entityID, err := strconv.Atoi(entityIDStr)
	if err != nil {
		http.Error(w, "Invalid entity ID", http.StatusBadRequest)
		return
	}

	auditLogs, err := h.auditService.GetAuditLogsByEntity(entityType, entityID)
	if err != nil {
		http.Error(w, "Error fetching audit logs", http.StatusInternalServerError)
		return
	}

	// Render entity audit log list
	h.renderEntityAuditLogList(w, r, auditLogs)
}

// renderAuditLogList renders the audit log list page
func (h *Handler) renderAuditLogList(w http.ResponseWriter, r *http.Request, auditLogs []*AuditLog) {
	// This will be implemented with HTML templates
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(auditLogs)
}

// renderUserAuditLogList renders the user audit log list page
func (h *Handler) renderUserAuditLogList(w http.ResponseWriter, r *http.Request, auditLogs []*AuditLog) {
	// This will be implemented with HTML templates
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(auditLogs)
}

// renderEntityAuditLogList renders the entity audit log list page
func (h *Handler) renderEntityAuditLogList(w http.ResponseWriter, r *http.Request, auditLogs []*AuditLog) {
	// This will be implemented with HTML templates
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(auditLogs)
}