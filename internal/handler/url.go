package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/samims/hcaas/internal/errors"
	"github.com/samims/hcaas/internal/model"
	"github.com/samims/hcaas/internal/service"
)

type URLHandler struct {
	svc    service.URLService
	logger *slog.Logger
}

func NewURLHandler(s service.URLService, logger *slog.Logger) *URLHandler {
	return &URLHandler{svc: s, logger: logger}
}

func (h *URLHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	urls, err := h.svc.GetAll(r.Context())
	if err != nil {
		h.logger.Error("GetAll failed", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(urls)
}

func (h *URLHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	url, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		if errors.IsNotFound(err) {
			h.logger.Warn("URL not found", "id", id)
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			h.logger.Error("GetByID failed", "id", id, "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	json.NewEncoder(w).Encode(url)
}

func (h *URLHandler) Add(w http.ResponseWriter, r *http.Request) {
	var url model.URL
	if err := json.NewDecoder(r.Body).Decode(&url); err != nil {
		h.logger.Warn("Invalid request body for Add")
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.svc.Add(r.Context(), url); err != nil {
		if errors.IsInternal(err) {
			h.logger.Warn("Duplicate or invalid Add", "url", url, "error", err)
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			h.logger.Error("Add failed", "url", url, "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	h.logger.Info("URL added", "url", url)
	w.WriteHeader(http.StatusCreated)
}

func (h *URLHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.logger.Warn("Invalid request body for UpdateStatus", "id", id)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.svc.UpdateStatus(r.Context(), id, body.Status); err != nil {
		if errors.IsNotFound(err) {
			h.logger.Warn("URL not found for update", "id", id)
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			h.logger.Error("UpdateStatus failed", "id", id, "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	h.logger.Info("URL status updated", "id", id, "status", body.Status)
	w.WriteHeader(http.StatusOK)
}
