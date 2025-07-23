package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"langlite-ingestion/internal/database"
)

type JobProcessor interface {
	Process(ctx context.Context, job *Job) (*JobResult, error)
	CanProcess(jobType JobType) bool
}

type EnrichTraceProcessor struct {
	db database.Service
}

func NewEnrichTraceProcessor(db database.Service) *EnrichTraceProcessor {
	return &EnrichTraceProcessor{db: db}
}

func (p *EnrichTraceProcessor) CanProcess(jobType JobType) bool {
	return jobType == JobTypeEnrichTrace
}

func (p *EnrichTraceProcessor) Process(ctx context.Context, job *Job) (*JobResult, error) {
	start := time.Now()

	traceData, err := p.extractTraceData(job.Payload)
	if err != nil {
		return &JobResult{
			Success:     false,
			Error:       fmt.Sprintf("failed to extract trace data: %v", err),
			Duration:    time.Since(start),
			ProcessedAt: time.Now().UTC(),
		}, nil
	}

	enrichedData, err := p.enrichTrace(ctx, traceData)
	if err != nil {
		return &JobResult{
			Success:     false,
			Error:       fmt.Sprintf("failed to enrich trace: %v", err),
			Duration:    time.Since(start),
			ProcessedAt: time.Now().UTC(),
		}, nil
	}

	return &JobResult{
		Success:     true,
		Data:        enrichedData,
		Duration:    time.Since(start),
		ProcessedAt: time.Now().UTC(),
	}, nil
}

func (p *EnrichTraceProcessor) extractTraceData(payload map[string]interface{}) (*database.TraceRequest, error) {
	traceDataRaw, exists := payload["trace_data"]
	if !exists {
		return nil, fmt.Errorf("trace_data not found in payload")
	}

	jsonData, err := json.Marshal(traceDataRaw)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal trace data: %w", err)
	}

	var traceData database.TraceRequest
	err = json.Unmarshal(jsonData, &traceData)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal trace data: %w", err)
	}

	return &traceData, nil
}

func (p *EnrichTraceProcessor) enrichTrace(ctx context.Context, trace *database.TraceRequest) (map[string]interface{}, error) {
	enriched := make(map[string]interface{})

	enriched["server_received_at"] = time.Now().UTC()
	enriched["enrichment_version"] = "1.0"

	if trace.EndTime != nil && !trace.StartTime.IsZero() {
		duration := trace.EndTime.Sub(trace.StartTime)
		enriched["duration_ms"] = duration.Milliseconds()
	}

	enriched["trace_id"] = trace.ID
	enriched["project_id"] = trace.ProjectID

	if trace.UserID != "" {
		enriched["user_context"] = map[string]interface{}{
			"user_id": trace.UserID,
		}
	}

	if trace.SessionID != "" {
		enriched["session_context"] = map[string]interface{}{
			"session_id": trace.SessionID,
		}
	}

	if trace.Metadata != nil {
		enriched["original_metadata"] = trace.Metadata
	}

	return enriched, nil
}

type StoreRawProcessor struct {
	db database.Service
}

func NewStoreRawProcessor(db database.Service) *StoreRawProcessor {
	return &StoreRawProcessor{db: db}
}

func (p *StoreRawProcessor) CanProcess(jobType JobType) bool {
	return jobType == JobTypeStoreRaw
}

func (p *StoreRawProcessor) Process(ctx context.Context, job *Job) (*JobResult, error) {
	start := time.Now()

	rawData, dataType, err := p.extractRawData(job.Payload)
	if err != nil {
		return &JobResult{
			Success:     false,
			Error:       fmt.Sprintf("failed to extract raw data: %v", err),
			Duration:    time.Since(start),
			ProcessedAt: time.Now().UTC(),
		}, nil
	}

	err = p.storeByType(ctx, dataType, rawData)
	if err != nil {
		return &JobResult{
			Success:     false,
			Error:       fmt.Sprintf("failed to store %s: %v", dataType, err),
			Duration:    time.Since(start),
			ProcessedAt: time.Now().UTC(),
		}, nil
	}

	return &JobResult{
		Success:     true,
		Data:        map[string]interface{}{"stored_type": dataType},
		Duration:    time.Since(start),
		ProcessedAt: time.Now().UTC(),
	}, nil
}

func (p *StoreRawProcessor) extractRawData(payload map[string]interface{}) (interface{}, string, error) {
	rawData, exists := payload["raw_data"]
	if !exists {
		return nil, "", fmt.Errorf("raw_data not found in payload")
	}

	dataType, exists := payload["data_type"]
	if !exists {
		return nil, "", fmt.Errorf("data_type not found in payload")
	}

	dataTypeStr, ok := dataType.(string)
	if !ok {
		return nil, "", fmt.Errorf("data_type must be a string")
	}

	return rawData, dataTypeStr, nil
}

