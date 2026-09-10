package config

import (
	"os"
	"strconv"
)

type Config struct {
	AppEnv        string
	AppPort       string
	DatabaseURL   string
	SessionSecret string
	CSRFSecret    string
}

func Load() *Config {
	return &Config{
		AppEnv:        getEnv("APP_ENV", "development"),
		AppPort:       getEnv("APP_PORT", "8080"),
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://user:password@localhost:5432/laboratory_management?sslmode=disable"),
		SessionSecret: getEnv("SESSION_SECRET", "change-this-secret-key"),
		CSRFSecret:    getEnv("CSRF_SECRET", "change-this-csrf-secret"),
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