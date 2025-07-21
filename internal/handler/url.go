package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/samims/hcaas/internal/errors"
	"github.com/samims/hcaas/internal/model"
	"github.com/samims/hcaas/internal/service"
)

type URLHandler struct {
	svc service.URLService
}

func NewURLHandler(s service.URLService) *URLHandler {
	return &URLHandler{svc: s}
}

func (h *URLHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	urls, err := h.svc.GetAll(r.Context())
	if err != nil {
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
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	json.NewEncoder(w).Encode(url)
}

func (h *URLHandler) Add(w http.ResponseWriter, r *http.Request) {
	var url model.URL
	if err := json.NewDecoder(r.Body).Decode(&url); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.svc.Add(r.Context(), url); err != nil {
		if errors.IsInternal(err) {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *URLHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.svc.UpdateStatus(r.Context(), id, body.Status); err != nil {
		if errors.IsNotFound(err) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusOK)
}
