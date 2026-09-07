package memorystorage

import (
	"context"
	"sync"
	"time"

	"github.com/maymaewa/home_work/hw12_13_14_15_calendar/internal/storage"
)

type Storage struct {
	mu     sync.RWMutex
	events map[string]storage.Event
}

func New() *Storage {
	return &Storage{
		events: make(map[string]storage.Event),
	}
}

func (s *Storage) CreateEvent(ctx context.Context, event storage.Event) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.events[event.ID]; exists {
		return storage.ErrEventAlreadyExists
	}

	for _, existing := range s.events {
		if eventsOverlap(existing, event) {
			return storage.ErrDateBusy
		}
	}

	s.events[event.ID] = event

	return nil
}

func (s *Storage) UpdateEvent(ctx context.Context, event storage.Event) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.events[event.ID]; !exists {
		return storage.ErrEventNotFound
	}

	for id, existing := range s.events {
		if id == event.ID {
			continue
		}

		if eventsOverlap(existing, event) {
			return storage.ErrDateBusy
		}
	}

	s.events[event.ID] = event

	return nil
}

func (s *Storage) DeleteEvent(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.events[id]; !exists {
		return storage.ErrEventNotFound
	}

	delete(s.events, id)

	return nil
}

func (s *Storage) ListEventsForDay(ctx context.Context, date time.Time) ([]storage.Event, error) {
	return s.listEvents(ctx, startOfDay(date), startOfDay(date).AddDate(0, 0, 1))
}

func (s *Storage) ListEventsForWeek(ctx context.Context, date time.Time) ([]storage.Event, error) {
	start := startOfWeek(date)

	return s.listEvents(ctx, start, start.AddDate(0, 0, 7))
}

func (s *Storage) ListEventsForMonth(ctx context.Context, date time.Time) ([]storage.Event, error) {
	start := startOfMonth(date)

	return s.listEvents(ctx, start, start.AddDate(0, 1, 0))
}

func (s *Storage) listEvents(ctx context.Context, start, end time.Time) ([]storage.Event, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	events := make([]storage.Event, 0)

	for _, event := range s.events {
		if event.StartAt.Before(end) && event.EndAt.After(start) {
			events = append(events, event)
		}
	}

	return events, nil
}

func eventsOverlap(first, second storage.Event) bool {
	return first.StartAt.Before(second.EndAt) &&
		second.StartAt.Before(first.EndAt)
}

func startOfDay(t time.Time) time.Time {
	return time.Date(
		t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location(),
	)
}

func startOfWeek(t time.Time) time.Time {
	t = startOfDay(t)

	daysSinceMonday := (int(t.Weekday()) + 6) % 7

	return t.AddDate(0, 0, -daysSinceMonday)
}

func startOfMonth(t time.Time) time.Time {
	return time.Date(
		t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location(),
	)
}
