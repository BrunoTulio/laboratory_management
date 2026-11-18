package web

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"laboratory-management/config"
	"laboratory-management/internal/audit"
	"laboratory-management/internal/auth"
	"laboratory-management/internal/booking"
	"laboratory-management/internal/class"
	"laboratory-management/internal/laboratory"
	"laboratory-management/internal/middleware"
	"laboratory-management/internal/notification"
	"laboratory-management/internal/registry"
	"laboratory-management/internal/user"
)

// Router handles HTTP routing
type Router struct {
	mux                 *http.ServeMux
	handler             http.Handler
	middlewares         []func(http.Handler) http.Handler
	db                  *sql.DB
	templates           *template.Template
	pageTemplates       map[string]*template.Template
	authService         *auth.Service
	userService         *user.Service
	bookingService      *booking.Service
	labService          *laboratory.Service
	notificationService *notification.Service
	auditService        *audit.Service
	registryRepo        *registry.Repository
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
	registryRepo := registry.NewRepository(db)

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

	// Load templates per page to avoid {{define "content"}} collisions
	pageTemplates := loadPageTemplates()

	mux := http.NewServeMux()

	router := &Router{
		mux:                 mux,
		handler:             mux,
		db:                  db,
		pageTemplates:       pageTemplates,
		authService:         authService,
		userService:         userService,
		bookingService:      bookingService,
		labService:          labService,
		notificationService: notificationService,
		auditService:        auditService,
		registryRepo:        registryRepo,
		userHandler:         userHandler,
		labHandler:          labHandler,
		classHandler:        classHandler,
		bookingHandler:      bookingHandler,
		notificationHandler: notificationHandler,
		auditHandler:        auditHandler,
		authHandler:         authHandler,
	}

	if db != nil {
		if err := authService.EnsureDefaultAdmin(config.Load()); err != nil {
			log.Printf("Warning: could not ensure default admin: %v", err)
		}
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

	// Authentication routes
	r.mux.HandleFunc("/login", r.loginPageHandler)
	r.mux.HandleFunc("/cadastro", r.registerDisabledHandler)
	r.mux.HandleFunc("/ativar-conta", r.activateAccountHandler)
	r.mux.HandleFunc("/logout", r.authHandler.LogoutHandler)

	// Protected routes (require authentication)
	r.mux.HandleFunc("/dashboard", middleware.AuthMiddleware(r.dashboardHandler))
	r.mux.HandleFunc("/perfil", middleware.AuthMiddleware(r.profilePageHandler))

	// Booking routes
	r.mux.HandleFunc("/agendamentos", middleware.AuthMiddleware(r.bookingListPageHandler))
	r.mux.HandleFunc("/agendamentos/novo", middleware.AuthMiddleware(r.bookingCreatePageHandler))
	r.mux.HandleFunc("/agendamentos/", middleware.AuthMiddleware(r.bookingDetailPageHandler))
	r.mux.HandleFunc("/disponibilidade", middleware.AuthMiddleware(r.availabilityPageHandler))

	// Notification routes
	r.mux.HandleFunc("/notificacoes", middleware.AuthMiddleware(r.notificationsPageHandler))
	r.mux.HandleFunc("/notificacoes/nao-lidas", middleware.AuthMiddleware(r.notificationHandler.UnreadCountHandler))
	r.mux.HandleFunc("/notificacoes/marcar-lida", middleware.AuthMiddleware(r.notificationHandler.MarkAsReadHandler))
	r.mux.HandleFunc("/notificacoes/marcar-todas-lidas", middleware.AuthMiddleware(r.notificationHandler.MarkAllAsReadHandler))

	// Agenda routes
	r.mux.HandleFunc("/agenda", middleware.AuthMiddleware(r.agendaHandler))
	r.mux.HandleFunc("/agenda/dia", middleware.AuthMiddleware(r.agendaDayHandler))
	r.mux.HandleFunc("/agenda/semana", middleware.AuthMiddleware(r.agendaWeekHandler))
	r.mux.HandleFunc("/agenda/mes", middleware.AuthMiddleware(r.agendaMonthHandler))

	// Cadastros route (admin only)
	r.mux.HandleFunc("/cadastros", middleware.AuthMiddleware(middleware.RequireAdmin(r.cadastrosHandler)))

	// History routes
	r.mux.HandleFunc("/historico", middleware.AuthMiddleware(r.historyHandler))

	// Admin routes (require authentication + admin role)
	r.mux.HandleFunc("/admin", middleware.AuthMiddleware(middleware.RequireAdmin(r.adminDashboardHandler)))
	r.mux.HandleFunc("/admin/usuarios", middleware.AuthMiddleware(middleware.RequireAdmin(r.adminUsersPageHandler)))
	r.mux.HandleFunc("/admin/usuarios/editar", middleware.AuthMiddleware(middleware.RequireAdmin(r.authHandler.AdminUserEditHandler)))
	r.mux.HandleFunc("/admin/agendamentos", middleware.AuthMiddleware(middleware.RequireAdmin(r.adminBookingsPageHandler)))
	r.mux.HandleFunc("/admin/agendamentos/pendentes", middleware.AuthMiddleware(middleware.RequireAdmin(r.adminPendingPageHandler)))
	r.mux.HandleFunc("/admin/agendamentos/aprovar", middleware.AuthMiddleware(middleware.RequireAdmin(r.bookingHandler.AdminApproveHandler)))
	r.mux.HandleFunc("/admin/agendamentos/rejeitar", middleware.AuthMiddleware(middleware.RequireAdmin(r.bookingHandler.AdminRejectHandler)))
	r.mux.HandleFunc("/admin/agendamentos/cancelar", middleware.AuthMiddleware(middleware.RequireAdmin(r.bookingHandler.AdminCancelHandler)))
	r.mux.HandleFunc("/admin/historico", middleware.AuthMiddleware(middleware.RequireAdmin(r.adminHistoryHandler)))
	r.mux.HandleFunc("/admin/auditoria", middleware.AuthMiddleware(middleware.RequireAdmin(r.adminAuditPageHandler)))

	// Laboratory management (admin)
	r.mux.HandleFunc("/admin/laboratorios", middleware.AuthMiddleware(middleware.RequireAdmin(r.labHandler.AdminLabListHandler)))
	r.mux.HandleFunc("/admin/laboratorios/criar", middleware.AuthMiddleware(middleware.RequireAdmin(r.labHandler.AdminLabCreateHandler)))
	r.mux.HandleFunc("/admin/laboratorios/editar", middleware.AuthMiddleware(middleware.RequireAdmin(r.labHandler.AdminLabEditHandler)))

	// Class management (admin)
	r.mux.HandleFunc("/admin/turmas", middleware.AuthMiddleware(middleware.RequireAdmin(r.classHandler.AdminClassListHandler)))
	r.mux.HandleFunc("/admin/turmas/criar", middleware.AuthMiddleware(middleware.RequireAdmin(r.classHandler.AdminClassCreateHandler)))
	r.mux.HandleFunc("/admin/turmas/editar", middleware.AuthMiddleware(middleware.RequireAdmin(r.classHandler.AdminClassEditHandler)))

	// Error pages
	r.mux.HandleFunc("/403", r.errorHandler)
	r.mux.HandleFunc("/404", r.notFoundHandler)
	r.mux.HandleFunc("/500", r.serverErrorHandler)

	// Catch-all for 404
	r.mux.HandleFunc("/", r.catchAllHandler)
}

// Handler functions for routes
func (r *Router) dashboardHandler(w http.ResponseWriter, req *http.Request) {
	now := time.Now()
	userID := middleware.GetUserIDFromContext(req)
	isAdmin := middleware.GetUserRoleFromContext(req) == middleware.RoleAdmin

	todayBookings, err := r.bookingService.GetBookingsByDate(now)
	if err != nil {
		log.Printf("Error loading today bookings: %v", err)
	}
	upcomingBookings, err := r.bookingService.GetUpcomingBookings(now, 20)
	if err != nil {
		log.Printf("Error loading upcoming bookings: %v", err)
	}

	if !isAdmin {
		todayBookings = filterBookingsByUser(todayBookings, userID)
		upcomingBookings = filterBookingsByUser(upcomingBookings, userID)
	}
	if len(upcomingBookings) > 8 {
		upcomingBookings = upcomingBookings[:8]
	}

	summaries, todayCount, freeSlots := BuildShiftSummaries(todayBookings, "", "", "", "")

	data := map[string]interface{}{
		"Title":            "Painel",
		"ActiveTab":        "painel",
		"IsAdmin":          isAdmin,
		"User":             getUserFromContext(req),
		"TodayFormatted":   FormatDatePTBR(now),
		"ShiftSummaries":   summaries,
		"UpcomingBookings": upcomingBookings,
		"Stats": map[string]int{
			"TodayBookings":    todayCount,
			"FreeSlotsToday":   freeSlots,
			"UpcomingBookings": len(upcomingBookings),
		},
	}

	r.renderTemplate(w, "dashboard", data)
}

func (r *Router) agendaHandler(w http.ResponseWriter, req *http.Request) {
	r.renderAgenda(w, req, "day")
}

func (r *Router) agendaDayHandler(w http.ResponseWriter, req *http.Request) {
	r.renderAgenda(w, req, "day")
}

func (r *Router) agendaWeekHandler(w http.ResponseWriter, req *http.Request) {
	r.renderAgenda(w, req, "week")
}

func (r *Router) renderAgenda(w http.ResponseWriter, req *http.Request, viewMode string) {
	targetDate := time.Now()
	if dateParam := req.URL.Query().Get("date"); dateParam != "" {
		if parsed, err := time.Parse("2006-01-02", dateParam); err == nil {
			targetDate = parsed
		}
	}

	filterShift := req.URL.Query().Get("shift")
	filterTeacher := req.URL.Query().Get("teacher")
	filterClass := req.URL.Query().Get("class_id")
	filterClassInt, _ := strconv.Atoi(filterClass)
	filterSubject := req.URL.Query().Get("subject")

	teachers, _ := r.registryRepo.ListTeachers()

	stepDays := 1
	if viewMode == "week" {
		stepDays = 7
	}

	userID := middleware.GetUserIDFromContext(req)
	isAdmin := middleware.GetUserRoleFromContext(req) == middleware.RoleAdmin

	var classOptions, subjectOptions []string
	if isAdmin {
		allUsers, _ := r.userService.GetAllUsers()
		seenC, seenS := map[string]bool{}, map[string]bool{}
		for _, u := range allUsers {
			for _, c := range u.ClassesList() {
				if !seenC[c] {
					seenC[c] = true
					classOptions = append(classOptions, c)
				}
			}
			for _, s := range u.SubjectsList() {
				if !seenS[s] {
					seenS[s] = true
					subjectOptions = append(subjectOptions, s)
				}
			}
		}
	} else if u, err := r.userService.GetUserByID(userID); err == nil && u != nil {
		classOptions = u.ClassesList()
		subjectOptions = u.SubjectsList()
	}
	if len(classOptions) == 0 {
		if cls, err := r.registryRepo.ListClasses(); err == nil {
			for _, c := range cls {
				classOptions = append(classOptions, c.Name)
			}
		}
	}
	if len(subjectOptions) == 0 {
		if subs, err := r.registryRepo.ListSubjects(); err == nil {
			for _, s := range subs {
				subjectOptions = append(subjectOptions, s.Name)
			}
		}
	}

	data := map[string]interface{}{
		"Title":          "Agenda do laboratório",
		"ActiveTab":      "agenda",
		"IsAdmin":        isAdmin,
		"ViewMode":       viewMode,
		"User":           getUserFromContext(req),
		"Laboratories":   getLaboratories(r),
		"CurrentDate":    targetDate,
		"CurrentDateISO": targetDate.Format("2006-01-02"),
		"PrevDateISO":    targetDate.AddDate(0, 0, -stepDays).Format("2006-01-02"),
		"NextDateISO":    targetDate.AddDate(0, 0, stepDays).Format("2006-01-02"),
		"TodayISO":       time.Now().Format("2006-01-02"),
		"TodayFormatted": FormatDatePTBR(targetDate),
		"Teachers":       teachers,
		"ClassOptions":   classOptions,
		"SubjectOptions": subjectOptions,
		"FilterShift":    filterShift,
		"FilterTeacher":  filterTeacher,
		"FilterClass":    filterClass,
		"FilterClassInt": filterClassInt,
		"FilterSubject":  filterSubject,
	}

	if viewMode == "week" {
		weekDates := BuildWeekDates(targetDate)
		var weekDays []WeekDaySchedule
		for _, d := range weekDates {
			dayBookings, err := r.bookingService.GetBookingsByDate(d)
			if err != nil {
				log.Printf("Error loading week bookings for %s: %v", d.Format("2006-01-02"), err)
			}
			if !isAdmin {
				dayBookings = filterBookingsByUser(dayBookings, userID)
			}
			summaries, booked, free := BuildShiftSummaries(dayBookings, filterShift, filterTeacher, filterClass, filterSubject)
			weekDays = append(weekDays, WeekDaySchedule{
				Date:          d,
				DateISO:       d.Format("2006-01-02"),
				DateFormatted: FormatDatePTBR(d),
				Summaries:     summaries,
				BookedCount:   booked,
				FreeCount:     free,
			})
		}
		data["WeekDays"] = weekDays
		data["TodayFormatted"] = FormatDatePTBR(weekDates[0]) + " — " + FormatDatePTBR(weekDates[len(weekDates)-1])
	} else {
		dayBookings, err := r.bookingService.GetBookingsByDate(targetDate)
		if err != nil {
			log.Printf("Error loading agenda bookings: %v", err)
		}
		if !isAdmin {
			dayBookings = filterBookingsByUser(dayBookings, userID)
		}
		summaries, _, _ := BuildShiftSummaries(dayBookings, filterShift, filterTeacher, filterClass, filterSubject)
		data["ShiftSummaries"] = summaries
	}

	r.renderTemplate(w, "agenda", data)
}

func (r *Router) cadastrosHandler(w http.ResponseWriter, req *http.Request) {
	renderPage := func(flashMsg, flashType string) {
		teachers, _ := r.registryRepo.ListTeachers()
		classes, _ := r.registryRepo.ListClassesWithSubjects()
		subjects, _ := r.registryRepo.ListSubjects()
		var usersList []*user.User
		if middleware.GetUserRoleFromContext(req) == middleware.RoleAdmin {
			usersList, _ = r.userService.GetAllUsers()
		}

		data := map[string]interface{}{
			"Title":        "Cadastros",
			"ActiveTab":    "cadastros",
			"User":         getUserFromContext(req),
			"CSRFToken":    middleware.GenerateCSRFToken(),
			"Teachers":     teachers,
			"Classes":      classes,
			"Subjects":     subjects,
			"Users":        usersList,
			"AdminEmail":   config.Load().AdminEmail,
			"IsAdmin":      middleware.GetUserRoleFromContext(req) == middleware.RoleAdmin,
			"FlashMessage": flashMsg,
			"FlashType":    flashType,
		}
		r.renderTemplate(w, "cadastros", data)
	}

	if req.Method == http.MethodPost {
		if err := req.ParseForm(); err != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		action := req.FormValue("action")
		itemType := req.FormValue("type")
		name := req.FormValue("name")
		id, _ := strconv.Atoi(req.FormValue("id"))

		if itemType == "class_subjects" {
			var subjectIDs []int
			for _, rawID := range req.Form["subject_ids"] {
				if sid, err := strconv.Atoi(rawID); err == nil && sid > 0 {
					subjectIDs = append(subjectIDs, sid)
				}
			}
			if err := r.registryRepo.SetClassSubjects(id, subjectIDs); err != nil {
				renderPage(err.Error(), "error")
				return
			}
			http.Redirect(w, req, "/cadastros", http.StatusSeeOther)
			return
		}

		if itemType == "user_status" && id > 0 {
			if middleware.GetUserRoleFromContext(req) != middleware.RoleAdmin {
				renderPage("Apenas a coordenação pode alterar usuários.", "error")
				return
			}
			newStatus := user.StatusInactive
			if action == "activate" {
				newStatus = user.StatusActive
			}
			if err := r.userService.SetUserStatus(id, newStatus, middleware.GetUserIDFromContext(req), config.Load().AdminEmail); err != nil {
				renderPage(err.Error(), "error")
				return
			}
			http.Redirect(w, req, "/cadastros", http.StatusSeeOther)
			return
		}

		if itemType == "user_delete" && id > 0 {
			if middleware.GetUserRoleFromContext(req) != middleware.RoleAdmin {
				renderPage("Apenas a coordenação pode excluir usuários.", "error")
				return
			}
			if err := r.userService.DeleteUserSafe(id, middleware.GetUserIDFromContext(req), config.Load().AdminEmail); err != nil {
				renderPage(err.Error(), "error")
				return
			}
			http.Redirect(w, req, "/cadastros", http.StatusSeeOther)
			return
		}

		if itemType == "user_invite" {
			if middleware.GetUserRoleFromContext(req) != middleware.RoleAdmin {
				renderPage("Apenas a coordenação pode cadastrar novos usuários.", "error")
				return
			}
			email := req.FormValue("email")
			role := req.FormValue("role")
			_, activationURL, err := r.authService.InviteUser(name, email, role, "", "")
			if err != nil {
				renderPage(err.Error(), "error")
				return
			}
			renderPage("Convite enviado para "+email+". Link de ativação: "+activationURL, "success")
			return
		}

		var err error
		switch {
		case action == "delete" && id > 0:
			switch itemType {
			case "teacher":
				err = r.registryRepo.DeleteTeacher(id)
			case "class":
				err = r.registryRepo.DeleteClass(id)
			case "subject":
				err = r.registryRepo.DeleteSubject(id)
			}
		case action == "update" && id > 0:
			switch itemType {
			case "teacher":
				err = r.registryRepo.UpdateTeacher(id, name)
			case "class":
				err = r.registryRepo.UpdateClass(id, name)
			case "subject":
				err = r.registryRepo.UpdateSubject(id, name)
			}
		default:
			switch itemType {
			case "teacher":
				err = r.registryRepo.AddTeacher(name)
			case "class":
				err = r.registryRepo.AddClass(name)
			case "subject":
				err = r.registryRepo.AddSubject(name)
			}
		}

		if err != nil {
			log.Printf("Error updating cadastros (%s/%s): %v", action, itemType, err)
		}

		http.Redirect(w, req, "/cadastros", http.StatusSeeOther)
		return
	}

	renderPage("", "")
}

func (r *Router) bookingCreatePageHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost && strings.Contains(req.Header.Get("Content-Type"), "application/json") {
		r.bookingHandler.CreateHandler(w, req)
		return
	}

	renderForm := func(flashMsg, flashType string, formData map[string]string) {
		labs, _ := r.labService.GetAllLaboratories()
		var classOptions, subjectOptions []string
		if u, err := r.userService.GetUserByID(middleware.GetUserIDFromContext(req)); err == nil && u != nil {
			classOptions = u.ClassesList()
			subjectOptions = u.SubjectsList()
		}
		if len(classOptions) == 0 {
			if cls, err := r.registryRepo.ListClasses(); err == nil {
				for _, c := range cls {
					classOptions = append(classOptions, c.Name)
				}
			}
		}
		if len(subjectOptions) == 0 {
			if subs, err := r.registryRepo.ListSubjects(); err == nil {
				for _, s := range subs {
					subjectOptions = append(subjectOptions, s.Name)
				}
			}
		}

		now := time.Now()
		if formData == nil {
			dateParam := req.URL.Query().Get("date")
			if dateParam == "" {
				dateParam = now.Format("2006-01-02")
			}
			shiftParam := req.URL.Query().Get("shift")
			if shiftParam == "" {
				shiftParam = "MATUTINO"
			}
			formData = map[string]string{
				"BookingDate": dateParam,
				"Shift":       shiftParam,
				"Lesson":      req.URL.Query().Get("lesson"),
				"StartTime":   req.URL.Query().Get("start"),
				"EndTime":     req.URL.Query().Get("end"),
			}
		}

		classSubjectsMap, _ := r.registryRepo.GetClassSubjectsMap()
		classSubjectsBytes, _ := json.Marshal(classSubjectsMap)

		data := map[string]interface{}{
			"Title":             "Novo agendamento",
			"ActiveTab":         "agenda",
			"IsAdmin":           middleware.GetUserRoleFromContext(req) == middleware.RoleAdmin,
			"User":              getUserFromContext(req),
			"CSRFToken":         middleware.GenerateCSRFToken(),
			"Laboratories":      labs,
			"ClassOptions":      classOptions,
			"SubjectOptions":    subjectOptions,
			"ClassSubjectsJSON": template.JS(classSubjectsBytes),
			"MinDate":           now.Format("2006-01-02"),
			"FormData":          formData,
			"FlashMessage":      flashMsg,
			"FlashType":         flashType,
		}
		r.renderTemplate(w, "booking-create", data)
	}

	if req.Method == http.MethodGet {
		renderForm("", "", nil)
		return
	}

	if req.Method == http.MethodPost {
		if err := req.ParseForm(); err != nil {
			renderForm("Dados do formulário inválidos.", "error", nil)
			return
		}

		formData := map[string]string{
			"LaboratoryID": req.FormValue("laboratory_id"),
			"BookingDate":  req.FormValue("booking_date"),
			"StartTime":    req.FormValue("start_time"),
			"EndTime":      req.FormValue("end_time"),
			"Shift":        req.FormValue("shift"),
			"TeacherName":  req.FormValue("teacher_name"),
			"ClassName":    req.FormValue("class_name"),
			"SubjectName":  req.FormValue("subject_name"),
			"Reason":       req.FormValue("reason"),
			"Observation":  req.FormValue("observation"),
		}

		labID, err := strconv.Atoi(formData["LaboratoryID"])
		if err != nil || labID <= 0 {
			labID = 1
		}

		bookingDate, err := time.Parse("2006-01-02", formData["BookingDate"])
		if err != nil {
			renderForm("Data inválida.", "error", formData)
			return
		}

		var startTime, endTime time.Time
		lessons := req.Form["lessons"]
		if len(lessons) > 0 {
			startTime, endTime, err = ResolveLessonRange(formData["Shift"], lessons)
			if err != nil {
				renderForm(err.Error(), "error", formData)
				return
			}
			formData["Lesson"] = lessons[0]
		} else if formData["StartTime"] != "" && formData["EndTime"] != "" {
			startTime, err = time.Parse("15:04", formData["StartTime"])
			if err != nil {
				renderForm("Horário de início inválido.", "error", formData)
				return
			}
			endTime, err = time.Parse("15:04", formData["EndTime"])
			if err != nil {
				renderForm("Horário de término inválido.", "error", formData)
				return
			}
		} else {
			renderForm("Selecione pelo menos uma aula.", "error", formData)
			return
		}

		var classIDPtr *int
		if cid, err := strconv.Atoi(formData["ClassID"]); err == nil && cid > 0 {
			classIDPtr = &cid
		}

		userID := middleware.GetUserIDFromContext(req)
		isAdmin := middleware.GetUserRoleFromContext(req) == middleware.RoleAdmin
		teacherName := strings.TrimSpace(formData["TeacherName"])
		if !isAdmin || teacherName == "" {
			if u, err := r.userService.GetUserByID(userID); err == nil && u != nil && u.Name != "" {
				teacherName = u.Name
			}
		}

		createReq := &booking.CreateBookingRequest{
			LaboratoryID: labID,
			ClassID:      classIDPtr,
			BookingDate:  bookingDate,
			StartTime:    startTime,
			EndTime:      endTime,
			Shift:        formData["Shift"],
			TeacherName:  teacherName,
			ClassName:    strings.TrimSpace(formData["ClassName"]),
			SubjectName:  formData["SubjectName"],
			Reason:       formData["Reason"],
			Observation:  formData["Observation"],
		}

		if _, err := r.bookingService.CreateBooking(createReq, userID); err != nil {
			renderForm(err.Error(), "error", formData)
			return
		}

		http.Redirect(w, req, "/agenda?date="+formData["BookingDate"], http.StatusSeeOther)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func (r *Router) loginPageHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost && strings.Contains(req.Header.Get("Content-Type"), "application/json") {
		r.authHandler.LoginHandler(w, req)
		return
	}

	renderLogin := func(flashMsg, flashType, email string) {
		r.renderAuthTemplate(w, "login", map[string]interface{}{
			"Title":        "Login",
			"Subtitle":     "Acesse sua conta para gerenciar agendamentos",
			"CSRFToken":    middleware.GenerateCSRFToken(),
			"FlashMessage": flashMsg,
			"FlashType":    flashType,
			"FormData": map[string]string{
				"Email": email,
			},
		})
	}

	if req.Method == http.MethodGet {
		renderLogin("", "", "")
		return
	}

	if req.Method == http.MethodPost {
		if err := req.ParseForm(); err != nil {
			renderLogin("Dados inválidos.", "error", "")
			return
		}

		email := strings.TrimSpace(req.FormValue("email"))
		password := req.FormValue("password")

		loggedUser, err := r.authService.Login(email, password)
		if err != nil {
			renderLogin("E-mail ou senha inválidos.", "error", email)
			return
		}

		session, err := r.authService.CreateSession(loggedUser)
		if err != nil {
			renderLogin("Erro ao criar sessão.", "error", email)
			return
		}

		middleware.SetSessionCookie(w, session.Token)
		http.Redirect(w, req, "/dashboard", http.StatusSeeOther)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func (r *Router) registerDisabledHandler(w http.ResponseWriter, req *http.Request) {
	r.renderAuthTemplate(w, "login", map[string]interface{}{
		"Title":        "Login",
		"CSRFToken":    middleware.GenerateCSRFToken(),
		"FlashMessage": "Cadastro público desativado. Solicite um convite à coordenação.",
		"FlashType":    "error",
		"FormData":     map[string]string{},
	})
}

func (r *Router) activateAccountHandler(w http.ResponseWriter, req *http.Request) {
	renderActivate := func(token, email string, isProf bool, classesCSV, subjectsCSV, flashMsg, flashType string) {
		availClasses, _ := r.registryRepo.ListClasses()
		availSubjects, _ := r.registryRepo.ListSubjects()
		r.renderAuthTemplate(w, "activate", map[string]interface{}{
			"Title":             "Finalizar cadastro",
			"CSRFToken":         middleware.GenerateCSRFToken(),
			"Token":             token,
			"UserEmail":         email,
			"IsProfessor":       isProf,
			"AvailableClasses":  availClasses,
			"AvailableSubjects": availSubjects,
			"ClassesCSV":        classesCSV,
			"SubjectsCSV":       subjectsCSV,
			"FlashMessage":      flashMsg,
			"FlashType":         flashType,
		})
	}

	if req.Method == http.MethodGet {
		token := strings.TrimSpace(req.URL.Query().Get("token"))
		u, err := r.authService.ValidateActivationToken(token)
		if err != nil {
			r.renderAuthTemplate(w, "login", map[string]interface{}{
				"Title":        "Login",
				"CSRFToken":    middleware.GenerateCSRFToken(),
				"FlashMessage": err.Error(),
				"FlashType":    "error",
				"FormData":     map[string]string{},
			})
			return
		}
		isProf := u.Role != nil && u.Role.Name == user.RoleProfessor
		renderActivate(token, u.Email, isProf, "", "", "", "")
		return
	}

	if req.Method == http.MethodPost {
		if err := req.ParseForm(); err != nil {
			renderActivate("", "", false, "", "", "Dados inválidos.", "error")
			return
		}
		token := strings.TrimSpace(req.FormValue("token"))
		password := req.FormValue("password")
		confirm := req.FormValue("confirm_password")
		classesCSV := req.FormValue("classes_csv")
		if selected := req.Form["classes"]; len(selected) > 0 {
			classesCSV = strings.Join(selected, ", ")
		}
		subjectsCSV := req.FormValue("subjects_csv")
		if selected := req.Form["subjects"]; len(selected) > 0 {
			subjectsCSV = strings.Join(selected, ", ")
		}

		u, _ := r.authService.ValidateActivationToken(token)
		isProf := u != nil && u.Role != nil && u.Role.Name == user.RoleProfessor
		email := ""
		if u != nil {
			email = u.Email
		}

		if password != confirm {
			renderActivate(token, email, isProf, classesCSV, subjectsCSV, "As senhas não coincidem.", "error")
			return
		}

		if err := r.authService.CompleteActivation(token, password, classesCSV, subjectsCSV); err != nil {
			renderActivate(token, email, isProf, classesCSV, subjectsCSV, err.Error(), "error")
			return
		}

		r.renderAuthTemplate(w, "login", map[string]interface{}{
			"Title":        "Login",
			"CSRFToken":    middleware.GenerateCSRFToken(),
			"FlashMessage": "Conta ativada com sucesso! Faça login com seu e-mail e senha.",
			"FlashType":    "success",
			"FormData":     map[string]string{},
		})
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func (r *Router) agendaMonthHandler(w http.ResponseWriter, req *http.Request) {
	r.renderAgenda(w, req, "day")
}

func (r *Router) profilePageHandler(w http.ResponseWriter, req *http.Request) {
	if strings.Contains(req.Header.Get("Accept"), "application/json") {
		r.authHandler.ProfileHandler(w, req)
		return
	}

	userID := middleware.GetUserIDFromContext(req)
	u, err := r.userService.GetUserByID(userID)
	if err != nil {
		u = &user.User{
			ID:    userID,
			Name:  middleware.GetUserEmailFromContext(req),
			Email: middleware.GetUserEmailFromContext(req),
			Role:  &user.Role{Name: middleware.GetUserRoleFromContext(req)},
		}
	}
	if u.Role == nil {
		u.Role = &user.Role{Name: middleware.GetUserRoleFromContext(req)}
	}

	r.renderTemplate(w, "profile", map[string]interface{}{
		"Title":     "Perfil",
		"User":      u,
		"CSRFToken": middleware.GenerateCSRFToken(),
	})
}

func (r *Router) bookingListPageHandler(w http.ResponseWriter, req *http.Request) {
	if strings.Contains(req.Header.Get("Accept"), "application/json") {
		r.bookingHandler.ListHandler(w, req)
		return
	}

	userID := middleware.GetUserIDFromContext(req)
	bookings, _ := r.bookingService.GetBookingsByUserID(userID)
	filterStatus := req.URL.Query().Get("status")
	filterDate := req.URL.Query().Get("date")

	var filtered []*booking.Booking
	for _, b := range bookings {
		if filterStatus != "" && b.Status != filterStatus {
			continue
		}
		if filterDate != "" && b.BookingDate.Format("2006-01-02") != filterDate {
			continue
		}
		filtered = append(filtered, b)
	}
	r.hydrateBookings(filtered)

	r.renderTemplate(w, "booking-list", map[string]interface{}{
		"Title":        "Meus Agendamentos",
		"User":         getUserFromContext(req),
		"Bookings":     filtered,
		"FilterStatus": filterStatus,
		"FilterDate":   filterDate,
	})
}

func (r *Router) bookingDetailPageHandler(w http.ResponseWriter, req *http.Request) {
	if strings.Contains(req.Header.Get("Accept"), "application/json") {
		r.bookingHandler.DetailHandler(w, req)
		return
	}

	idStr := req.URL.Query().Get("id")
	if idStr == "" {
		idStr = strings.TrimPrefix(req.URL.Path, "/agendamentos/")
		idStr = strings.Trim(idStr, "/")
	}
	bookingID, err := strconv.Atoi(idStr)
	if err != nil || bookingID <= 0 {
		r.notFoundHandler(w, req)
		return
	}

	b, err := r.bookingService.GetBookingByID(bookingID)
	if err != nil {
		r.notFoundHandler(w, req)
		return
	}

	userID := middleware.GetUserIDFromContext(req)
	userRole := middleware.GetUserRoleFromContext(req)
	if b.UserID != userID && userRole != middleware.RoleAdmin {
		r.errorHandler(w, req)
		return
	}

	r.hydrateBookings([]*booking.Booking{b})
	r.renderTemplate(w, "booking-detail", map[string]interface{}{
		"Title":   "Detalhes do Agendamento",
		"User":    getUserFromContext(req),
		"Booking": b,
	})
}

func (r *Router) availabilityPageHandler(w http.ResponseWriter, req *http.Request) {
	if req.URL.Query().Get("start_time") != "" && req.URL.Query().Get("end_time") != "" {
		r.bookingHandler.AvailabilityHandler(w, req)
		return
	}

	targetDate := time.Now()
	if dateStr := req.URL.Query().Get("date"); dateStr != "" {
		if parsed, err := time.Parse("2006-01-02", dateStr); err == nil {
			targetDate = parsed
		}
	}

	dayBookings, _ := r.bookingService.GetBookingsByDate(targetDate)
	summaries, _, _ := BuildShiftSummaries(dayBookings, "", "", "", "")
	labs, _ := r.labService.GetAllLaboratories()

	r.renderTemplate(w, "availability", map[string]interface{}{
		"Title":          "Disponibilidade",
		"ActiveTab":      "agenda",
		"User":           getUserFromContext(req),
		"Laboratories":   labs,
		"Date":           targetDate.Format("2006-01-02"),
		"TodayFormatted": FormatDatePTBR(targetDate),
		"ShiftSummaries": summaries,
	})
}

func (r *Router) notificationsPageHandler(w http.ResponseWriter, req *http.Request) {
	if strings.Contains(req.Header.Get("Accept"), "application/json") {
		r.notificationHandler.ListHandler(w, req)
		return
	}

	userID := middleware.GetUserIDFromContext(req)
	notifications, _ := r.notificationService.GetNotificationsByUserID(userID)
	unreadCount, _ := r.notificationService.GetUnreadCount(userID)

	r.renderTemplate(w, "notifications", map[string]interface{}{
		"Title":         "Notificações",
		"User":          getUserFromContext(req),
		"Notifications": notifications,
		"UnreadCount":   unreadCount,
		"CSRFToken":     middleware.GenerateCSRFToken(),
	})
}

func (r *Router) historyHandler(w http.ResponseWriter, req *http.Request) {
	userID := middleware.GetUserIDFromContext(req)
	bookings, _ := r.bookingService.GetBookingsByUserID(userID)

	dateFrom := req.URL.Query().Get("date_from")
	dateTo := req.URL.Query().Get("date_to")
	status := req.URL.Query().Get("status")

	var filtered []*booking.Booking
	for _, b := range bookings {
		d := b.BookingDate.Format("2006-01-02")
		if dateFrom != "" && d < dateFrom {
			continue
		}
		if dateTo != "" && d > dateTo {
			continue
		}
		if status != "" && b.Status != status {
			continue
		}
		filtered = append(filtered, b)
	}
	r.hydrateBookings(filtered)

	data := map[string]interface{}{
		"Title":          "Histórico",
		"User":           getUserFromContext(req),
		"FilterDateFrom": dateFrom,
		"FilterDateTo":   dateTo,
		"FilterStatus":   status,
		"Bookings":       filtered,
	}

	r.renderTemplate(w, "history", data)
}

func (r *Router) adminDashboardHandler(w http.ResponseWriter, req *http.Request) {
	users, _ := r.userService.GetAllUsers()
	allBookings, _ := r.bookingService.GetAllBookings()
	pendingBookings, _ := r.bookingService.GetPendingBookings()
	upcomingBookings, _ := r.bookingService.GetUpcomingBookings(time.Now(), 20)
	r.hydrateBookings(pendingBookings)

	approvedCount := 0
	for _, b := range allBookings {
		if b.Status == booking.StatusApproved {
			approvedCount++
		}
	}

	data := map[string]interface{}{
		"Title": "Dashboard Administrativo",
		"User":  getUserFromContext(req),
		"Stats": map[string]int{
			"TotalUsers":       len(users),
			"PendingBookings":  len(pendingBookings),
			"ApprovedBookings": approvedCount,
			"UpcomingBookings": len(upcomingBookings),
		},
		"PendingBookings": pendingBookings,
	}

	r.renderTemplate(w, "admin-dashboard", data)
}

func (r *Router) adminUsersPageHandler(w http.ResponseWriter, req *http.Request) {
	if strings.Contains(req.Header.Get("Accept"), "application/json") {
		r.userHandler.AdminUserListHandler(w, req)
		return
	}

	users, _ := r.userService.GetAllUsers()
	filterRole := req.URL.Query().Get("role")
	filterStatus := req.URL.Query().Get("status")
	filterSearch := strings.ToLower(strings.TrimSpace(req.URL.Query().Get("search")))

	var filtered []*user.User
	for _, u := range users {
		if u.Role == nil {
			u.Role = &user.Role{Name: ""}
		}
		if filterRole != "" && u.Role.Name != filterRole {
			continue
		}
		if filterStatus != "" && u.Status != filterStatus {
			continue
		}
		if filterSearch != "" && !strings.Contains(strings.ToLower(u.Name), filterSearch) && !strings.Contains(strings.ToLower(u.Email), filterSearch) {
			continue
		}
		filtered = append(filtered, u)
	}

	r.renderTemplate(w, "admin-users", map[string]interface{}{
		"Title":        "Gerenciar Usuários",
		"User":         getUserFromContext(req),
		"Users":        filtered,
		"FilterRole":   filterRole,
		"FilterStatus": filterStatus,
		"FilterSearch": req.URL.Query().Get("search"),
	})
}

func (r *Router) adminBookingsPageHandler(w http.ResponseWriter, req *http.Request) {
	if strings.Contains(req.Header.Get("Accept"), "application/json") {
		r.bookingHandler.AdminBookingListHandler(w, req)
		return
	}

	allBookings, _ := r.bookingService.GetAllBookings()
	pendingBookings, _ := r.bookingService.GetPendingBookings()
	r.hydrateBookings(allBookings)
	r.hydrateBookings(pendingBookings)

	var approved, rejected []*booking.Booking
	for _, b := range allBookings {
		switch b.Status {
		case booking.StatusApproved:
			approved = append(approved, b)
		case booking.StatusRejected:
			rejected = append(rejected, b)
		}
	}

	r.renderTemplate(w, "admin-bookings", map[string]interface{}{
		"Title":            "Gerenciar Agendamentos",
		"User":             getUserFromContext(req),
		"PendingCount":     len(pendingBookings),
		"PendingBookings":  pendingBookings,
		"AllBookings":      allBookings,
		"ApprovedBookings": approved,
		"RejectedBookings": rejected,
	})
}

func (r *Router) adminPendingPageHandler(w http.ResponseWriter, req *http.Request) {
	if strings.Contains(req.Header.Get("Accept"), "application/json") {
		r.bookingHandler.AdminPendingListHandler(w, req)
		return
	}
	r.adminBookingsPageHandler(w, req)
}

func (r *Router) adminHistoryHandler(w http.ResponseWriter, req *http.Request) {
	allBookings, _ := r.bookingService.GetAllBookings()
	r.hydrateBookings(allBookings)

	r.renderTemplate(w, "admin-bookings", map[string]interface{}{
		"Title":       "Histórico Administrativo",
		"User":        getUserFromContext(req),
		"AllBookings": allBookings,
		"Bookings":    allBookings,
	})
}

func (r *Router) adminAuditPageHandler(w http.ResponseWriter, req *http.Request) {
	if strings.Contains(req.Header.Get("Accept"), "application/json") {
		r.auditHandler.ListHandler(w, req)
		return
	}

	logs, _ := r.auditService.GetAllAuditLogs()
	filterAction := req.URL.Query().Get("action")
	filterEntity := req.URL.Query().Get("entity")
	dateFrom := req.URL.Query().Get("date_from")
	dateTo := req.URL.Query().Get("date_to")

	var filtered []*audit.AuditLog
	for _, l := range logs {
		if filterAction != "" && l.Action != filterAction {
			continue
		}
		if filterEntity != "" && l.EntityType != filterEntity {
			continue
		}
		d := l.CreatedAt.Format("2006-01-02")
		if dateFrom != "" && d < dateFrom {
			continue
		}
		if dateTo != "" && d > dateTo {
			continue
		}
		filtered = append(filtered, l)
	}

	r.renderTemplate(w, "admin-audit", map[string]interface{}{
		"Title":          "Auditoria do Sistema",
		"User":           getUserFromContext(req),
		"AuditLogs":      filtered,
		"FilterAction":   filterAction,
		"FilterEntity":   filterEntity,
		"FilterDateFrom": dateFrom,
		"FilterDateTo":   dateTo,
	})
}

func (r *Router) hydrateBookings(bookings []*booking.Booking) {
	for _, b := range bookings {
		if b.Laboratory == nil {
			b.Laboratory = &laboratory.Laboratory{ID: b.LaboratoryID, Name: "Laboratório Principal"}
		}
		if b.User == nil {
			name := b.TeacherName
			if name == "" {
				name = "Usuário"
			}
			b.User = &user.User{ID: b.UserID, Name: name}
		}
	}
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

func loadPageTemplates() map[string]*template.Template {
	result := make(map[string]*template.Template)
	funcs := template.FuncMap{
		"lower": strings.ToLower,
	}

	baseSet, err := template.New("base").Funcs(funcs).ParseFiles(
		"templates/layouts/base.html",
		"templates/components/navbar.html",
		"templates/components/footer.html",
		"templates/components/sidebar.html",
		"templates/components/form-field.html",
	)
	if err != nil {
		log.Printf("Warning: could not load base layout templates: %v", err)
		return result
	}

	authSet, err := template.New("auth").Funcs(funcs).ParseFiles(
		"templates/layouts/auth.html",
		"templates/components/form-field.html",
	)
	if err != nil {
		log.Printf("Warning: could not load auth layout templates: %v", err)
		return result
	}

	pages, _ := filepath.Glob("templates/pages/*.html")
	for _, pagePath := range pages {
		name := strings.TrimSuffix(filepath.Base(pagePath), ".html")
		if name == "login" || name == "register" || name == "activate" {
			cloned, _ := authSet.Clone()
			result[name] = template.Must(cloned.ParseFiles(pagePath))
		} else {
			cloned, _ := baseSet.Clone()
			result[name] = template.Must(cloned.ParseFiles(pagePath))
		}
	}

	return result
}

// Helper functions
func (r *Router) renderTemplate(w http.ResponseWriter, templateName string, data interface{}) {
	tmpl, ok := r.pageTemplates[templateName]
	if !ok {
		log.Printf("Template %s not found", templateName)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if err := tmpl.ExecuteTemplate(w, "base.html", data); err != nil {
		log.Printf("Error rendering template %s: %v", templateName, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (r *Router) renderAuthTemplate(w http.ResponseWriter, templateName string, data interface{}) {
	tmpl, ok := r.pageTemplates[templateName]
	if !ok {
		log.Printf("Auth template %s not found", templateName)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if err := tmpl.ExecuteTemplate(w, "auth.html", data); err != nil {
		log.Printf("Error rendering auth template %s: %v", templateName, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func getUserFromContext(req *http.Request) map[string]interface{} {
	email := middleware.GetUserEmailFromContext(req)
	name := email
	if idx := strings.Index(email, "@"); idx > 0 {
		name = email[:idx]
	}
	if name == "" {
		name = "Usuário"
	}
	return map[string]interface{}{
		"ID":    middleware.GetUserIDFromContext(req),
		"Name":  name,
		"Email": email,
		"Role":  middleware.GetUserRoleFromContext(req),
	}
}

func getLaboratories(r *Router) []interface{} {
	return []interface{}{}
}

func getCurrentDate() time.Time {
	return time.Now()
}

func filterBookingsByUser(bookings []*booking.Booking, userID int) []*booking.Booking {
	var out []*booking.Booking
	for _, b := range bookings {
		if b.UserID == userID {
			out = append(out, b)
		}
	}
	return out
}
