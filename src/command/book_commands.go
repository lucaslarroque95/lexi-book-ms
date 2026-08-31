package command

import (
	"context"
	"io"

	"lexi/books/models"
	"lexi/books/repositories"
)

// ObjectStorage is the narrow slice of storage.ObjectStorage the book
// commands need — just enough to upload a file and undo that upload.
type ObjectStorage interface {
	Upload(ctx context.Context, key string, body io.Reader, size int64, contentType string) error
	Delete(ctx context.Context, key string) error
}

// IngestionNotifier tells the RAG worker that a book's file is ready to be
// processed into chunks.
type IngestionNotifier interface {
	NotifyBookCreated(ctx context.Context, book, fileKey string) error
}

// UploadFile uploads a book's file to object storage under a pre-generated
// key. Undo deletes it, so a failure later in the chain doesn't leave an
// orphaned object in the bucket.
type UploadFile struct {
	Storage     ObjectStorage
	Key         string
	File        io.Reader
	Size        int64
	ContentType string
}

func (c *UploadFile) Execute(ctx context.Context) error {
	return c.Storage.Upload(ctx, c.Key, c.File, c.Size, c.ContentType)
}

func (c *UploadFile) Undo(ctx context.Context) {
	_ = c.Storage.Delete(ctx, c.Key)
}

// CreateBookRow creates the book row (ID and file key already set by the
// caller). Undo deletes the row. Created holds the result once Execute
// succeeds.
type CreateBookRow struct {
	Repository repositories.BookRepository
	Book       models.Book
	Created    models.Book
}

func (c *CreateBookRow) Execute(ctx context.Context) error {
	created, err := c.Repository.Create(c.Book)
	if err != nil {
		return err
	}
	c.Created = created
	return nil
}

func (c *CreateBookRow) Undo(ctx context.Context) {
	_ = c.Repository.Delete(c.Created.ID)
}

// NotifyIngestion tells the RAG worker to ingest the book's file. Nothing to
// undo: if Execute fails no message went out, and once it succeeds there's
// no way to recall a message already published.
type NotifyIngestion struct {
	Notifier IngestionNotifier
	Book     string
	FileKey  string
}

func (c *NotifyIngestion) Execute(ctx context.Context) error {
	return c.Notifier.NotifyBookCreated(ctx, c.Book, c.FileKey)
}

func (c *NotifyIngestion) Undo(ctx context.Context) {}
