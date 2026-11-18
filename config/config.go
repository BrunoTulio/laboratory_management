package config

import (
	"os"
	"strconv"
)

type Config struct {
	AppEnv        string
	AppPort       string
	BaseURL       string
	DatabasePath  string
	SessionSecret string
	CSRFSecret    string
	AdminName     string
	AdminEmail    string
	AdminPassword string
	SMTPHost      string
	SMTPPort      string
	SMTPUser      string
	SMTPPassword  string
	SMTPFrom      string
}

func Load() *Config {
	port := getEnv("APP_PORT", "8080")
	return &Config{
		AppEnv:        getEnv("APP_ENV", "development"),
		AppPort:       port,
		BaseURL:       getEnv("APP_BASE_URL", "http://localhost:"+port),
		DatabasePath:  getEnv("DATABASE_PATH", "laboratory.db"),
		SessionSecret: getEnv("SESSION_SECRET", "change-this-secret-key"),
		CSRFSecret:    getEnv("CSRF_SECRET", "change-this-csrf-secret"),
		AdminName:     getEnv("ADMIN_NAME", "Coordenação CETEP"),
		AdminEmail:    getEnv("ADMIN_EMAIL", "admin@cetep.local"),
		AdminPassword: getEnv("ADMIN_PASSWORD", "Admin@123"),
		SMTPHost:      getEnv("SMTP_HOST", ""),
		SMTPPort:      getEnv("SMTP_PORT", "587"),
		SMTPUser:      getEnv("SMTP_USER", ""),
		SMTPPassword:  getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:      getEnv("SMTP_FROM", "noreply@cetep.local"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

func IsDevelopment() bool {
	return Load().AppEnv == "development"
}

func IsProduction() bool {
	return Load().AppEnv == "production"
}
