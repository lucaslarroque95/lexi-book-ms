package routes

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"lexi/books/models"
	"lexi/books/schemas"
	"lexi/books/service"

	"github.com/gin-gonic/gin"
)

type BookHandler struct {
	service *service.BookService
}

func NewBookHandler(service *service.BookService) *BookHandler {
	return &BookHandler{service: service}
}

// GetBooks godoc
// @Summary      List books
// @Description  Lists books, optionally filtered by serie, universe (resolved via the book's serie), or position within the serie
// @Tags         books
// @Produce      json
// @Param        serieId          query     string  false  "Filter by serie ID"
// @Param        universeId       query     string  false  "Filter by universe ID"
// @Param        bookPosition     query     int     false  "Filter by exact position in the serie"
// @Param        maxBookPosition  query     int     false  "Filter by position <= this value"
// @Success      200  {array}   schemas.BookRead
// @Failure      400  {object}  schemas.ErrorResponse
// @Failure      500  {object}  schemas.ErrorResponse
// @Security     BearerAuth
// @Router       /books [get]
func (h *BookHandler) GetBooks(c *gin.Context) {
	filter, err := parseBookFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	books, err := h.service.ListBooks(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch books"})
		return
	}

	c.JSON(http.StatusOK, toBookReadList(books))
}

// GetBookByID godoc
// @Summary      Get a book by ID
// @Tags         books
// @Produce      json
// @Param        id  path      string  true  "Book ID"
// @Success      200  {object}  schemas.BookRead
// @Failure      404  {object}  schemas.ErrorResponse
// @Security     BearerAuth
// @Router       /books/{id} [get]
func (h *BookHandler) GetBookByID(c *gin.Context) {
	id := c.Param("id")

	book, err := h.service.GetBook(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	}

	c.JSON(http.StatusOK, toBookRead(book))
}

// GetBookByName godoc
// @Summary      Get a book by name
// @Tags         books
// @Produce      json
// @Param        name  path      string  true  "Book name"
// @Success      200  {object}  schemas.BookRead
// @Failure      404  {object}  schemas.ErrorResponse
// @Security     BearerAuth
// @Router       /books/name/{name} [get]
func (h *BookHandler) GetBookByName(c *gin.Context) {
	name := c.Param("name")

	book, err := h.service.GetBookByName(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	}

	c.JSON(http.StatusOK, toBookRead(book))
}

// GetBookDownloadURL godoc
// @Summary      Get a time-limited download link for a book's file
// @Description  Presigns a short-lived URL (15 min) straight to the bucket. Fails with 400 if the book has no file attached yet.
// @Tags         books
// @Produce      json
// @Param        id  path      string  true  "Book ID"
// @Success      200  {object}  schemas.BookDownloadURL
// @Failure      400  {object}  schemas.ErrorResponse  "book has no file"
// @Failure      404  {object}  schemas.ErrorResponse
// @Security     BearerAuth
// @Router       /books/{id}/download-url [get]
func (h *BookHandler) GetBookDownloadURL(c *gin.Context) {
	id := c.Param("id")

	url, expiresAt, err := h.service.GetBookDownloadURL(id)
	if err != nil {
		if errors.Is(err, service.ErrBookHasNoFile) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "book has no file"})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	}

	c.JSON(http.StatusOK, schemas.BookDownloadURL{DownloadURL: url, ExpiresAt: expiresAt})
}

// CreateBook godoc
// @Summary      Create a book
// @Description  Accepts multipart/form-data with a "file" part: uploads it to object storage, saves the resulting fileKey, and notifies the RAG worker to ingest it. Also accepts a plain application/json body shaped like the request schema for metadata-only creation (no file, no ingestion notify) — send it as JSON instead of multipart in that case.
// @Tags         books
// @Accept       multipart/form-data
// @Produce      json
// @Param        book           formData  string  true   "Book name"
// @Param        serieId        formData  string  false  "Serie ID"
// @Param        seriePosition  formData  int     false  "Position within the serie"
// @Param        file           formData  file    true   "The book's file (e.g. epub)"
// @Success      201  {object}  schemas.BookRead
// @Failure      400  {object}  schemas.ErrorResponse
// @Failure      500  {object}  schemas.ErrorResponse
// @Security     BearerAuth
// @Router       /books [post]
func (h *BookHandler) CreateBook(c *gin.Context) {
	if strings.HasPrefix(c.ContentType(), "multipart/form-data") {
		h.createBookWithFile(c)
		return
	}

	var payload schemas.BookCreate
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	book := models.Book{
		Name:          payload.Name,
		UserID:        c.GetString("userId"),
		SerieID:       payload.SerieID,
		SeriePosition: payload.SeriePosition,
	}

	created, err := h.service.CreateBook(book)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create book"})
		return
	}

	c.JSON(http.StatusCreated, toBookRead(created))
}

