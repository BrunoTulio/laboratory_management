package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"sync"

	"laboratory-management/config"
)

// CSRFMiddleware provides CSRF protection
type CSRFMiddleware struct {
	tokens map[string]bool
	mu     sync.RWMutex
	secret string
}

var defaultCSRF = NewCSRFMiddleware(config.Load().CSRFSecret)

// GenerateCSRFToken generates a CSRF token using the shared middleware store.
func GenerateCSRFToken() string {
	token, err := defaultCSRF.GenerateToken()
	if err != nil {
		return ""
	}
	return token
}

// ValidateCSRFToken validates a CSRF token against the shared middleware store.
func ValidateCSRFToken(token string) bool {
	return defaultCSRF.ValidateToken(token)
}

// NewCSRFMiddleware creates a new CSRF middleware
func NewCSRFMiddleware(secret string) *CSRFMiddleware {
	return &CSRFMiddleware{
		tokens: make(map[string]bool),
		secret: secret,
	}
}

// GenerateToken generates a new CSRF token
func (m *CSRFMiddleware) GenerateToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	token := base64.URLEncoding.EncodeToString(b)
	
	m.mu.Lock()
	m.tokens[token] = true
	m.mu.Unlock()
	
	return token, nil
}

// ValidateToken validates a CSRF token
func (m *CSRFMiddleware) ValidateToken(token string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.tokens[token]
}

// CSRF is a middleware that adds CSRF protection
func CSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip CSRF for GET, HEAD, OPTIONS, TRACE
		if r.Method == http.MethodGet || r.Method == http.MethodHead || 
		   r.Method == http.MethodOptions || r.Method == http.MethodTrace {
			next.ServeHTTP(w, r)
			return
		}

		// Check CSRF token for POST, PUT, DELETE, etc.
		token := r.FormValue("csrf_token")
		if token == "" {
			token = r.Header.Get("X-CSRF-Token")
		}

		if token == "" || !defaultCSRF.ValidateToken(token) {
			http.Error(w, "Invalid or missing CSRF token", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}