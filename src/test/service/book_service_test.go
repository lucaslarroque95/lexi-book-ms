package service_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"lexi/books/models"
	"lexi/books/service"
	"lexi/books/test/testutil"
)

func TestCreateBook_ReturnsBookWithID(t *testing.T) {
	bookRepo := testutil.NewFakeBookRepository()
	svc := service.NewBookService(bookRepo, testutil.NewFakeObjectStorage(), testutil.NewFakeIngestionNotifier())

	created, err := svc.CreateBook(models.Book{Name: "Mistborn"})
	if err != nil {
		t.Fatalf("CreateBook returned error: %v", err)
	}
	if created.ID == "" {
		t.Fatalf("expected created book to have an ID")
	}
}

func TestGetBook_NotFound(t *testing.T) {
	bookRepo := testutil.NewFakeBookRepository()
	svc := service.NewBookService(bookRepo, testutil.NewFakeObjectStorage(), testutil.NewFakeIngestionNotifier())

	if _, err := svc.GetBook("missing"); err == nil {
		t.Fatalf("expected an error for a missing book")
	}
}

func TestGetBookByName_ReturnsMatch(t *testing.T) {
	bookRepo := testutil.NewFakeBookRepository()
	bookRepo.Create(models.Book{Name: "Mistborn"})

	svc := service.NewBookService(bookRepo, testutil.NewFakeObjectStorage(), testutil.NewFakeIngestionNotifier())

	found, err := svc.GetBookByName("Mistborn")
	if err != nil {
		t.Fatalf("GetBookByName returned error: %v", err)
	}
	if found.Name != "Mistborn" {
		t.Fatalf("expected to find Mistborn, got %q", found.Name)
	}
}

func TestListBooks_ReturnsAllBooks(t *testing.T) {
	bookRepo := testutil.NewFakeBookRepository()
	bookRepo.Create(models.Book{Name: "Mistborn"})
	bookRepo.Create(models.Book{Name: "The Well of Ascension"})

	svc := service.NewBookService(bookRepo, testutil.NewFakeObjectStorage(), testutil.NewFakeIngestionNotifier())

	books, err := svc.ListBooks(models.BookFilter{})
	if err != nil {
		t.Fatalf("ListBooks returned error: %v", err)
	}
	if len(books) != 2 {
		t.Fatalf("expected 2 books, got %d", len(books))
	}
}

func TestListBooks_FiltersBySerieID(t *testing.T) {
	bookRepo := testutil.NewFakeBookRepository()
	bookRepo.Create(models.Book{Name: "Mistborn", SerieID: "mistborn-serie"})
	bookRepo.Create(models.Book{Name: "Elantris", SerieID: "other-serie"})

	svc := service.NewBookService(bookRepo, testutil.NewFakeObjectStorage(), testutil.NewFakeIngestionNotifier())

	books, err := svc.ListBooks(models.BookFilter{SerieID: "mistborn-serie"})
	if err != nil {
		t.Fatalf("ListBooks returned error: %v", err)
	}
	if len(books) != 1 || books[0].Name != "Mistborn" {
		t.Fatalf("expected only Mistborn, got %v", books)
	}
}

func TestListBooks_FiltersByMaxBookPosition(t *testing.T) {
	bookRepo := testutil.NewFakeBookRepository()
	bookRepo.Create(models.Book{Name: "Book 1", SeriePosition: 1})
	bookRepo.Create(models.Book{Name: "Book 2", SeriePosition: 2})
	bookRepo.Create(models.Book{Name: "Book 3", SeriePosition: 3})

	svc := service.NewBookService(bookRepo, testutil.NewFakeObjectStorage(), testutil.NewFakeIngestionNotifier())

	maxPosition := 2
	books, err := svc.ListBooks(models.BookFilter{MaxBookPosition: &maxPosition})
	if err != nil {
		t.Fatalf("ListBooks returned error: %v", err)
	}
	if len(books) != 2 {
		t.Fatalf("expected 2 books with position <= 2, got %d", len(books))
	}
}

