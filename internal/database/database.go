package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
)

// Service represents a service that interacts with a database.
type Service interface {
	Health() map[string]string
	CreateTrace(TraceRequest) error
	CreateGeneration(GenerationRequest) error
	CreateSpan(SpanRequest) error
	UpdateSpan(spanID string, req SpanUpdateRequest) error
	TraceExists(traceID string) bool
	SpanExists(spanID string) bool
	CreateEvent(EventRequest) error
	CreateScore(ScoreRequest) error
	GenerationExists(generationID string) bool

	Close() error
}

type service struct {
	db *sql.DB
}

var (
	database   = os.Getenv("LANGLITE_DB_DATABASE")
	password   = os.Getenv("LANGLITE_DB_PASSWORD")
	username   = os.Getenv("LANGLITE_DB_USERNAME")
	port       = os.Getenv("LANGLITE_DB_PORT")
	host       = os.Getenv("LANGLITE_DB_HOST")
	schema     = os.Getenv("LANGLITE_DB_SCHEMA")
	dbInstance *service
)

func New() Service {
	if dbInstance != nil {
		return dbInstance
	}
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&search_path=%s", username, password, host, port, database, schema)
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatal(err)
	}
	dbInstance = &service{
		db: db,
	}
	return dbInstance
}

func (s *service) CreateTrace(tr TraceRequest) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `INSERT INTO traces (id, name, metadata, tags, user_id, session_id, start_time, end_time)
          VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	var metadata []byte
	var err error
	if tr.Metadata != nil {
		metadata, err = json.Marshal(tr.Metadata)
		if err != nil {
			return fmt.Errorf("Failed to marshal metadata: %w", err)
		}
	}

	_, err = s.db.ExecContext(ctx, query, tr.ID, tr.Name, metadata, tr.Tags, tr.UserID, tr.SessionID, tr.StartTime, tr.EndTime)
	if err != nil {
		return fmt.Errorf("Failed to create trace: %w", err)
	}

	return nil
}

func (s *service) CreateGeneration(gr GenerationRequest) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `INSERT INTO generations (id, trace_id, name, input, output, model, prompt_tokens, completion_tokes, total_tokens, metadata, start_time, end_time)
          VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	var metadata []byte
	var err error
	if gr.Metadata != nil {
		metadata, err = json.Marshal(gr.Metadata)
		if err != nil {
			return fmt.Errorf("Failed to marshal metadata: %w", err)
		}
	}

	var promptTokens, completionTokens, totalTokens interface{}
	if gr.Usage != nil {
		promptTokens = gr.Usage.PromptTokens
		completionTokens = gr.Usage.CompletionTokens
		totalTokens = gr.Usage.TotalTokens
	}

	_, err = s.db.ExecContext(ctx, query, gr.ID, gr.TraceID, gr.Name, gr.Input, gr.Output, gr.Model,
		promptTokens, completionTokens, totalTokens, metadata, gr.StartTime, gr.EndTime)
	if err != nil {
		return fmt.Errorf("failed to create generation: %w", err)
	}

	return nil
}

