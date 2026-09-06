package ports

import (
	"context"
	"time"
)

type ObjectStorage interface {
	PresignPutObject(ctx context.Context, bucketName, objectName string, expiry time.Duration) (string, error)
	PresignGetObject(ctx context.Context, bucketName, objectName string, expiry time.Duration) (string, error)
}
