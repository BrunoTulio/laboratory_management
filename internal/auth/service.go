package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"laboratory-management/config"
	"laboratory-management/internal/mailer"
	"laboratory-management/internal/middleware"
	"laboratory-management/internal/user"
	"log"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Service handles authentication operations
type Service struct {
	userRepo *user.Repository
}

// NewService creates a new authentication service
func NewService(userRepo *user.Repository) *Service {
	return &Service{userRepo: userRepo}
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("error hashing password: %w", err)
	}
	return string(bytes), nil
}

// CheckPassword checks if a password matches a hash
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateSessionToken generates a random session token
func GenerateSessionToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("error generating session token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// Login authenticates a user and returns the user if successful
func (s *Service) Login(email, password string) (*user.User, error) {
	// Get user by email
	foundUser, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check password
	if !CheckPassword(password, foundUser.PasswordHash) {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check if user is active
	if foundUser.Status != user.StatusActive {
		return nil, fmt.Errorf("user account is not active")
	}

	return foundUser, nil
}

// Register creates a new user
func (s *Service) Register(req *user.CreateUserRequest) (*user.User, error) {
	// Validate input
	if req.Name == "" || req.Email == "" || req.Password == "" {
		return nil, fmt.Errorf("name, email, and password are required")
	}

	// Check if email already exists
	_, err := s.userRepo.GetByEmail(req.Email)
	if err == nil {
		return nil, fmt.Errorf("email already registered")
	}

	// Hash password
	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("error processing password: %w", err)
	}

	// Create user
	newUser := &user.User{
		Name:               req.Name,
		Email:              req.Email,
		PasswordHash:       hashedPassword,
		RoleID:             req.RoleID,
		RegistrationNumber: req.RegistrationNumber,
		Phone:              req.Phone,
		Status:             user.StatusActive,
	}

	err = s.userRepo.Create(newUser)
	if err != nil {
		return nil, fmt.Errorf("error creating user: %w", err)
	}

	// Remove password hash from response
	newUser.PasswordHash = ""

	return newUser, nil
}

// ValidateSession validates a signed session token and returns the active user
func (s *Service) ValidateSession(token string) (*user.User, error) {
	userID, _, _, ok := middleware.VerifySessionToken(token, config.Load().SessionSecret)
	if !ok {
		return nil, fmt.Errorf("invalid or expired session token")
	}

	foundUser, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	if foundUser.Status != user.StatusActive {
		return nil, fmt.Errorf("user account is not active")
	}

	return foundUser, nil
}

// Session represents a user session
type Session struct {
	Token     string
	UserID    int
	UserEmail string
	UserName  string
	Role      string
	ExpiresAt time.Time
	CreatedAt time.Time
}

// CreateSession creates a new signed session for a user
func (s *Service) CreateSession(u *user.User) (*Session, error) {
	roleName := ""
	if u.Role != nil {
		roleName = u.Role.Name
	}

	ttl := 24 * time.Hour
	token := middleware.SignSessionToken(u.ID, u.Email, roleName, ttl, config.Load().SessionSecret)

	session := &Session{
		Token:     token,
		UserID:    u.ID,
		UserEmail: u.Email,
		UserName:  u.Name,
		Role:      roleName,
		ExpiresAt: time.Now().Add(ttl),
		CreatedAt: time.Now(),
	}

	return session, nil
}

// EnsureDefaultAdmin seeds the default DIRECAO administrator account if it does not exist.
func (s *Service) EnsureDefaultAdmin(cfg *config.Config) error {
	if _, err := s.userRepo.GetByEmail(cfg.AdminEmail); err == nil {
		return nil
	}

	adminRole, err := s.userRepo.GetRoleByName(user.RoleAdmin)
	if err != nil {
		return fmt.Errorf("default admin role not found: %w", err)
	}

	hashed, err := HashPassword(cfg.AdminPassword)
	if err != nil {
		return err
	}

	adminUser := &user.User{
		Name:         cfg.AdminName,
		Email:        cfg.AdminEmail,
		PasswordHash: hashed,
		RoleID:       adminRole.ID,
		Status:       user.StatusActive,
	}
	if err := s.userRepo.Create(adminUser); err != nil {
		return fmt.Errorf("error creating default admin: %w", err)
	}

	log.Printf("Default admin user created: %s", cfg.AdminEmail)
	return nil
}

// InviteUser creates a pending user and sends an activation link by email so the user can set their password.
func (s *Service) InviteUser(name, email, roleName, registrationNumber, phone string) (*user.User, string, error) {
	name = strings.TrimSpace(name)
	email = strings.TrimSpace(strings.ToLower(email))
	if name == "" || email == "" {
		return nil, "", fmt.Errorf("nome e e-mail são obrigatórios")
	}

	if _, err := s.userRepo.GetByEmail(email); err == nil {
		return nil, "", fmt.Errorf("e-mail já cadastrado")
	}

	if roleName == "" {
		roleName = user.RoleProfessor
	}
	role, err := s.userRepo.GetRoleByName(roleName)
	if err != nil {
		return nil, "", fmt.Errorf("perfil inválido: %w", err)
	}

	token, err := GenerateSessionToken()
	if err != nil {
		return nil, "", err
	}

	invited := &user.User{
		Name:               name,
		Email:              email,
		RoleID:             role.ID,
		RegistrationNumber: registrationNumber,
		Phone:              phone,
		Status:             user.StatusPending,
	}
	expiresAt := time.Now().Add(48 * time.Hour)
	if err := s.userRepo.CreateInvited(invited, token, expiresAt); err != nil {
		return nil, "", fmt.Errorf("erro ao cadastrar usuário convidado: %w", err)
	}

	cfg := config.Load()
	activationURL := fmt.Sprintf("%s/ativar-conta?token=%s", strings.TrimRight(cfg.BaseURL, "/"), token)
	if err := mailer.SendActivationEmail(email, name, activationURL); err != nil {
		return invited, activationURL, fmt.Errorf("usuário criado, mas falha ao enviar e-mail: %w", err)
	}

	return invited, activationURL, nil
}

// ValidateActivationToken checks if an activation token exists and is not expired.
func (s *Service) ValidateActivationToken(token string) (*user.User, error) {
	u, expiresAt, err := s.userRepo.GetByActivationToken(token)
	if err != nil {
		return nil, fmt.Errorf("link de ativação inválido")
	}
	if time.Now().After(expiresAt) {
		return nil, fmt.Errorf("link de ativação expirado")
	}
	return u, nil
}

// CompleteActivation sets the invited user's password and optional professor classes/subjects, then activates the account.
func (s *Service) CompleteActivation(token, password, classesCSV, subjectsCSV string) error {
	if len(password) < 6 {
		return fmt.Errorf("a senha deve ter pelo menos 6 caracteres")
	}

	u, err := s.ValidateActivationToken(token)
	if err != nil {
		return err
	}

	hashed, err := HashPassword(password)
	if err != nil {
		return err
	}

	return s.userRepo.ActivateWithPassword(u.ID, hashed, strings.TrimSpace(classesCSV), strings.TrimSpace(subjectsCSV))
}
