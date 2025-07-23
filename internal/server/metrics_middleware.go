package server

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (s *Server) MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.metrics == nil {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()
		endpoint := r.URL.Path
		method := r.Method

		s.metrics.RecordHTTPRequestInFlight(method, endpoint, 1)
		defer s.metrics.RecordHTTPRequestInFlight(method, endpoint, -1)

		wrapped := &responseWriter{ResponseWriter: w, statusCode: 200}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		s.metrics.RecordHTTPRequest(method, endpoint, wrapped.statusCode, duration)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (s *Server) DatabaseMetricsCollector() {
	if s.metrics == nil {
		return
	}

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		stats := s.db.Health()

		if openStr, ok := stats["open_connections"]; ok {
			if open, err := strconv.ParseFloat(openStr, 64); err == nil {
				s.metrics.UpdateDatabaseConnections(open, 0, 0) // Simplified for now
			}
		}
	}
}

func (s *Server) QueueMetricsCollector() {
	if s.metrics == nil || s.queueClient == nil {
		return
	}

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		ctx := context.Background()
		stats, err := s.queueClient.GetQueueStats(ctx)
		if err != nil {
			continue
		}

		for queueName, depth := range stats {
			priority, jobType := parseQueueName(queueName)
			s.metrics.UpdateQueueDepth(queueName, priority, jobType, float64(depth))
		}
	}
}

func parseQueueName(queueName string) (priority, jobType string) {
	if queueName == "dead_letter" || queueName == "delayed" {
		return queueName, queueName
	}

	parts := strings.Split(queueName, ":")
	if len(parts) == 2 {
		return parts[0], parts[1]
	}

	return "unknown", "unknown"
}
