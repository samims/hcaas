package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/samims/hcaas/internal/handler"
)

func NewRouter(h *handler.URLHandler) http.Handler {
	r := chi.NewRouter()

	r.Route("/urls", func(r chi.Router) {
		r.Get("/", h.GetAll)
		r.Get("/{id}", h.GetByID)
		r.Post("/", h.Add)
		r.Put("/{id}", h.UpdateStatus)
	})
	return r
}
