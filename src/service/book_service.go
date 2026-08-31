package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"lexi/books/command"
	"lexi/books/models"
	"lexi/books/repositories"

	"github.com/google/uuid"
)

const downloadURLExpiry = 15 * time.Minute

var ErrBookHasNoFile = errors.New("book has no file")

// ObjectStorage stores a book's file and generates a time-limited download
// link for it. It extends command.ObjectStorage with the one extra
// capability BookService needs outside of any command.
type ObjectStorage interface {
	command.ObjectStorage
	PresignDownloadURL(ctx context.Context, key string, expiry time.Duration) (string, time.Time, error)
}

type BookService struct {
	repository repositories.BookRepository
	storage    ObjectStorage
	notifier   command.IngestionNotifier
}

func NewBookService(repository repositories.BookRepository, storage ObjectStorage, notifier command.IngestionNotifier) *BookService {
	return &BookService{repository: repository, storage: storage, notifier: notifier}
}

func (rs *BookService) CreateBook(Book models.Book) (models.Book, error) {
	return rs.repository.Create(Book)
}

// CreateBookWithFile uploads the book's file to object storage, creates the
// book row, and notifies the RAG worker to ingest it — as a single unit via
// the command pattern: the book ID and file key are decided upfront so each
// step is self-contained, and if any step fails, every step that already
// succeeded is undone (uploaded file removed, book row deleted) so the book
// never ends up half-created.
func (rs *BookService) CreateBookWithFile(book models.Book, file io.Reader, size int64, filename, contentType string) (models.Book, error) {
	book.ID = uuid.NewString()
	book.FileKey = fmt.Sprintf("books/%s/%s", book.ID, filename)

	upload := &command.UploadFile{Storage: rs.storage, Key: book.FileKey, File: file, Size: size, ContentType: contentType}
	createRow := &command.CreateBookRow{Repository: rs.repository, Book: book}
	notify := &command.NotifyIngestion{Notifier: rs.notifier, Book: book.Name, FileKey: book.FileKey}

	if err := command.Run(context.Background(), upload, createRow, notify); err != nil {
		return models.Book{}, err
	}

	return createRow.Created, nil
}

func (rs *BookService) GetBook(BookID string) (models.Book, error) {
	return rs.repository.Get(BookID)
}

func (rs *BookService) GetBookByName(name string) (models.Book, error) {
	return rs.repository.GetByName(name)
}

func (rs *BookService) ListBooks(filter models.BookFilter) ([]models.Book, error) {
	return rs.repository.GetAll(filter)
}

func (rs *BookService) UpdateBook(BookID string, Book models.Book) (models.Book, error) {
	return rs.repository.Update(BookID, Book)
}

func (rs *BookService) DeleteBook(BookID string) error {
	return rs.repository.Delete(BookID)
}

func (rs *BookService) GetBookDownloadURL(bookID string) (string, time.Time, error) {
	book, err := rs.repository.Get(bookID)
	if err != nil {
		return "", time.Time{}, err
	}
	if book.FileKey == "" {
		return "", time.Time{}, ErrBookHasNoFile
	}

	return rs.storage.PresignDownloadURL(context.Background(), book.FileKey, downloadURLExpiry)
}
