package server

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"langlite-ingestion/internal/database"
)

type RateLimiter struct {
	redis *redis.Client
}

func NewRateLimiter(redisClient *redis.Client) *RateLimiter {
	return &RateLimiter{
		redis: redisClient,
	}
}

func (rl *RateLimiter) RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" || r.URL.Path == "/" {
			next.ServeHTTP(w, r)
			return
		}

		authCtx, ok := GetAuthContext(r)
		if !ok {
			next.ServeHTTP(w, r)
			return
		}

		allowed, resetTime, err := rl.checkRateLimit(r.Context(), authCtx.APIKeyID)
		if err != nil {
			// Log error but don't block request on Redis failure
			// In production, we might want to fail open or closed based on requirements
			next.ServeHTTP(w, r)
			return
		}

		if !allowed {
			w.Header().Set("X-RateLimit-Limit", "1000") // this should come from the API key config
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))
			w.Header().Set("Retry-After", strconv.FormatInt(int64(time.Until(resetTime).Seconds()), 10))

			errorResp := database.ErrorResponse{
				Error:   "Rate limit exceeded",
				Message: "Too many requests. Please try again later.",
				Code:    http.StatusTooManyRequests,
			}
			encode(w, r, http.StatusTooManyRequests, errorResp)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) checkRateLimit(ctx context.Context, apiKeyID string) (bool, time.Time, error) {
	now := time.Now()

	// Production rate limits (should eventually come from database API key config)
	minuteLimit := int64(1000)
	hourLimit := int64(10000)

	minuteKey := fmt.Sprintf("rate_limit:minute:%s:%d", apiKeyID, now.Unix()/60)
	minuteCount, err := rl.redis.Incr(ctx, minuteKey).Result()
	if err != nil {
		return false, time.Time{}, err
	}

	if minuteCount == 1 {
		rl.redis.Expire(ctx, minuteKey, time.Minute)
	}

	if minuteCount > minuteLimit {
		resetTime := time.Unix((now.Unix()/60+1)*60, 0)
		return false, resetTime, nil
	}

	hourKey := fmt.Sprintf("rate_limit:hour:%s:%d", apiKeyID, now.Unix()/3600)
	hourCount, err := rl.redis.Incr(ctx, hourKey).Result()
	if err != nil {
		return false, time.Time{}, err
	}

	if hourCount == 1 {
		rl.redis.Expire(ctx, hourKey, time.Hour)
	}

	if hourCount > hourLimit {
		resetTime := time.Unix((now.Unix()/3600+1)*3600, 0)
		return false, resetTime, nil
	}

	return true, time.Time{}, nil
}

func (rl *RateLimiter) GetRateLimitStatus(ctx context.Context, apiKeyID string) (minuteUsed, hourUsed int64, err error) {
	now := time.Now()

	minuteKey := fmt.Sprintf("rate_limit:minute:%s:%d", apiKeyID, now.Unix()/60)
	hourKey := fmt.Sprintf("rate_limit:hour:%s:%d", apiKeyID, now.Unix()/3600)

	minuteUsed, err = rl.redis.Get(ctx, minuteKey).Int64()
	if err == redis.Nil {
		minuteUsed = 0
		err = nil
	} else if err != nil {
		return 0, 0, err
	}

	hourUsed, err = rl.redis.Get(ctx, hourKey).Int64()
	if err == redis.Nil {
		hourUsed = 0
		err = nil
	} else if err != nil {
		return 0, 0, err
	}

	return minuteUsed, hourUsed, nil
}
