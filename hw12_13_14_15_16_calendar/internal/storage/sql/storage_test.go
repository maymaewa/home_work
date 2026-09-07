package sqlstorage

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/maymaewa/home_work/hw12_13_14_15_calendar/internal/storage"
)

func TestStorage_CreateEvent(t *testing.T) {
	s := newTestStorage(t)
	ctx := context.Background()

	event := testEvent("create-1", "10:00", "11:00")

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
	s := newTestStorage(t)
	ctx := context.Background()

	event := testEvent("duplicate-1", "10:00", "11:00")

	if err := s.CreateEvent(ctx, event); err != nil {
		t.Fatalf("first CreateEvent() error = %v", err)
	}

	err := s.CreateEvent(ctx, event)

	// Пока этот тест ожидаемо будет падать,
	// потому что обработку duplicate ID мы ещё не добавили.
	if !errors.Is(err, storage.ErrEventAlreadyExists) {
		t.Fatalf("expected ErrEventAlreadyExists, got %v", err)
	}
}

func TestStorage_UpdateEvent(t *testing.T) {
	s := newTestStorage(t)
	ctx := context.Background()

	event := testEvent("update-1", "10:00", "11:00")

	if err := s.CreateEvent(ctx, event); err != nil {
		t.Fatalf("CreateEvent() error = %v", err)
	}

	event.Title = "Updated event"

	err := s.UpdateEvent(ctx, event)
	if err != nil {
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
	s := newTestStorage(t)
	ctx := context.Background()

	event := testEvent("unknown", "10:00", "11:00")

	err := s.UpdateEvent(ctx, event)

	if !errors.Is(err, storage.ErrEventNotFound) {
		t.Fatalf("expected ErrEventNotFound, got %v", err)
	}
}

func TestStorage_DeleteEvent(t *testing.T) {
	s := newTestStorage(t)
	ctx := context.Background()

	event := testEvent("delete-1", "10:00", "11:00")

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
	s := newTestStorage(t)
	ctx := context.Background()

	err := s.DeleteEvent(ctx, "unknown")

	if !errors.Is(err, storage.ErrEventNotFound) {
		t.Fatalf("expected ErrEventNotFound, got %v", err)
	}
}

func TestStorage_ListEventsForDay(t *testing.T) {
	s := newTestStorage(t)
	ctx := context.Background()

	events := []storage.Event{
		testEvent("day-1", "10:00", "11:00"),
		testEvent("day-2", "14:00", "15:00"),
		testEvent("day-3", "16:00", "17:00"),
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
	s := newTestStorage(t)
	ctx := context.Background()

	events := []storage.Event{
		testEvent("week-1", "2026-09-01 10:00", "2026-09-01 11:00"),
		testEvent("week-2", "2026-09-03 10:00", "2026-09-03 11:00"),
		testEvent("week-3", "2026-09-06 10:00", "2026-09-06 11:00"),
		{
			ID:      "week-4",
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
	s := newTestStorage(t)
	ctx := context.Background()

	events := []storage.Event{
		testEvent("month-1", "2026-09-01 10:00", "2026-09-01 11:00"),
		testEvent("month-2", "2026-09-15 10:00", "2026-09-15 11:00"),
		testEvent("month-3", "2026-09-30 10:00", "2026-09-30 11:00"),
		{
			ID:      "month-4",
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

func TestStorage_NotifyBefore(t *testing.T) {
	s := newTestStorage(t)
	ctx := context.Background()

	notifyBefore := 15 * time.Minute

	event := testEvent("notify-1", "10:00", "11:00")
	event.NotifyBefore = &notifyBefore

	if err := s.CreateEvent(ctx, event); err != nil {
		t.Fatalf("CreateEvent() error = %v", err)
	}

	events, err := s.ListEventsForDay(ctx, event.StartAt)
	if err != nil {
		t.Fatalf("ListEventsForDay() error = %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	if events[0].NotifyBefore == nil {
		t.Fatal("expected NotifyBefore to be non-nil")
	}

	if *events[0].NotifyBefore != notifyBefore {
		t.Fatalf(
			"expected NotifyBefore %v, got %v",
			notifyBefore,
			*events[0].NotifyBefore,
		)
	}
}

func TestStorage_NotifyBefore_Nil(t *testing.T) {
	s := newTestStorage(t)
	ctx := context.Background()

	event := testEvent("notify-nil-1", "10:00", "11:00")

	if err := s.CreateEvent(ctx, event); err != nil {
		t.Fatalf("CreateEvent() error = %v", err)
	}

	events, err := s.ListEventsForDay(ctx, event.StartAt)
	if err != nil {
		t.Fatalf("ListEventsForDay() error = %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	if events[0].NotifyBefore != nil {
		t.Fatalf("expected NotifyBefore to be nil, got %v", *events[0].NotifyBefore)
	}
}

func TestStorage_ContextCancelled(t *testing.T) {
	s := newTestStorage(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	event := testEvent("cancelled-1", "10:00", "11:00")

	if err := s.CreateEvent(ctx, event); err == nil {
		t.Fatal("expected error, got nil")
	}

	if err := s.UpdateEvent(ctx, event); err == nil {
		t.Fatal("expected error, got nil")
	}

	if err := s.DeleteEvent(ctx, event.ID); err == nil {
		t.Fatal("expected error, got nil")
	}

	if _, err := s.ListEventsForDay(ctx, event.StartAt); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func newTestStorage(t *testing.T) *Storage {
	t.Helper()

	s := New(Config{
		Host:     "localhost",
		Port:     5432,
		Database: "postgres",
		Username: "postgres",
		Password: "postgres",
	})

	if err := s.Connect(context.Background()); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}

	t.Cleanup(func() {
		_ = s.Close(context.Background())
	})

	if _, err := s.db.Exec(`DELETE FROM events`); err != nil {
		t.Fatalf("cleanup events error = %v", err)
	}

	return s
}

func testEvent(id, start, end string) storage.Event {
	return storage.Event{
		ID:      id,
		Title:   "Test event",
		StartAt: parseTime(start),
		EndAt:   parseTime(end),
		UserID:  "test-user",
	}
}

func parseTime(value string) time.Time {
	layout := "2006-01-02 15:04"

	if len(value) == 5 {
		value = "2026-09-03 " + value
	}

	t, err := time.Parse(layout, value)
	if err != nil {
		panic(err)
	}

	return t
}

func TestStorage_CreateEvent_DateBusy(t *testing.T) {
	s := newTestStorage(t)
	ctx := context.Background()

	first := testEvent("busy-1", "10:00", "11:00")

	if err := s.CreateEvent(ctx, first); err != nil {
		t.Fatalf("first CreateEvent() error = %v", err)
	}

	second := testEvent("busy-2", "10:30", "11:30")

	err := s.CreateEvent(ctx, second)

	if !errors.Is(err, storage.ErrDateBusy) {
		t.Fatalf("expected ErrDateBusy, got %v", err)
	}
}

func TestStorage_UpdateEvent_DateBusy(t *testing.T) {
	s := newTestStorage(t)
	ctx := context.Background()

	first := testEvent("update-busy-1", "10:00", "11:00")
	second := testEvent("update-busy-2", "12:00", "13:00")

	if err := s.CreateEvent(ctx, first); err != nil {
		t.Fatalf("CreateEvent(first) error = %v", err)
	}

	if err := s.CreateEvent(ctx, second); err != nil {
		t.Fatalf("CreateEvent(second) error = %v", err)
	}

	second.StartAt = parseTime("2026-09-03 10:30")
	second.EndAt = parseTime("2026-09-03 11:30")

	err := s.UpdateEvent(ctx, second)

	if !errors.Is(err, storage.ErrDateBusy) {
		t.Fatalf("expected ErrDateBusy, got %v", err)
	}
}

func TestStorage_CreateEvent_AdjacentEvents(t *testing.T) {
	s := newTestStorage(t)
	ctx := context.Background()

	first := testEvent("adjacent-1", "10:00", "11:00")
	second := testEvent("adjacent-2", "11:00", "12:00")

	if err := s.CreateEvent(ctx, first); err != nil {
		t.Fatalf("CreateEvent(first) error = %v", err)
	}

	if err := s.CreateEvent(ctx, second); err != nil {
		t.Fatalf("CreateEvent(second) error = %v", err)
	}

	events, err := s.ListEventsForDay(ctx, first.StartAt)
	if err != nil {
		t.Fatalf("ListEventsForDay() error = %v", err)
	}

	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
}
