package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/securedlinq/backend/internal/service"
	"github.com/securedlinq/backend/pkg/agora"
)

// AgoraHandler handles Agora-related HTTP requests
type AgoraHandler struct {
	agoraClient      *agora.Client
	recordingService *service.RecordingService
}

// NewAgoraHandler creates a new Agora handler
func NewAgoraHandler(agoraClient *agora.Client, recordingService *service.RecordingService) *AgoraHandler {
	return &AgoraHandler{
		agoraClient:      agoraClient,
		recordingService: recordingService,
	}
}

// TokenRequest represents a token generation request
// UID can be sent as either a number or string
type TokenRequest struct {
	ChannelName string      `json:"channelName" binding:"required"`
	UID         interface{} `json:"uid" binding:"required"` // Accept both string and number
	Role        string      `json:"role"`
}

// GenerateToken generates an Agora RTC token
func (h *AgoraHandler) GenerateToken(c *gin.Context) {
	var req TokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "channelName and uid are required"})
		return
	}

	// Convert UID to string (handles both number and string)
	var uidStr string
	switch v := req.UID.(type) {
	case string:
		uidStr = v
	case float64:
		// JSON numbers are parsed as float64
		uidStr = fmt.Sprintf("%.0f", v)
	case int:
		uidStr = fmt.Sprintf("%d", v)
	case int64:
		uidStr = fmt.Sprintf("%d", v)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "uid must be a number or string"})
		return
	}

	if uidStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "uid cannot be empty"})
		return
	}

	// Determine role
	role := agora.RolePublisher
	if req.Role == "subscriber" {
		role = agora.RoleSubscriber
	}

	// Token expires in 24 hours
	expireSeconds := uint32(86400)

	token, err := agora.GenerateRTCToken(
		h.agoraClient.GetAppID(),
		h.agoraClient.GetAppCertificate(),
		req.ChannelName,
		uidStr,
		role,
		expireSeconds,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":          token,
		"appId":          h.agoraClient.GetAppID(),
		"channelName":    req.ChannelName,
		"uid":            uidStr,
		"expirationTime": time.Now().Unix() + int64(expireSeconds),
	})
}

// StartRecordingRequest represents a recording start request
type StartRecordingRequest struct {
	RoomID      string `json:"roomId" binding:"required"`
	ChannelName string `json:"channelName" binding:"required"`
	UID         string `json:"uid" binding:"required"`
	Token       string `json:"token" binding:"required"`
}

// StartRecording starts cloud recording
func (h *AgoraHandler) StartRecording(c *gin.Context) {
	var req StartRecordingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	result, err := h.recordingService.StartRecording(&service.StartRecordingRequest{
		RoomID:      req.RoomID,
		ChannelName: req.ChannelName,
		UID:         req.UID,
		Token:       req.Token,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// StopRecordingRequest represents a recording stop request
type StopRecordingRequest struct {
	ResourceID  string `json:"resourceId" binding:"required"`
	SID         string `json:"sid" binding:"required"`
	ChannelName string `json:"channelName" binding:"required"`
	UID         string `json:"uid" binding:"required"`
}

// StopRecording stops cloud recording
func (h *AgoraHandler) StopRecording(c *gin.Context) {
	var req StopRecordingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "resourceId, sid, channelName, and uid are required"})
		return
	}

	result, err := h.recordingService.StopRecording(&service.StopRecordingRequest{
		ResourceID:  req.ResourceID,
		SID:         req.SID,
		ChannelName: req.ChannelName,
		UID:         req.UID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// QueryRecordingRequest represents a recording query request
type QueryRecordingRequest struct {
	ResourceID string `json:"resourceId" binding:"required"`
	SID        string `json:"sid" binding:"required"`
}

// QueryRecording queries recording status
func (h *AgoraHandler) QueryRecording(c *gin.Context) {
	resourceID := c.Query("resourceId")
	sid := c.Query("sid")

	if resourceID == "" || sid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "resourceId and sid are required"})
		return
	}

	result, err := h.recordingService.QueryRecording(resourceID, sid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
