package service

import (
	"fmt"
	"sync"

	"github.com/securedlinq/backend/internal/models"
	"github.com/securedlinq/backend/internal/repository"
	"github.com/securedlinq/backend/pkg/agora"
)

// RecordingService handles recording business logic
type RecordingService struct {
	meetingRepo *repository.MeetingRepository
	galleryRepo *repository.GalleryRepository
	agoraClient *agora.Client
	// In-memory storage for active recordings (resourceId -> recordingInfo)
	activeRecordings map[string]*ActiveRecording
	mu               sync.RWMutex
}

// ActiveRecording stores info about an active recording session
type ActiveRecording struct {
	ResourceID  string
	SID         string
	ChannelName string
	UID         string
	LoadID      uint
	LoadNumber  string
}

// NewRecordingService creates a new recording service
func NewRecordingService(
	meetingRepo *repository.MeetingRepository,
	galleryRepo *repository.GalleryRepository,
	agoraClient *agora.Client,
) *RecordingService {
	return &RecordingService{
		meetingRepo:      meetingRepo,
		galleryRepo:      galleryRepo,
		agoraClient:      agoraClient,
		activeRecordings: make(map[string]*ActiveRecording),
	}
}

// StartRecordingRequest represents a request to start recording
type StartRecordingRequest struct {
	RoomID      string `json:"roomId"`
	ChannelName string `json:"channelName"`
	UID         string `json:"uid"`
	Token       string `json:"token"`
}

// StartRecordingResponse represents a response from starting recording
type StartRecordingResponse struct {
	Success     bool   `json:"success"`
	ResourceID  string `json:"resourceId"`
	SID         string `json:"sid"`
	RecordingID string `json:"recordingId"`
	CName       string `json:"cname"`
	UID         string `json:"uid"`
}

// StopRecordingRequest represents a request to stop recording
type StopRecordingRequest struct {
	ResourceID  string `json:"resourceId"`
	SID         string `json:"sid"`
	ChannelName string `json:"channelName"`
	UID         string `json:"uid"`
}

// StopRecordingResponse represents a response from stopping recording
type StopRecordingResponse struct {
	Success  bool     `json:"success"`
	FileName string   `json:"fileName,omitempty"`
	S3Key    string   `json:"s3Key,omitempty"`
	S3URL    string   `json:"s3Url,omitempty"`
	FileList []string `json:"fileList,omitempty"`
	FileSize int64    `json:"fileSize,omitempty"`
	Duration int      `json:"duration,omitempty"`
	Status   string   `json:"status,omitempty"`
	Warning  string   `json:"warning,omitempty"`
}

// StartRecording starts a new recording
func (s *RecordingService) StartRecording(req *StartRecordingRequest) (*StartRecordingResponse, error) {
	// Get meeting room to get load info
	meeting, err := s.meetingRepo.GetByRoomID(req.RoomID)
	if err != nil {
		return nil, fmt.Errorf("meeting room not found: %w", err)
	}

	// Get load number from meeting room
	loadNumber := ""
	if meeting.LoadNumber.Valid {
		loadNumber = meeting.LoadNumber.String
	}

	// Start recording via Agora
	result, err := s.agoraClient.StartRecording(req.ChannelName, req.UID, req.Token, loadNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to start Agora recording: %w", err)
	}

	// Store active recording info in memory
	s.mu.Lock()
	s.activeRecordings[result.SID] = &ActiveRecording{
		ResourceID:  result.ResourceID,
		SID:         result.SID,
		ChannelName: req.ChannelName,
		UID:         req.UID,
		LoadID:      meeting.LoadID,
		LoadNumber:  loadNumber,
	}
	s.mu.Unlock()

	return &StartRecordingResponse{
		Success:     true,
		ResourceID:  result.ResourceID,
		SID:         result.SID,
		RecordingID: result.SID,
		CName:       req.ChannelName,
		UID:         req.UID,
	}, nil
}

// StopRecording stops an active recording
func (s *RecordingService) StopRecording(req *StopRecordingRequest) (*StopRecordingResponse, error) {
	// Get recording info from memory using SID as key
	s.mu.RLock()
	recording, exists := s.activeRecordings[req.SID]
	s.mu.RUnlock()

	// Validate recording exists
	if !exists {
		return nil, fmt.Errorf("recording not found for sid: %s. Make sure the recording was started and the SID matches.", req.SID)
	}

	// Validate ResourceID matches
	if recording.ResourceID != req.ResourceID {
		return nil, fmt.Errorf("resourceId mismatch. Expected: %s, Got: %s", recording.ResourceID, req.ResourceID)
	}

	// Validate ChannelName matches
	if recording.ChannelName != req.ChannelName {
		return nil, fmt.Errorf("channelName mismatch. Expected: %s, Got: %s", recording.ChannelName, req.ChannelName)
	}

	// Validate UID matches (critical for Agora API)
	if recording.UID != req.UID {
		return nil, fmt.Errorf("UID mismatch. The UID used to stop recording (%s) must match the UID used to start recording (%s)", req.UID, recording.UID)
	}

	// Stop recording via Agora using the provided UID
	result, err := s.agoraClient.StopRecording(req.ResourceID, req.SID, req.UID, req.ChannelName)
	if err != nil {
		return nil, fmt.Errorf("failed to stop Agora recording: %w", err)
	}

	// Save video recording to gallery if S3 key is available
	if result.S3Key != "" && recording.LoadID > 0 {
		gallery := &models.Gallery{
			LoadID:            recording.LoadID,
			FileName:          result.FileName,
			S3Key:             "", // Empty for video recordings (screenshots use this)
			VideoRecordingKey: result.S3Key,
		}
		if err := s.galleryRepo.Create(gallery); err != nil {
			// Log error but don't fail the request - recording is already stopped
			fmt.Printf("Warning: Failed to save video recording to gallery: %v\n", err)
		}
	}

	// Remove from active recordings
	s.mu.Lock()
	delete(s.activeRecordings, req.SID)
	s.mu.Unlock()

	return &StopRecordingResponse{
		Success:  true,
		FileName: result.FileName,
		S3Key:    result.S3Key,
		S3URL:    result.S3URL,
		FileList: result.FileList,
		FileSize: result.FileSize,
		Duration: result.Duration,
		Status:   "completed",
	}, nil
}

// QueryRecording queries the status of a recording
func (s *RecordingService) QueryRecording(resourceID, sid string) (map[string]interface{}, error) {
	return s.agoraClient.QueryRecording(resourceID, sid)
}
