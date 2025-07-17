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

	// Configure CORS with environment variables
	allowedOrigins := []string{"http://localhost:3000", "http://localhost:8080", "https://app.langlite.com"}
	if origins := os.Getenv("LANGLITE_CORS_ORIGINS"); origins != "" {
		allowedOrigins = strings.Split(origins, ",")
		for i := range allowedOrigins {
			allowedOrigins[i] = strings.TrimSpace(allowedOrigins[i])
		}
	}

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Requested-With"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/", s.HelloWorldHandler)

	r.Get("/health", s.healthHandler)

	// langlite ingestion routes
	r.Post("/v1/trace", s.CreateTrace)
	r.Post("/v1/generation", s.CreateGeneration)
	r.Post("/v1/span", s.CreateSpan)
	r.Post("/v1/span/{id}", s.CreateSpan)
	r.Post("/v1/trace/batch", s.TraceBatchHandler)
	r.Post("/v1/event", s.EventHandler)
	r.Post("/v1/score", s.ScoreHandler)

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
	health := s.db.Health()
	
	// Set appropriate status code based on health
	statusCode := http.StatusOK
	if health["status"] == "down" {
		statusCode = http.StatusServiceUnavailable
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	jsonResp, _ := json.Marshal(health)
	_, _ = w.Write(jsonResp)
}
