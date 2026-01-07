package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/securedlinq/backend/internal/service"
)

// MeetingHandler handles meeting HTTP requests
type MeetingHandler struct {
	meetingService *service.MeetingService
}

// NewMeetingHandler creates a new meeting handler
func NewMeetingHandler(meetingService *service.MeetingService) *MeetingHandler {
	return &MeetingHandler{
		meetingService: meetingService,
	}
}

// CreateMeetingRequest represents a request to create a meeting
type CreateMeetingRequest struct {
	LoadID uint `json:"load_id" binding:"required"`
}

// CreateMeeting creates or retrieves a meeting room based on load_id
func (h *MeetingHandler) CreateMeeting(c *gin.Context) {
	var req CreateMeetingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "load_id is required"})
		return
	}

	if req.LoadID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "load_id must be greater than 0"})
		return
	}

	meetingRoom, err := h.meetingService.GetOrCreateMeetingRoom(req.LoadID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"room":    meetingRoom,
		"message": "Meeting room ready",
	})
}

// GetMeetingByRoomID gets a meeting room by room ID
func (h *MeetingHandler) GetMeetingByRoomID(c *gin.Context) {
	roomID := c.Query("roomId")
	if roomID == "" || roomID == "undefined" || roomID == "null" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "roomId is required"})
		return
	}

	meetingRoom, err := h.meetingService.GetMeetingRoomByRoomID(roomID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Link Invalid or Expired. This meeting link has been invalidated. Please request a new link from the admin.",
		})
		return
	}

	// Update last joined timestamp
	_ = h.meetingService.UpdateLastJoined(roomID)

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"meetingRoom": meetingRoom,
	})
}

// EndMeetingRequest represents a request to end a meeting
type EndMeetingRequest struct {
	RoomID string `json:"roomId" binding:"required"`
}

// EndMeeting ends a meeting room
func (h *MeetingHandler) EndMeeting(c *gin.Context) {
	var req EndMeetingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "roomId is required"})
		return
	}

	if err := h.meetingService.EndMeeting(req.RoomID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Meeting room not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Meeting room ended",
	})
}
