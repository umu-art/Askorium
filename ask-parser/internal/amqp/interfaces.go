package amqp

import "context"

type MessagePublisher interface {
	Publish(ctx context.Context, body []byte) error
	PublishWithKey(ctx context.Context, routingKey string, body []byte) error
}

type MessageConsumer interface {
	Queue() string
}
