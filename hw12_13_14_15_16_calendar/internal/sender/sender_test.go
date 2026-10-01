package sender

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/maymaewa/home_work/hw12_13_14_15_calendar/internal/notification"
	"github.com/maymaewa/home_work/hw12_13_14_15_calendar/internal/queue"
	"github.com/stretchr/testify/require"
)

type mockSenderLogger struct {
	infoMessages  []string
	errorMessages []string
}

func (m *mockSenderLogger) Info(message string) {
	m.infoMessages = append(m.infoMessages, message)
}

func (m *mockSenderLogger) Error(message string) {
	m.errorMessages = append(m.errorMessages, message)
}

func TestSender_Process(t *testing.T) {
	logger := &mockSenderLogger{}
	s := New(nil, logger)

	message := queue.Message{
		Body: []byte(`{
			"event_id": "event-1",
			"title": "Test event",
			"date": "2026-10-01T21:45:00+03:00",
			"user_id": "user-1"
		}`),
	}

	err := s.process(message)

	require.NoError(t, err)
	require.Len(t, logger.infoMessages, 1)
	require.Contains(t, logger.infoMessages[0], "event_id=event-1")
	require.Contains(t, logger.infoMessages[0], `title="Test event"`)
	require.Contains(t, logger.infoMessages[0], "user_id=user-1")
}

func TestSender_Process_InvalidJSON(t *testing.T) {
	logger := &mockSenderLogger{}
	s := New(nil, logger)

	message := queue.Message{
		Body: []byte(`not json`),
	}

	err := s.process(message)

	require.Error(t, err)
	require.Contains(t, err.Error(), "unmarshal notification")
}

func TestSender_Process_NotificationFields(t *testing.T) {
	logger := &mockSenderLogger{}
	s := New(nil, logger)

	date := time.Date(
		2026,
		time.October,
		1,
		21,
		45,
		0,
		0,
		time.UTC,
	)

	notification := notification.Notification{
		EventID: "event-42",
		Title:   "Important event",
		Date:    date,
		UserID:  "user-42",
	}

	body, err := json.Marshal(notification)
	require.NoError(t, err)

	err = s.process(queue.Message{Body: body})
	require.NoError(t, err)

	require.Len(t, logger.infoMessages, 1)
	require.Contains(t, logger.infoMessages[0], "event_id=event-42")
	require.Contains(t, logger.infoMessages[0], `title="Important event"`)
	require.Contains(t, logger.infoMessages[0], "user_id=user-42")
}
