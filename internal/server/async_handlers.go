package server

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"langlite-ingestion/internal/database"
	"langlite-ingestion/internal/queue"
)

func (s *Server) CreateTraceAsync(w http.ResponseWriter, r *http.Request) {
	authCtx, ok := GetAuthContext(r)
	if !ok {
		errorResp := database.ErrorResponse{
			Error:   "Authentication required",
			Message: "Valid API key required",
			Code:    http.StatusUnauthorized,
		}
		encode(w, r, http.StatusUnauthorized, errorResp)
		return
	}

	req, problems, err := decodeValid[database.TraceRequest](r)
	if err != nil {
		if len(problems) > 0 {
			errorResp := database.ErrorResponse{
				Error:    "Validation failed",
				Message:  "The request contains invalid data",
				Code:     http.StatusBadRequest,
				Problems: problems,
			}
			encode(w, r, http.StatusBadRequest, errorResp)
			return
		}

		errorResp := database.ErrorResponse{
			Error:   "Invalid request",
			Message: "Could not parse request body",
			Code:    http.StatusBadRequest,
		}
		encode(w, r, http.StatusBadRequest, errorResp)
		return
	}

	req.ProjectID = authCtx.ProjectID

	if req.ID == "" {
		req.ID = uuid.New().String()
	}

	if req.StartTime.IsZero() {
		req.StartTime = time.Now().UTC()
	}

	if s.queueClient == nil {
		if err := s.db.CreateTrace(req); err != nil {
			errorResp := database.ErrorResponse{
				Error:   "Database error",
				Message: "Failed to create trace",
				Code:    http.StatusInternalServerError,
			}
			encode(w, r, http.StatusInternalServerError, errorResp)
			return
		}
	} else {
		err = s.enqueueTraceJobs(r, req)
		if err != nil {
			if err := s.db.CreateTrace(req); err != nil {
				errorResp := database.ErrorResponse{
					Error:   "Processing error",
					Message: "Failed to process trace",
					Code:    http.StatusInternalServerError,
				}
				encode(w, r, http.StatusInternalServerError, errorResp)
				return
			}
		}
	}

	response := database.SuccessResponse{
		ID:     req.ID,
		Status: "accepted",
	}

	encode(w, r, http.StatusAccepted, response)
}

func (s *Server) enqueueTraceJobs(r *http.Request, req database.TraceRequest) error {
	ctx := r.Context()

	// Job 1: Store raw data (high priority)
	storePayload := map[string]interface{}{
		"project_id": req.ProjectID,
		"trace_id":   req.ID,
		"raw_data":   req,
		"data_type":  "trace",
	}

	_, err := s.queueClient.Enqueue(ctx, queue.JobTypeStoreRaw, queue.QueueHigh, storePayload)
	if err != nil {
		return err
	}

	// Job 2: Enrich trace data (medium priority)
	enrichPayload := map[string]interface{}{
		"project_id": req.ProjectID,
		"trace_id":   req.ID,
		"trace_data": req,
	}

	_, err = s.queueClient.Enqueue(ctx, queue.JobTypeEnrichTrace, queue.QueueMedium, enrichPayload)
	if err != nil {
		return err
	}

	// Job 3: Export to analytics (low priority)
	analyticsPayload := map[string]interface{}{
		"project_id":  req.ProjectID,
		"trace_id":    req.ID,
		"export_data": req,
		"export_type": "clickhouse",
	}

	_, err = s.queueClient.Enqueue(ctx, queue.JobTypeAnalyticsExport, queue.QueueLow, analyticsPayload)
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) CreateGenerationAsync(w http.ResponseWriter, r *http.Request) {
	req, problems, err := decodeValid[database.GenerationRequest](r)
	if err != nil {
		if len(problems) > 0 {
			errorResp := database.ErrorResponse{
				Error:    "Validation failed",
				Message:  "The request contains invalid data",
				Code:     http.StatusBadRequest,
				Problems: problems,
			}
			encode(w, r, http.StatusBadRequest, errorResp)
			return
		}

		errorResp := database.ErrorResponse{
			Error:   "Invalid request",
			Message: "Could not parse request body",
			Code:    http.StatusBadRequest,
		}
		encode(w, r, http.StatusBadRequest, errorResp)
		return
	}

	if req.ID == "" {
		req.ID = uuid.New().String()
	}

	if req.StartTime.IsZero() {
		req.StartTime = time.Now().UTC()
	}

	if s.queueClient == nil {
		if err := s.db.CreateGeneration(req); err != nil {
			errorResp := database.ErrorResponse{
				Error:   "Database error",
				Message: "Failed to create generation",
				Code:    http.StatusInternalServerError,
			}
			encode(w, r, http.StatusInternalServerError, errorResp)
			return
		}
	} else {
		err = s.enqueueGenerationJobs(r, req)
		if err != nil {
			if err := s.db.CreateGeneration(req); err != nil {
				errorResp := database.ErrorResponse{
					Error:   "Processing error",
					Message: "Failed to process generation",
					Code:    http.StatusInternalServerError,
				}
				encode(w, r, http.StatusInternalServerError, errorResp)
				return
			}
		}
	}

	response := database.SuccessResponse{
		ID:     req.ID,
		Status: "accepted",
	}

	encode(w, r, http.StatusAccepted, response)
}

