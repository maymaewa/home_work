package rabbitmq

import (
	"context"
	"fmt"

	"github.com/maymaewa/home_work/hw12_13_14_15_calendar/internal/queue"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	Queue    string
}

type RabbitMQ struct {
	connection *amqp.Connection
	channel    *amqp.Channel
	queue      amqp.Queue
}

func New(config Config) (*RabbitMQ, error) {
	address := fmt.Sprintf(
		"amqp://%s:%s@%s:%d/",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
	)

	connection, err := amqp.Dial(address)
	if err != nil {
		return nil, fmt.Errorf("connect to rabbitmq: %w", err)
	}

	channel, err := connection.Channel()
	if err != nil {
		_ = connection.Close()
		return nil, fmt.Errorf("open rabbitmq channel: %w", err)
	}

	q, err := channel.QueueDeclare(
		config.Queue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		_ = channel.Close()
		_ = connection.Close()

		return nil, fmt.Errorf("declare rabbitmq queue: %w", err)
	}

	return &RabbitMQ{
		connection: connection,
		channel:    channel,
		queue:      q,
	}, nil
}

func (r *RabbitMQ) Publish(ctx context.Context, message queue.Message) error {
	err := r.channel.PublishWithContext(
		ctx,
		"",
		r.queue.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        message.Body,
		},
	)
	if err != nil {
		return fmt.Errorf("publish message: %w", err)
	}

	return nil
}

func (r *RabbitMQ) Consume(ctx context.Context) (<-chan queue.Message, error) {
	messages, err := r.channel.Consume(
		r.queue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("consume messages: %w", err)
	}

	result := make(chan queue.Message)

	go func() {
		defer close(result)

		for {
			select {
			case <-ctx.Done():
				return

			case message, ok := <-messages:
				if !ok {
					return
				}

				select {
				case result <- queue.Message{Body: message.Body}:
					if err := message.Ack(false); err != nil {
						return
					}
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return result, nil
}

func (r *RabbitMQ) Close() error {
	var firstErr error

	if err := r.channel.Close(); err != nil {
		firstErr = err
	}

	if err := r.connection.Close(); err != nil && firstErr == nil {
		firstErr = err
	}

	return firstErr
}

var _ queue.Queue = (*RabbitMQ)(nil)
