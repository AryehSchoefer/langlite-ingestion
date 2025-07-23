package queue

import (
	"encoding/json"
	"time"

	"langlite-ingestion/internal/database"
)

type JobType string

const (
	JobTypeEnrichTrace     JobType = "enrich_trace"
	JobTypeStoreRaw        JobType = "store_raw"
	JobTypeAnalyticsExport JobType = "analytics_export"
)

type QueuePriority string

const (
	QueueHigh   QueuePriority = "high"
	QueueMedium QueuePriority = "medium"
	QueueLow    QueuePriority = "low"
)

type Job struct {
	ID          string                 `json:"id"`
	Type        JobType                `json:"type"`
	Priority    QueuePriority          `json:"priority"`
	Payload     map[string]interface{} `json:"payload"`
	CreatedAt   time.Time              `json:"created_at"`
	ProcessedAt *time.Time             `json:"processed_at,omitempty"`
	Attempts    int                    `json:"attempts"`
	MaxAttempts int                    `json:"max_attempts"`
	Error       string                 `json:"error,omitempty"`
}

type JobPayload struct {
	// Common fields
	ProjectID string `json:"project_id"`
	TraceID   string `json:"trace_id,omitempty"`

	// For enrich_trace jobs
	TraceData *database.TraceRequest `json:"trace_data,omitempty"`

	// For store_raw jobs
	RawData  interface{} `json:"raw_data,omitempty"`
	DataType string      `json:"data_type,omitempty"` // "trace", "span", "generation", etc.

	// For analytics_export jobs
	ExportData interface{} `json:"export_data,omitempty"`
	ExportType string      `json:"export_type,omitempty"` // "clickhouse", "warehouse", etc.
}

func (j *Job) ToJSON() (string, error) {
	data, err := json.Marshal(j)
	return string(data), err
}

func FromJSON(data string) (*Job, error) {
	var job Job
	err := json.Unmarshal([]byte(data), &job)
	return &job, err
}

func GetQueueName(jobType JobType, priority QueuePriority) string {
	return string(priority) + ":" + string(jobType)
}

type JobResult struct {
	Success     bool          `json:"success"`
	Error       string        `json:"error,omitempty"`
	Data        interface{}   `json:"data,omitempty"`
	Duration    time.Duration `json:"duration"`
	ProcessedAt time.Time     `json:"processed_at"`
}
