package repository

import (
	"github.com/securedlinq/backend/internal/models"
	"gorm.io/gorm"
)

// LoadRepository handles load database operations
type LoadRepository struct {
	db *gorm.DB
}

// NewLoadRepository creates a new load repository
func NewLoadRepository(db *gorm.DB) *LoadRepository {
	return &LoadRepository{db: db}
}

// Create creates a new load
func (r *LoadRepository) Create(load *models.Load) error {
	return r.db.Create(load).Error
}

// GetByID gets a load by ID
func (r *LoadRepository) GetByID(id uint) (*models.Load, error) {
	var load models.Load
	err := r.db.Preload("Driver").Where("id = ?", id).First(&load).Error
	if err != nil {
		return nil, err
	}
	return &load, nil
}

// GetByLoadNumber gets a load by load number
func (r *LoadRepository) GetByLoadNumber(loadNumber string) (*models.Load, error) {
	var load models.Load
	err := r.db.Preload("Driver").Where("load_number = ?", loadNumber).First(&load).Error
	if err != nil {
		return nil, err
	}
	return &load, nil
}

// GetAll gets all loads with pagination
func (r *LoadRepository) GetAll(page, pageSize int) ([]models.Load, int64, error) {
	var loads []models.Load
	var total int64

	offset := (page - 1) * pageSize

	err := r.db.Model(&models.Load{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Preload("Driver").Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&loads).Error
	if err != nil {
		return nil, 0, err
	}

	return loads, total, nil
}

// GetByDriverID gets all loads assigned to a specific driver
func (r *LoadRepository) GetByDriverID(driverID uint, page, pageSize int) ([]models.Load, int64, error) {
	var loads []models.Load
	var total int64

	offset := (page - 1) * pageSize

	err := r.db.Model(&models.Load{}).Where("driver_id = ?", driverID).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Preload("Driver").Where("driver_id = ?", driverID).Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&loads).Error
	if err != nil {
		return nil, 0, err
	}

	return loads, total, nil
}

// GetByStatus gets loads by status
func (r *LoadRepository) GetByStatus(status string, page, pageSize int) ([]models.Load, int64, error) {
	var loads []models.Load
	var total int64

	offset := (page - 1) * pageSize

	err := r.db.Model(&models.Load{}).Where("status = ?", status).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Preload("Driver").Where("status = ?", status).Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&loads).Error
	if err != nil {
		return nil, 0, err
	}

	return loads, total, nil
}

// Update updates a load
func (r *LoadRepository) Update(load *models.Load) error {
	return r.db.Save(load).Error
}

// UpdateStatus updates a load's status
func (r *LoadRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&models.Load{}).Where("id = ?", id).Update("status", status).Error
}

// AssignDriver assigns a driver to a load
func (r *LoadRepository) AssignDriver(loadID uint, driverID uint) error {
	return r.db.Model(&models.Load{}).Where("id = ?", loadID).Updates(map[string]interface{}{
		"driver_id": driverID,
		"status":    "Assigned",
	}).Error
}

// MarkCompleted marks a load as completed
func (r *LoadRepository) MarkCompleted(id uint) error {
	return r.db.Model(&models.Load{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":       "Completed",
		"completed_at": gorm.Expr("NOW()"),
	}).Error
}

// SetMeetingStarted marks that a meeting has been started for this load
func (r *LoadRepository) SetMeetingStarted(id uint, started bool) error {
	return r.db.Model(&models.Load{}).Where("id = ?", id).Update("meeting_started", started).Error
}

// Delete deletes a load
func (r *LoadRepository) Delete(id uint) error {
	return r.db.Delete(&models.Load{}, id).Error
}

// LoadNumberExists checks if a load number already exists
func (r *LoadRepository) LoadNumberExists(loadNumber string) bool {
	var count int64
	r.db.Model(&models.Load{}).Where("load_number = ?", loadNumber).Count(&count)
	return count > 0
}

