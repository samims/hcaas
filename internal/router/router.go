package router

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/samims/hcaas/internal/handler"
)

func NewRouter(h *handler.URLHandler) http.Handler {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Route("/urls", func(r chi.Router) {
		r.Get("/", h.GetAll)
		r.Get("/{id}", h.GetByID)
		r.Post("/", h.Add)
		r.Put("/{id}", h.UpdateStatus)
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})
	return r
}
