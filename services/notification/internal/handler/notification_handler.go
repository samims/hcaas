package handler

import (
	"encoding/json"
	"net/http"

	"github.com/samims/hcaas/services/notification/internal/model"
	"github.com/samims/hcaas/services/notification/internal/service"
)

type NotificationHandler struct {
	service service.NotificationService
}

func NewNotificationHandler(s service.NotificationService) *NotificationHandler {
	return &NotificationHandler{service: s}
}

func (h *NotificationHandler) Notify(w http.ResponseWriter, r *http.Request) {
	var notification model.Notification
	if err := json.NewDecoder(r.Body).Decode(&notification); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	err := h.service.SendNotification(r.Context(), &notification)
	if err != nil {
		http.Error(w, "failed to send notification", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Notification sent"))
}
