package repositories

import "lexi/books/models"

type UniverseRepository interface {
	Create(universe models.Universe) (models.Universe, error)
	Get(id string) (models.Universe, error)
	GetAll() ([]models.Universe, error)
	Update(id string, universe models.Universe) (models.Universe, error)
	Delete(id string) error
}
