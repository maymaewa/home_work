package memorystorage

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/maymaewa/home_work/hw12_13_14_15_calendar/internal/storage"
)

func TestStorage_CreateEvent(t *testing.T) {
	ctx := context.Background()
	s := New()

	event := testEvent("1", "10:00", "11:00")

	err := s.CreateEvent(ctx, event)
	if err != nil {
		t.Fatalf("CreateEvent() error = %v", err)
	}

	events, err := s.ListEventsForDay(ctx, event.StartAt)
	if err != nil {
		t.Fatalf("ListEventsForDay() error = %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	if events[0].ID != event.ID {
		t.Fatalf("expected event ID %q, got %q", event.ID, events[0].ID)
	}
}

func TestStorage_CreateEvent_DuplicateID(t *testing.T) {
	ctx := context.Background()
	s := New()

	event := testEvent("1", "10:00", "11:00")

	if err := s.CreateEvent(ctx, event); err != nil {
		t.Fatalf("first CreateEvent() error = %v", err)
	}

	err := s.CreateEvent(ctx, event)

	if !errors.Is(err, storage.ErrEventAlreadyExists) {
		t.Fatalf("expected ErrEventAlreadyExists, got %v", err)
	}
}

func TestStorage_CreateEvent_DateBusy(t *testing.T) {
	ctx := context.Background()
	s := New()

	first := testEvent("1", "10:00", "11:00")
	second := testEvent("2", "10:30", "11:30")

	if err := s.CreateEvent(ctx, first); err != nil {
		t.Fatalf("first CreateEvent() error = %v", err)
	}

	err := s.CreateEvent(ctx, second)

	if !errors.Is(err, storage.ErrDateBusy) {
		t.Fatalf("expected ErrDateBusy, got %v", err)
	}
}

func TestStorage_CreateEvent_AdjacentEvents(t *testing.T) {
	ctx := context.Background()
	s := New()

	first := testEvent("1", "10:00", "11:00")
	second := testEvent("2", "11:00", "12:00")

	if err := s.CreateEvent(ctx, first); err != nil {
		t.Fatalf("first CreateEvent() error = %v", err)
	}

	if err := s.CreateEvent(ctx, second); err != nil {
		t.Fatalf("second CreateEvent() error = %v", err)
	}
}

func TestStorage_UpdateEvent(t *testing.T) {
	ctx := context.Background()
	s := New()

	event := testEvent("1", "10:00", "11:00")

	if err := s.CreateEvent(ctx, event); err != nil {
		t.Fatalf("CreateEvent() error = %v", err)
	}

	event.Title = "Updated event"

	if err := s.UpdateEvent(ctx, event); err != nil {
		t.Fatalf("UpdateEvent() error = %v", err)
	}

	events, err := s.ListEventsForDay(ctx, event.StartAt)
	if err != nil {
		t.Fatalf("ListEventsForDay() error = %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	if events[0].Title != "Updated event" {
		t.Fatalf("expected updated title, got %q", events[0].Title)
	}
}

func TestStorage_UpdateEvent_NotFound(t *testing.T) {
	ctx := context.Background()
	s := New()

	err := s.UpdateEvent(ctx, testEvent("unknown", "10:00", "11:00"))

	if !errors.Is(err, storage.ErrEventNotFound) {
		t.Fatalf("expected ErrEventNotFound, got %v", err)
	}
}

func TestStorage_UpdateEvent_DateBusy(t *testing.T) {
	ctx := context.Background()
	s := New()

	first := testEvent("1", "10:00", "11:00")
	second := testEvent("2", "12:00", "13:00")

	if err := s.CreateEvent(ctx, first); err != nil {
		t.Fatalf("first CreateEvent() error = %v", err)
	}

	if err := s.CreateEvent(ctx, second); err != nil {
		t.Fatalf("second CreateEvent() error = %v", err)
	}

	second.StartAt = parseTime("10:30")
	second.EndAt = parseTime("11:30")

	err := s.UpdateEvent(ctx, second)

	if !errors.Is(err, storage.ErrDateBusy) {
		t.Fatalf("expected ErrDateBusy, got %v", err)
	}
}

func TestStorage_DeleteEvent(t *testing.T) {
	ctx := context.Background()
	s := New()

	event := testEvent("1", "10:00", "11:00")

	if err := s.CreateEvent(ctx, event); err != nil {
		t.Fatalf("CreateEvent() error = %v", err)
	}

	if err := s.DeleteEvent(ctx, event.ID); err != nil {
		t.Fatalf("DeleteEvent() error = %v", err)
	}

	events, err := s.ListEventsForDay(ctx, event.StartAt)
	if err != nil {
		t.Fatalf("ListEventsForDay() error = %v", err)
	}

	if len(events) != 0 {
		t.Fatalf("expected 0 events, got %d", len(events))
	}
}

func TestStorage_DeleteEvent_NotFound(t *testing.T) {
	ctx := context.Background()
	s := New()

	err := s.DeleteEvent(ctx, "unknown")

	if !errors.Is(err, storage.ErrEventNotFound) {
		t.Fatalf("expected ErrEventNotFound, got %v", err)
	}
}

func TestStorage_ListEventsForDay(t *testing.T) {
	ctx := context.Background()
	s := New()

	events := []storage.Event{
		testEvent("1", "10:00", "11:00"),
		testEvent("2", "14:00", "15:00"),
		testEvent("3", "16:00", "17:00"),
	}

	for _, event := range events {
		if err := s.CreateEvent(ctx, event); err != nil {
			t.Fatalf("CreateEvent() error = %v", err)
		}
	}

	result, err := s.ListEventsForDay(ctx, parseTime("2026-09-03 00:00"))
	if err != nil {
		t.Fatalf("ListEventsForDay() error = %v", err)
	}

	if len(result) != 3 {
		t.Fatalf("expected 3 events, got %d", len(result))
	}
}

func TestStorage_ListEventsForWeek(t *testing.T) {
	ctx := context.Background()
	s := New()

	events := []storage.Event{
		testEvent("1", "2026-09-01 10:00", "2026-09-01 11:00"),
		testEvent("2", "2026-09-03 10:00", "2026-09-03 11:00"),
		testEvent("3", "2026-09-06 10:00", "2026-09-06 11:00"),
		{
			ID:      "4",
			Title:   "Next week",
			StartAt: parseTime("2026-09-07 10:00"),
			EndAt:   parseTime("2026-09-07 11:00"),
		},
	}

	for _, event := range events {
		if err := s.CreateEvent(ctx, event); err != nil {
			t.Fatalf("CreateEvent() error = %v", err)
		}
	}

	result, err := s.ListEventsForWeek(ctx, parseTime("2026-09-03 12:00"))
	if err != nil {
		t.Fatalf("ListEventsForWeek() error = %v", err)
	}

	if len(result) != 3 {
		t.Fatalf("expected 3 events, got %d", len(result))
	}
}

func TestStorage_ListEventsForMonth(t *testing.T) {
	ctx := context.Background()
	s := New()

	events := []storage.Event{
		testEvent("1", "2026-09-01 10:00", "2026-09-01 11:00"),
		testEvent("2", "2026-09-15 10:00", "2026-09-15 11:00"),
		testEvent("3", "2026-09-30 10:00", "2026-09-30 11:00"),
		{
			ID:      "4",
			Title:   "Next month",
			StartAt: parseTime("2026-10-01 10:00"),
			EndAt:   parseTime("2026-10-01 11:00"),
		},
	}

	for _, event := range events {
		if err := s.CreateEvent(ctx, event); err != nil {
			t.Fatalf("CreateEvent() error = %v", err)
		}
	}

	result, err := s.ListEventsForMonth(ctx, parseTime("2026-09-15 12:00"))
	if err != nil {
		t.Fatalf("ListEventsForMonth() error = %v", err)
	}

	if len(result) != 3 {
		t.Fatalf("expected 3 events, got %d", len(result))
	}
}

func TestStorage_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	s := New()
	event := testEvent("1", "10:00", "11:00")

	if err := s.CreateEvent(ctx, event); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}

	if err := s.UpdateEvent(ctx, event); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}

	if err := s.DeleteEvent(ctx, event.ID); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}

	if _, err := s.ListEventsForDay(ctx, event.StartAt); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestStorage_Concurrent(t *testing.T) {
	s := New()
	ctx := context.Background()

	const count = 100

	var wg sync.WaitGroup

	for i := 0; i < count; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			start := time.Date(2026, 9, 3, 0, i, 0, 0, time.UTC)

			event := storage.Event{
				ID:      string(rune(i + 1)),
				Title:   "Concurrent event",
				StartAt: start,
				EndAt:   start.Add(time.Minute),
			}

			_ = s.CreateEvent(ctx, event)
		}(i)
	}

	wg.Wait()

	events, err := s.ListEventsForDay(ctx, parseTime("2026-09-03 00:00"))
	if err != nil {
		t.Fatalf("ListEventsForDay() error = %v", err)
	}

	if len(events) != count {
		t.Fatalf("expected %d events, got %d", count, len(events))
	}
}

func testEvent(id, start, end string) storage.Event {
	return storage.Event{
		ID:      id,
		Title:   "Test event",
		StartAt: parseTime(start),
		EndAt:   parseTime(end),
	}
}

func parseTime(value string) time.Time {
	layout := "2006-01-02 15:04"

	if len(value) == 5 {
		return parseTime("2026-09-03 " + value)
	}

	t, err := time.Parse(layout, value)
	if err != nil {
		panic(err)
	}

	return t
}
