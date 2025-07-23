package server

import (
	"net/http"

	"langlite-ingestion/internal/database"
)

func (s *Server) RateLimitStatusHandler(w http.ResponseWriter, r *http.Request) {
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

	if s.rateLimiter == nil {
		response := map[string]interface{}{
			"rate_limiting_enabled": false,
			"message":               "Rate limiting is disabled (Redis not available)",
		}
		encode(w, r, http.StatusOK, response)
		return
	}

	minuteUsed, hourUsed, err := s.rateLimiter.GetRateLimitStatus(r.Context(), authCtx.APIKeyID)
	if err != nil {
		errorResp := database.ErrorResponse{
			Error:   "Rate limit check failed",
			Message: "Could not retrieve rate limit status",
			Code:    http.StatusInternalServerError,
		}
		encode(w, r, http.StatusInternalServerError, errorResp)
		return
	}

	response := map[string]interface{}{
		"rate_limiting_enabled": true,
		"limits": map[string]interface{}{
			"per_minute": 1000,
			"per_hour":   10000,
		},
		"usage": map[string]interface{}{
			"current_minute": minuteUsed,
			"current_hour":   hourUsed,
		},
		"remaining": map[string]interface{}{
			"current_minute": 1000 - minuteUsed,
			"current_hour":   10000 - hourUsed,
		},
	}

	encode(w, r, http.StatusOK, response)
}
