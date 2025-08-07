package model

import (
	"time"
)

// Notification struct represents a notification
// This shall match the message model consumed by notification service
type Notification struct {
	UrlID     string    `json:"url_id"`
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
