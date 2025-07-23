package database

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"
	"unicode"
)

type TraceRequest struct {
	ID        string         `json:"id,omitempty"`
	ProjectID string         `json:"project_id,omitempty"`
	Name      string         `json:"name"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	Tags      []string       `json:"tags,omitempty"`
	UserID    string         `json:"user_id,omitempty"`
	SessionID string         `json:"session_id,omitempty"`
	StartTime time.Time      `json:"start_time,omitempty"`
	EndTime   *time.Time     `json:"end_time,omitempty"`
}

func (tr TraceRequest) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)

	if strings.TrimSpace(tr.Name) == "" {
		problems["name"] = "name is required and cannot be empty"
	}

	if len(tr.Name) > 255 {
		problems["name"] = "name cannot exceed 255 characters"
	}

	if tr.EndTime != nil && !tr.StartTime.IsZero() && tr.EndTime.Before(tr.StartTime) {
		problems["end_time"] = "end_time cannot be before start_time"
	}

	if tr.UserID != "" && !isAlphanumeric(tr.UserID) {
		problems["user_id"] = "user_id must contain only letters and numbers"
	}

	for i, tag := range tr.Tags {
		if strings.TrimSpace(tag) == "" {
			problems[fmt.Sprintf("tags[%d]", i)] = "tag cannot be empty"
		}
	}

	return problems
}

type GenerationRequest struct {
	ID        string         `json:"id,omitempty"`
	TraceID   string         `json:"trace_id"`
	Name      string         `json:"name,omitempty"`
	Input     string         `json:"input"`
	Output    string         `json:"output,omitempty"`
	Model     string         `json:"model"`
	Usage     *UsageMetrics  `json:"usage,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	StartTime time.Time      `json:"start_time,omitempty"`
	EndTime   *time.Time     `json:"end_time,omitempty"`
}

func (gr GenerationRequest) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)

	if strings.TrimSpace(gr.TraceID) == "" {
		problems["trace_id"] = "trace_id is required"
	}

	if strings.TrimSpace(gr.Input) == "" {
		problems["input"] = "input is required"
	}

	if strings.TrimSpace(gr.Model) == "" {
		problems["model"] = "model is required"
	}

	return problems
}

type UsageMetrics struct {
	PromptTokens     int `json:"prompt_tokens,omitempty"`
	CompletionTokens int `json:"completion_tokens,omitempty"`
	TotalTokens      int `json:"total_tokens,omitempty"`
}

type SpanRequest struct {
	ID        string         `json:"id,omitempty"`
	TraceID   string         `json:"trace_id"`
	ParentID  string         `json:"parent_id,omitempty"`
	Name      string         `json:"name"`
	Type      string         `json:"type,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	StartTime time.Time      `json:"start_time,omitempty"`
	EndTime   *time.Time     `json:"end_time,omitempty"`
}

func (sr SpanRequest) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)

	if strings.TrimSpace(sr.TraceID) == "" {
		problems["trace_id"] = "trace_id is required"
	}

	if strings.TrimSpace(sr.Name) == "" {
		problems["name"] = "name is required and cannot be empty"
	}

	if len(sr.Name) > 255 {
		problems["name"] = "name cannot exceed 255 characters"
	}

	if sr.EndTime != nil && !sr.StartTime.IsZero() && sr.EndTime.Before(sr.StartTime) {
		problems["end_time"] = "end_time cannot be before start_time"
	}

	if sr.TraceID != "" && !isValidID(sr.TraceID) {
		problems["trace_id"] = "trace_id must be a valid UUID or alphanumeric string"
	}

	if sr.ParentID != "" && !isValidID(sr.ParentID) {
		problems["parent_id"] = "parent_id must be a valid UUID or alphanumeric string"
	}

	if sr.Type != "" {
		validTypes := map[string]bool{
			"db":      true,
			"http":    true,
			"llm":     true,
			"cache":   true,
			"auth":    true,
			"custom":  true,
			"compute": true,
		}
		if !validTypes[strings.ToLower(sr.Type)] {
			problems["type"] = "type must be one of: db, http, llm, cache, auth, custom, compute"
		}
	}

	return problems
}

type EventRequest struct {
	ID        string         `json:"id,omitempty"`
	TraceID   string         `json:"trace_id"`
	SpanID    string         `json:"span_id,omitempty"`
	Name      string         `json:"name"`
	Level     string         `json:"level,omitempty"`
	Message   string         `json:"message"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	Timestamp time.Time      `json:"timestamp,omitempty"`
}

