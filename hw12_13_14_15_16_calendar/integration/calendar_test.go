package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/require"
)

const requestTimeout = 5 * time.Second

type eventRequest struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	StartAt      string `json:"start_at"`
	EndAt        string `json:"end_at"`
	Description  string `json:"description"`
	UserID       string `json:"user_id"`
	NotifyBefore string `json:"notify_before"`
}

type eventResponse struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	StartAt      string `json:"start_at"`
	EndAt        string `json:"end_at"`
	Description  string `json:"description"`
	UserID       string `json:"user_id"`
	NotifyBefore string `json:"notify_before"`
}

type notificationStatus struct {
	EventID string `json:"event_id"`
	Status  string `json:"status"`
	Error   string `json:"error,omitempty"`
}

func calendarURL() string {
	url := os.Getenv("CALENDAR_URL")
	if url == "" {
		return "http://localhost:8888"
	}

	return url
}

func doRequest(
	ctx context.Context,
	method string,
	url string,
	body *bytes.Reader,
) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("create HTTP request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send HTTP request: %w", err)
	}

	return resp, nil
}

func waitForCalendar(t *testing.T) {
	t.Helper()

	require.Eventually(t, func() bool {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		resp, err := doRequest(
			ctx,
			http.MethodGet,
			calendarURL()+"/hello",
			nil,
		)
		if err != nil {
			return false
		}

		defer resp.Body.Close()

		return resp.StatusCode == http.StatusOK
	}, 30*time.Second, 500*time.Millisecond)
}

func createEvent(t *testing.T, event eventRequest) eventResponse {
	t.Helper()

	body, err := json.Marshal(event)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	resp, err := doRequest(
		ctx,
		http.MethodPost,
		calendarURL()+"/events",
		bytes.NewReader(body),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var result eventResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))

	return result
}

func TestCalendarIsAvailable(t *testing.T) {
	waitForCalendar(t)
}

func TestCalendarAPI(t *testing.T) {
	waitForCalendar(t)

	now := time.Now().UTC()

	// Делаем событие позже, чтобы оно не пересекалось
	// с событием из TestNotificationStatus.
	startAt := now.Add(2 * time.Minute)
	endAt := startAt.Add(time.Hour)

	event := eventRequest{
		ID:           "integration-event-1",
		Title:        "Integration test event",
		StartAt:      startAt.Format(time.RFC3339),
		EndAt:        endAt.Format(time.RFC3339),
		Description:  "Created by integration test",
		UserID:       "integration-user",
		NotifyBefore: "5s",
	}

	created := createEvent(t, event)

	require.Equal(t, event.ID, created.ID)
	require.Equal(t, event.Title, created.Title)
	require.Equal(t, event.UserID, created.UserID)

	date := startAt.Format("2006-01-02")

	t.Run("day", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
		defer cancel()

		resp, err := doRequest(
			ctx,
			http.MethodGet,
			calendarURL()+"/events/day?date="+date,
			nil,
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		var events []eventResponse
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&events))

		require.NotEmpty(t, events)
		require.Equal(t, event.ID, events[0].ID)
	})

	t.Run("week", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
		defer cancel()

		resp, err := doRequest(
			ctx,
			http.MethodGet,
			calendarURL()+"/events/week?date="+date,
			nil,
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		var events []eventResponse
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&events))

		require.NotEmpty(t, events)
		require.Equal(t, event.ID, events[0].ID)
	})

	t.Run("month", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
		defer cancel()

		resp, err := doRequest(
			ctx,
			http.MethodGet,
			calendarURL()+"/events/month?date="+date,
			nil,
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		var events []eventResponse
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&events))

		require.NotEmpty(t, events)
		require.Equal(t, event.ID, events[0].ID)
	})

	t.Run("duplicate_event", func(t *testing.T) {
		body, err := json.Marshal(event)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
		defer cancel()

		resp, err := doRequest(
			ctx,
			http.MethodPost,
			calendarURL()+"/events",
			bytes.NewReader(body),
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusConflict, resp.StatusCode)
	})

	t.Run("delete_not_found", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
		defer cancel()

		resp, err := doRequest(
			ctx,
			http.MethodDelete,
			calendarURL()+"/events/not-existing-event",
			nil,
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

func TestNotificationStatus(t *testing.T) {
	waitForCalendar(t)

	statusQueue := os.Getenv("RABBITMQ_STATUS_QUEUE")
	if statusQueue == "" {
		statusQueue = "calendar.notification.status"
	}

	host := os.Getenv("RABBITMQ_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("RABBITMQ_PORT")
	if port == "" {
		port = "5672"
	}

	username := os.Getenv("RABBITMQ_USERNAME")
	if username == "" {
		username = "rabbit"
	}

	password := os.Getenv("RABBITMQ_PASSWORD")
	if password == "" {
		password = "password"
	}

	address := fmt.Sprintf(
		"amqp://%s:%s@%s:%s/",
		username,
		password,
		host,
		port,
	)

	conn, err := amqp.Dial(address)
	require.NoError(t, err)
	defer conn.Close()

	ch, err := conn.Channel()
	require.NoError(t, err)
	defer ch.Close()

	_, err = ch.QueueDeclare(
		statusQueue,
		true,
		false,
		false,
		false,
		nil,
	)
	require.NoError(t, err)

	messages, err := ch.Consume(
		statusQueue,
		"integration-test",
		false,
		false,
		false,
		false,
		nil,
	)
	require.NoError(t, err)

	now := time.Now().UTC()

	event := eventRequest{
		ID:           "integration-notification-event",
		Title:        "Notification test",
		StartAt:      now.Add(10 * time.Second).Format(time.RFC3339),
		EndAt:        now.Add(time.Minute).Format(time.RFC3339),
		UserID:       "integration-user",
		NotifyBefore: "5s",
	}

	createEvent(t, event)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			t.Fatal("notification status was not received")

		case message, ok := <-messages:
			if !ok {
				t.Fatal("notification status channel was closed")
			}

			var status notificationStatus

			err := json.Unmarshal(message.Body, &status)
			require.NoError(t, err)

			if status.EventID != event.ID {
				require.NoError(t, message.Ack(false))
				continue
			}

			require.Equal(t, "sent", status.Status)
			require.Empty(t, status.Error)

			require.NoError(t, message.Ack(false))

			return
		}
	}
}
