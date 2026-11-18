package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"laboratory-management/config"
	"laboratory-management/internal/middleware"
)

func TestSessionTokenSignAndVerify(t *testing.T) {
	secret := "test-secret-key"
	token := middleware.SignSessionToken(42, "camila@cetep.ba.gov.br", middleware.RoleProfessor, time.Hour, secret)

	userID, email, role, ok := middleware.VerifySessionToken(token, secret)
	if !ok {
		t.Fatalf("expected valid token to verify")
	}
	if userID != 42 || email != "camila@cetep.ba.gov.br" || role != middleware.RoleProfessor {
		t.Fatalf("unexpected claims: id=%d email=%s role=%s", userID, email, role)
	}

	// Tampered token
	if _, _, _, ok := middleware.VerifySessionToken(token+"tampered", secret); ok {
		t.Errorf("expected tampered token to fail verification")
	}

	// Wrong secret
	if _, _, _, ok := middleware.VerifySessionToken(token, "wrong-secret"); ok {
		t.Errorf("expected token signed with different secret to fail")
	}

	// Expired token
	expired := middleware.SignSessionToken(42, "camila@cetep.ba.gov.br", middleware.RoleProfessor, -time.Minute, secret)
	if _, _, _, ok := middleware.VerifySessionToken(expired, secret); ok {
		t.Errorf("expected expired token to fail verification")
	}

	// Malformed token
	if _, _, _, ok := middleware.VerifySessionToken("not-a-valid-token", secret); ok {
		t.Errorf("expected malformed token to fail verification")
	}
}

func TestAuthMiddleware(t *testing.T) {
	protected := middleware.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		id := middleware.GetUserIDFromContext(r)
		email := middleware.GetUserEmailFromContext(r)
		role := middleware.GetUserRoleFromContext(r)
		if id != 7 || email != "ana@cetep.ba.gov.br" || role != middleware.RoleAdmin {
			t.Errorf("unexpected context values: id=%d email=%s role=%s", id, email, role)
		}
		w.WriteHeader(http.StatusOK)
	})

	// 1. Missing cookie -> redirect to /login
	reqNoCookie := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	recNoCookie := httptest.NewRecorder()
	protected.ServeHTTP(recNoCookie, reqNoCookie)
	if recNoCookie.Code != http.StatusSeeOther || recNoCookie.Header().Get("Location") != "/login" {
		t.Errorf("expected redirect to /login without cookie, got status=%d location=%s", recNoCookie.Code, recNoCookie.Header().Get("Location"))
	}

	// 2. Invalid cookie -> redirect to /login
	reqBadCookie := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	reqBadCookie.AddCookie(&http.Cookie{Name: "session_token", Value: "forged.token"})
	recBadCookie := httptest.NewRecorder()
	protected.ServeHTTP(recBadCookie, reqBadCookie)
	if recBadCookie.Code != http.StatusSeeOther {
		t.Errorf("expected redirect for invalid cookie, got %d", recBadCookie.Code)
	}

	// 3. Valid signed cookie -> 200 OK with context populated
	token := middleware.SignSessionToken(7, "ana@cetep.ba.gov.br", middleware.RoleAdmin, time.Hour, config.Load().SessionSecret)
	reqValid := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	reqValid.AddCookie(&http.Cookie{Name: "session_token", Value: token})
	recValid := httptest.NewRecorder()
	protected.ServeHTTP(recValid, reqValid)
	if recValid.Code != http.StatusOK {
		t.Errorf("expected 200 OK for valid cookie, got %d", recValid.Code)
	}
}

func TestCSRFMiddleware(t *testing.T) {
	handler := middleware.CSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// GET allowed without token
	reqGet := httptest.NewRequest(http.MethodGet, "/cadastros", nil)
	recGet := httptest.NewRecorder()
	handler.ServeHTTP(recGet, reqGet)
	if recGet.Code != http.StatusOK {
		t.Errorf("expected GET to pass CSRF, got %d", recGet.Code)
	}

	// POST without token -> 403
	reqPostEmpty := httptest.NewRequest(http.MethodPost, "/cadastros", nil)
	recPostEmpty := httptest.NewRecorder()
	handler.ServeHTTP(recPostEmpty, reqPostEmpty)
	if recPostEmpty.Code != http.StatusForbidden {
		t.Errorf("expected 403 for POST without CSRF token, got %d", recPostEmpty.Code)
	}

	// POST with forged token -> 403
	formForged := url.Values{"csrf_token": {"forged-token"}}
	reqPostForged := httptest.NewRequest(http.MethodPost, "/cadastros", strings.NewReader(formForged.Encode()))
	reqPostForged.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recPostForged := httptest.NewRecorder()
	handler.ServeHTTP(recPostForged, reqPostForged)
	if recPostForged.Code != http.StatusForbidden {
		t.Errorf("expected 403 for forged CSRF token, got %d", recPostForged.Code)
	}

	// POST with valid token via form -> 200
	validToken := middleware.GenerateCSRFToken()
	formValid := url.Values{"csrf_token": {validToken}}
	reqPostValid := httptest.NewRequest(http.MethodPost, "/cadastros", strings.NewReader(formValid.Encode()))
	reqPostValid.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recPostValid := httptest.NewRecorder()
	handler.ServeHTTP(recPostValid, reqPostValid)
	if recPostValid.Code != http.StatusOK {
		t.Errorf("expected 200 for valid form CSRF token, got %d", recPostValid.Code)
	}

	// POST with valid token via X-CSRF-Token header -> 200
	reqHeaderValid := httptest.NewRequest(http.MethodPost, "/cadastros", nil)
	reqHeaderValid.Header.Set("X-CSRF-Token", validToken)
	recHeaderValid := httptest.NewRecorder()
	handler.ServeHTTP(recHeaderValid, reqHeaderValid)
	if recHeaderValid.Code != http.StatusOK {
		t.Errorf("expected 200 for valid header CSRF token, got %d", recHeaderValid.Code)
	}
}
