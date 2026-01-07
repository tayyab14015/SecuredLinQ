package service

import (
	"errors"

	"github.com/securedlinq/backend/internal/models"
	"github.com/securedlinq/backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// DriverService handles driver business logic
type DriverService struct {
	driverRepo *repository.DriverRepository
}

// NewDriverService creates a new driver service
func NewDriverService(driverRepo *repository.DriverRepository) *DriverService {
	return &DriverService{
		driverRepo: driverRepo,
	}
}

// RegisterDriverRequest represents driver registration data
type RegisterDriverRequest struct {
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required,min=6"`
	PhoneNumber string `json:"phone_number" binding:"required"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Email       string `json:"email"`
}

// RegisterDriver registers a new driver
func (s *DriverService) RegisterDriver(req *RegisterDriverRequest) (*models.Driver, error) {
	// Check if username already exists
	if s.driverRepo.UsernameExists(req.Username) {
		return nil, errors.New("username already exists")
	}

	// Check if phone number already exists
	if s.driverRepo.PhoneNumberExists(req.PhoneNumber) {
		return nil, errors.New("phone number already registered")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	driver := &models.Driver{
		Username:     req.Username,
		PasswordHash: string(hashedPassword),
		PhoneNumber:  req.PhoneNumber,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		IsActive:     true,
	}

	if req.Email != "" {
		driver.Email.String = req.Email
		driver.Email.Valid = true
	}

	if err := s.driverRepo.Create(driver); err != nil {
		return nil, err
	}

	return driver, nil
}

// ValidateCredentials validates driver login credentials
func (s *DriverService) ValidateCredentials(username, password string) (*models.Driver, error) {
	driver, err := s.driverRepo.GetByUsername(username)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if !driver.IsActive {
		return nil, errors.New("account is deactivated")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(driver.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	return driver, nil
}

// GetDriverByID gets a driver by ID
func (s *DriverService) GetDriverByID(id uint) (*models.Driver, error) {
	return s.driverRepo.GetByID(id)
}

// GetAllDrivers gets all drivers with pagination
func (s *DriverService) GetAllDrivers(page, pageSize int) ([]models.Driver, int64, error) {
	return s.driverRepo.GetAll(page, pageSize)
}

// UpdateDriver updates a driver's information
func (s *DriverService) UpdateDriver(driver *models.Driver) error {
	return s.driverRepo.Update(driver)
}

// DeactivateDriver deactivates a driver account
func (s *DriverService) DeactivateDriver(id uint) error {
	driver, err := s.driverRepo.GetByID(id)
	if err != nil {
		return err
	}
	driver.IsActive = false
	return s.driverRepo.Update(driver)
}

// ActivateDriver activates a driver account
func (s *DriverService) ActivateDriver(id uint) error {
	driver, err := s.driverRepo.GetByID(id)
	if err != nil {
		return err
	}
	driver.IsActive = true
	return s.driverRepo.Update(driver)
}

