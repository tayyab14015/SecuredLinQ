package repository

import (
	"github.com/securedlinq/backend/internal/models"
	"gorm.io/gorm"
)

// GalleryRepository handles gallery database operations
type GalleryRepository struct {
	db *gorm.DB
}

// NewGalleryRepository creates a new gallery repository
func NewGalleryRepository(db *gorm.DB) *GalleryRepository {
	return &GalleryRepository{db: db}
}

// Create creates a new gallery entry
func (r *GalleryRepository) Create(gallery *models.Gallery) error {
	return r.db.Create(gallery).Error
}

// GetByLoadID gets all gallery entries for a specific load
func (r *GalleryRepository) GetByLoadID(loadID uint) ([]models.Gallery, error) {
	var galleries []models.Gallery
	err := r.db.Where("load_id = ?", loadID).
		Order("created_at DESC").
		Find(&galleries).Error
	return galleries, err
}

// GetByLoadIDs gets all gallery entries for multiple loads
func (r *GalleryRepository) GetByLoadIDs(loadIDs []uint) ([]models.Gallery, error) {
	var galleries []models.Gallery
	err := r.db.Where("load_id IN ?", loadIDs).
		Order("created_at DESC").
		Find(&galleries).Error
	return galleries, err
}

// Delete deletes a gallery entry by ID
func (r *GalleryRepository) Delete(id uint) error {
	return r.db.Delete(&models.Gallery{}, id).Error
}

// DeleteByLoadID deletes all gallery entries for a specific load
func (r *GalleryRepository) DeleteByLoadID(loadID uint) error {
	return r.db.Where("load_id = ?", loadID).Delete(&models.Gallery{}).Error
}

