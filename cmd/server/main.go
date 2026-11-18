package main

import (
	"log"
	"net/http"
	"os"

	"laboratory-management/internal/database"
	"laboratory-management/internal/middleware"
	"laboratory-management/internal/web"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Failed dot to env %v", err)
		}
	}

	// Initialize database
	db, err := database.Initialize()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Run database migrations
	if err := database.RunMigrations(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize router
	router := web.NewRouter(db)

	// Apply middleware
	router.Use(middleware.Logging)
	router.Use(middleware.Session)
	router.Use(middleware.CSRF)

	// Get port from environment or use default
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
