package storage

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// ObjectStorage stores and generates time-limited download links for objects
// kept in an S3-compatible bucket (MinIO in this project).
type ObjectStorage struct {
	client        *s3.Client
	presignClient *s3.PresignClient
	bucket        string
}

// NewObjectStorageFromEnv builds an ObjectStorage from MINIO_ENDPOINT,
// MINIO_ACCESS_KEY, MINIO_SECRET_KEY, and MINIO_BUCKET, falling back to this
// project's local-dev MinIO defaults for whichever aren't set.
func NewObjectStorageFromEnv() *ObjectStorage {
	return NewObjectStorage(
		getEnv("MINIO_ENDPOINT", "http://localhost:9000"),
		getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		getEnv("MINIO_SECRET_KEY", "minioadmin123"),
		getEnv("MINIO_BUCKET", "lexi-books"),
	)
}

func NewObjectStorage(endpoint, accessKey, secretKey, bucket string) *ObjectStorage {
	client := s3.New(s3.Options{
		BaseEndpoint: aws.String(endpoint),
		Region:       "us-east-1",
		Credentials:  credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		UsePathStyle: true,
	})

	return &ObjectStorage{client: client, presignClient: s3.NewPresignClient(client), bucket: bucket}
}

func (s *ObjectStorage) Upload(ctx context.Context, key string, body io.Reader, size int64, contentType string) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		Body:          body,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(contentType),
	})
	return err
}

func (s *ObjectStorage) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return err
}

func (s *ObjectStorage) PresignDownloadURL(ctx context.Context, key string, expiry time.Duration) (string, time.Time, error) {
	request, err := s.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", time.Time{}, err
	}

	return request.URL, time.Now().Add(expiry), nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
