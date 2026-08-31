package service

import (
	"lexi/books/models"
	"lexi/books/repositories"
)

// SerieUnlinker clears a serie's link to a universe that's being deleted. A
// narrow view of repositories.SerieRepository so UniverseService only
// depends on what it actually needs.
type SerieUnlinker interface {
	ClearUniverse(universeID string) error
}

type UniverseService struct {
	repository repositories.UniverseRepository
	series     SerieUnlinker
}

func NewUniverseService(repository repositories.UniverseRepository, series SerieUnlinker) *UniverseService {
	return &UniverseService{repository: repository, series: series}
}

func (us *UniverseService) CreateUniverse(universe models.Universe) (models.Universe, error) {
	return us.repository.Create(universe)
}

func (us *UniverseService) GetUniverse(universeID string) (models.Universe, error) {
	return us.repository.Get(universeID)
}

func (us *UniverseService) ListUniverses() ([]models.Universe, error) {
	return us.repository.GetAll()
}

func (us *UniverseService) UpdateUniverse(universeID string, universe models.Universe) (models.Universe, error) {
	return us.repository.Update(universeID, universe)
}

// DeleteUniverse unlinks every serie that pointed to this universe (their
// universeId and universePosition go back to unset) before deleting it, so
// no serie is ever left referencing a universe that no longer exists.
func (us *UniverseService) DeleteUniverse(universeID string) error {
	if err := us.series.ClearUniverse(universeID); err != nil {
		return err
	}
	return us.repository.Delete(universeID)
}
