package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/samims/hcaas/services/notification/internal/handler"
	"github.com/samims/hcaas/services/notification/internal/service"
)

func main() {
	r := chi.NewRouter()

	svc := service.NewNotificationService()
	h := handler.NewNotificationHandler(svc)

	r.Post("/notify", h.Notify)

	err := http.ListenAndServe(":8082", r)
	if err != nil {
		log.Fatalf("Error starting server %v", err)
	}

}
