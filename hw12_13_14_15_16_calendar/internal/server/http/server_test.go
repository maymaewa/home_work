package internalhttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/maymaewa/home_work/hw12_13_14_15_calendar/internal/app"
	"github.com/maymaewa/home_work/hw12_13_14_15_calendar/internal/logger"
	memorystorage "github.com/maymaewa/home_work/hw12_13_14_15_calendar/internal/storage/memory"
	"github.com/stretchr/testify/require"
)

const (
	testHost = "127.0.0.1"
)

func TestCreateAndListEvent(t *testing.T) {
	storage := memorystorage.New()
	logg := logger.New("error")
	calendar := app.New(logg, storage)

	server := NewServer(
		logg,
		calendar,
		Config{
			Host: testHost,
			Port: 0,
		},
	)

	createRequest := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		"/events",
		strings.NewReader(`{
		"title":"HTTP test event",
		"start_at":"2026-09-25T15:00:00Z",
		"end_at":"2026-09-25T16:00:00Z",
		"description":"created through HTTP",
		"user_id":"user-1",
		"notify_before":"30m"
	}`),
	)
	createRequest.Header.Set("Content-Type", "application/json")

	createResponse := httptest.NewRecorder()

	server.server.Handler.ServeHTTP(createResponse, createRequest)

	require.Equal(t, http.StatusCreated, createResponse.Code)
	require.Contains(t, createResponse.Body.String(), `"id"`)
	require.Contains(t, createResponse.Body.String(), `"HTTP test event"`)

	listRequest := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"/events/day?date=2026-09-25",
		nil,
	)

	listResponse := httptest.NewRecorder()

	server.server.Handler.ServeHTTP(listResponse, listRequest)

	require.Equal(t, http.StatusOK, listResponse.Code)
	require.Contains(t, listResponse.Body.String(), `"HTTP test event"`)
	require.Contains(t, listResponse.Body.String(), `"user-1"`)
}

func TestUpdateAndDeleteEvent(t *testing.T) {
	storage := memorystorage.New()
	logg := logger.New("error")
	calendar := app.New(logg, storage)

	server := NewServer(
		logg,
		calendar,
		Config{
			Host: testHost,
			Port: 0,
		},
	)

	createRequest := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		"/events",
		strings.NewReader(`{
			"id":"http-crud-event",
			"title":"Original title",
			"start_at":"2026-09-26T15:00:00Z",
			"end_at":"2026-09-26T16:00:00Z",
			"user_id":"user-1"
		}`),
	)
	createRequest.Header.Set("Content-Type", "application/json")

	createResponse := httptest.NewRecorder()

	server.server.Handler.ServeHTTP(createResponse, createRequest)

	require.Equal(t, http.StatusCreated, createResponse.Code)

	updateRequest := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodPut,
		"/events/http-crud-event",
		strings.NewReader(`{
			"title":"Updated title",
			"start_at":"2026-09-26T15:00:00Z",
			"end_at":"2026-09-26T16:00:00Z",
			"description":"updated description",
			"user_id":"user-1",
			"notify_before":"15m"
		}`),
	)
	updateRequest.Header.Set("Content-Type", "application/json")

	updateResponse := httptest.NewRecorder()

	server.server.Handler.ServeHTTP(updateResponse, updateRequest)

	require.Equal(t, http.StatusOK, updateResponse.Code)
	require.Contains(t, updateResponse.Body.String(), `"Updated title"`)
	require.Contains(t, updateResponse.Body.String(), `"updated description"`)
	require.Contains(t, updateResponse.Body.String(), `"15m0s"`)

	deleteRequest := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodDelete,
		"/events/http-crud-event",
		nil,
	)

	deleteResponse := httptest.NewRecorder()

	server.server.Handler.ServeHTTP(deleteResponse, deleteRequest)

	require.Equal(t, http.StatusNoContent, deleteResponse.Code)

	listRequest := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"/events/day?date=2026-09-26",
		nil,
	)

	listResponse := httptest.NewRecorder()

	server.server.Handler.ServeHTTP(listResponse, listRequest)

	require.Equal(t, http.StatusOK, listResponse.Code)
	require.NotContains(t, listResponse.Body.String(), "http-crud-event")
}

func TestEventErrors(t *testing.T) {
	storage := memorystorage.New()
	logg := logger.New("error")
	calendar := app.New(logg, storage)

	server := NewServer(
		logg,
		calendar,
		Config{
			Host: testHost,
			Port: 0,
		},
	)

	t.Run("delete_not_found", func(t *testing.T) {
		request := httptest.NewRequestWithContext(
			context.Background(),
			http.MethodDelete,
			"/events/not-found",
			nil,
		)

		response := httptest.NewRecorder()

		server.server.Handler.ServeHTTP(response, request)

		require.Equal(t, http.StatusNotFound, response.Code)
	})

	t.Run("create_duplicate", func(t *testing.T) {
		body := `{
			"id":"duplicate-event",
			"title":"Duplicate event",
			"start_at":"2026-09-27T15:00:00Z",
			"end_at":"2026-09-27T16:00:00Z",
			"user_id":"user-1"
		}`

		request := httptest.NewRequestWithContext(
			context.Background(),
			http.MethodPost,
			"/events",
			strings.NewReader(body),
		)
		request.Header.Set("Content-Type", "application/json")

		response := httptest.NewRecorder()

		server.server.Handler.ServeHTTP(response, request)

		require.Equal(t, http.StatusCreated, response.Code)

		request = httptest.NewRequestWithContext(
			context.Background(),
			http.MethodPost,
			"/events",
			strings.NewReader(body),
		)
		request.Header.Set("Content-Type", "application/json")

		response = httptest.NewRecorder()

		server.server.Handler.ServeHTTP(response, request)

		require.Equal(t, http.StatusConflict, response.Code)
	})
}