func (er EventRequest) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)

	if strings.TrimSpace(er.TraceID) == "" {
		problems["trace_id"] = "trace_id is required"
	}

	if strings.TrimSpace(er.Name) == "" {
		problems["name"] = "name is required and cannot be empty"
	}

	if len(er.Name) > 255 {
		problems["name"] = "name cannot exceed 255 characters"
	}

	if strings.TrimSpace(er.Message) == "" {
		problems["message"] = "message is required and cannot be empty"
	}

	if len(er.Message) > 10000 {
		problems["message"] = "message cannot exceed 10000 characters"
	}

	if er.Level != "" {
		validLevels := map[string]bool{
			"debug": true,
			"info":  true,
			"warn":  true,
			"error": true,
		}
		if !validLevels[strings.ToLower(er.Level)] {
			problems["level"] = "level must be one of: debug, info, warn, error"
		}
	}

	if er.TraceID != "" && !isValidID(er.TraceID) {
		problems["trace_id"] = "trace_id must be a valid UUID or alphanumeric string"
	}

	if er.SpanID != "" && !isValidID(er.SpanID) {
		problems["span_id"] = "span_id must be a valid UUID or alphanumeric string"
	}

	return problems
}

type ScoreRequest struct {
	ID           string         `json:"id,omitempty"`
	TraceID      string         `json:"trace_id,omitempty"`
	GenerationID string         `json:"generation_id,omitempty"`
	Name         string         `json:"name"`
	Value        float64        `json:"value"`
	Source       string         `json:"source,omitempty"`
	Comment      string         `json:"comment,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty"`
	Timestamp    time.Time      `json:"timestamp,omitempty"`
}

func (scr ScoreRequest) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)

	if strings.TrimSpace(scr.Name) == "" {
		problems["name"] = "name is required and cannot be empty"
	}

	if len(scr.Name) > 255 {
		problems["name"] = "name cannot exceed 255 characters"
	}

	if strings.TrimSpace(scr.TraceID) == "" && strings.TrimSpace(scr.GenerationID) == "" {
		problems["trace_id"] = "either trace_id or generation_id must be provided"
	}

	if math.IsNaN(scr.Value) || math.IsInf(scr.Value, 0) {
		problems["value"] = "value must be a valid number"
	}

	if scr.Value < 0 || scr.Value > 1 {
		problems["value"] = "value must be between 0 and 1"
	}

	if scr.Source != "" {
		validSources := map[string]bool{
			"human":     true,
			"llm":       true,
			"heuristic": true,
			"automated": true,
		}
		if !validSources[strings.ToLower(scr.Source)] {
			problems["source"] = "source must be one of: human, llm, heuristic, automated"
		}
	}

	if scr.TraceID != "" && !isValidID(scr.TraceID) {
		problems["trace_id"] = "trace_id must be a valid UUID or alphanumeric string"
	}

	if scr.GenerationID != "" && !isValidID(scr.GenerationID) {
		problems["generation_id"] = "generation_id must be a valid UUID or alphanumeric string"
	}

	if len(scr.Comment) > 1000 {
		problems["comment"] = "comment cannot exceed 1000 characters"
	}

	return problems
}

