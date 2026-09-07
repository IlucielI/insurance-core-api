package ports

import (
	"context"
	"time"

	"github.com/nats-io/nats.go"
)

type MessageBus interface {
	PublishJSON(ctx context.Context, subject string, payload any) error
	RequestJSON(ctx context.Context, subject string, payload any, result any, timeout time.Duration) error
}

type MessageSubscriber interface {
	Subscribe(subject string, handler func(context.Context, []byte) error) (*nats.Subscription, error)
	QueueSubscribe(subject, queueGroup string, handler func(context.Context, []byte) error) (*nats.Subscription, error)
}
