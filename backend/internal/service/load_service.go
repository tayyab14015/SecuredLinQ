package service

import (
	"database/sql"
	"errors"
	"time"

	"github.com/securedlinq/backend/internal/models"
	"github.com/securedlinq/backend/internal/repository"
)

// LoadService handles load business logic
type LoadService struct {
	loadRepo   *repository.LoadRepository
	driverRepo *repository.DriverRepository
}

// NewLoadService creates a new load service
func NewLoadService(loadRepo *repository.LoadRepository, driverRepo *repository.DriverRepository) *LoadService {
	return &LoadService{
		loadRepo:   loadRepo,
		driverRepo: driverRepo,
	}
}

// CreateLoadRequest represents load creation data
type CreateLoadRequest struct {
	LoadNumber      string     `json:"load_number" binding:"required"`
	Description     string     `json:"description"`
	PickupAddress   string     `json:"pickup_address"`
	DeliveryAddress string     `json:"delivery_address"`
	ScheduledDate   *time.Time `json:"scheduled_date"`
	CreatedByID     int        `json:"created_by_id"`
}

// CreateLoad creates a new load
func (s *LoadService) CreateLoad(req *CreateLoadRequest) (*models.Load, error) {
	// Check if load number already exists
	if s.loadRepo.LoadNumberExists(req.LoadNumber) {
		return nil, errors.New("load number already exists")
	}

	load := &models.Load{
		LoadNumber:  req.LoadNumber,
		Status:      "Unassigned",
		CreatedByID: req.CreatedByID,
	}

	if req.Description != "" {
		load.Description = sql.NullString{String: req.Description, Valid: true}
	}

	if req.PickupAddress != "" {
		load.PickupAddress = sql.NullString{String: req.PickupAddress, Valid: true}
	}

	if req.DeliveryAddress != "" {
		load.DeliveryAddress = sql.NullString{String: req.DeliveryAddress, Valid: true}
	}

	if req.ScheduledDate != nil {
		load.ScheduledDate = sql.NullTime{Time: *req.ScheduledDate, Valid: true}
	}

	if err := s.loadRepo.Create(load); err != nil {
		return nil, err
	}

	return load, nil
}

// GetLoadByID gets a load by ID
func (s *LoadService) GetLoadByID(id uint) (*models.Load, error) {
	return s.loadRepo.GetByID(id)
}

// GetLoadByNumber gets a load by load number
func (s *LoadService) GetLoadByNumber(loadNumber string) (*models.Load, error) {
	return s.loadRepo.GetByLoadNumber(loadNumber)
}

// GetAllLoads gets all loads with pagination
func (s *LoadService) GetAllLoads(page, pageSize int) ([]models.Load, int64, error) {
	return s.loadRepo.GetAll(page, pageSize)
}

// GetLoadsByDriverID gets loads assigned to a driver
func (s *LoadService) GetLoadsByDriverID(driverID uint, page, pageSize int) ([]models.Load, int64, error) {
	return s.loadRepo.GetByDriverID(driverID, page, pageSize)
}

// GetLoadsByStatus gets loads by status
func (s *LoadService) GetLoadsByStatus(status string, page, pageSize int) ([]models.Load, int64, error) {
	return s.loadRepo.GetByStatus(status, page, pageSize)
}

// AssignDriverToLoad assigns a driver to a load
func (s *LoadService) AssignDriverToLoad(loadID uint, driverID uint) error {
	// Verify driver exists
	_, err := s.driverRepo.GetByID(driverID)
	if err != nil {
		return errors.New("driver not found")
	}

	// Verify load exists
	load, err := s.loadRepo.GetByID(loadID)
	if err != nil {
		return errors.New("load not found")
	}

	// Check load is not already completed
	if load.Status == "Completed" {
		return errors.New("cannot assign driver to completed load")
	}

	return s.loadRepo.AssignDriver(loadID, driverID)
}

// UpdateLoadStatus updates a load's status
func (s *LoadService) UpdateLoadStatus(loadID uint, status string, driverID uint) error {
	load, err := s.loadRepo.GetByID(loadID)
	if err != nil {
		return errors.New("load not found")
	}

	// Verify driver owns this load
	if !load.DriverID.Valid || uint(load.DriverID.Int64) != driverID {
		return errors.New("you are not authorized to update this load")
	}

	// Validate status transitions
	validStatuses := map[string][]string{
		"Assigned":  {"Completed"},
		"Completed": {}, // Cannot change from completed
	}

	allowedNextStatuses, exists := validStatuses[load.Status]
	if !exists {
		return errors.New("invalid current status")
	}

	statusAllowed := false
	for _, allowed := range allowedNextStatuses {
		if allowed == status {
			statusAllowed = true
			break
		}
	}

	if !statusAllowed {
		return errors.New("invalid status transition")
	}

	if status == "Completed" {
		return s.loadRepo.MarkCompleted(loadID)
	}

	return s.loadRepo.UpdateStatus(loadID, status)
}

// MarkLoadCompleted marks a load as completed by the assigned driver
func (s *LoadService) MarkLoadCompleted(loadID uint, driverID uint) error {
	return s.UpdateLoadStatus(loadID, "Completed", driverID)
}

// StartMeeting marks that a meeting has been started for this load (admin only)
func (s *LoadService) StartMeeting(loadID uint) (*models.Load, error) {
	load, err := s.loadRepo.GetByID(loadID)
	if err != nil {
		return nil, errors.New("load not found")
	}

	if load.Status != "Completed" {
		return nil, errors.New("can only start meeting for completed loads")
	}

	if err := s.loadRepo.SetMeetingStarted(loadID, true); err != nil {
		return nil, err
	}

	// Refresh load data
	return s.loadRepo.GetByID(loadID)
}

// DeleteLoad deletes a load
func (s *LoadService) DeleteLoad(id uint) error {
	return s.loadRepo.Delete(id)
}

