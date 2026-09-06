package s3

import (
	"context"
	"net/url"
	"testing"
	"time"
)

func TestNewClientValidatesConfig(t *testing.T) {
	tests := []Config{
		{},
		{Endpoint: "s3.example.com"},
		{Endpoint: "s3.example.com", AccessKey: "access"},
	}

	for _, tt := range tests {
		if _, err := NewClient(tt); err == nil {
			t.Fatalf("NewClient(%+v) error = nil, want error", tt)
		}
	}
}

func TestPresignGetObject(t *testing.T) {
	client := &Client{
		presignExpiry: time.Hour,
		presignGetObject: func(ctx context.Context, bucketName, objectName string, expiry time.Duration, reqParams url.Values) (*url.URL, error) {
			if bucketName != "insurance-files" || objectName != "documents/abc.pdf" {
				t.Fatalf("unexpected presign args: %s %s", bucketName, objectName)
			}
			if expiry != 15*time.Minute {
				t.Fatalf("unexpected expiry: %s", expiry)
			}
			return url.Parse("https://s3.example.com/insurance-files/documents/abc.pdf?X-Amz-Signature=test")
		},
	}

	link, err := client.PresignGetObject(context.Background(), "insurance-files", "documents/abc.pdf", 15*time.Minute)
	if err != nil {
		t.Fatalf("PresignGetObject() error = %v", err)
	}
	if link != "https://s3.example.com/insurance-files/documents/abc.pdf?X-Amz-Signature=test" {
		t.Fatalf("PresignGetObject() = %q", link)
	}
}

func TestPresignGetObjectUsesDefaultExpiry(t *testing.T) {
	client := &Client{
		presignExpiry: 45 * time.Minute,
		presignGetObject: func(ctx context.Context, bucketName, objectName string, expiry time.Duration, reqParams url.Values) (*url.URL, error) {
			if expiry != 45*time.Minute {
				t.Fatalf("unexpected expiry: %s", expiry)
			}
			return url.Parse("https://s3.example.com/file")
		},
	}

	if _, err := client.PresignGetObject(context.Background(), "bucket", "object", 0); err != nil {
		t.Fatalf("PresignGetObject() error = %v", err)
	}
}

func TestPresignPutObject(t *testing.T) {
	client := &Client{
		presignExpiry: time.Hour,
		presignPutObject: func(ctx context.Context, bucketName, objectName string, expiry time.Duration, reqParams url.Values) (*url.URL, error) {
			if bucketName != "insurance-files" || objectName != "documents/abc.pdf" {
				t.Fatalf("unexpected presign args: %s %s", bucketName, objectName)
			}
			if expiry != 30*time.Minute {
				t.Fatalf("unexpected expiry: %s", expiry)
			}
			return url.Parse("https://s3.example.com/upload?X-Amz-Signature=test")
		},
	}

	link, err := client.PresignPutObject(context.Background(), "insurance-files", "documents/abc.pdf", 30*time.Minute)
	if err != nil {
		t.Fatalf("PresignPutObject() error = %v", err)
	}
	if link != "https://s3.example.com/upload?X-Amz-Signature=test" {
		t.Fatalf("PresignPutObject() = %q", link)
	}
}

func TestPresignPutObjectRejectsTooLongExpiry(t *testing.T) {
	client := &Client{
		presignExpiry: time.Hour,
		presignPutObject: func(ctx context.Context, bucketName, objectName string, expiry time.Duration, reqParams url.Values) (*url.URL, error) {
			t.Fatal("presignPutObject should not be called for invalid expiry")
			return nil, nil
		},
	}

	if _, err := client.PresignPutObject(context.Background(), "bucket", "object", 200*time.Hour); err == nil {
		t.Fatal("PresignPutObject() error = nil, want expiry validation error")
	}
}

func TestPresignPutObjectUsesManualFallback(t *testing.T) {
	client := &Client{
		presignPutObject: func(ctx context.Context, bucketName, objectName string, expiry time.Duration, reqParams url.Values) (*url.URL, error) {
			if expiry != time.Hour {
				t.Fatalf("unexpected expiry: %s", expiry)
			}
			return url.Parse("https://s3.example.com/upload?X-Amz-Signature=test")
		},
	}

	if _, err := client.PresignPutObject(context.Background(), "bucket", "object", 0); err != nil {
		t.Fatalf("PresignPutObject() error = %v", err)
	}
}
