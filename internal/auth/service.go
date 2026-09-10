package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"laboratory-management/internal/user"
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

// ValidateSession validates a session token (placeholder for session management)
func (s *Service) ValidateSession(token string) (*user.User, error) {
	// In a real implementation, this would check the session against a storage mechanism
	// For now, this is a placeholder
	return nil, fmt.Errorf("session validation not implemented")
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

// CreateSession creates a new session for a user
func (s *Service) CreateSession(u *user.User) (*Session, error) {
	token, err := GenerateSessionToken()
	if err != nil {
		return nil, err
	}

	roleName := ""
	if u.Role != nil {
		roleName = u.Role.Name
	}

	session := &Session{
		Token:     token,
		UserID:    u.ID,
		UserEmail: u.Email,
		UserName:  u.Name,
		Role:      roleName,
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24 hour session
		CreatedAt: time.Now(),
	}

	return session, nil
}