func (p *StoreRawProcessor) storeByType(ctx context.Context, dataType string, rawData interface{}) error {
	switch dataType {
	case "trace":
		return p.storeTrace(ctx, rawData)
	case "span":
		return p.storeSpan(ctx, rawData)
	case "generation":
		return p.storeGeneration(ctx, rawData)
	case "event":
		return p.storeEvent(ctx, rawData)
	case "score":
		return p.storeScore(ctx, rawData)
	default:
		return fmt.Errorf("unsupported data type: %s", dataType)
	}
}

func (p *StoreRawProcessor) storeTrace(ctx context.Context, rawData interface{}) error {
	jsonData, err := json.Marshal(rawData)
	if err != nil {
		return fmt.Errorf("failed to marshal trace data: %w", err)
	}

	var trace database.TraceRequest
	err = json.Unmarshal(jsonData, &trace)
	if err != nil {
		return fmt.Errorf("failed to unmarshal trace data: %w", err)
	}

	return p.db.CreateTrace(trace)
}

func (p *StoreRawProcessor) storeSpan(ctx context.Context, rawData interface{}) error {
	jsonData, err := json.Marshal(rawData)
	if err != nil {
		return fmt.Errorf("failed to marshal span data: %w", err)
	}

	var span database.SpanRequest
	err = json.Unmarshal(jsonData, &span)
	if err != nil {
		return fmt.Errorf("failed to unmarshal span data: %w", err)
	}

	return p.db.CreateSpan(span)
}

func (p *StoreRawProcessor) storeGeneration(ctx context.Context, rawData interface{}) error {
	jsonData, err := json.Marshal(rawData)
	if err != nil {
		return fmt.Errorf("failed to marshal generation data: %w", err)
	}

	var generation database.GenerationRequest
	err = json.Unmarshal(jsonData, &generation)
	if err != nil {
		return fmt.Errorf("failed to unmarshal generation data: %w", err)
	}

	return p.db.CreateGeneration(generation)
}

func (p *StoreRawProcessor) storeEvent(ctx context.Context, rawData interface{}) error {
	jsonData, err := json.Marshal(rawData)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	var event database.EventRequest
	err = json.Unmarshal(jsonData, &event)
	if err != nil {
		return fmt.Errorf("failed to unmarshal event data: %w", err)
	}

	return p.db.CreateEvent(event)
}

func (p *StoreRawProcessor) storeScore(ctx context.Context, rawData interface{}) error {
	jsonData, err := json.Marshal(rawData)
	if err != nil {
		return fmt.Errorf("failed to marshal score data: %w", err)
	}

	var score database.ScoreRequest
	err = json.Unmarshal(jsonData, &score)
	if err != nil {
		return fmt.Errorf("failed to unmarshal score data: %w", err)
	}

	return p.db.CreateScore(score)
}

// TODO: The following is not implemented yet and just placeholder

type AnalyticsExportProcessor struct {
	// this should have ClickHouse client
}

func NewAnalyticsExportProcessor() *AnalyticsExportProcessor {
	return &AnalyticsExportProcessor{}
}

func (p *AnalyticsExportProcessor) CanProcess(jobType JobType) bool {
	return jobType == JobTypeAnalyticsExport
}

func (p *AnalyticsExportProcessor) Process(ctx context.Context, job *Job) (*JobResult, error) {
	start := time.Now()

	// For now, this is a placeholder implementation:
	// In production, this would export to ClickHouse or other analytics systems

	exportData, exportType, err := p.extractExportData(job.Payload)
	if err != nil {
		return &JobResult{
			Success:     false,
			Error:       fmt.Sprintf("failed to extract export data: %v", err),
			Duration:    time.Since(start),
			ProcessedAt: time.Now().UTC(),
		}, nil
	}

	// Simulate export processing
	err = p.simulateExport(ctx, exportType, exportData)
	if err != nil {
		return &JobResult{
			Success:     false,
			Error:       fmt.Sprintf("failed to export to %s: %v", exportType, err),
			Duration:    time.Since(start),
			ProcessedAt: time.Now().UTC(),
		}, nil
	}

	return &JobResult{
		Success:     true,
		Data:        map[string]interface{}{"exported_to": exportType},
		Duration:    time.Since(start),
		ProcessedAt: time.Now().UTC(),
	}, nil
}

func (p *AnalyticsExportProcessor) extractExportData(payload map[string]interface{}) (interface{}, string, error) {
	exportData, exists := payload["export_data"]
	if !exists {
		return nil, "", fmt.Errorf("export_data not found in payload")
	}

	exportType, exists := payload["export_type"]
	if !exists {
		exportType = "clickhouse" // Default export type
	}

	exportTypeStr, ok := exportType.(string)
	if !ok {
		return nil, "", fmt.Errorf("export_type must be a string")
	}

	return exportData, exportTypeStr, nil
}

func (p *AnalyticsExportProcessor) simulateExport(ctx context.Context, exportType string, data interface{}) error {
	time.Sleep(100 * time.Millisecond)

	// In production, this would:
	// 1. Transform data for analytics format
	// 2. Connect to ClickHouse/warehouse
	// 3. Batch insert the data
	// 4. Handle export errors and retries

	fmt.Printf("Simulated export to %s completed\n", exportType)
	return nil
}