func (s *Server) enqueueGenerationJobs(r *http.Request, req database.GenerationRequest) error {
	ctx := r.Context()

	storePayload := map[string]interface{}{
		"trace_id":  req.TraceID,
		"raw_data":  req,
		"data_type": "generation",
	}

	_, err := s.queueClient.Enqueue(ctx, queue.JobTypeStoreRaw, queue.QueueHigh, storePayload)
	if err != nil {
		return err
	}

	analyticsPayload := map[string]interface{}{
		"trace_id":    req.TraceID,
		"export_data": req,
		"export_type": "clickhouse",
	}

	_, err = s.queueClient.Enqueue(ctx, queue.JobTypeAnalyticsExport, queue.QueueLow, analyticsPayload)
	return err
}

// Similar async handlers for other endpoints...

func (s *Server) CreateSpanAsync(w http.ResponseWriter, r *http.Request) {
	req, problems, err := decodeValid[database.SpanRequest](r)
	if err != nil {
		if len(problems) > 0 {
			errorResp := database.ErrorResponse{
				Error:    "Validation failed",
				Message:  "The request contains invalid data",
				Code:     http.StatusBadRequest,
				Problems: problems,
			}
			encode(w, r, http.StatusBadRequest, errorResp)
			return
		}

		errorResp := database.ErrorResponse{
			Error:   "Invalid request",
			Message: "Could not parse request body",
			Code:    http.StatusBadRequest,
		}
		encode(w, r, http.StatusBadRequest, errorResp)
		return
	}

	if req.ID == "" {
		req.ID = uuid.New().String()
	}

	if req.StartTime.IsZero() {
		req.StartTime = time.Now().UTC()
	}

	if !s.db.TraceExists(req.TraceID) {
		errorResp := database.ErrorResponse{
			Error:   "Invalid trace",
			Message: "The specified trace_id does not exist",
			Code:    http.StatusBadRequest,
		}
		encode(w, r, http.StatusBadRequest, errorResp)
		return
	}

	if req.ParentID != "" && !s.db.SpanExists(req.ParentID) {
		errorResp := database.ErrorResponse{
			Error:   "Invalid parent span",
			Message: "The specified parent_id does not exist",
			Code:    http.StatusBadRequest,
		}
		encode(w, r, http.StatusBadRequest, errorResp)
		return
	}

	if s.queueClient == nil {
		if err := s.db.CreateSpan(req); err != nil {
			errorResp := database.ErrorResponse{
				Error:   "Database error",
				Message: "Failed to create span",
				Code:    http.StatusInternalServerError,
			}
			encode(w, r, http.StatusInternalServerError, errorResp)
			return
		}
	} else {
		err = s.enqueueSpanJobs(r, req)
		if err != nil {
			if err := s.db.CreateSpan(req); err != nil {
				errorResp := database.ErrorResponse{
					Error:   "Processing error",
					Message: "Failed to process span",
					Code:    http.StatusInternalServerError,
				}
				encode(w, r, http.StatusInternalServerError, errorResp)
				return
			}
		}
	}

	response := database.SuccessResponse{
		ID:     req.ID,
		Status: "accepted",
	}

	encode(w, r, http.StatusAccepted, response)
}

func (s *Server) enqueueSpanJobs(r *http.Request, req database.SpanRequest) error {
	ctx := r.Context()

	storePayload := map[string]interface{}{
		"trace_id":  req.TraceID,
		"raw_data":  req,
		"data_type": "span",
	}

	_, err := s.queueClient.Enqueue(ctx, queue.JobTypeStoreRaw, queue.QueueHigh, storePayload)
	if err != nil {
		return err
	}

	analyticsPayload := map[string]interface{}{
		"trace_id":    req.TraceID,
		"export_data": req,
		"export_type": "clickhouse",
	}

	_, err = s.queueClient.Enqueue(ctx, queue.JobTypeAnalyticsExport, queue.QueueLow, analyticsPayload)
	return err
}
