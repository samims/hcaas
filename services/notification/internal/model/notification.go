package model

// Notification struct holds the Notification information
type Notification struct {
	UrlId     string `json:"url_id"`
	OldStatus string `json:"old_status"`
	NewStatus string `json:"new_status"`
	TimeStamp string `json:"timestamp"`
}
