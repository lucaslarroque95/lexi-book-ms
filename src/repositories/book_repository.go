package repositories

import "lexi/books/models"

type BookRepository interface {
	Create(book models.Book) (models.Book, error)
	Get(id string) (models.Book, error)
	GetAll(filter models.BookFilter) ([]models.Book, error)
	GetByName(name string) (models.Book, error)
	Update(id string, role models.Book) (models.Book, error)
	Delete(id string) error
	// ClearSerie unlinks every book pointing at serieID: their serieId and
	// seriePosition go back to unset. Called when that serie is deleted.
	ClearSerie(serieID string) error
}
