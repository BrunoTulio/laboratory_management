package web

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strings"

	"laboratory-management/internal/audit"
	"laboratory-management/internal/auth"
	"laboratory-management/internal/booking"
	"laboratory-management/internal/class"
	"laboratory-management/internal/laboratory"
	"laboratory-management/internal/middleware"
	"laboratory-management/internal/notification"
	"laboratory-management/internal/user"
)

// Router handles HTTP routing
type Router struct {
	mux                 *http.ServeMux
	handler             http.Handler
	middlewares         []func(http.Handler) http.Handler
	db                  *sql.DB
	templates           *template.Template
	authService         *auth.Service
	userHandler         *user.Handler
	labHandler          *laboratory.Handler
	classHandler        *class.Handler
	bookingHandler      *booking.Handler
	notificationHandler *notification.Handler
	auditHandler        *audit.Handler
	authHandler         *auth.Handler
}

// NewRouter creates a new router
func NewRouter(db *sql.DB) *Router {
	// Initialize repositories
	userRepo := user.NewRepository(db)
	labRepo := laboratory.NewRepository(db)
	classRepo := class.NewRepository(db)
	bookingRepo := booking.NewRepository(db)
	notificationRepo := notification.NewRepository(db)
	auditRepo := audit.NewRepository(db)

	// Initialize services
	authService := auth.NewService(userRepo)
	userService := user.NewService(userRepo, auditRepo)
	labService := laboratory.NewService(labRepo, auditRepo)
	classService := class.NewService(classRepo, auditRepo)
	bookingService := booking.NewService(bookingRepo, auditRepo, notificationRepo)
	notificationService := notification.NewService(notificationRepo)
	auditService := audit.NewService(auditRepo)

	// Initialize handlers
	authHandler := auth.NewHandler(userService, authService)
	userHandler := user.NewHandler(userService)
	labHandler := laboratory.NewHandler(labService)
	classHandler := class.NewHandler(classService)
	bookingHandler := booking.NewHandler(bookingService)
	notificationHandler := notification.NewHandler(notificationService)
	auditHandler := audit.NewHandler(auditService)

	// Load templates
	templates := template.Must(
		template.New("").Funcs(template.FuncMap{
			"lower": strings.ToLower,
		}).ParseGlob("templates/**/*.html"),
	)

	mux := http.NewServeMux()

	router := &Router{
		mux:                 mux,
		handler:             mux,
		db:                  db,
		templates:           templates,
		authService:         authService,
		userHandler:         userHandler,
		labHandler:          labHandler,
		classHandler:        classHandler,
		bookingHandler:      bookingHandler,
		notificationHandler: notificationHandler,
		auditHandler:        auditHandler,
		authHandler:         authHandler,
	}

	// Setup routes
	router.setupRoutes()

	return router
}

// Use registra um middleware para ser aplicado a todas as requisições.
// Os middlewares são executados na ordem em que foram adicionados
// (o primeiro Use() é o mais externo).
func (r *Router) Use(mw func(http.Handler) http.Handler) {
	r.middlewares = append(r.middlewares, mw)
	r.rebuildHandler()
}

// ServeHTTP implements http.Handler interface
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.handler.ServeHTTP(w, req)
}

// rebuildHandler reaplica a cadeia de middlewares sobre o mux.
func (r *Router) rebuildHandler() {
	var h http.Handler = r.mux
	// aplica em ordem reversa para que o primeiro middleware adicionado
	// seja o mais externo da cadeia
	for i := len(r.middlewares) - 1; i >= 0; i-- {
		h = r.middlewares[i](h)
	}
	r.handler = h
}

