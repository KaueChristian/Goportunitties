package repository

import (
	"github.com/KaueChristian/Goportunitties/schemas"
	"gorm.io/gorm"
)

// OpeningRepository defines data access operations for Opening.
type OpeningRepository interface {
	Create(opening *schemas.Opening) error
	FindByID(id string) (*schemas.Opening, error)
	FindAll() ([]schemas.Opening, error)
	Update(opening *schemas.Opening) error
	Delete(opening *schemas.Opening) error
}

type gormOpeningRepository struct {
	db *gorm.DB
}

// NewOpeningRepository creates a new GORM-backed OpeningRepository.
func NewOpeningRepository(db *gorm.DB) OpeningRepository {
	return &gormOpeningRepository{db: db}
}

func (r *gormOpeningRepository) Create(opening *schemas.Opening) error {
	return r.db.Create(opening).Error
}

func (r *gormOpeningRepository) FindByID(id string) (*schemas.Opening, error) {
	opening := &schemas.Opening{}
	if err := r.db.First(opening, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return opening, nil
}

func (r *gormOpeningRepository) FindAll() ([]schemas.Opening, error) {
	openings := []schemas.Opening{}
	if err := r.db.Find(&openings).Error; err != nil {
		return nil, err
	}
	return openings, nil
}

func (r *gormOpeningRepository) Update(opening *schemas.Opening) error {
	return r.db.Save(opening).Error
}

func (r *gormOpeningRepository) Delete(opening *schemas.Opening) error {
	return r.db.Delete(opening).Error
}