func TestUpdateBook_ChangesName(t *testing.T) {
	bookRepo := testutil.NewFakeBookRepository()
	created, _ := bookRepo.Create(models.Book{Name: "Mistborn"})

	svc := service.NewBookService(bookRepo, testutil.NewFakeObjectStorage(), testutil.NewFakeIngestionNotifier())

	updated, err := svc.UpdateBook(created.ID, models.Book{Name: "Mistborn: The Final Empire"})
	if err != nil {
		t.Fatalf("UpdateBook returned error: %v", err)
	}
	if updated.Name != "Mistborn: The Final Empire" {
		t.Fatalf("expected book name to be updated, got %q", updated.Name)
	}
}

func TestDeleteBook_RemovesBook(t *testing.T) {
	bookRepo := testutil.NewFakeBookRepository()
	created, _ := bookRepo.Create(models.Book{Name: "Mistborn"})

	svc := service.NewBookService(bookRepo, testutil.NewFakeObjectStorage(), testutil.NewFakeIngestionNotifier())

	if err := svc.DeleteBook(created.ID); err != nil {
		t.Fatalf("DeleteBook returned error: %v", err)
	}
	if _, err := svc.GetBook(created.ID); err == nil {
		t.Fatalf("expected the book to be deleted")
	}
}

func TestCreateBookWithFile_UploadsAndNotifies(t *testing.T) {
	bookRepo := testutil.NewFakeBookRepository()
	objectStorage := testutil.NewFakeObjectStorage()
	notifier := testutil.NewFakeIngestionNotifier()
	svc := service.NewBookService(bookRepo, objectStorage, notifier)

	created, err := svc.CreateBookWithFile(
		models.Book{Name: "Mistborn"},
		strings.NewReader("fake epub bytes"),
		15,
		"mistborn.epub",
		"application/epub+zip",
	)
	if err != nil {
		t.Fatalf("CreateBookWithFile returned error: %v", err)
	}

	expectedKey := "books/" + created.ID + "/mistborn.epub"
	if created.FileKey != expectedKey {
		t.Fatalf("expected file key %q, got %q", expectedKey, created.FileKey)
	}
	if len(objectStorage.UploadedKeys) != 1 || objectStorage.UploadedKeys[0] != expectedKey {
		t.Fatalf("expected the file to be uploaded under %q, got %v", expectedKey, objectStorage.UploadedKeys)
	}
	if len(notifier.Notified) != 1 || notifier.Notified[0] != "Mistborn" {
		t.Fatalf("expected the ingestion worker to be notified about Mistborn, got %v", notifier.Notified)
	}
}

func TestCreateBookWithFile_UploadFailure_NothingIsCreated(t *testing.T) {
	bookRepo := testutil.NewFakeBookRepository()
	objectStorage := testutil.NewFakeObjectStorage()
	objectStorage.UploadErr = errors.New("bucket unreachable")
	notifier := testutil.NewFakeIngestionNotifier()
	svc := service.NewBookService(bookRepo, objectStorage, notifier)

	if _, err := svc.CreateBookWithFile(models.Book{Name: "Mistborn"}, strings.NewReader("x"), 1, "m.epub", "text/plain"); err == nil {
		t.Fatalf("expected the upload error to propagate")
	}
	if len(bookRepo.Books) != 0 {
		t.Fatalf("expected no book row when the upload fails, got %v", bookRepo.Books)
	}
	if len(notifier.Notified) != 0 {
		t.Fatalf("expected no notification when the upload fails, got %v", notifier.Notified)
	}
}

