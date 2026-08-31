package service

import (
	"lexi/books/models"
	"lexi/books/repositories"
)

// BookUnlinker clears a book's link to a serie that's being deleted. A
// narrow view of repositories.BookRepository so SerieService only depends
// on what it actually needs.
type BookUnlinker interface {
	ClearSerie(serieID string) error
}

type SerieService struct {
	repository repositories.SerieRepository
	books      BookUnlinker
}

func NewSerieService(repository repositories.SerieRepository, books BookUnlinker) *SerieService {
	return &SerieService{repository: repository, books: books}
}

func (ss *SerieService) CreateSerie(serie models.Serie) (models.Serie, error) {
	return ss.repository.Create(serie)
}

func (ss *SerieService) GetSerie(serieID string) (models.Serie, error) {
	return ss.repository.Get(serieID)
}

func (ss *SerieService) ListSeries(filter models.SerieFilter) ([]models.Serie, error) {
	return ss.repository.GetAll(filter)
}

func (ss *SerieService) UpdateSerie(serieID string, serie models.Serie) (models.Serie, error) {
	return ss.repository.Update(serieID, serie)
}

// DeleteSerie unlinks every book that pointed to this serie (their serieId
// and seriePosition go back to unset) before deleting it, so no book is
// ever left referencing a serie that no longer exists.
func (ss *SerieService) DeleteSerie(serieID string) error {
	if err := ss.books.ClearSerie(serieID); err != nil {
		return err
	}
	return ss.repository.Delete(serieID)
}
