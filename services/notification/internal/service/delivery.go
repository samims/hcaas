package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/samims/hcaas/services/notification/internal/model"
)

// DeliveryService handles the actual delivery of notification
type DeliveryService interface {
	Deliver(ctx context.Context, n *model.Notification) error
}

// deliveryService is an implementation of DeliveryService
type deliveryService struct {
	log *slog.Logger
}

// NewDeliveryService creates a new delivery service instance
func NewDeliveryService(log *slog.Logger) DeliveryService {
	return &deliveryService{log: log}
}

// Deliver simulates the delivery of a notification
func (s *deliveryService) Deliver(ctx context.Context, n *model.Notification) error {
	s.log.Info("Simulating delivery", slog.String("type", n.Type), slog.String("url_id", n.UrlId))
	time.Sleep(500 * time.Millisecond) // simulating network latency for now
	return nil
}
