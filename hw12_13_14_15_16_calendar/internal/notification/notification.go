package notification

import "time"

type Notification struct {
	EventID string    `json:"event_id"`
	Title   string    `json:"title"`
	Date    time.Time `json:"date"`
	UserID  string    `json:"user_id"`
}

type Status struct {
	EventID string `json:"event_id"`
	Status  string `json:"status"`
	Error   string `json:"error,omitempty"`
}
