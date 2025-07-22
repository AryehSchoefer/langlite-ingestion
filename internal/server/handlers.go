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

func (s *Server) BatchHandler(w http.ResponseWriter, r *http.Request) {
	batchReq, problems, err := decodeValid[database.BatchRequest](r)
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

	totalItems := len(batchReq.Traces) + len(batchReq.Spans) + len(batchReq.Generations) + len(batchReq.Events) + len(batchReq.Scores)
	response := database.BatchResponse{
		Results: make([]database.BatchResult, 0, totalItems),
	}

	index := 0

	for _, traceReq := range batchReq.Traces {
		result := s.processBatchTrace(traceReq, index, r)
		response.Results = append(response.Results, result)
		if result.Status == "success" {
			response.Summary.Succeeded++
		} else {
			response.Summary.Failed++
		}
		index++
	}

	for _, spanReq := range batchReq.Spans {
		result := s.processBatchSpan(spanReq, index, r)
		response.Results = append(response.Results, result)
		if result.Status == "success" {
			response.Summary.Succeeded++
		} else {
			response.Summary.Failed++
		}
		index++
	}

	for _, genReq := range batchReq.Generations {
		result := s.processBatchGeneration(genReq, index, r)
		response.Results = append(response.Results, result)
		if result.Status == "success" {
			response.Summary.Succeeded++
		} else {
			response.Summary.Failed++
		}
		index++
	}

	for _, eventReq := range batchReq.Events {
		result := s.processBatchEvent(eventReq, index, r)
		response.Results = append(response.Results, result)
		if result.Status == "success" {
			response.Summary.Succeeded++
		} else {
			response.Summary.Failed++
		}
		index++
	}

	for _, scoreReq := range batchReq.Scores {
		result := s.processBatchScore(scoreReq, index, r)
		response.Results = append(response.Results, result)
		if result.Status == "success" {
			response.Summary.Succeeded++
		} else {
			response.Summary.Failed++
		}
		index++
	}

	response.Summary.Total = totalItems

	statusCode := http.StatusCreated
	if response.Summary.Failed > 0 {
		statusCode = http.StatusMultiStatus
	}

	encode(w, r, statusCode, response)
}

func (s *Server) processBatchTrace(traceReq database.TraceRequest, index int, r *http.Request) database.BatchResult {
	result := database.BatchResult{
		Index:  index,
		Status: "error",
	}

	if problems := traceReq.Valid(r.Context()); len(problems) > 0 {
		var errorMessages []string
		for field, problem := range problems {
			errorMessages = append(errorMessages, fmt.Sprintf("%s: %s", field, problem))
		}
		result.Error = "Validation failed: " + strings.Join(errorMessages, ", ")
		return result
	}

	if traceReq.ID == "" {
		traceReq.ID = uuid.New().String()
	}

	if traceReq.StartTime.IsZero() {
		traceReq.StartTime = time.Now().UTC()
	}

	if err := s.db.CreateTrace(traceReq); err != nil {
		result.Error = "Database error: " + err.Error()
		return result
	}

	result.ID = traceReq.ID
	result.Status = "success"
	result.Error = ""
	return result
}

func (s *Server) processBatchSpan(spanReq database.SpanRequest, index int, r *http.Request) database.BatchResult {
	result := database.BatchResult{
		Index:  index,
		Status: "error",
	}

	if problems := spanReq.Valid(r.Context()); len(problems) > 0 {
		var errorMessages []string
		for field, problem := range problems {
			errorMessages = append(errorMessages, fmt.Sprintf("%s: %s", field, problem))
		}
		result.Error = "Validation failed: " + strings.Join(errorMessages, ", ")
		return result
	}

	if spanReq.ID == "" {
		spanReq.ID = uuid.New().String()
	}

	if spanReq.StartTime.IsZero() {
		spanReq.StartTime = time.Now().UTC()
	}

	if !s.db.TraceExists(spanReq.TraceID) {
		result.Error = "Invalid trace: The specified trace_id does not exist"
		return result
	}

	if spanReq.ParentID != "" && !s.db.SpanExists(spanReq.ParentID) {
		result.Error = "Invalid parent span: The specified parent_id does not exist"
		return result
	}

	if err := s.db.CreateSpan(spanReq); err != nil {
		result.Error = "Database error: " + err.Error()
		return result
	}

	result.ID = spanReq.ID
	result.Status = "success"
	result.Error = ""
	return result
}

