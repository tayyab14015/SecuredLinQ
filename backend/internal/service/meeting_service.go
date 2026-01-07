package service

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/securedlinq/backend/internal/config"
	"github.com/securedlinq/backend/internal/models"
	"github.com/securedlinq/backend/internal/repository"
)

// MeetingService handles meeting business logic
type MeetingService struct {
	meetingRepo *repository.MeetingRepository
	loadRepo    *repository.LoadRepository
	config      *config.Config
}

// NewMeetingService creates a new meeting service
func NewMeetingService(
	meetingRepo *repository.MeetingRepository,
	loadRepo *repository.LoadRepository,
	cfg *config.Config,
) *MeetingService {
	return &MeetingService{
		meetingRepo: meetingRepo,
		loadRepo:    loadRepo,
		config:      cfg,
	}
}

// MeetingRoomInfo represents meeting room information for API responses
type MeetingRoomInfo struct {
	ID           uint   `json:"id"`
	LoadID       uint   `json:"load_id"`
	RoomID       string `json:"roomId"`
	ChannelName  string `json:"channelName"`
	MeetingLink  string `json:"meetingLink"`
	LoadNumber   string `json:"load_number,omitempty"`
	SaveType     string `json:"save_type,omitempty"`
	Status       string `json:"status"`
	CreatedAt    string `json:"created_at"`
	LastJoinedAt string `json:"lastJoinedAt,omitempty"`
}

// GetOrCreateMeetingRoom gets an existing meeting room or creates a new one based on load_id
func (s *MeetingService) GetOrCreateMeetingRoom(loadID uint) (*MeetingRoomInfo, error) {
	// Get the Load to verify it exists and get load number
	load, err := s.loadRepo.GetByID(loadID)
	if err != nil {
		return nil, fmt.Errorf("load not found: %w", err)
	}

	// Check if meeting room already exists for this load
	existingRoom, err := s.meetingRepo.GetByLoadID(loadID)
	if err == nil && existingRoom != nil && existingRoom.Status == "active" {
		return s.roomToInfo(existingRoom), nil
	}

	// Generate room_id: load_id + random identifier
	randomID := generateRoomID() // This generates a shortened UUID
	roomID := fmt.Sprintf("load_%d_%s", loadID, randomID)

	// Generate channel name as full UUID
	channelName := uuid.New().String() // Full UUID for channel name

	// Generate meeting link
	meetingLink := fmt.Sprintf("%s/join/%s", s.config.Server.BaseURL, roomID)

	// Use the actual load number
	loadNumber := load.LoadNumber

	// Create new meeting room
	room, err := s.meetingRepo.CreateByLoadID(loadID, roomID, channelName, meetingLink, loadNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to create meeting room: %w", err)
	}

	return s.roomToInfo(room), nil
}

// GetMeetingRoomByRoomID gets a meeting room by room ID
func (s *MeetingService) GetMeetingRoomByRoomID(roomID string) (*MeetingRoomInfo, error) {
	room, err := s.meetingRepo.GetByRoomID(roomID)
	if err != nil {
		return nil, fmt.Errorf("meeting room not found or expired: %w", err)
	}
	return s.roomToInfo(room), nil
}

// UpdateLastJoined updates the last joined timestamp
func (s *MeetingService) UpdateLastJoined(roomID string) error {
	return s.meetingRepo.UpdateLastJoined(roomID)
}

// EndMeeting ends a meeting room
func (s *MeetingService) EndMeeting(roomID string) error {
	return s.meetingRepo.EndMeeting(roomID)
}

// Helper functions

func (s *MeetingService) roomToInfo(room *models.MeetingRoom) *MeetingRoomInfo {
	info := &MeetingRoomInfo{
		ID:          room.ID,
		LoadID:      room.LoadID,
		RoomID:      room.RoomID,
		ChannelName: room.ChannelName,
		MeetingLink: room.MeetingLink,
		Status:      room.Status,
		CreatedAt:   room.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if room.LoadNumber.Valid {
		info.LoadNumber = room.LoadNumber.String
	}
	if room.SaveType.Valid {
		info.SaveType = room.SaveType.String
	}
	if room.LastJoinedAt.Valid {
		info.LastJoinedAt = room.LastJoinedAt.Time.Format("2006-01-02T15:04:05Z07:00")
	}
	return info
}

func generateRoomID() string {
	id := uuid.New().String()
	return strings.ReplaceAll(id, "-", "")[:12]
}
