package sender

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/maymaewa/home_work/hw12_13_14_15_calendar/internal/notification"
	"github.com/maymaewa/home_work/hw12_13_14_15_calendar/internal/queue"
)

type Logger interface {
	Info(msg string)
	Error(msg string)
}

type Sender struct {
	consumer queue.Consumer
	logger   Logger
}

func New(consumer queue.Consumer, logger Logger) *Sender {
	return &Sender{
		consumer: consumer,
		logger:   logger,
	}
}

func (s *Sender) Start(ctx context.Context) error {
	messages, err := s.consumer.Consume(ctx)
	if err != nil {
		return fmt.Errorf("start consuming messages: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil

		case message, ok := <-messages:
			if !ok {
				return nil
			}

			if err := s.process(message); err != nil {
				s.logger.Error(err.Error())
			}
		}
	}
}

func (s *Sender) process(message queue.Message) error {
	var notification notification.Notification

	if err := json.Unmarshal(message.Body, &notification); err != nil {
		return fmt.Errorf("unmarshal notification: %w", err)
	}

	s.logger.Info(
		fmt.Sprintf(
			"notification: event_id=%s title=%q date=%s user_id=%s",
			notification.EventID,
			notification.Title,
			notification.Date.Format("2006-01-02 15:04:05"),
			notification.UserID,
		),
	)

	return nil
}
