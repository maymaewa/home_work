package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/maymaewa/home_work/hw12_13_14_15_calendar/internal/notification"
	"github.com/maymaewa/home_work/hw12_13_14_15_calendar/internal/queue"
	"github.com/maymaewa/home_work/hw12_13_14_15_calendar/internal/storage"
)

type Application interface {
	ListEventsForNotification(
		ctx context.Context,
		from time.Time,
		to time.Time,
	) ([]storage.Event, error)

	DeleteEventsOlderThan(
		ctx context.Context,
		before time.Time,
	) error
}

type Logger interface {
	Info(msg string)
	Error(msg string)
}

type Scheduler struct {
	app       Application
	publisher queue.Publisher
	logger    Logger
	interval  time.Duration
}

func New(
	app Application,
	publisher queue.Publisher,
	logger Logger,
	interval time.Duration,
) *Scheduler {
	return &Scheduler{
		app:       app,
		publisher: publisher,
		logger:    logger,
		interval:  interval,
	}
}

func (s *Scheduler) Start(ctx context.Context) error {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	lastRun := time.Now().Add(-s.interval)

	for {
		if err := s.run(ctx, lastRun, time.Now()); err != nil {
			s.logger.Error(err.Error())
		}

		lastRun = time.Now()

		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

func (s *Scheduler) run(
	ctx context.Context,
	from time.Time,
	to time.Time,
) error {
	events, err := s.app.ListEventsForNotification(ctx, from, to)
	if err != nil {
		return fmt.Errorf("get events for notification: %w", err)
	}

	for _, event := range events {
		notification := notification.Notification{
			EventID: event.ID,
			Title:   event.Title,
			Date:    event.StartAt,
			UserID:  event.UserID,
		}

		body, err := json.Marshal(notification)
		if err != nil {
			return fmt.Errorf("marshal notification: %w", err)
		}

		if err := s.publisher.Publish(ctx, queue.Message{
			Body: body,
		}); err != nil {
			return fmt.Errorf("publish notification: %w", err)
		}

		s.logger.Info(
			fmt.Sprintf(
				"notification published: event_id=%s user_id=%s",
				event.ID,
				event.UserID,
			),
		)
	}

	before := to.AddDate(-1, 0, 0)

	if err := s.app.DeleteEventsOlderThan(ctx, before); err != nil {
		return fmt.Errorf("delete old events: %w", err)
	}

	return nil
}
