package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/samims/hcaas/services/auth/internal/service"
)

const (
	KeyError = "error"
)

type AuthHandler struct {
	authSvc service.AuthService
	logger  *slog.Logger
}

func NewAuthHandler(authSvc service.AuthService, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{authSvc: authSvc, logger: logger}
}

// inline error responder
func respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{KeyError: message})
}

// Register handles User Registration/Signup
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("Invalid register payload", slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	user, err := h.authSvc.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid payload")
		return
	}

	_, token, err := h.authSvc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		respondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func (h *AuthHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Get User handler")
	email := r.URL.Query().Get("email")

	if email == "" {
		http.Error(w, "missing email query param", http.StatusBadRequest)
		return
	}
	user, err := h.authSvc.GetUserByEmail(r.Context(), email)

	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(user)

}

func (h *AuthHandler) Validate(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		respondError(w, http.StatusUnauthorized, "missing token")
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	userID, err := h.authSvc.ValidateToken(token)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "invalid token")
	}

	resp := map[string]string{"user_id": userID}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
