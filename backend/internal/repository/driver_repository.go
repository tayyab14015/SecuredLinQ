package repository

import (
	"github.com/securedlinq/backend/internal/models"
	"gorm.io/gorm"
)

// DriverRepository handles driver database operations
type DriverRepository struct {
	db *gorm.DB
}

// NewDriverRepository creates a new driver repository
func NewDriverRepository(db *gorm.DB) *DriverRepository {
	return &DriverRepository{db: db}
}

// Create creates a new driver
func (r *DriverRepository) Create(driver *models.Driver) error {
	return r.db.Create(driver).Error
}

// GetByID gets a driver by ID
func (r *DriverRepository) GetByID(id uint) (*models.Driver, error) {
	var driver models.Driver
	err := r.db.Where("id = ?", id).First(&driver).Error
	if err != nil {
		return nil, err
	}
	return &driver, nil
}

// GetByUsername gets a driver by username
func (r *DriverRepository) GetByUsername(username string) (*models.Driver, error) {
	var driver models.Driver
	err := r.db.Where("username = ?", username).First(&driver).Error
	if err != nil {
		return nil, err
	}
	return &driver, nil
}

// GetByPhoneNumber gets a driver by phone number
func (r *DriverRepository) GetByPhoneNumber(phoneNumber string) (*models.Driver, error) {
	var driver models.Driver
	err := r.db.Where("phone_number = ?", phoneNumber).First(&driver).Error
	if err != nil {
		return nil, err
	}
	return &driver, nil
}

// GetAll gets all drivers with pagination
func (r *DriverRepository) GetAll(page, pageSize int) ([]models.Driver, int64, error) {
	var drivers []models.Driver
	var total int64

	offset := (page - 1) * pageSize

	err := r.db.Model(&models.Driver{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&drivers).Error
	if err != nil {
		return nil, 0, err
	}

	return drivers, total, nil
}

// Update updates a driver
func (r *DriverRepository) Update(driver *models.Driver) error {
	return r.db.Save(driver).Error
}

// Delete deletes a driver
func (r *DriverRepository) Delete(id uint) error {
	return r.db.Delete(&models.Driver{}, id).Error
}

// UsernameExists checks if a username already exists
func (r *DriverRepository) UsernameExists(username string) bool {
	var count int64
	r.db.Model(&models.Driver{}).Where("username = ?", username).Count(&count)
	return count > 0
}

// PhoneNumberExists checks if a phone number already exists
func (r *DriverRepository) PhoneNumberExists(phoneNumber string) bool {
	var count int64
	r.db.Model(&models.Driver{}).Where("phone_number = ?", phoneNumber).Count(&count)
	return count > 0
}