// setupRoutes configures all HTTP routes
func (r *Router) setupRoutes() {
	// Static files
	r.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Authentication routes (no middleware)
	r.mux.HandleFunc("/login", r.authHandler.LoginHandler)
	r.mux.HandleFunc("/cadastro", r.authHandler.RegisterHandler)
	r.mux.HandleFunc("/logout", r.authHandler.LogoutHandler)

	// Protected routes (require authentication)
	r.mux.HandleFunc("/dashboard", middleware.AuthMiddleware(r.dashboardHandler))
	r.mux.HandleFunc("/perfil", middleware.AuthMiddleware(r.authHandler.ProfileHandler))

	// Booking routes
	r.mux.HandleFunc("/agendamentos", middleware.AuthMiddleware(r.bookingHandler.ListHandler))
	r.mux.HandleFunc("/agendamentos/novo", middleware.AuthMiddleware(r.bookingHandler.CreateHandler))
	r.mux.HandleFunc("/agendamentos/", middleware.AuthMiddleware(r.bookingHandler.DetailHandler))
	r.mux.HandleFunc("/disponibilidade", middleware.AuthMiddleware(r.bookingHandler.AvailabilityHandler))

	// Notification routes
	r.mux.HandleFunc("/notificacoes", middleware.AuthMiddleware(r.notificationHandler.ListHandler))
	r.mux.HandleFunc("/notificacoes/nao-lidas", middleware.AuthMiddleware(r.notificationHandler.UnreadCountHandler))
	r.mux.HandleFunc("/notificacoes/marcar-lida", middleware.AuthMiddleware(r.notificationHandler.MarkAsReadHandler))
	r.mux.HandleFunc("/notificacoes/marcar-todas-lidas", middleware.AuthMiddleware(r.notificationHandler.MarkAllAsReadHandler))

	// Agenda routes
	r.mux.HandleFunc("/agenda", middleware.AuthMiddleware(r.agendaHandler))
	r.mux.HandleFunc("/agenda/dia", middleware.AuthMiddleware(r.agendaDayHandler))
	r.mux.HandleFunc("/agenda/semana", middleware.AuthMiddleware(r.agendaWeekHandler))
	r.mux.HandleFunc("/agenda/mes", middleware.AuthMiddleware(r.agendaMonthHandler))

	// History routes
	r.mux.HandleFunc("/historico", middleware.AuthMiddleware(r.historyHandler))

	// Admin routes (require admin role)
	r.mux.HandleFunc("/admin", middleware.RequireAdmin(r.adminDashboardHandler))
	r.mux.HandleFunc("/admin/usuarios", middleware.RequireAdmin(r.userHandler.AdminUserListHandler))
	r.mux.HandleFunc("/admin/usuarios/editar", middleware.RequireAdmin(r.authHandler.AdminUserEditHandler))
	r.mux.HandleFunc("/admin/agendamentos", middleware.RequireAdmin(r.bookingHandler.AdminBookingListHandler))
	r.mux.HandleFunc("/admin/agendamentos/pendentes", middleware.RequireAdmin(r.bookingHandler.AdminPendingListHandler))
	r.mux.HandleFunc("/admin/agendamentos/aprovar", middleware.RequireAdmin(r.bookingHandler.AdminApproveHandler))
	r.mux.HandleFunc("/admin/agendamentos/rejeitar", middleware.RequireAdmin(r.bookingHandler.AdminRejectHandler))
	r.mux.HandleFunc("/admin/agendamentos/cancelar", middleware.RequireAdmin(r.bookingHandler.AdminCancelHandler))
	r.mux.HandleFunc("/admin/historico", middleware.RequireAdmin(r.adminHistoryHandler))
	r.mux.HandleFunc("/admin/auditoria", middleware.RequireAdmin(r.auditHandler.ListHandler))

	// Laboratory management (admin)
	r.mux.HandleFunc("/admin/laboratorios", middleware.RequireAdmin(r.labHandler.AdminLabListHandler))
	r.mux.HandleFunc("/admin/laboratorios/criar", middleware.RequireAdmin(r.labHandler.AdminLabCreateHandler))
	r.mux.HandleFunc("/admin/laboratorios/editar", middleware.RequireAdmin(r.labHandler.AdminLabEditHandler))

	// Class management (admin)
	r.mux.HandleFunc("/admin/turmas", middleware.RequireAdmin(r.classHandler.AdminClassListHandler))
	r.mux.HandleFunc("/admin/turmas/criar", middleware.RequireAdmin(r.classHandler.AdminClassCreateHandler))
	r.mux.HandleFunc("/admin/turmas/editar", middleware.RequireAdmin(r.classHandler.AdminClassEditHandler))

	// Error pages
	r.mux.HandleFunc("/403", r.errorHandler)
	r.mux.HandleFunc("/404", r.notFoundHandler)
	r.mux.HandleFunc("/500", r.serverErrorHandler)

	// Catch-all for 404
	r.mux.HandleFunc("/", r.catchAllHandler)
}