type GenerationUpdateRequest struct {
	Output   string         `json:"output,omitempty"`
	Usage    *UsageMetrics  `json:"usage,omitempty"`
	EndTime  *time.Time     `json:"end_time,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

func (gur GenerationUpdateRequest) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)

	hasUpdate := false

	if gur.Output != "" {
		hasUpdate = true
		if len(gur.Output) > 100000 {
			problems["output"] = "output cannot exceed 100000 characters"
		}
	}

	if gur.Usage != nil {
		hasUpdate = true
		if gur.Usage.PromptTokens < 0 {
			problems["usage.prompt_tokens"] = "prompt_tokens cannot be negative"
		}
		if gur.Usage.CompletionTokens < 0 {
			problems["usage.completion_tokens"] = "completion_tokens cannot be negative"
		}
		if gur.Usage.TotalTokens < 0 {
			problems["usage.total_tokens"] = "total_tokens cannot be negative"
		}
		if gur.Usage.TotalTokens > 0 && gur.Usage.PromptTokens > 0 && gur.Usage.CompletionTokens > 0 {
			if gur.Usage.TotalTokens != gur.Usage.PromptTokens+gur.Usage.CompletionTokens {
				problems["usage.total_tokens"] = "total_tokens should equal prompt_tokens + completion_tokens"
			}
		}
	}

	if gur.EndTime != nil {
		hasUpdate = true

		if gur.EndTime.After(time.Now().UTC()) {
			problems["end_time"] = "end_time cannot be in the future"
		}
	}

	if gur.Metadata != nil {
		hasUpdate = true

		if len(gur.Metadata) > 100 {
			problems["metadata"] = "metadata cannot have more than 100 keys"
		}
	}

	if !hasUpdate {
		problems["update"] = "at least one field must be provided for update"
	}

	return problems
}

type SpanUpdateRequest struct {
	EndTime  *time.Time     `json:"end_time,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

func (sur SpanUpdateRequest) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)

	hasUpdate := false

	if sur.EndTime != nil {
		hasUpdate = true
		if sur.EndTime.After(time.Now().UTC()) {
			problems["end_time"] = "end_time cannot be in the future"
		}
	}

	if sur.Metadata != nil {
		hasUpdate = true
		if len(sur.Metadata) > 100 {
			problems["metadata"] = "metadata cannot have more than 100 keys"
		}
	}

	if !hasUpdate {
		problems["update"] = "at least one field must be provided for update"
	}

	return problems
}

type SuccessResponse struct {
	ID      string `json:"id"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

type ErrorResponse struct {
	Error    string            `json:"error"`
	Message  string            `json:"message"`
	Code     int               `json:"code"`
	Problems map[string]string `json:"problems,omitempty"`
}

func isAlphanumeric(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func isValidID(id string) bool {
	if id == "" {
		return false
	}

	if len(id) == 36 && strings.Count(id, "-") == 4 {
		return true
	}

	for _, r := range id {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '-' && r != '_' {
			return false
		}
	}

	return len(id) >= 3 && len(id) <= 255
}

type BatchTraceRequest struct {
	Traces []TraceRequest `json:"traces"`
}

func (btr BatchTraceRequest) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)

	// Validate batch size
	if len(btr.Traces) == 0 {
		problems["traces"] = "at least one trace must be provided"
	}

	if len(btr.Traces) > 100 {
		problems["traces"] = "maximum 100 traces allowed per batch"
	}

	return problems
}

type BatchResult struct {
	Index  int    `json:"index"`
	ID     string `json:"id,omitempty"`
	Status string `json:"status"` // "success", "error"
	Error  string `json:"error,omitempty"`
}

type BatchResponse struct {
	Results []BatchResult `json:"results"`
	Summary BatchSummary  `json:"summary"`
}

type BatchSummary struct {
	Total     int `json:"total"`
	Succeeded int `json:"succeeded"`
	Failed    int `json:"failed"`
}

type BatchRequest struct {
	Traces      []TraceRequest      `json:"traces,omitempty"`
	Spans       []SpanRequest       `json:"spans,omitempty"`
	Generations []GenerationRequest `json:"generations,omitempty"`
	Events      []EventRequest      `json:"events,omitempty"`
	Scores      []ScoreRequest      `json:"scores,omitempty"`
}

func (br BatchRequest) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)

	totalItems := len(br.Traces) + len(br.Spans) + len(br.Generations) + len(br.Events) + len(br.Scores)

	if totalItems == 0 {
		problems["batch"] = "at least one item must be provided"
	}

	if totalItems > 1000 {
		problems["batch"] = "maximum 1000 items allowed per batch"
	}

	return problems
}

type Project struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type APIKey struct {
	ID                 string     `json:"id"`
	ProjectID          string     `json:"project_id"`
	KeyHash            string     `json:"-"`
	Name               string     `json:"name"`
	LastUsedAt         *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt          *time.Time `json:"expires_at,omitempty"`
	RateLimitPerMinute int        `json:"rate_limit_per_minute"`
	RateLimitPerHour   int        `json:"rate_limit_per_hour"`
	IsActive           bool       `json:"is_active"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

type AuthContext struct {
	ProjectID string
	APIKeyID  string
}
