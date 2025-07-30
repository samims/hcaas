package middleware

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

func AuthMiddleware(authServiceURL string, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				logger.Warn("Unauthorized: missing or malformed token")
				http.Error(w, "Unauthorized: missing token", http.StatusUnauthorized)
				return
			}
			token := strings.TrimPrefix(authHeader, "Bearer ")

			req, err := http.NewRequest(http.MethodGet, authServiceURL+"auth/validate", nil)
			if err != nil {
				logger.Error("Failed to create request to auth service", "error", err)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			req.Header.Set("Authorization", "Bearer "+token)
			client := http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				logger.Error("Failed to call auth service", "error", err)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				bodyBytes, _ := io.ReadAll(resp.Body)
				logger.Warn("Auth service validation failed", "status", resp.StatusCode, "body", string(bodyBytes))
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			var data struct {
				UserID string `json:"user_id"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
				logger.Error("Failed to decode auth service response", "error", err)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), "userID", data.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
