package repositories

import "lexi/books/models"

type SerieRepository interface {
	Create(serie models.Serie) (models.Serie, error)
	Get(id string) (models.Serie, error)
	GetAll(filter models.SerieFilter) ([]models.Serie, error)
	Update(id string, serie models.Serie) (models.Serie, error)
	Delete(id string) error
	// ClearUniverse unlinks every serie pointing at universeID: their
	// universeId and universePosition go back to unset. Called when that
	// universe is deleted.
	ClearUniverse(universeID string) error
}
