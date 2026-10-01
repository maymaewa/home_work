package queue

import "context"

type Message struct {
	Body []byte
}

type Publisher interface {
	Publish(ctx context.Context, message Message) error
}

type Consumer interface {
	Consume(ctx context.Context) (<-chan Message, error)
}

type Queue interface {
	Publisher
	Consumer
	Close() error
}
