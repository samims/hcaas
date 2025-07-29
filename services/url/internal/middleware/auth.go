package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

func AuthMiddleware(authServiceURL string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "Unauthorized: missing token", http.StatusUnauthorized)
				return
			}
			token := strings.TrimPrefix(authHeader, "Bearer ")

			// Call the auth service
			//reqBody, _ := json.Marshal(map[string]string{"token": token})
			req, err := http.NewRequest("GET", authServiceURL+"auth/validate", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			client := http.Client{}
			resp, err := client.Do(req)

			//resp, err := http.Post(authServiceURL+"/auth/validate", "application/json", bytes.NewReader(reqBody))
			if err != nil || resp.StatusCode != http.StatusOK {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			var data struct {
				UserID string `json:"user_id"`
			}
			json.NewDecoder(resp.Body).Decode(&data)

			ctx := context.WithValue(r.Context(), "userID", data.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
