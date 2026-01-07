package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/securedlinq/backend/internal/config"
	"github.com/securedlinq/backend/internal/models"
	"github.com/securedlinq/backend/internal/repository"
	"github.com/securedlinq/backend/pkg/s3"
)

// MediaHandler handles media HTTP requests
type MediaHandler struct {
	s3Client     *s3.Client
	galleryRepo  *repository.GalleryRepository
	meetingRepo  *repository.MeetingRepository
	config       *config.Config
}

// NewMediaHandler creates a new media handler
func NewMediaHandler(s3Client *s3.Client, galleryRepo *repository.GalleryRepository, meetingRepo *repository.MeetingRepository, cfg *config.Config) *MediaHandler {
	return &MediaHandler{
		s3Client:    s3Client,
		galleryRepo: galleryRepo,
		meetingRepo: meetingRepo,
		config:      cfg,
	}
}

// GetLoadMedia gets all media for a load
func (h *MediaHandler) GetLoadMedia(c *gin.Context) {
	loadNumber := c.Query("loadNumber")
	if loadNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "loadNumber is required"})
		return
	}

	media, err := h.s3Client.ListLoadMedia(loadNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Transform for frontend
	transformedMedia := make([]map[string]interface{}, len(media))
	for i, m := range media {
		transformedMedia[i] = map[string]interface{}{
			"id":         m.Key,
			"type":       m.Type,
			"step":       m.Step,
			"timestamp":  m.LastModified,
			"fileName":   m.FileName,
			"size":       m.Size,
			"loadNumber": m.LoadNumber,
			"signedUrl":  m.SignedURL,
			"s3Key":      m.Key,
			"uri":        m.SignedURL,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"media":   transformedMedia,
	})
}

// SaveScreenshotRequest represents a screenshot save request
type SaveScreenshotRequest struct {
	Screenshot string `json:"screenshot" binding:"required"`
	RoomID     string `json:"roomId,omitempty"` // Optional: to get load_id from meeting room
	LoadID     uint   `json:"loadId,omitempty"` // Optional: direct load_id
}

// SaveScreenshot saves a screenshot to S3 and gallery
func (h *MediaHandler) SaveScreenshot(c *gin.Context) {
	var req SaveScreenshotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "screenshot is required"})
		return
	}

	// Get load_id from room_id or direct load_id
	var loadID uint
	var loadNumber string = "unknown"
	
	if req.LoadID > 0 {
		loadID = req.LoadID
	} else if req.RoomID != "" {
		// Get meeting room to find load_id
		meetingRoom, err := h.meetingRepo.GetByRoomID(req.RoomID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room_id"})
			return
		}
		if meetingRoom.LoadID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "meeting room has no associated load"})
			return
		}
		loadID = meetingRoom.LoadID
		// Get load number from meeting room if available
		if meetingRoom.LoadNumber.Valid {
			loadNumber = meetingRoom.LoadNumber.String
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "load_id or room_id is required"})
		return
	}

	// Upload to S3
	result, err := h.s3Client.UploadBase64Image(loadNumber, req.Screenshot)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !result.Success {
		c.JSON(http.StatusBadRequest, gin.H{"error": result.Error})
		return
	}

	// Save to gallery
	gallery := &models.Gallery{
		LoadID:   loadID,
		FileName: result.Key,
		S3Key:    result.Key,
	}
	if err := h.galleryRepo.Create(gallery); err != nil {
		// Log error but don't fail the request - screenshot is already uploaded
		fmt.Printf("Warning: Failed to save screenshot to gallery: %v\n", err)
	}

	// Construct direct S3 URL (bucket is public)
	// Format: https://{bucket}.s3.{region}.amazonaws.com/{key}
	directURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s",
		h.config.AWS.S3BucketName,
		h.config.AWS.Region,
		result.Key,
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"s3Key":   result.Key,
		"url":     directURL,
		"id":      gallery.ID,
	})
}

// GetSignedURLRequest represents a signed URL request
type GetSignedURLRequest struct {
	Key string `json:"key" binding:"required"`
}

// GetSignedURL gets a signed URL for an S3 object
func (h *MediaHandler) GetSignedURL(c *gin.Context) {
	key := c.Query("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key is required"})
		return
	}

	url, err := h.s3Client.GetSignedURL(key, 3600)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"url":     url,
	})
}

// GetScreenshotsByLoad gets all screenshots for a specific load
func (h *MediaHandler) GetScreenshotsByLoad(c *gin.Context) {
	loadIDStr := c.Query("loadId")
	if loadIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "loadId is required"})
		return
	}

	var loadID uint
	if _, err := fmt.Sscanf(loadIDStr, "%d", &loadID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid loadId"})
		return
	}

	galleries, err := h.galleryRepo.GetByLoadID(loadID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Construct direct S3 URLs for both screenshots and videos
	screenshots := make([]map[string]interface{}, 0, len(galleries))
	for _, gallery := range galleries {
		var directURL string
		var mediaType string
		
		// Determine if it's a video or screenshot
		if gallery.VideoRecordingKey != "" {
			// Video recording
			directURL = fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s",
				h.config.AWS.S3BucketName,
				h.config.AWS.Region,
				gallery.VideoRecordingKey,
			)
			mediaType = "video"
		} else if gallery.S3Key != "" {
			// Screenshot
			directURL = fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s",
				h.config.AWS.S3BucketName,
				h.config.AWS.Region,
				gallery.S3Key,
			)
			mediaType = "image"
		} else {
			// Skip entries with no media
			continue
		}
		
		screenshots = append(screenshots, map[string]interface{}{
			"id":        gallery.ID,
			"loadId":    gallery.LoadID,
			"fileName":  gallery.FileName,
			"s3Key":     gallery.S3Key,
			"videoKey":  gallery.VideoRecordingKey,
			"url":       directURL,
			"type":      mediaType,
			"createdAt": gallery.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"screenshots": screenshots,
	})
}

