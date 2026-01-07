package repository

import (
	"database/sql"
	"time"

	"github.com/securedlinq/backend/internal/models"
	"gorm.io/gorm"
)

// MeetingRepository handles meeting room database operations
type MeetingRepository struct {
	*Repository
}

// NewMeetingRepository creates a new meeting repository
func NewMeetingRepository(db *gorm.DB) *MeetingRepository {
	return &MeetingRepository{
		Repository: NewRepository(db),
	}
}

// GetByID retrieves a meeting room by ID
func (r *MeetingRepository) GetByID(id uint) (*models.MeetingRoom, error) {
	var meeting models.MeetingRoom
	err := r.db.First(&meeting, id).Error
	if err != nil {
		return nil, err
	}
	return &meeting, nil
}

// GetByRoomID retrieves an active meeting room by room ID
func (r *MeetingRepository) GetByRoomID(roomID string) (*models.MeetingRoom, error) {
	var meeting models.MeetingRoom
	err := r.db.Where("roomId = ? AND status = ?", roomID, "active").First(&meeting).Error
	if err != nil {
		return nil, err
	}
	return &meeting, nil
}

// GetByLoadID retrieves an active meeting room by load ID
func (r *MeetingRepository) GetByLoadID(loadID uint) (*models.MeetingRoom, error) {
	var meeting models.MeetingRoom
	err := r.db.Where("load_id = ? AND status = ?", loadID, "active").
		Order("created_at DESC").First(&meeting).Error
	if err != nil {
		return nil, err
	}
	return &meeting, nil
}

// GetByGuestID retrieves an active meeting room by guest ID (legacy method)
func (r *MeetingRepository) GetByGuestID(guestID int) (*models.MeetingRoom, error) {
	var meeting models.MeetingRoom
	err := r.db.Where("guest_id = ? AND status = ?", guestID, "active").
		Order("created_at DESC").First(&meeting).Error
	if err != nil {
		return nil, err
	}
	return &meeting, nil
}

// GetByChannelName retrieves an active meeting room by channel name
func (r *MeetingRepository) GetByChannelName(channelName string) (*models.MeetingRoom, error) {
	var meeting models.MeetingRoom
	err := r.db.Where("channelName LIKE ? AND status = ?", channelName+"%", "active").
		Order("created_at DESC").First(&meeting).Error
	if err != nil {
		return nil, err
	}
	return &meeting, nil
}

// Create creates a new meeting room
func (r *MeetingRepository) Create(meeting *models.MeetingRoom) error {
	meeting.CreatedAt = time.Now()
	meeting.Status = "active"
	return r.db.Create(meeting).Error
}

// UpdateOrCreate updates an existing meeting room or creates a new one
func (r *MeetingRepository) UpdateOrCreate(guestID int, roomID, channelName, meetingLink string, loadNumber, saveType *string) (*models.MeetingRoom, error) {
	var existingRoom models.MeetingRoom
	err := r.db.Where("guest_id = ?", guestID).Order("created_at DESC").First(&existingRoom).Error

	if err == gorm.ErrRecordNotFound {
		// Create new meeting room (legacy method - using guestID)
		newRoom := &models.MeetingRoom{
			RoomID:      roomID,
			ChannelName: channelName,
			MeetingLink: meetingLink,
			Status:      "active",
			CreatedAt:   time.Now(),
		}
		if loadNumber != nil {
			newRoom.LoadNumber = sql.NullString{String: *loadNumber, Valid: true}
		}
		if saveType != nil {
			newRoom.SaveType = sql.NullString{String: *saveType, Valid: true}
		}
		if err := r.db.Create(newRoom).Error; err != nil {
			return nil, err
		}
		return newRoom, nil
	} else if err != nil {
		return nil, err
	}

	// Update existing meeting room
	updates := map[string]interface{}{
		"roomId":       roomID,
		"channelName":  channelName,
		"meetingLink":  meetingLink,
		"status":       "active",
		"created_at":   time.Now(),
		"lastJoinedAt": nil,
	}
	if loadNumber != nil {
		updates["load_number"] = *loadNumber
	}
	if saveType != nil {
		updates["save_type"] = *saveType
	}

	if err := r.db.Model(&existingRoom).Updates(updates).Error; err != nil {
		return nil, err
	}

	// Reload the updated record
	r.db.First(&existingRoom, existingRoom.ID)
	return &existingRoom, nil
}

// UpdateLastJoined updates the last joined timestamp
func (r *MeetingRepository) UpdateLastJoined(roomID string) error {
	return r.db.Model(&models.MeetingRoom{}).
		Where("roomId = ?", roomID).
		Update("lastJoinedAt", time.Now()).Error
}

// EndMeeting ends a meeting room by room ID
func (r *MeetingRepository) EndMeeting(roomID string) error {
	return r.db.Model(&models.MeetingRoom{}).
		Where("roomId = ?", roomID).
		Update("status", "ended").Error
}

// CreateByLoadID creates a new meeting room for a load
func (r *MeetingRepository) CreateByLoadID(loadID uint, roomID, channelName, meetingLink, loadNumber string) (*models.MeetingRoom, error) {
	newRoom := &models.MeetingRoom{
		LoadID:      loadID,
		RoomID:      roomID,
		ChannelName: channelName,
		MeetingLink: meetingLink,
		Status:      "active",
		CreatedAt:   time.Now(),
	}
	if loadNumber != "" {
		newRoom.LoadNumber = sql.NullString{String: loadNumber, Valid: true}
	}
	if err := r.db.Create(newRoom).Error; err != nil {
		return nil, err
	}
	return newRoom, nil
}

// InvalidateAllActiveMeetings invalidates all active meetings for a guest (legacy method)
func (r *MeetingRepository) InvalidateAllActiveMeetings(guestID int) (int64, error) {
	result := r.db.Model(&models.MeetingRoom{}).
		Where("guest_id = ? AND status = ?", guestID, "active").
		Update("status", "ended")
	return result.RowsAffected, result.Error
}
