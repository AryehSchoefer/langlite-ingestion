package server

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(s.AuthMiddleware)

	if s.rateLimiter != nil {
		r.Use(s.rateLimiter.RateLimitMiddleware)
	}

	allowedOrigins := []string{"http://localhost:3000", "http://localhost:8080", "https://app.langlite.com"}
	if origins := os.Getenv("LANGLITE_CORS_ORIGINS"); origins != "" {
		allowedOrigins = strings.Split(origins, ",")
		for i := range allowedOrigins {
			allowedOrigins[i] = strings.TrimSpace(allowedOrigins[i])
		}
	}

	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Requested-With"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/", s.HelloWorldHandler)

	r.Get("/health", s.healthHandler)
	r.Get("/rate-limit-status", s.RateLimitStatusHandler)
	r.Post("/reset-rate-limit", s.ResetRateLimitHandler)

	// langlite ingestion routes
	r.Post("/api/v1/traces", s.CreateTrace)
	r.Post("/api/v1/generations", s.CreateGeneration)
	r.Post("/api/v1/spans", s.CreateSpan)
	r.Post("/api/v1/spans/{id}", s.UpdateSpan)
	r.Post("/api/v1/events", s.EventHandler)
	r.Post("/api/v1/scores", s.ScoreHandler)
	r.Post("/api/v1/batch", s.BatchHandler)

	return r
}

func (s *Server) HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	resp := make(map[string]string)
	resp["message"] = "Hello World"

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	_, _ = w.Write(jsonResp)
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	jsonResp, _ := json.Marshal(s.db.Health())
	_, _ = w.Write(jsonResp)
}
