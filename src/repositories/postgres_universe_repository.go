package repositories

import (
	"lexi/books/db"
	"lexi/books/models"

	"gorm.io/gorm"
)

type PostgresUniverseRepository struct {
	gormDB *gorm.DB
}

func NewPostgresUniverseRepository(gormDB *gorm.DB) *PostgresUniverseRepository {
	return &PostgresUniverseRepository{gormDB: gormDB}
}

func (pur *PostgresUniverseRepository) Create(universe models.Universe) (models.Universe, error) {
	ormUniverse := toUniverseORM(universe)
	if err := pur.gormDB.Create(&ormUniverse).Error; err != nil {
		return models.Universe{}, err
	}
	return toUniverseDomain(ormUniverse), nil
}

func (pur *PostgresUniverseRepository) Get(id string) (models.Universe, error) {
	var ormUniverse db.Universe
	if err := pur.gormDB.Where("id = ?", id).First(&ormUniverse).Error; err != nil {
		return models.Universe{}, err
	}
	return toUniverseDomain(ormUniverse), nil
}

func (pur *PostgresUniverseRepository) GetAll() ([]models.Universe, error) {
	var ormUniverses []db.Universe
	if err := pur.gormDB.Find(&ormUniverses).Error; err != nil {
		return nil, err
	}

	universes := make([]models.Universe, len(ormUniverses))
	for i, ormUniverse := range ormUniverses {
		universes[i] = toUniverseDomain(ormUniverse)
	}
	return universes, nil
}

func (pur *PostgresUniverseRepository) Update(id string, universe models.Universe) (models.Universe, error) {
	ormUniverse := toUniverseORM(universe)
	ormUniverse.ID = id

	if err := pur.gormDB.Model(&db.Universe{}).Where("id = ?", id).Updates(&ormUniverse).Error; err != nil {
		return models.Universe{}, err
	}
	return pur.Get(id)
}

func (pur *PostgresUniverseRepository) Delete(id string) error {
	return pur.gormDB.Where("id = ?", id).Delete(&db.Universe{}).Error
}

func toUniverseDomain(ormUniverse db.Universe) models.Universe {
	return models.Universe{
		ID:     ormUniverse.ID,
		UserID: ormUniverse.UserID,
		Name:   ormUniverse.Name,
	}
}

func toUniverseORM(universe models.Universe) db.Universe {
	return db.Universe{
		ID:     universe.ID,
		UserID: universe.UserID,
		Name:   universe.Name,
	}
}
