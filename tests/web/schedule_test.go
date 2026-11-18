package web_test

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"laboratory-management/config"
	"laboratory-management/internal/audit"
	"laboratory-management/internal/auth"
	"laboratory-management/internal/booking"
	"laboratory-management/internal/middleware"
	"laboratory-management/internal/notification"
	"laboratory-management/internal/user"
	"laboratory-management/internal/web"

	_ "modernc.org/sqlite"
)

func TestBuildShiftSummariesEmpty(t *testing.T) {
	summaries, booked, free := web.BuildShiftSummaries(nil, "", "", "", "")

	if booked != 0 {
		t.Errorf("expected 0 booked slots, got %d", booked)
	}
	if free != 16 {
		t.Errorf("expected 16 free slots, got %d", free)
	}
	if len(summaries) != 3 {
		t.Fatalf("expected 3 shift summaries, got %d", len(summaries))
	}

	if summaries[0].FreeCount != 5 || summaries[1].FreeCount != 5 || summaries[2].FreeCount != 6 {
		t.Errorf("unexpected shift free counts: morning=%d afternoon=%d night=%d",
			summaries[0].FreeCount, summaries[1].FreeCount, summaries[2].FreeCount)
	}
}

func TestBuildShiftSummariesWithBookingsAndFilters(t *testing.T) {
	classID := 10
	start1, _ := time.Parse("15:04", "13:00")
	end1, _ := time.Parse("15:04", "14:30")

	start2, _ := time.Parse("15:04", "18:50")
	end2, _ := time.Parse("15:04", "19:20")

	bookings := []*booking.Booking{
		{
			ID:          1,
			ClassID:     &classID,
			StartTime:   start1,
			EndTime:     end1,
			TeacherName: "Camila",
			SubjectName: "Programação",
		},
		{
			ID:          2,
			StartTime:   start2,
			EndTime:     end2,
			TeacherName: "Ana Souza",
			SubjectName: "Redes",
		},
	}

	// Unfiltered: 13:00-14:30 occupies 2 afternoon slots, 18:50-19:20 occupies 1 night slot -> 3 booked, 13 free
	summaries, booked, free := web.BuildShiftSummaries(bookings, "", "", "", "")
	if booked != 3 || free != 13 {
		t.Fatalf("expected 3 booked and 13 free, got booked=%d free=%d", booked, free)
	}
	if summaries[1].BookedCount != 2 || summaries[2].BookedCount != 1 {
		t.Errorf("unexpected shift counts: afternoon=%d night=%d", summaries[1].BookedCount, summaries[2].BookedCount)
	}

	// Filter by teacher "Camila" + classID "10" + subject "Programação"
	filtered, bookedFiltered, freeFiltered := web.BuildShiftSummaries(bookings, "VESPERTINO", "Camila", "10", "Programação")
	if bookedFiltered != 2 || freeFiltered != 14 {
		t.Errorf("expected 2 booked and 14 free with teacher filter, got booked=%d free=%d", bookedFiltered, freeFiltered)
	}
	if filtered[0].ShowInFilter || !filtered[1].ShowInFilter || filtered[2].ShowInFilter {
		t.Errorf("expected only VESPERTINO shift to have ShowInFilter=true")
	}
}

func TestBuildWeekDates(t *testing.T) {
	// Tuesday 2026-11-17 -> Monday 2026-11-16 through Saturday 2026-11-21
	tuesday := time.Date(2026, 11, 17, 10, 0, 0, 0, time.UTC)
	days := web.BuildWeekDates(tuesday)

	if len(days) != 6 {
		t.Fatalf("expected 6 days in week, got %d", len(days))
	}
	if days[0].Format("2006-01-02") != "2026-11-16" {
		t.Errorf("expected Monday 2026-11-16, got %s", days[0].Format("2006-01-02"))
	}
	if days[5].Format("2006-01-02") != "2026-11-21" {
		t.Errorf("expected Saturday 2026-11-21, got %s", days[5].Format("2006-01-02"))
	}

	// Sunday 2026-11-22 -> still belongs to Monday 2026-11-16..2026-11-21
	sunday := time.Date(2026, 11, 22, 10, 0, 0, 0, time.UTC)
	sundayWeek := web.BuildWeekDates(sunday)
	if sundayWeek[0].Format("2006-01-02") != "2026-11-16" {
		t.Errorf("expected Sunday to resolve to Monday 2026-11-16, got %s", sundayWeek[0].Format("2006-01-02"))
	}
}

