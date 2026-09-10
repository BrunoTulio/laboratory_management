package middleware

import (
	"context"
	"net/http"
	"strings"
)

// Context keys
type contextKey string

const (
	UserIDKey    contextKey = "userID"
	UserRoleKey  contextKey = "userRole"
	UserEmailKey contextKey = "userEmail"
)

// User role constants (defined here to avoid circular dependency)
const (
	RoleProfessor = "PROFESSOR"
	RoleStudent   = "ALUNO"
	RoleAdmin     = "DIRECAO"
)

// AuthMiddleware checks if user is authenticated
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check for session cookie
		cookie, err := r.Cookie("session_token")
		if err != nil {
			if err == http.ErrNoCookie {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Validate session (this is a simplified version)
		// In production, you would validate against a session store
		if cookie.Value == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// For now, we'll parse user info from the cookie
		// In production, this should come from a validated session
		// This is a placeholder - proper session management needed
		userID, userEmail, userRole := parseSessionCookie(cookie.Value)

		if userID == 0 {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Add user info to context
		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		ctx = context.WithValue(ctx, UserEmailKey, userEmail)
		ctx = context.WithValue(ctx, UserRoleKey, userRole)

		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// RequireRole checks if user has the required role
func RequireRole(allowedRoles ...string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			userRole := r.Context().Value(UserRoleKey)
			if userRole == nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			roleStr := userRole.(string)
			allowed := false
			for _, allowedRole := range allowedRoles {
				if roleStr == allowedRole {
					allowed = true
					break
				}
			}

			if !allowed {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		}
	}
}

// RequireAdmin checks if user is an admin
func RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return RequireRole(RoleAdmin)(next)
}

// RequireTeacherOrAdmin checks if user is a teacher or admin
func RequireTeacherOrAdmin(next http.HandlerFunc) http.HandlerFunc {
	return RequireRole(RoleProfessor, RoleAdmin)(next)
}

// parseSessionCookie parses session info from cookie (placeholder)
// In production, this should validate against a session store
func parseSessionCookie(cookieValue string) (int, string, string) {
	// This is a simplified placeholder
	// In production, implement proper session validation
	parts := strings.Split(cookieValue, "|")
	if len(parts) >= 3 {
		// format: userID|email|role
		// This is NOT secure - just for development
		return 1, parts[1], parts[2] // placeholder
	}
	return 0, "", ""
}

// GetUserIDFromContext retrieves user ID from request context
func GetUserIDFromContext(r *http.Request) int {
	if userID := r.Context().Value(UserIDKey); userID != nil {
		return userID.(int)
	}
	return 0
}

// GetUserRoleFromContext retrieves user role from request context
func GetUserRoleFromContext(r *http.Request) string {
	if userRole := r.Context().Value(UserRoleKey); userRole != nil {
		return userRole.(string)
	}
	return ""
}

// GetUserEmailFromContext retrieves user email from request context
func GetUserEmailFromContext(r *http.Request) string {
	if userEmail := r.Context().Value(UserEmailKey); userEmail != nil {
		return userEmail.(string)
	}
	return ""
}