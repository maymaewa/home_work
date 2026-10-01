package internalhttp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/maymaewa/home_work/hw12_13_14_15_calendar/internal/storage"
)

const dateFormat = "2006-01-02"

type eventRequest struct {
	ID           string `json:"id,omitempty"`
	Title        string `json:"title"`
	StartAt      string `json:"start_at"`
	EndAt        string `json:"end_at"`
	Description  string `json:"description,omitempty"`
	UserID       string `json:"user_id"`
	NotifyBefore string `json:"notify_before,omitempty"`
}

type eventResponse struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	StartAt      string `json:"start_at"`
	EndAt        string `json:"end_at"`
	Description  string `json:"description,omitempty"`
	UserID       string `json:"user_id"`
	NotifyBefore string `json:"notify_before,omitempty"`
}

func (s *Server) createEvent(w http.ResponseWriter, r *http.Request) {
	var request eventRequest

	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	model, err := eventFromRequest(request)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	createdEvent, err := s.app.CreateEvent(r.Context(), model)
	if err != nil {
		writeStorageError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, eventToResponse(createdEvent))
}

func (s *Server) updateEvent(w http.ResponseWriter, r *http.Request) {
	id, ok := eventIDFromPath(r)
	if !ok {
		writeError(w, http.StatusBadRequest, errors.New("event id is required"))
		return
	}

	var request eventRequest

	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	model, err := eventFromRequest(request)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	model.ID = id

	if err := s.app.UpdateEvent(r.Context(), model); err != nil {
		writeStorageError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, eventToResponse(model))
}

func (s *Server) deleteEvent(w http.ResponseWriter, r *http.Request) {
	id, ok := eventIDFromPath(r)
	if !ok {
		writeError(w, http.StatusBadRequest, errors.New("event id is required"))
		return
	}

	if err := s.app.DeleteEvent(r.Context(), id); err != nil {
		writeStorageError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) listEventsForDay(w http.ResponseWriter, r *http.Request) {
	s.listEvents(w, r, s.app.ListEventsForDay)
}

func (s *Server) listEventsForWeek(w http.ResponseWriter, r *http.Request) {
	s.listEvents(w, r, s.app.ListEventsForWeek)
}

func (s *Server) listEventsForMonth(w http.ResponseWriter, r *http.Request) {
	s.listEvents(w, r, s.app.ListEventsForMonth)
}

type listEventsFunc func(context.Context, time.Time) ([]storage.Event, error)

func (s *Server) listEvents(
	w http.ResponseWriter,
	r *http.Request,
	listFunc listEventsFunc,
) {
	dateValue := r.URL.Query().Get("date")
	if dateValue == "" {
		writeError(w, http.StatusBadRequest, errors.New("date is required"))
		return
	}

	date, err := time.Parse(dateFormat, dateValue)
	if err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid date format, expected YYYY-MM-DD"))
		return
	}

	events, err := listFunc(r.Context(), date)
	if err != nil {
		writeStorageError(w, err)
		return
	}

	response := make([]eventResponse, 0, len(events))

	for _, model := range events {
		response = append(response, eventToResponse(model))
	}

	writeJSON(w, http.StatusOK, response)
}

func eventFromRequest(request eventRequest) (storage.Event, error) {
	startAt, err := time.Parse(time.RFC3339, request.StartAt)
	if err != nil {
		return storage.Event{}, errors.New("invalid start_at")
	}

	endAt, err := time.Parse(time.RFC3339, request.EndAt)
	if err != nil {
		return storage.Event{}, errors.New("invalid end_at")
	}

	var notifyBefore *time.Duration

	if request.NotifyBefore != "" {
		duration, err := time.ParseDuration(request.NotifyBefore)
		if err != nil {
			return storage.Event{}, errors.New("invalid notify_before")
		}

		notifyBefore = &duration
	}

	return storage.Event{
		ID:           request.ID,
		Title:        request.Title,
		StartAt:      startAt,
		EndAt:        endAt,
		Description:  request.Description,
		UserID:       request.UserID,
		NotifyBefore: notifyBefore,
	}, nil
}

func eventToResponse(model storage.Event) eventResponse {
	response := eventResponse{
		ID:          model.ID,
		Title:       model.Title,
		StartAt:     model.StartAt.Format(time.RFC3339),
		EndAt:       model.EndAt.Format(time.RFC3339),
		Description: model.Description,
		UserID:      model.UserID,
	}

	if model.NotifyBefore != nil {
		response.NotifyBefore = model.NotifyBefore.String()
	}

	return response
}

func eventIDFromPath(r *http.Request) (string, bool) {
	id := strings.TrimPrefix(r.URL.Path, "/events/")

	if id == "" || id == r.URL.Path {
		return "", false
	}

	return id, true
}

func decodeJSON(r *http.Request, target any) error {
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(target); err != nil {
		return errors.New("invalid JSON")
	}

	return nil
}

func writeJSON(w http.ResponseWriter, statusCode int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, statusCode int, err error) {
	writeJSON(w, statusCode, map[string]string{
		"error": err.Error(),
	})
}

func writeStorageError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, storage.ErrEventNotFound):
		writeError(w, http.StatusNotFound, err)

	case errors.Is(err, storage.ErrEventAlreadyExists):
		writeError(w, http.StatusConflict, err)

	case errors.Is(err, storage.ErrDateBusy):
		writeError(w, http.StatusConflict, err)

	default:
		writeError(w, http.StatusInternalServerError, err)
	}
}
