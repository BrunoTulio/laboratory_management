package middleware

import (
	"net/http"
)

// Session middleware handles session management
func Session(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for existing session
		cookie, err := r.Cookie("session_token")
		if err == nil && cookie.Value != "" {
			// Session exists, could validate here
			// For now, we'll let the auth middleware handle validation
		}

		next.ServeHTTP(w, r)
	})
}

// SetSessionCookie sets a session cookie
func SetSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		MaxAge:   86400, // 24 hours
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	})
}

// ClearSessionCookie clears a session cookie
func ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
}