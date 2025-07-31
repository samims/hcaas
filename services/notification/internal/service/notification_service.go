package service

import (
	"context"
	"fmt"

	"github.com/samims/hcaas/services/notification/internal/model"
)

type NotificationService interface {
	SendNotification(ctx context.Context, notification *model.Notification) error
}

type notificationService struct {
}

func NewNotificationService() NotificationService {
	return &notificationService{}
}

// SendNotification sends the notification
func (s *notificationService) SendNotification(ctx context.Context, notification *model.Notification) error {
	// Simulate sending notification service
	fmt.Println("Notification: URL", notification.UrlId)
	return nil
}
