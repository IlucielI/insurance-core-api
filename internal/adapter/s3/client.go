package s3

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const maxPresignExpiry = 7 * 24 * time.Hour

type Config struct {
	Endpoint       string
	AccessKey      string
	SecretKey      string
	Region         string
	UseSSL         bool
	ForcePathStyle bool
	PresignExpiry  time.Duration
}

type Client struct {
	client           *minio.Client
	presignExpiry    time.Duration
	presignGetObject func(ctx context.Context, bucketName, objectName string, expiry time.Duration, reqParams url.Values) (*url.URL, error)
	presignPutObject func(ctx context.Context, bucketName, objectName string, expiry time.Duration, reqParams url.Values) (*url.URL, error)
}

func NewClient(config Config) (*Client, error) {
	endpoint := strings.TrimSpace(config.Endpoint)
	if endpoint == "" {
		return nil, errors.New("Endpoint is required")
	}
	accessKey := strings.TrimSpace(config.AccessKey)
	if accessKey == "" {
		return nil, errors.New("AccessKey is required")
	}
	secretKey := strings.TrimSpace(config.SecretKey)
	if secretKey == "" {
		return nil, errors.New("SecretKey is required")
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:        credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure:       config.UseSSL,
		Region:       strings.TrimSpace(config.Region),
		BucketLookup: bucketLookup(config.ForcePathStyle),
	})
	if err != nil {
		return nil, err
	}

	return &Client{
		client:           client,
		presignExpiry:    normalizeExpiry(config.PresignExpiry),
		presignGetObject: client.PresignedGetObject,
		presignPutObject: client.PresignedPutObject,
	}, nil
}

func (client *Client) EnsureBucketExists(ctx context.Context, bucketName string) error {
	if client == nil || client.client == nil {
		return errors.New("storage client is required")
	}
	bucketName = strings.TrimSpace(bucketName)
	if bucketName == "" {
		return errors.New("bucketName is required")
	}

	exists, err := client.client.BucketExists(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("check bucket exists: %w", err)
	}
	if exists {
		return nil
	}

	if err := client.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{}); err != nil {
		return fmt.Errorf("create bucket: %w", err)
	}

	return nil
}

func (client *Client) PresignPutObject(ctx context.Context, bucketName, objectName string, expiry time.Duration) (string, error) {
	bucketName = strings.TrimSpace(bucketName)
	if bucketName == "" {
		return "", errors.New("bucketName is required")
	}
	objectName = strings.TrimSpace(objectName)
	if objectName == "" {
		return "", errors.New("objectName is required")
	}

	expires, err := client.resolvePresignExpiry(expiry)
	if err != nil {
		return "", err
	}

	presignedURL, err := client.presignPutObject(ctx, bucketName, objectName, expires, nil)
	if err != nil {
		return "", fmt.Errorf("presign put object: %w", err)
	}

	return presignedURL.String(), nil
}

func (client *Client) PresignGetObject(ctx context.Context, bucketName, objectName string, expiry time.Duration) (string, error) {
	bucketName = strings.TrimSpace(bucketName)
	if bucketName == "" {
		return "", errors.New("bucketName is required")
	}
	objectName = strings.TrimSpace(objectName)
	if objectName == "" {
		return "", errors.New("objectName is required")
	}

	expires, err := client.resolvePresignExpiry(expiry)
	if err != nil {
		return "", err
	}

	presignedURL, err := client.presignGetObject(ctx, bucketName, objectName, expires, nil)
	if err != nil {
		return "", fmt.Errorf("presign get object: %w", err)
	}

	return presignedURL.String(), nil
}

func normalizeExpiry(expiry time.Duration) time.Duration {
	if expiry <= 0 {
		return time.Hour
	}
	if expiry > maxPresignExpiry {
		return maxPresignExpiry
	}
	return expiry
}

func (client *Client) resolvePresignExpiry(expiry time.Duration) (time.Duration, error) {
	if expiry > maxPresignExpiry {
		return 0, errors.New("expiry exceeds maximum allowed duration")
	}
	if expiry > 0 {
		return expiry, nil
	}

	if client != nil && client.presignExpiry > 0 {
		if client.presignExpiry > maxPresignExpiry {
			return 0, errors.New("expiry exceeds maximum allowed duration")
		}
		return client.presignExpiry, nil
	}

	return time.Hour, nil
}

func bucketLookup(forcePathStyle bool) minio.BucketLookupType {
	if forcePathStyle {
		return minio.BucketLookupPath
	}

	return minio.BucketLookupAuto
}
