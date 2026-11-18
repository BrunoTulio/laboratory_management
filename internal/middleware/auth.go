package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"laboratory-management/config"
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

		userID, userEmail, userRole, ok := VerifySessionToken(cookie.Value, config.Load().SessionSecret)

		if !ok || userID == 0 {
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

// SignSessionToken creates an HMAC-SHA256 signed session token containing user identity and expiration.
func SignSessionToken(userID int, email, role string, ttl time.Duration, secret string) string {
	exp := time.Now().Add(ttl).Unix()
	payload := fmt.Sprintf("%d|%s|%s|%d", userID, email, role, exp)
	encoded := base64.RawURLEncoding.EncodeToString([]byte(payload))

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(encoded))
	sig := hex.EncodeToString(mac.Sum(nil))

	return encoded + "." + sig
}

// VerifySessionToken validates the HMAC signature and expiration of a session token.
func VerifySessionToken(token, secret string) (int, string, string, bool) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return 0, "", "", false
	}

	encoded, sig := parts[0], parts[1]
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(encoded))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(sig), []byte(expectedSig)) {
		return 0, "", "", false
	}

	raw, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return 0, "", "", false
	}

	fields := strings.Split(string(raw), "|")
	if len(fields) != 4 {
		return 0, "", "", false
	}

	userID, err := strconv.Atoi(fields[0])
	if err != nil || userID <= 0 {
		return 0, "", "", false
	}

	exp, err := strconv.ParseInt(fields[3], 10, 64)
	if err != nil || time.Now().Unix() > exp {
		return 0, "", "", false
	}

	return userID, fields[1], fields[2], true
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