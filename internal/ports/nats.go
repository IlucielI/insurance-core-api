package ports

import (
	"context"
	"time"
)

type MessageBus interface {
	PublishJSON(ctx context.Context, subject string, payload any) error
	RequestJSON(ctx context.Context, subject string, payload any, result any, timeout time.Duration) error
}