func TestFormatDatePTBR(t *testing.T) {
	date := time.Date(2026, 11, 17, 0, 0, 0, 0, time.UTC)
	got := web.FormatDatePTBR(date)
	want := "ter., 17 de nov. de 2026"
	if got != want {
		t.Errorf("FormatDatePTBR() = %q, want %q", got, want)
	}
}

func TestRouterTemplatesAndGuards(t *testing.T) {
	t.Chdir("../..")

	router := web.NewRouter(nil)
	router.Use(middleware.Session)
	router.Use(middleware.CSRF)

	// 1. GET /login renders HTML with CSRF token
	reqLogin := httptest.NewRequest(http.MethodGet, "/login", nil)
	recLogin := httptest.NewRecorder()
	router.ServeHTTP(recLogin, reqLogin)
	if recLogin.Code != http.StatusOK || !strings.Contains(recLogin.Body.String(), "csrf_token") {
		t.Fatalf("expected GET /login to render HTML with csrf_token, got status=%d", recLogin.Code)
	}

	// 2. GET /cadastro is disabled (no public registration)
	reqReg := httptest.NewRequest(http.MethodGet, "/cadastro", nil)
	recReg := httptest.NewRecorder()
	router.ServeHTTP(recReg, reqReg)
	if recReg.Code != http.StatusOK || !strings.Contains(recReg.Body.String(), "Cadastro público desativado") {
		t.Fatalf("expected GET /cadastro to show disabled notice, got status=%d", recReg.Code)
	}

	// 3. GET /admin without session -> redirect to /login
	reqAdminAnon := httptest.NewRequest(http.MethodGet, "/admin", nil)
	recAdminAnon := httptest.NewRecorder()
	router.ServeHTTP(recAdminAnon, reqAdminAnon)
	if recAdminAnon.Code != http.StatusSeeOther || recAdminAnon.Header().Get("Location") != "/login" {
		t.Fatalf("expected anonymous GET /admin to redirect to /login, got %d", recAdminAnon.Code)
	}

	// 4. GET /admin and /cadastros with PROFESSOR role -> 403 Forbidden
	profToken := middleware.SignSessionToken(10, "prof@cetep.ba.gov.br", middleware.RoleProfessor, time.Hour, config.Load().SessionSecret)
	reqAdminProf := httptest.NewRequest(http.MethodGet, "/admin", nil)
	reqAdminProf.AddCookie(&http.Cookie{Name: "session_token", Value: profToken})
	recAdminProf := httptest.NewRecorder()
	router.ServeHTTP(recAdminProf, reqAdminProf)
	if recAdminProf.Code != http.StatusForbidden {
		t.Fatalf("expected PROFESSOR GET /admin to return 403, got %d", recAdminProf.Code)
	}

	reqCadProf := httptest.NewRequest(http.MethodGet, "/cadastros", nil)
	reqCadProf.AddCookie(&http.Cookie{Name: "session_token", Value: profToken})
	recCadProf := httptest.NewRecorder()
	router.ServeHTTP(recCadProf, reqCadProf)
	if recCadProf.Code != http.StatusForbidden {
		t.Fatalf("expected PROFESSOR GET /cadastros to return 403, got %d", recCadProf.Code)
	}
}