// Handler functions for routes
func (r *Router) dashboardHandler(w http.ResponseWriter, req *http.Request) {
	// Render dashboard page
	data := map[string]interface{}{
		"Title": "Dashboard",
		"User":  getUserFromContext(req),
		"Stats": map[string]int{
			"UpcomingBookings":    0,
			"PendingBookings":     0,
			"UnreadNotifications": 0,
			"TotalBookings":       0,
		},
	}

	r.renderTemplate(w, "dashboard", data)
}

func (r *Router) agendaHandler(w http.ResponseWriter, req *http.Request) {
	data := map[string]interface{}{
		"Title":        "Agenda",
		"User":         getUserFromContext(req),
		"Laboratories": getLaboratories(r),
		"CurrentDate":  getCurrentDate(),
	}

	r.renderTemplate(w, "agenda", data)
}

func (r *Router) agendaDayHandler(w http.ResponseWriter, req *http.Request) {
	// Similar to agendaHandler but for day view
	r.agendaHandler(w, req)
}

func (r *Router) agendaWeekHandler(w http.ResponseWriter, req *http.Request) {
	// Similar to agendaHandler but for week view
	r.agendaHandler(w, req)
}

func (r *Router) agendaMonthHandler(w http.ResponseWriter, req *http.Request) {
	// Similar to agendaHandler but for month view
	r.agendaHandler(w, req)
}

func (r *Router) historyHandler(w http.ResponseWriter, req *http.Request) {
	data := map[string]interface{}{
		"Title":          "Histórico",
		"User":           getUserFromContext(req),
		"FilterDateFrom": "",
		"FilterDateTo":   "",
		"FilterStatus":   "",
		"Bookings":       []interface{}{},
	}

	r.renderTemplate(w, "history", data)
}

func (r *Router) adminDashboardHandler(w http.ResponseWriter, req *http.Request) {
	data := map[string]interface{}{
		"Title": "Dashboard Administrativo",
		"User":  getUserFromContext(req),
		"Stats": map[string]int{
			"TotalUsers":       0,
			"PendingBookings":  0,
			"ApprovedBookings": 0,
			"UpcomingBookings": 0,
		},
		"PendingBookings": []interface{}{},
	}

	r.renderTemplate(w, "admin-dashboard", data)
}

func (r *Router) adminHistoryHandler(w http.ResponseWriter, req *http.Request) {
	data := map[string]interface{}{
		"Title":    "Histórico Administrativo",
		"User":     getUserFromContext(req),
		"Bookings": []interface{}{},
	}

	r.renderTemplate(w, "admin-bookings", data)
}

func (r *Router) errorHandler(w http.ResponseWriter, req *http.Request) {
	data := map[string]interface{}{
		"Title": "Acesso Negado",
	}
	r.renderTemplate(w, "403", data)
}

func (r *Router) notFoundHandler(w http.ResponseWriter, req *http.Request) {
	data := map[string]interface{}{
		"Title": "Página Não Encontrada",
	}
	r.renderTemplate(w, "404", data)
}

func (r *Router) serverErrorHandler(w http.ResponseWriter, req *http.Request) {
	data := map[string]interface{}{
		"Title": "Erro Interno do Servidor",
	}
	r.renderTemplate(w, "500", data)
}

func (r *Router) catchAllHandler(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path == "/" {
		// Root path - redirect to login or dashboard
		http.Redirect(w, req, "/login", http.StatusSeeOther)
		return
	}

	// Any other path - 404
	r.notFoundHandler(w, req)
}

// Helper functions
func (r *Router) renderTemplate(w http.ResponseWriter, templateName string, data interface{}) {
	err := r.templates.ExecuteTemplate(w, templateName+".html", data)
	if err != nil {
		log.Printf("Error rendering template %s: %v", templateName, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func getUserFromContext(req *http.Request) map[string]interface{} {
	return map[string]interface{}{
		"ID":    middleware.GetUserIDFromContext(req),
		"Name":  "Usuário",          // Placeholder
		"Email": "user@example.com", // Placeholder
		"Role":  middleware.GetUserRoleFromContext(req),
	}
}

func getLaboratories(r *Router) []interface{} {
	// Placeholder - should fetch from database
	return []interface{}{}
}

func getCurrentDate() interface{} {
	// Return current date
	return "2024-01-01" // Placeholder
}
