package routes_test

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"lexi/books/models"
	"lexi/books/routes"
	"lexi/books/service"
	"lexi/books/test/testutil"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newBookHandler(t *testing.T) (*routes.BookHandler, *testutil.FakeBookRepository) {
	t.Helper()

	bookRepo := testutil.NewFakeBookRepository()
	svc := service.NewBookService(bookRepo, testutil.NewFakeObjectStorage(), testutil.NewFakeIngestionNotifier())
	return routes.NewBookHandler(svc), bookRepo
}

func TestCreateBook_Success(t *testing.T) {
	handler, bookRepo := newBookHandler(t)

	router := gin.New()
	router.POST("/books", handler.CreateBook)

	body, _ := json.Marshal(map[string]string{"book": "Mistborn"})
	req := httptest.NewRequest(http.MethodPost, "/books", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	if len(bookRepo.Books) != 1 {
		t.Fatalf("expected 1 book to be persisted, got %d", len(bookRepo.Books))
	}
}

func TestCreateBook_WithFile_Success(t *testing.T) {
	handler, bookRepo := newBookHandler(t)

	router := gin.New()
	router.POST("/books", handler.CreateBook)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	writer.WriteField("book", "Mistborn")
	writer.WriteField("seriePosition", "1")
	fileWriter, err := writer.CreateFormFile("file", "mistborn.epub")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	fileWriter.Write([]byte("fake epub bytes"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/books", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	fileKey, _ := resp["fileKey"].(string)
	if fileKey == "" || !strings.HasSuffix(fileKey, "/mistborn.epub") {
		t.Fatalf("expected a fileKey ending in /mistborn.epub, got %v", resp)
	}
	if len(bookRepo.Books) != 1 {
		t.Fatalf("expected 1 book to be persisted, got %d", len(bookRepo.Books))
	}
}

func TestCreateBook_WithFile_MissingFileReturns400(t *testing.T) {
	handler, _ := newBookHandler(t)

	router := gin.New()
	router.POST("/books", handler.CreateBook)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	writer.WriteField("book", "Mistborn")
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/books", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCreateBook_MissingNameReturns400(t *testing.T) {
	handler, _ := newBookHandler(t)

	router := gin.New()
	router.POST("/books", handler.CreateBook)

	body, _ := json.Marshal(map[string]string{})
	req := httptest.NewRequest(http.MethodPost, "/books", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestGetBookByID_NotFoundReturns404(t *testing.T) {
	handler, _ := newBookHandler(t)

	router := gin.New()
	router.GET("/books/:id", handler.GetBookByID)

	req := httptest.NewRequest(http.MethodGet, "/books/missing", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestGetBookDownloadURL_Success(t *testing.T) {
	handler, bookRepo := newBookHandler(t)
	created, _ := bookRepo.Create(models.Book{Name: "Mistborn", FileKey: "mistborn.epub"})

	router := gin.New()
	router.GET("/books/:id/download-url", handler.GetBookDownloadURL)

	req := httptest.NewRequest(http.MethodGet, "/books/"+created.ID+"/download-url", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["downloadUrl"] == "" || resp["downloadUrl"] == nil {
		t.Fatalf("expected a non-empty downloadUrl, got %v", resp)
	}
}

func TestGetBookDownloadURL_NoFileReturns400(t *testing.T) {
	handler, bookRepo := newBookHandler(t)
	created, _ := bookRepo.Create(models.Book{Name: "Mistborn"})

	router := gin.New()
	router.GET("/books/:id/download-url", handler.GetBookDownloadURL)

	req := httptest.NewRequest(http.MethodGet, "/books/"+created.ID+"/download-url", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestGetBookDownloadURL_NotFoundReturns404(t *testing.T) {
	handler, _ := newBookHandler(t)

	router := gin.New()
	router.GET("/books/:id/download-url", handler.GetBookDownloadURL)

	req := httptest.NewRequest(http.MethodGet, "/books/missing/download-url", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestGetBookByName_Success(t *testing.T) {
	handler, bookRepo := newBookHandler(t)
	bookRepo.Create(models.Book{Name: "Mistborn"})

	router := gin.New()
	router.GET("/books/name/:name", handler.GetBookByName)

	req := httptest.NewRequest(http.MethodGet, "/books/name/Mistborn", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["book"] != "Mistborn" {
		t.Fatalf("expected book Mistborn, got %v", resp)
	}
}

func TestGetBookByName_NotFoundReturns404(t *testing.T) {
	handler, _ := newBookHandler(t)

	router := gin.New()
	router.GET("/books/name/:name", handler.GetBookByName)

	req := httptest.NewRequest(http.MethodGet, "/books/name/missing", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestGetBooks_FiltersBySerieID(t *testing.T) {
	handler, bookRepo := newBookHandler(t)
	bookRepo.Create(models.Book{Name: "Mistborn", SerieID: "mistborn-serie"})
	bookRepo.Create(models.Book{Name: "Elantris", SerieID: "other-serie"})

	router := gin.New()
	router.GET("/books", handler.GetBooks)

	req := httptest.NewRequest(http.MethodGet, "/books?serieId=mistborn-serie", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp) != 1 || resp[0]["book"] != "Mistborn" {
		t.Fatalf("expected only Mistborn, got %v", resp)
	}
}

func TestGetBooks_InvalidBookPositionReturns400(t *testing.T) {
	handler, _ := newBookHandler(t)

	router := gin.New()
	router.GET("/books", handler.GetBooks)

	req := httptest.NewRequest(http.MethodGet, "/books?bookPosition=not-a-number", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestUpdateBook_ChangesName(t *testing.T) {
	handler, bookRepo := newBookHandler(t)
	created, _ := bookRepo.Create(models.Book{Name: "Mistborn"})

	router := gin.New()
	router.PUT("/books/:id", handler.UpdateBook)

	body, _ := json.Marshal(map[string]string{"book": "Mistborn: The Final Empire"})
	req := httptest.NewRequest(http.MethodPut, "/books/"+created.ID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["book"] != "Mistborn: The Final Empire" {
		t.Fatalf("expected book name to be updated, got %v", resp)
	}
}

func TestDeleteBook_Success(t *testing.T) {
	handler, bookRepo := newBookHandler(t)
	created, _ := bookRepo.Create(models.Book{Name: "Mistborn"})

	router := gin.New()
	router.DELETE("/books/:id", handler.DeleteBook)

	req := httptest.NewRequest(http.MethodDelete, "/books/"+created.ID, nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}