func TestEmbeddedSQLiteDefaultAdminAndInviteFlow(t *testing.T) {
	t.Chdir("../..")

	dbPath := filepath.Join(t.TempDir(), "test_lab.db")
	t.Setenv("DATABASE_PATH", dbPath)
	t.Setenv("APP_BASE_URL", "http://localhost:8088")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	defer db.Close()

	// Run schema migrations
	for _, mig := range []string{"migrations/001_initial_schema.sql", "migrations/002_school_registry_and_subjects.sql"} {
		sqlBytes, err := os.ReadFile(mig)
		if err != nil {
			t.Fatalf("failed to read %s: %v", mig, err)
		}
		if _, err := db.Exec(string(sqlBytes)); err != nil {
			t.Fatalf("failed to exec %s: %v", mig, err)
		}
	}

	userRepo := user.NewRepository(db)
	authSvc := auth.NewService(userRepo)
	cfg := config.Load()

	// 1. Seed default admin
	if err := authSvc.EnsureDefaultAdmin(cfg); err != nil {
		t.Fatalf("EnsureDefaultAdmin failed: %v", err)
	}
	adminUser, err := authSvc.Login(cfg.AdminEmail, cfg.AdminPassword)
	if err != nil || adminUser.Role == nil || adminUser.Role.Name != user.RoleAdmin {
		t.Fatalf("expected default admin login to succeed with DIRECAO role, err=%v", err)
	}

	// 2. Invite new user (pending until password is set)
	invited, activationURL, err := authSvc.InviteUser("Camila", "camila@cetep.ba.gov.br", user.RoleProfessor, "", "")
	if err != nil {
		t.Fatalf("InviteUser failed: %v", err)
	}
	if invited.Status != user.StatusPending {
		t.Fatalf("expected invited user status PENDING, got %s", invited.Status)
	}

	// Pending user cannot log in before setting password
	if _, err := authSvc.Login("camila@cetep.ba.gov.br", "MinhaSenha123"); err == nil {
		t.Fatalf("expected login to fail before activation")
	}

	// 3. Extract token from activation URL and test HTTP GET /ativar-conta + POST /ativar-conta
	parsedURL, err := url.Parse(activationURL)
	if err != nil {
		t.Fatalf("invalid activation URL %q: %v", activationURL, err)
	}
	token := parsedURL.Query().Get("token")

	router := web.NewRouter(db)
	router.Use(middleware.Session)
	router.Use(middleware.CSRF)

	reqActGet := httptest.NewRequest(http.MethodGet, "/ativar-conta?token="+url.QueryEscape(token), nil)
	recActGet := httptest.NewRecorder()
	router.ServeHTTP(recActGet, reqActGet)
	if recActGet.Code != http.StatusOK || !strings.Contains(recActGet.Body.String(), "Turmas que você leciona") {
		t.Fatalf("expected GET /ativar-conta to render professor classes/subjects fields, got status=%d", recActGet.Code)
	}

	csrfToken := middleware.GenerateCSRFToken()
	actForm := url.Values{
		"csrf_token":       {csrfToken},
		"token":            {token},
		"classes_csv":      {"2º ano · Informática, 3º ano · Redes"},
		"subjects_csv":     {"Programação, Banco de Dados"},
		"password":         {"MinhaSenha123"},
		"confirm_password": {"MinhaSenha123"},
	}
	reqActPost := httptest.NewRequest(http.MethodPost, "/ativar-conta", strings.NewReader(actForm.Encode()))
	reqActPost.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recActPost := httptest.NewRecorder()
	router.ServeHTTP(recActPost, reqActPost)
	if recActPost.Code != http.StatusOK || !strings.Contains(recActPost.Body.String(), "Conta ativada com sucesso") {
		t.Fatalf("expected POST /ativar-conta to activate account, got status=%d", recActPost.Code)
	}

	// 4. Activated user can log in with chosen password and has configured classes/subjects
	activatedUser, err := authSvc.Login("camila@cetep.ba.gov.br", "MinhaSenha123")
	if err != nil || activatedUser.Status != user.StatusActive {
		t.Fatalf("expected activated user to log in, err=%v", err)
	}
	if len(activatedUser.ClassesList()) != 2 || len(activatedUser.SubjectsList()) != 2 {
		t.Fatalf("expected 2 classes and 2 subjects saved for professor, got classes=%v subjects=%v",
			activatedUser.ClassesList(), activatedUser.SubjectsList())
	}

	// 5. Deactivate & Reactivate user, while protecting primary admin
	userSvc := user.NewService(userRepo, audit.NewRepository(db))
	if err := userSvc.SetUserStatus(adminUser.ID, user.StatusInactive, adminUser.ID, cfg.AdminEmail); err == nil {
		t.Fatalf("expected primary admin deactivation to be blocked")
	}
	if err := userSvc.DeleteUserSafe(adminUser.ID, adminUser.ID, cfg.AdminEmail); err == nil {
		t.Fatalf("expected primary admin deletion to be blocked")
	}

	if err := userSvc.SetUserStatus(activatedUser.ID, user.StatusInactive, adminUser.ID, cfg.AdminEmail); err != nil {
		t.Fatalf("failed to deactivate user: %v", err)
	}
	if _, err := authSvc.Login("camila@cetep.ba.gov.br", "MinhaSenha123"); err == nil {
		t.Fatalf("expected login to fail when user is INACTIVE")
	}
	if err := userSvc.SetUserStatus(activatedUser.ID, user.StatusActive, adminUser.ID, cfg.AdminEmail); err != nil {
		t.Fatalf("failed to reactivate user: %v", err)
	}

	// 6. Create booking for user and verify DeleteUserSafe removes user + bookings
	bookingSvc := booking.NewService(booking.NewRepository(db), audit.NewRepository(db), notification.NewRepository(db))
	sTime, _ := time.Parse("15:04", "07:30")
	eTime, _ := time.Parse("15:04", "08:15")
	if _, err := bookingSvc.CreateBooking(&booking.CreateBookingRequest{
		LaboratoryID: 1,
		BookingDate:  time.Now(),
		StartTime:    sTime,
		EndTime:      eTime,
		Shift:        "MATUTINO",
		Reason:       "Aula teste",
	}, activatedUser.ID); err != nil {
		t.Fatalf("failed to create booking: %v", err)
	}

	if err := userSvc.DeleteUserSafe(activatedUser.ID, adminUser.ID, cfg.AdminEmail); err != nil {
		t.Fatalf("DeleteUserSafe failed: %v", err)
	}
	remaining, err := bookingSvc.GetBookingsByUserID(activatedUser.ID)
	if err != nil || len(remaining) != 0 {
		t.Fatalf("expected bookings to be removed on user delete, got len=%d err=%v", len(remaining), err)
	}
}

func TestResolveLessonRange(t *testing.T) {
	// Manhã: 1ª e 2ª aula -> 07:30 às 09:00
	s1, e1, err := web.ResolveLessonRange("MATUTINO", []string{"1", "2"})
	if err != nil || s1.Format("15:04") != "07:30" || e1.Format("15:04") != "09:00" {
		t.Fatalf("expected 07:30-09:00 for MATUTINO 1,2, got %s-%s err=%v", s1.Format("15:04"), e1.Format("15:04"), err)
	}

	// Noite: 6ª aula -> 21:20 às 21:50
	s2, e2, err := web.ResolveLessonRange("NOTURNO", []string{"6"})
	if err != nil || s2.Format("15:04") != "21:20" || e2.Format("15:04") != "21:50" {
		t.Fatalf("expected 21:20-21:50 for NOTURNO 6, got %s-%s err=%v", s2.Format("15:04"), e2.Format("15:04"), err)
	}

	// Sem aulas selecionadas -> erro
	if _, _, err := web.ResolveLessonRange("VESPERTINO", nil); err == nil {
		t.Fatalf("expected error when no lessons are selected")
	}
}
