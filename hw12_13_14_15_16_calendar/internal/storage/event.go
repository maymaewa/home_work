package storage

import (
	"errors"
	"time"
)

var (
	ErrEventNotFound      = errors.New("event not found")
	ErrEventAlreadyExists = errors.New("event already exists")
	ErrDateBusy           = errors.New("date is busy")
)

type Event struct {
	ID           string
	Title        string
	StartAt      time.Time
	EndAt        time.Time
	Description  string
	UserID       string
	NotifyBefore *time.Duration
}
