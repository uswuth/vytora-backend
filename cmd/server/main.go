package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/uswuth/vytora-backend/internal/config"
	"github.com/uswuth/vytora-backend/internal/database"
	"github.com/uswuth/vytora-backend/internal/handlers"
	middlewareauth "github.com/uswuth/vytora-backend/internal/middleware/auth"
	"github.com/uswuth/vytora-backend/internal/repository"
	"github.com/uswuth/vytora-backend/internal/services"
)

func main() {
	// Load config
	cfg := config.Load()

	// Connect to database
	if err := database.Connect(cfg.DatabaseURL); err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer database.Close()

	// Repositories
	userRepo := repository.NewUserRepository(database.Pool)

	// Services
	jwtService := services.NewJWTService(cfg.JWTSecret, cfg.JWTExpiryHours)

	// Handlers
	authHandler := handlers.NewAuthHandler(userRepo, jwtService)

	// Router
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.RequestID)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Public routes
	r.Post("/api/v1/login", authHandler.Login)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Protected routes (example)
	r.Group(func(r chi.Router) {
		r.Use(middlewareauth.AuthMiddleware(jwtService))
		r.Get("/api/v1/me", func(w http.ResponseWriter, r *http.Request) {
			claims := r.Context().Value(middlewareauth.UserContextKey).(*services.Claims)
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"user_id":"%s","code":"%s","email":"%s","role":"%s"}`,
				claims.UserID, claims.Code, claims.Email, claims.Role)
		})
	})

	fmt.Println("VRMP server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
