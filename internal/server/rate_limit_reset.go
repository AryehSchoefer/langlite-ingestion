package server

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"langlite-ingestion/internal/database"
)

// ResetRateLimitHandler resets rate limits for the authenticated API key (development only)
func (s *Server) ResetRateLimitHandler(w http.ResponseWriter, r *http.Request) {
	if os.Getenv("APP_ENV") != "development" && os.Getenv("APP_ENV") != "dev" {
		errorResp := database.ErrorResponse{
			Error:   "Not allowed",
			Message: "Rate limit reset is only available in development mode",
			Code:    http.StatusForbidden,
		}
		encode(w, r, http.StatusForbidden, errorResp)
		return
	}

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
		errorResp := database.ErrorResponse{
			Error:   "Rate limiting disabled",
			Message: "Redis not available",
			Code:    http.StatusServiceUnavailable,
		}
		encode(w, r, http.StatusServiceUnavailable, errorResp)
		return
	}

	now := time.Now()
	keysToDelete := []string{
		fmt.Sprintf("rate_limit:minute:%s:%d", authCtx.APIKeyID, now.Unix()/60),
		fmt.Sprintf("rate_limit:minute:%s:%d", authCtx.APIKeyID, now.Unix()/60-1), // Previous minute
		fmt.Sprintf("rate_limit:hour:%s:%d", authCtx.APIKeyID, now.Unix()/3600),
		fmt.Sprintf("rate_limit:hour:%s:%d", authCtx.APIKeyID, now.Unix()/3600-1), // Previous hour
	}

	deletedCount := 0
	for _, key := range keysToDelete {
		deleted, err := s.redis.Del(r.Context(), key).Result()
		if err == nil {
			deletedCount += int(deleted)
		}
	}

	response := map[string]interface{}{
		"message":      "Rate limits reset successfully",
		"keys_deleted": deletedCount,
		"api_key_id":   authCtx.APIKeyID,
	}

	encode(w, r, http.StatusOK, response)
}
