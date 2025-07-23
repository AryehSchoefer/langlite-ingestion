package server

import (
	"net/http"

	"langlite-ingestion/internal/database"
)

func (s *Server) QueueStatusHandler(w http.ResponseWriter, r *http.Request) {
	if s.queueClient == nil {
		response := map[string]interface{}{
			"queue_enabled": false,
			"message":       "Queue system is disabled (Redis not available)",
		}
		encode(w, r, http.StatusOK, response)
		return
	}

	stats, err := s.queueClient.GetQueueStats(r.Context())
	if err != nil {
		errorResp := database.ErrorResponse{
			Error:   "Queue status error",
			Message: "Could not retrieve queue statistics",
			Code:    http.StatusInternalServerError,
		}
		encode(w, r, http.StatusInternalServerError, errorResp)
		return
	}

	response := map[string]interface{}{
		"queue_enabled": true,
		"queues":        stats,
		"summary": map[string]interface{}{
			"total_pending":   calculateTotalPending(stats),
			"high_priority":   calculatePriorityTotal(stats, "high"),
			"medium_priority": calculatePriorityTotal(stats, "medium"),
			"low_priority":    calculatePriorityTotal(stats, "low"),
		},
	}

	encode(w, r, http.StatusOK, response)
}

func (s *Server) WorkerStatusHandler(w http.ResponseWriter, r *http.Request) {
	if s.workerPool == nil {
		response := map[string]interface{}{
			"workers_enabled": false,
			"message":         "Worker pool is disabled (Redis not available)",
		}
		encode(w, r, http.StatusOK, response)
		return
	}

	stats, err := s.workerPool.GetStats(r.Context())
	if err != nil {
		errorResp := database.ErrorResponse{
			Error:   "Worker status error",
			Message: "Could not retrieve worker statistics",
			Code:    http.StatusInternalServerError,
		}
		encode(w, r, http.StatusInternalServerError, errorResp)
		return
	}

	response := map[string]interface{}{
		"workers_enabled": true,
		"stats":           stats,
	}

	encode(w, r, http.StatusOK, response)
}

func (s *Server) ProcessingStatusHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"async_processing_enabled": s.queueClient != nil,
		"redis_connected":          s.redis != nil,
	}

	if s.queueClient != nil {
		queueStats, err := s.queueClient.GetQueueStats(r.Context())
		if err == nil {
			response["queue_stats"] = queueStats
		}
	}

	if s.workerPool != nil {
		workerStats, err := s.workerPool.GetStats(r.Context())
		if err == nil {
			response["worker_stats"] = workerStats
		}
	}

	if s.queueClient != nil {
		response["processing_mode"] = "async"
		response["response_status"] = "202 Accepted (async processing)"
	} else {
		response["processing_mode"] = "sync"
		response["response_status"] = "201 Created (immediate processing)"
	}

	encode(w, r, http.StatusOK, response)
}

func calculateTotalPending(stats map[string]int64) int64 {
	var total int64
	for queueName, count := range stats {
		if queueName != "dead_letter" && queueName != "delayed" {
			total += count
		}
	}
	return total
}

func calculatePriorityTotal(stats map[string]int64, priority string) int64 {
	var total int64
	for queueName, count := range stats {
		if len(queueName) > len(priority) && queueName[:len(priority)] == priority {
			total += count
		}
	}
	return total
}