func (s *Server) processBatchGeneration(genReq database.GenerationRequest, index int, r *http.Request) database.BatchResult {
	result := database.BatchResult{
		Index:  index,
		Status: "error",
	}

	if problems := genReq.Valid(r.Context()); len(problems) > 0 {
		var errorMessages []string
		for field, problem := range problems {
			errorMessages = append(errorMessages, fmt.Sprintf("%s: %s", field, problem))
		}
		result.Error = "Validation failed: " + strings.Join(errorMessages, ", ")
		return result
	}

	if genReq.ID == "" {
		genReq.ID = uuid.New().String()
	}

	if genReq.StartTime.IsZero() {
		genReq.StartTime = time.Now().UTC()
	}

	if err := s.db.CreateGeneration(genReq); err != nil {
		result.Error = "Database error: " + err.Error()
		return result
	}

	result.ID = genReq.ID
	result.Status = "success"
	result.Error = ""
	return result
}

func (s *Server) processBatchEvent(eventReq database.EventRequest, index int, r *http.Request) database.BatchResult {
	result := database.BatchResult{
		Index:  index,
		Status: "error",
	}

	if problems := eventReq.Valid(r.Context()); len(problems) > 0 {
		var errorMessages []string
		for field, problem := range problems {
			errorMessages = append(errorMessages, fmt.Sprintf("%s: %s", field, problem))
		}
		result.Error = "Validation failed: " + strings.Join(errorMessages, ", ")
		return result
	}

	if eventReq.ID == "" {
		eventReq.ID = uuid.New().String()
	}

	if eventReq.Timestamp.IsZero() {
		eventReq.Timestamp = time.Now().UTC()
	}

	if eventReq.Level == "" {
		eventReq.Level = "info"
	}

	if !s.db.TraceExists(eventReq.TraceID) {
		result.Error = "Invalid trace: The specified trace_id does not exist"
		return result
	}

	if eventReq.SpanID != "" && !s.db.SpanExists(eventReq.SpanID) {
		result.Error = "Invalid span: The specified span_id does not exist"
		return result
	}

	if err := s.db.CreateEvent(eventReq); err != nil {
		result.Error = "Database error: " + err.Error()
		return result
	}

	result.ID = eventReq.ID
	result.Status = "success"
	result.Error = ""
	return result
}

func (s *Server) processBatchScore(scoreReq database.ScoreRequest, index int, r *http.Request) database.BatchResult {
	result := database.BatchResult{
		Index:  index,
		Status: "error",
	}

	if problems := scoreReq.Valid(r.Context()); len(problems) > 0 {
		var errorMessages []string
		for field, problem := range problems {
			errorMessages = append(errorMessages, fmt.Sprintf("%s: %s", field, problem))
		}
		result.Error = "Validation failed: " + strings.Join(errorMessages, ", ")
		return result
	}

	if scoreReq.ID == "" {
		scoreReq.ID = uuid.New().String()
	}

	if scoreReq.Timestamp.IsZero() {
		scoreReq.Timestamp = time.Now().UTC()
	}

	if scoreReq.Source == "" {
		scoreReq.Source = "human"
	}

	if scoreReq.TraceID != "" && !s.db.TraceExists(scoreReq.TraceID) {
		result.Error = "Invalid trace: The specified trace_id does not exist"
		return result
	}

	if scoreReq.GenerationID != "" && !s.db.GenerationExists(scoreReq.GenerationID) {
		result.Error = "Invalid generation: The specified generation_id does not exist"
		return result
	}

	if err := s.db.CreateScore(scoreReq); err != nil {
		result.Error = "Database error: " + err.Error()
		return result
	}

	result.ID = scoreReq.ID
	result.Status = "success"
	result.Error = ""
	return result
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
