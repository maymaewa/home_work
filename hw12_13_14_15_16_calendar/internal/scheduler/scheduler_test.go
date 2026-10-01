package scheduler

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/maymaewa/home_work/hw12_13_14_15_calendar/internal/notification"
	"github.com/maymaewa/home_work/hw12_13_14_15_calendar/internal/queue"
	"github.com/maymaewa/home_work/hw12_13_14_15_calendar/internal/storage"
	"github.com/stretchr/testify/require"
)

type mockApplication struct {
	events        []storage.Event
	listErr       error
	deleteErr     error
	deletedBefore time.Time
}

func (m *mockApplication) ListEventsForNotification(
	_ context.Context,
	_, _ time.Time,
) ([]storage.Event, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}

	return m.events, nil
}

func (m *mockApplication) DeleteEventsOlderThan(
	_ context.Context,
	before time.Time,
) error {
	m.deletedBefore = before

	return m.deleteErr
}

type mockPublisher struct {
	messages []queue.Message
	err      error
}

func (m *mockPublisher) Publish(_ context.Context, message queue.Message) error {
	if m.err != nil {
		return m.err
	}

	m.messages = append(m.messages, message)

	return nil
}

type mockLogger struct {
	infoMessages  []string
	errorMessages []string
}

func (m *mockLogger) Info(message string) {
	m.infoMessages = append(m.infoMessages, message)
}

func (m *mockLogger) Error(message string) {
	m.errorMessages = append(m.errorMessages, message)
}

func TestScheduler_Run_PublishesNotification(t *testing.T) {
	now := time.Now()

	app := &mockApplication{
		events: []storage.Event{
			{
				ID:           "event-1",
				Title:        "Test event",
				StartAt:      now.Add(10 * time.Minute),
				EndAt:        now.Add(11 * time.Minute),
				UserID:       "user-1",
				NotifyBefore: durationPtr(5 * time.Minute),
			},
		},
	}

	publisher := &mockPublisher{}
	logger := &mockLogger{}

	s := New(app, publisher, logger, time.Minute)

	from := now
	to := now.Add(time.Minute)

	err := s.run(context.Background(), from, to)

	require.NoError(t, err)
	require.Len(t, publisher.messages, 1)

	var got notification.Notification

	err = json.Unmarshal(publisher.messages[0].Body, &got)
	require.NoError(t, err)

	require.Equal(t, "event-1", got.EventID)
	require.Equal(t, "Test event", got.Title)
	require.Equal(t, "user-1", got.UserID)
	require.True(t, now.Add(10*time.Minute).Equal(got.Date))

	require.Len(t, logger.infoMessages, 1)
}

func TestScheduler_Run_DeletesOldEvents(t *testing.T) {
	now := time.Now()

	app := &mockApplication{}
	publisher := &mockPublisher{}
	logger := &mockLogger{}

	s := New(app, publisher, logger, time.Minute)

	err := s.run(
		context.Background(),
		now.Add(-time.Minute),
		now,
	)

	require.NoError(t, err)

	require.Equal(
		t,
		now.AddDate(-1, 0, 0),
		app.deletedBefore,
	)
}

func TestScheduler_Run_ReturnsListError(t *testing.T) {
	expectedErr := errors.New("list events error")

	app := &mockApplication{
		listErr: expectedErr,
	}

	publisher := &mockPublisher{}
	logger := &mockLogger{}

	s := New(app, publisher, logger, time.Minute)

	err := s.run(
		context.Background(),
		time.Now().Add(-time.Minute),
		time.Now(),
	)

	require.ErrorIs(t, err, expectedErr)
	require.Empty(t, publisher.messages)
}

func TestScheduler_Run_ReturnsPublishError(t *testing.T) {
	expectedErr := errors.New("publish error")

	app := &mockApplication{
		events: []storage.Event{
			{
				ID:      "event-1",
				Title:   "Test event",
				StartAt: time.Now().Add(time.Minute),
				UserID:  "user-1",
			},
		},
	}

	publisher := &mockPublisher{
		err: expectedErr,
	}

	logger := &mockLogger{}

	s := New(app, publisher, logger, time.Minute)

	err := s.run(
		context.Background(),
		time.Now().Add(-time.Minute),
		time.Now(),
	)

	require.ErrorIs(t, err, expectedErr)
}

func TestScheduler_Run_ReturnsDeleteError(t *testing.T) {
	expectedErr := errors.New("delete events error")

	app := &mockApplication{
		deleteErr: expectedErr,
	}

	publisher := &mockPublisher{}
	logger := &mockLogger{}

	s := New(app, publisher, logger, time.Minute)

	err := s.run(
		context.Background(),
		time.Now().Add(-time.Minute),
		time.Now(),
	)

	require.ErrorIs(t, err, expectedErr)
}

func durationPtr(value time.Duration) *time.Duration {
	return &value
}
