package repositories

import (
	"lexi/books/db"
	"lexi/books/models"

	"gorm.io/gorm"
)

type PostgresSerieRepository struct {
	gormDB *gorm.DB
}

func NewPostgresSerieRepository(gormDB *gorm.DB) *PostgresSerieRepository {
	return &PostgresSerieRepository{gormDB: gormDB}
}

func (psr *PostgresSerieRepository) Create(serie models.Serie) (models.Serie, error) {
	ormSerie := toSerieORM(serie)
	if err := psr.gormDB.Create(&ormSerie).Error; err != nil {
		return models.Serie{}, err
	}
	return toSerieDomain(ormSerie), nil
}

func (psr *PostgresSerieRepository) Get(id string) (models.Serie, error) {
	var ormSerie db.Serie
	if err := psr.gormDB.Where("id = ?", id).First(&ormSerie).Error; err != nil {
		return models.Serie{}, err
	}
	return toSerieDomain(ormSerie), nil
}

func (psr *PostgresSerieRepository) GetAll(filter models.SerieFilter) ([]models.Serie, error) {
	query := psr.gormDB
	if filter.UniverseID != "" {
		query = query.Where("universe_id = ?", filter.UniverseID)
	}
	if filter.UniversePosition != nil {
		query = query.Where("universe_position = ?", *filter.UniversePosition)
	}
	if filter.MaxUniversePosition != nil {
		query = query.Where("universe_position <= ?", *filter.MaxUniversePosition)
	}

	var ormSeries []db.Serie
	if err := query.Find(&ormSeries).Error; err != nil {
		return nil, err
	}

	series := make([]models.Serie, len(ormSeries))
	for i, ormSerie := range ormSeries {
		series[i] = toSerieDomain(ormSerie)
	}
	return series, nil
}

func (psr *PostgresSerieRepository) Update(id string, serie models.Serie) (models.Serie, error) {
	ormSerie := toSerieORM(serie)
	ormSerie.ID = id

	if err := psr.gormDB.Model(&db.Serie{}).Where("id = ?", id).Updates(&ormSerie).Error; err != nil {
		return models.Serie{}, err
	}
	return psr.Get(id)
}

func (psr *PostgresSerieRepository) Delete(id string) error {
	return psr.gormDB.Where("id = ?", id).Delete(&db.Serie{}).Error
}

func (psr *PostgresSerieRepository) ClearUniverse(universeID string) error {
	return psr.gormDB.Model(&db.Serie{}).
		Where("universe_id = ?", universeID).
		Updates(map[string]any{"universe_id": nil, "universe_position": nil}).Error
}

func toSerieDomain(ormSerie db.Serie) models.Serie {
	return models.Serie{
		ID:               ormSerie.ID,
		UserID:           ormSerie.UserID,
		Name:             ormSerie.Name,
		UniverseID:       derefOrEmpty(ormSerie.UniverseID),
		UniversePosition: derefOrZero(ormSerie.UniversePosition),
	}
}

func toSerieORM(serie models.Serie) db.Serie {
	return db.Serie{
		ID:               serie.ID,
		UserID:           serie.UserID,
		Name:             serie.Name,
		UniverseID:       nilIfEmpty(serie.UniverseID),
		UniversePosition: nilIfZero(serie.UniversePosition),
	}
}
