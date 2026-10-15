package sender

import (
	"context"
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

type mockSenderPublisher struct {
	messages []queue.Message
}

func (m *mockSenderPublisher) Publish(_ context.Context, message queue.Message) error {
	m.messages = append(m.messages, message)

	return nil
}

func TestSender_Process(t *testing.T) {
	logger := &mockSenderLogger{}
	publisher := &mockSenderPublisher{}
	s := New(nil, publisher, logger)

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

	require.Len(t, publisher.messages, 1)

	var status notification.Status
	err = json.Unmarshal(publisher.messages[0].Body, &status)
	require.NoError(t, err)

	require.Equal(t, "event-1", status.EventID)
	require.Equal(t, "sent", status.Status)
	require.Empty(t, status.Error)
}

func TestSender_Process_InvalidJSON(t *testing.T) {
	logger := &mockSenderLogger{}
	publisher := &mockSenderPublisher{}
	s := New(nil, publisher, logger)

	message := queue.Message{
		Body: []byte(`not json`),
	}

	err := s.process(message)

	require.Error(t, err)
	require.Contains(t, err.Error(), "unmarshal notification")
	require.Empty(t, publisher.messages)
}

func TestSender_Process_NotificationFields(t *testing.T) {
	logger := &mockSenderLogger{}
	publisher := &mockSenderPublisher{}
	s := New(nil, publisher, logger)

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

	notif := notification.Notification{
		EventID: "event-42",
		Title:   "Important event",
		Date:    date,
		UserID:  "user-42",
	}

	body, err := json.Marshal(notif)
	require.NoError(t, err)

	err = s.process(queue.Message{Body: body})
	require.NoError(t, err)

	require.Len(t, logger.infoMessages, 1)
	require.Contains(t, logger.infoMessages[0], "event_id=event-42")
	require.Contains(t, logger.infoMessages[0], `title="Important event"`)
	require.Contains(t, logger.infoMessages[0], "user_id=user-42")

	require.Len(t, publisher.messages, 1)

	var status notification.Status
	err = json.Unmarshal(publisher.messages[0].Body, &status)
	require.NoError(t, err)

	require.Equal(t, "event-42", status.EventID)
	require.Equal(t, "sent", status.Status)
	require.Empty(t, status.Error)
}
