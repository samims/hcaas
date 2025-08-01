package model

import "time"

type Notification struct {
	ID        int       `json:"id" db:"id"`
	UrlId     string    `json:"url_id" db:"url_id"`
	Type      string    `json:"type" db:"type"` // email, sms, webhook
	Message   string    `json:"message" db:"message"`
	Status    string    `json:"status" db:"status"` // pending, sent, failed
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

const (
	StatusPending = "pending"
	StatusSent    = "sent"
	StatusFailed  = "failed"
)