// createBookWithFile handles a multipart create: the book's file travels in
// the request, gets uploaded to object storage, and the RAG worker is
// notified to ingest it — all before the book is considered created.
func (h *BookHandler) createBookWithFile(c *gin.Context) {
	name := c.PostForm("book")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "book is required"})
		return
	}

	seriePosition := 0
	if raw := c.PostForm("seriePosition"); raw != "" {
		position, err := strconv.Atoi(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid seriePosition"})
			return
		}
		seriePosition = position
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not read file"})
		return
	}
	defer file.Close()

	book := models.Book{
		Name:          name,
		UserID:        c.GetString("userId"),
		SerieID:       c.PostForm("serieId"),
		SeriePosition: seriePosition,
	}

	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	created, err := h.service.CreateBookWithFile(book, file, fileHeader.Size, fileHeader.Filename, contentType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create book"})
		return
	}

	c.JSON(http.StatusCreated, toBookRead(created))
}

// UpdateBook godoc
// @Summary      Update a book
// @Tags         books
// @Accept       json
// @Produce      json
// @Param        id       path  string             true  "Book ID"
// @Param        payload  body  schemas.BookUpdate  true  "Fields to update"
// @Success      200  {object}  schemas.BookRead
// @Failure      400  {object}  schemas.ErrorResponse
// @Failure      404  {object}  schemas.ErrorResponse
// @Failure      500  {object}  schemas.ErrorResponse
// @Security     BearerAuth
// @Router       /books/{id} [put]
func (h *BookHandler) UpdateBook(c *gin.Context) {
	id := c.Param("id")

	var payload schemas.BookUpdate
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existing, err := h.service.GetBook(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	}

	if payload.Name != nil {
		existing.Name = *payload.Name
	}
	if payload.FileKey != nil {
		existing.FileKey = *payload.FileKey
	}

	updated, err := h.service.UpdateBook(id, existing)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update book"})
		return
	}

	c.JSON(http.StatusOK, toBookRead(updated))
}

// DeleteBook godoc
// @Summary      Delete a book
// @Tags         books
// @Param        id  path  string  true  "Book ID"
// @Success      204  "No Content"
// @Failure      500  {object}  schemas.ErrorResponse
// @Security     BearerAuth
// @Router       /books/{id} [delete]
func (h *BookHandler) DeleteBook(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeleteBook(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete book"})
		return
	}

	c.Status(http.StatusNoContent)
}

func parseBookFilter(c *gin.Context) (models.BookFilter, error) {
	filter := models.BookFilter{
		SerieID: c.Query("serieId"),
		// UniverseID is filtered indirectly: a book's universe is whatever
		// universe its serie belongs to, resolved via a join at query time.
		UniverseID: c.Query("universeId"),
	}

	if raw := c.Query("bookPosition"); raw != "" {
		position, err := strconv.Atoi(raw)
		if err != nil {
			return models.BookFilter{}, err
		}
		filter.BookPosition = &position
	}

	if raw := c.Query("maxBookPosition"); raw != "" {
		position, err := strconv.Atoi(raw)
		if err != nil {
			return models.BookFilter{}, err
		}
		filter.MaxBookPosition = &position
	}

	return filter, nil
}

func toBookRead(book models.Book) schemas.BookRead {
	return schemas.BookRead{
		ID:            book.ID,
		Name:          book.Name,
		SerieID:       book.SerieID,
		SeriePosition: book.SeriePosition,
		FileKey:       book.FileKey,
	}
}

func toBookReadList(books []models.Book) []schemas.BookRead {
	result := make([]schemas.BookRead, len(books))
	for i, book := range books {
		result[i] = toBookRead(book)
	}
	return result
}
