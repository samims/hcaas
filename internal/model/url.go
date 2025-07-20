package model

import "time"

type URL struct {
	ID        string    `json:"id"`
	Address   string    `json:"address"`
	Status    string    `json:"status"`     // "up" or "down"
	CheckedAt time.Time `json:"checked_at"` // last checked time
}
