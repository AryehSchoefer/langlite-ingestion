package server

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"langlite-ingestion/internal/database"
)

func (s *Server) CreateTrace(w http.ResponseWriter, r *http.Request) {
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

	if req.ID == "" {
		req.ID = uuid.New().String()
	}

	if req.StartTime.IsZero() {
		req.StartTime = time.Now().UTC()
	}

	if err := s.db.CreateTrace(req); err != nil {
		errorResp := database.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to create trace",
			Code:    http.StatusInternalServerError,
		}
		encode(w, r, http.StatusInternalServerError, errorResp)
		return
	}

	response := database.SuccessResponse{
		ID:     req.ID,
		Status: "created",
	}

	encode(w, r, http.StatusCreated, response)
}

func (s *Server) CreateGeneration(w http.ResponseWriter, r *http.Request) {
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

	if err := s.db.CreateGeneration(req); err != nil {
		errorResp := database.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to create generation",
			Code:    http.StatusInternalServerError,
		}
		encode(w, r, http.StatusInternalServerError, errorResp)
		return
	}

	response := database.SuccessResponse{
		ID:     req.ID,
		Status: "created",
	}

	encode(w, r, http.StatusCreated, response)
}

func (s *Server) CreateSpan(w http.ResponseWriter, r *http.Request) {
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

	if err := s.db.CreateSpan(req); err != nil {
		errorResp := database.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to create span",
			Code:    http.StatusInternalServerError,
		}
		encode(w, r, http.StatusInternalServerError, errorResp)
		return
	}

	response := database.SuccessResponse{
		ID:     req.ID,
		Status: "created",
	}

	encode(w, r, http.StatusCreated, response)
}

func (s *Server) UpdateSpan(w http.ResponseWriter, r *http.Request) {
	spanID := r.PathValue("id")
	if spanID == "" {
		errorResp := database.ErrorResponse{
			Error:   "Missing span ID",
			Message: "Span ID is required in the URL path",
			Code:    http.StatusBadRequest,
		}
		encode(w, r, http.StatusBadRequest, errorResp)
		return
	}

	req, problems, err := decodeValid[database.SpanUpdateRequest](r)
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

	if req.EndTime == nil && req.Metadata != nil {
		now := time.Now().UTC()
		req.EndTime = &now
	}

	if err := s.db.UpdateSpan(spanID, req); err != nil {
		if err.Error() == "span not found" {
			errorResp := database.ErrorResponse{
				Error:   "Span not found",
				Message: "The specified span does not exist",
				Code:    http.StatusNotFound,
			}
			encode(w, r, http.StatusNotFound, errorResp)
			return
		}

		errorResp := database.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to update span",
			Code:    http.StatusInternalServerError,
		}
		encode(w, r, http.StatusInternalServerError, errorResp)
		return
	}

	response := database.SuccessResponse{
		ID:     spanID,
		Status: "updated",
	}

	encode(w, r, http.StatusOK, response)
}

func (s *Server) TraceBatchHandler(w http.ResponseWriter, r *http.Request) {
	batchReq, problems, err := decodeValid[database.BatchTraceRequest](r)
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

	response := database.BatchResponse{
		Results: make([]database.BatchResult, len(batchReq.Traces)),
	}

	for i, traceReq := range batchReq.Traces {
		result := database.BatchResult{
			Index:  i,
			Status: "error",
		}

		if problems := traceReq.Valid(r.Context()); len(problems) > 0 {
			var errorMessages []string
			for field, problem := range problems {
				errorMessages = append(errorMessages, fmt.Sprintf("%s: %s", field, problem))
			}
			result.Error = "Validation failed: " + strings.Join(errorMessages, ", ")
			response.Results[i] = result
			response.Summary.Failed++
			continue
		}

		if traceReq.ID == "" {
			traceReq.ID = uuid.New().String()
		}

		if traceReq.StartTime.IsZero() {
			traceReq.StartTime = time.Now().UTC()
		}

		if err := s.db.CreateTrace(traceReq); err != nil {
			result.Error = "Database error: " + err.Error()
			response.Results[i] = result
			response.Summary.Failed++
			continue
		}

		result.ID = traceReq.ID
		result.Status = "success"
		result.Error = ""
		response.Results[i] = result
		response.Summary.Succeeded++
	}

	response.Summary.Total = len(batchReq.Traces)

	statusCode := http.StatusCreated
	if response.Summary.Failed > 0 {
		statusCode = http.StatusMultiStatus
	}

	encode(w, r, statusCode, response)
}

func (s *Server) EventHandler(w http.ResponseWriter, r *http.Request) {
	req, problems, err := decodeValid[database.EventRequest](r)
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

	if req.Timestamp.IsZero() {
		req.Timestamp = time.Now().UTC()
	}

	if req.Level == "" {
		req.Level = "info"
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

	if req.SpanID != "" && !s.db.SpanExists(req.SpanID) {
		errorResp := database.ErrorResponse{
			Error:   "Invalid span",
			Message: "The specified span_id does not exist",
			Code:    http.StatusBadRequest,
		}
		encode(w, r, http.StatusBadRequest, errorResp)
		return
	}

	if err := s.db.CreateEvent(req); err != nil {
		errorResp := database.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to create event",
			Code:    http.StatusInternalServerError,
		}
		encode(w, r, http.StatusInternalServerError, errorResp)
		return
	}

	response := database.SuccessResponse{
		ID:     req.ID,
		Status: "created",
	}

	encode(w, r, http.StatusCreated, response)
}

func (s *Server) ScoreHandler(w http.ResponseWriter, r *http.Request) {
	req, problems, err := decodeValid[database.ScoreRequest](r)
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

	if req.Timestamp.IsZero() {
		req.Timestamp = time.Now().UTC()
	}

	if req.Source == "" {
		req.Source = "human"
	}

	if req.TraceID != "" && !s.db.TraceExists(req.TraceID) {
		errorResp := database.ErrorResponse{
			Error:   "Invalid trace",
			Message: "The specified trace_id does not exist",
			Code:    http.StatusBadRequest,
		}
		encode(w, r, http.StatusBadRequest, errorResp)
		return
	}

	if req.GenerationID != "" && !s.db.GenerationExists(req.GenerationID) {
		errorResp := database.ErrorResponse{
			Error:   "Invalid generation",
			Message: "The specified generation_id does not exist",
			Code:    http.StatusBadRequest,
		}
		encode(w, r, http.StatusBadRequest, errorResp)
		return
	}

	if err := s.db.CreateScore(req); err != nil {
		errorResp := database.ErrorResponse{
			Error:   "Database error",
			Message: "Failed to create score",
			Code:    http.StatusInternalServerError,
		}
		encode(w, r, http.StatusInternalServerError, errorResp)
		return
	}

	response := database.SuccessResponse{
		ID:     req.ID,
		Status: "created",
	}

	encode(w, r, http.StatusCreated, response)
}
