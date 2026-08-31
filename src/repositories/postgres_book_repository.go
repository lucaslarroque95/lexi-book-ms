package repositories

import (
	"lexi/books/db"
	"lexi/books/models"

	"gorm.io/gorm"
)

type PostgresBookRepository struct {
	gormDB *gorm.DB
}

func NewPostgresBookRepository(gormDB *gorm.DB) *PostgresBookRepository {
	return &PostgresBookRepository{gormDB: gormDB}
}

func (prr *PostgresBookRepository) Create(Book models.Book) (models.Book, error) {
	ormBook := toBookORM(Book)
	if err := prr.gormDB.Create(&ormBook).Error; err != nil {
		return models.Book{}, err
	}
	return toBookDomain(ormBook), nil
}

func (prr *PostgresBookRepository) Get(id string) (models.Book, error) {
	var ormBook db.Book
	if err := prr.gormDB.Where("id = ?", id).First(&ormBook).Error; err != nil {
		return models.Book{}, err
	}
	return toBookDomain(ormBook), nil
}

func (prr *PostgresBookRepository) GetByName(name string) (models.Book, error) {
	var ormBook db.Book
	err := prr.gormDB.Where("name = ?", name).First(&ormBook).Error

	if err != nil {
		return models.Book{}, err
	}
	return toBookDomain(ormBook), nil
}

func (prr *PostgresBookRepository) GetAll(filter models.BookFilter) ([]models.Book, error) {
	query := prr.gormDB
	if filter.SerieID != "" {
		query = query.Where("serie_id = ?", filter.SerieID)
	}
	if filter.UniverseID != "" {
		// A book has no universe of its own: it inherits it from its serie.
		query = query.Select("books.*").
			Joins("JOIN series ON series.id = books.serie_id").
			Where("series.universe_id = ?", filter.UniverseID)
	}
	if filter.BookPosition != nil {
		query = query.Where("serie_position = ?", *filter.BookPosition)
	}
	if filter.MaxBookPosition != nil {
		query = query.Where("serie_position <= ?", *filter.MaxBookPosition)
	}

	var ormBooks []db.Book
	if err := query.Find(&ormBooks).Error; err != nil {
		return nil, err
	}

	books := make([]models.Book, len(ormBooks))
	for i, ormBook := range ormBooks {
		books[i] = toBookDomain(ormBook)
	}
	return books, nil
}

func (prr *PostgresBookRepository) Update(id string, Book models.Book) (models.Book, error) {
	ormBook := toBookORM(Book)
	ormBook.ID = id

	if err := prr.gormDB.Model(&db.Book{}).Where("id = ?", id).Updates(&ormBook).Error; err != nil {
		return models.Book{}, err
	}
	return prr.Get(id)
}

func (prr *PostgresBookRepository) Delete(id string) error {
	return prr.gormDB.Where("id = ?", id).Delete(&db.Book{}).Error
}

func (prr *PostgresBookRepository) ClearSerie(serieID string) error {
	return prr.gormDB.Model(&db.Book{}).
		Where("serie_id = ?", serieID).
		Updates(map[string]any{"serie_id": nil, "serie_position": nil}).Error
}

func toBookDomain(ormBook db.Book) models.Book {
	return models.Book{
		ID:            ormBook.ID,
		UserID:        ormBook.UserID,
		Name:          ormBook.Name,
		SerieID:       derefOrEmpty(ormBook.SerieID),
		SeriePosition: derefOrZero(ormBook.SeriePosition),
		FileKey:       ormBook.FileKey,
	}
}

func toBookORM(book models.Book) db.Book {
	return db.Book{
		ID:            book.ID,
		UserID:        book.UserID,
		Name:          book.Name,
		SerieID:       nilIfEmpty(book.SerieID),
		SeriePosition: nilIfZero(book.SeriePosition),
		FileKey:       book.FileKey,
	}
}

// nilIfEmpty and derefOrEmpty bridge optional uuid foreign keys: the domain
// layer represents "no relation" as "", but a uuid column rejects "" and
// needs SQL NULL instead.
func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func derefOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// nilIfZero and derefOrZero do the same for a position within a parent that
// may not be set (or may have just been cleared by the parent's deletion).
func nilIfZero(i int) *int {
	if i == 0 {
		return nil
	}
	return &i
}

func derefOrZero(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}
