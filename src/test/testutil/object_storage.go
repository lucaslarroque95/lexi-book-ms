package testutil

import (
	"context"
	"io"
	"time"

	"lexi/books/command"
	"lexi/books/service"
)

var (
	_ service.ObjectStorage     = (*FakeObjectStorage)(nil)
	_ command.IngestionNotifier = (*FakeIngestionNotifier)(nil)
)

// FakeObjectStorage is an in-memory service.ObjectStorage for tests, no bucket required.
type FakeObjectStorage struct {
	PresignErr error
	UploadErr  error
	DeleteErr  error

	UploadedKeys []string
	DeletedKeys  []string
}

func NewFakeObjectStorage() *FakeObjectStorage {
	return &FakeObjectStorage{}
}

func (f *FakeObjectStorage) Upload(ctx context.Context, key string, body io.Reader, size int64, contentType string) error {
	if f.UploadErr != nil {
		return f.UploadErr
	}
	f.UploadedKeys = append(f.UploadedKeys, key)
	return nil
}

func (f *FakeObjectStorage) Delete(ctx context.Context, key string) error {
	if f.DeleteErr != nil {
		return f.DeleteErr
	}
	f.DeletedKeys = append(f.DeletedKeys, key)
	return nil
}

func (f *FakeObjectStorage) PresignDownloadURL(ctx context.Context, key string, expiry time.Duration) (string, time.Time, error) {
	if f.PresignErr != nil {
		return "", time.Time{}, f.PresignErr
	}
	return "https://fake-bucket.local/" + key, time.Now().Add(expiry), nil
}

// FakeIngestionNotifier is an in-memory command.IngestionNotifier for tests, no broker required.
type FakeIngestionNotifier struct {
	NotifyErr error

	Notified []string
}

func NewFakeIngestionNotifier() *FakeIngestionNotifier {
	return &FakeIngestionNotifier{}
}

func (f *FakeIngestionNotifier) NotifyBookCreated(ctx context.Context, book, fileKey string) error {
	if f.NotifyErr != nil {
		return f.NotifyErr
	}
	f.Notified = append(f.Notified, book)
	return nil
}