func TestCreateBookWithFile_CreateFailure_UploadedFileIsRemoved(t *testing.T) {
	bookRepo := testutil.NewFakeBookRepository()
	bookRepo.CreateErr = errors.New("database unreachable")
	objectStorage := testutil.NewFakeObjectStorage()
	notifier := testutil.NewFakeIngestionNotifier()
	svc := service.NewBookService(bookRepo, objectStorage, notifier)

	if _, err := svc.CreateBookWithFile(models.Book{Name: "Mistborn"}, strings.NewReader("x"), 1, "m.epub", "text/plain"); err == nil {
		t.Fatalf("expected the create error to propagate")
	}
	if len(objectStorage.UploadedKeys) != 1 || len(objectStorage.DeletedKeys) != 1 || objectStorage.UploadedKeys[0] != objectStorage.DeletedKeys[0] {
		t.Fatalf("expected the uploaded file to be deleted, uploaded=%v deleted=%v", objectStorage.UploadedKeys, objectStorage.DeletedKeys)
	}
	if len(notifier.Notified) != 0 {
		t.Fatalf("expected no notification when creating the book row fails, got %v", notifier.Notified)
	}
}

func TestCreateBookWithFile_NotifyFailure_BookAndFileAreRemoved(t *testing.T) {
	bookRepo := testutil.NewFakeBookRepository()
	objectStorage := testutil.NewFakeObjectStorage()
	notifier := testutil.NewFakeIngestionNotifier()
	notifier.NotifyErr = errors.New("broker unreachable")
	svc := service.NewBookService(bookRepo, objectStorage, notifier)

	if _, err := svc.CreateBookWithFile(models.Book{Name: "Mistborn"}, strings.NewReader("x"), 1, "m.epub", "text/plain"); err == nil {
		t.Fatalf("expected the notify error to propagate")
	}
	if len(bookRepo.Books) != 0 {
		t.Fatalf("expected the book row to be rolled back, got %v", bookRepo.Books)
	}
	if len(objectStorage.UploadedKeys) != 1 || len(objectStorage.DeletedKeys) != 1 || objectStorage.UploadedKeys[0] != objectStorage.DeletedKeys[0] {
		t.Fatalf("expected the uploaded file to be deleted, uploaded=%v deleted=%v", objectStorage.UploadedKeys, objectStorage.DeletedKeys)
	}
}

func TestGetBookDownloadURL_Success(t *testing.T) {
	bookRepo := testutil.NewFakeBookRepository()
	created, _ := bookRepo.Create(models.Book{Name: "Mistborn", FileKey: "mistborn.epub"})

	svc := service.NewBookService(bookRepo, testutil.NewFakeObjectStorage(), testutil.NewFakeIngestionNotifier())

	url, expiresAt, err := svc.GetBookDownloadURL(created.ID)
	if err != nil {
		t.Fatalf("GetBookDownloadURL returned error: %v", err)
	}
	if url == "" {
		t.Fatalf("expected a non-empty download URL")
	}
	if !expiresAt.After(time.Now()) {
		t.Fatalf("expected expiresAt to be in the future, got %v", expiresAt)
	}
}

func TestGetBookDownloadURL_NoFile(t *testing.T) {
	bookRepo := testutil.NewFakeBookRepository()
	created, _ := bookRepo.Create(models.Book{Name: "Mistborn"})

	svc := service.NewBookService(bookRepo, testutil.NewFakeObjectStorage(), testutil.NewFakeIngestionNotifier())

	if _, _, err := svc.GetBookDownloadURL(created.ID); !errors.Is(err, service.ErrBookHasNoFile) {
		t.Fatalf("expected ErrBookHasNoFile, got %v", err)
	}
}

func TestGetBookDownloadURL_NotFound(t *testing.T) {
	bookRepo := testutil.NewFakeBookRepository()
	svc := service.NewBookService(bookRepo, testutil.NewFakeObjectStorage(), testutil.NewFakeIngestionNotifier())

	if _, _, err := svc.GetBookDownloadURL("missing"); err == nil {
		t.Fatalf("expected an error for a missing book")
	}
}
