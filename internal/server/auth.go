package server

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"langlite-ingestion/internal/database"
)

type contextKey string

const AuthContextKey contextKey = "auth"

func (s *Server) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" || r.URL.Path == "/" || r.URL.Path == "/metrics" {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			errorResp := database.ErrorResponse{
				Error:   "Missing authorization",
				Message: "Authorization header is required",
				Code:    http.StatusUnauthorized,
			}
			encode(w, r, http.StatusUnauthorized, errorResp)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			errorResp := database.ErrorResponse{
				Error:   "Invalid authorization format",
				Message: "Authorization header must be 'Bearer <token>'",
				Code:    http.StatusUnauthorized,
			}
			encode(w, r, http.StatusUnauthorized, errorResp)
			return
		}

		apiKey := parts[1]
		if apiKey == "" {
			errorResp := database.ErrorResponse{
				Error:   "Missing API key",
				Message: "API key cannot be empty",
				Code:    http.StatusUnauthorized,
			}
			encode(w, r, http.StatusUnauthorized, errorResp)
			return
		}

		hasher := sha256.New()
		hasher.Write([]byte(apiKey))
		keyHash := hex.EncodeToString(hasher.Sum(nil))

		validatedKey, err := s.db.ValidateAPIKey(keyHash)
		if err != nil {
			errorResp := database.ErrorResponse{
				Error:   "Invalid API key",
				Message: "The provided API key is invalid or expired",
				Code:    http.StatusUnauthorized,
			}
			encode(w, r, http.StatusUnauthorized, errorResp)
			return
		}

		// Update last used timestamp (async to not slow down request)
		go func() {
			_ = s.db.UpdateAPIKeyLastUsed(validatedKey.ID)
		}()

		authCtx := database.AuthContext{
			ProjectID: validatedKey.ProjectID,
			APIKeyID:  validatedKey.ID,
		}

		ctx := context.WithValue(r.Context(), AuthContextKey, authCtx)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetAuthContext(r *http.Request) (*database.AuthContext, bool) {
	authCtx, ok := r.Context().Value(AuthContextKey).(database.AuthContext)
	if !ok {
		return nil, false
	}
	return &authCtx, true
}
