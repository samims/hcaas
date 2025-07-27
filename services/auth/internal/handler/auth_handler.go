package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/samims/hcaas/services/auth/internal/service"
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
	json.NewEncoder(w).Encode(map[string]string{"error": message})
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
		respondError(w, http.StatusBadRequest, "Invalid request payload")
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