func (s *service) CreateSpan(sr SpanRequest) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `INSERT INTO spans (id, trace_id, parent_id, name, type, metadata, start_time, end_time)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	var metadata []byte
	var err error
	if sr.Metadata != nil {
		metadata, err = json.Marshal(sr.Metadata)
		if err != nil {
			return fmt.Errorf("Failed to marshal metadata: %w", err)
		}
	}

	var parentID interface{}
	if sr.ParentID != "" {
		parentID = sr.ParentID
	}

	_, err = s.db.ExecContext(ctx, query, sr.ID, sr.TraceID, parentID, sr.Name, sr.Type, metadata, sr.StartTime, sr.EndTime)
	if err != nil {
		return fmt.Errorf("Failed to create span: %w", err)
	}

	return nil
}

func (s *service) UpdateSpan(spanID string, req SpanUpdateRequest) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var exists bool
	err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM spans WHERE id = $1)", spanID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("Failed to check span existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("Span not found")
	}

	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.EndTime != nil {
		setParts = append(setParts, fmt.Sprintf("end_time = $%d", argIndex))
		args = append(args, *req.EndTime)
		argIndex++
	}

	if req.Metadata != nil {
		metadata, err := json.Marshal(req.Metadata)
		if err != nil {
			return fmt.Errorf("Failed to marshal metadata: %w", err)
		}
		setParts = append(setParts, fmt.Sprintf("metadata = $%d", argIndex))
		args = append(args, metadata)
		argIndex++
	}

	if len(setParts) == 0 {
		return fmt.Errorf("No fields to update")
	}

	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now().UTC())
	argIndex++

	query := fmt.Sprintf("UPDATE spans SET %s WHERE id = $%d",
		strings.Join(setParts, ", "), argIndex)

	args = append(args, spanID)

	_, err = s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("Failed to update span: %w", err)
	}

	return nil
}

func (s *service) TraceExists(traceID string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM traces WHERE id = $1)"
	err := s.db.QueryRowContext(ctx, query, traceID).Scan(&exists)
	if err != nil {
		log.Printf("Error checking trace existence: %v", err)
		return false
	}

	return exists
}

func (s *service) SpanExists(spanID string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM spans WHERE id = $1)"
	err := s.db.QueryRowContext(ctx, query, spanID).Scan(&exists)
	if err != nil {
		log.Printf("Error checking span existence: %v", err)
		return false
	}
	return exists
}

func (s *service) CreateEvent(er EventRequest) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `INSERT INTO events (id, trace_id, span_id, name, level, message, metadata, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	var metadata []byte
	var err error
	if er.Metadata != nil {
		metadata, err = json.Marshal(er.Metadata)
		if err != nil {
			return fmt.Errorf("Failed to marshal metadata: %w", err)
		}
	}

	var spanID interface{}
	if er.SpanID != "" {
		spanID = er.SpanID
	}

	_, err = s.db.ExecContext(ctx, query, er.ID, er.TraceID, spanID, er.Name, er.Level, er.Message, metadata, er.Timestamp)
	if err != nil {
		return fmt.Errorf("Failed to create event: %w", err)
	}

	return nil
}

func (s *service) CreateScore(scr ScoreRequest) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `INSERT INTO scores (id, trace_id, generation_id, name, value, source, comment, metadata, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	var metadata []byte
	var err error
	if scr.Metadata != nil {
		metadata, err = json.Marshal(scr.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	var traceID interface{}
	if scr.TraceID != "" {
		traceID = scr.TraceID
	}

	var generationID interface{}
	if scr.GenerationID != "" {
		generationID = scr.GenerationID
	}

	_, err = s.db.ExecContext(ctx, query, scr.ID, traceID, generationID, scr.Name, scr.Value, scr.Source, scr.Comment, metadata, scr.Timestamp)
	if err != nil {
		return fmt.Errorf("failed to create score: %w", err)
	}

	return nil
}

func (s *service) GenerationExists(generationID string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM generations WHERE id = $1)"
	err := s.db.QueryRowContext(ctx, query, generationID).Scan(&exists)
	if err != nil {
		log.Printf("Error checking generation existence: %v", err)
		return false
	}

	return exists
}

func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stats := make(map[string]string)

	err := s.db.PingContext(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		log.Fatalf("db down: %v", err)
		return stats
	}

	stats["status"] = "up"
	stats["message"] = "It's healthy"

	dbStats := s.db.Stats()
	stats["open_connections"] = strconv.Itoa(dbStats.OpenConnections)
	stats["in_use"] = strconv.Itoa(dbStats.InUse)
	stats["idle"] = strconv.Itoa(dbStats.Idle)
	stats["wait_count"] = strconv.FormatInt(dbStats.WaitCount, 10)
	stats["wait_duration"] = dbStats.WaitDuration.String()
	stats["max_idle_closed"] = strconv.FormatInt(dbStats.MaxIdleClosed, 10)
	stats["max_lifetime_closed"] = strconv.FormatInt(dbStats.MaxLifetimeClosed, 10)

	// Evaluate stats to provide a health message
	if dbStats.OpenConnections > 40 { // Assuming 50 is the max for this example
		stats["message"] = "The database is experiencing heavy load."
	}

	if dbStats.WaitCount > 1000 {
		stats["message"] = "The database has a high number of wait events, indicating potential bottlenecks."
	}

	if dbStats.MaxIdleClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many idle connections are being closed, consider revising the connection pool settings."
	}

	if dbStats.MaxLifetimeClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many connections are being closed due to max lifetime, consider increasing max lifetime or revising the connection usage pattern."
	}

	return stats
}

func (s *service) Close() error {
	log.Printf("Disconnected from database: %s", database)
	return s.db.Close()
}
