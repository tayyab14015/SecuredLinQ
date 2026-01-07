package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/securedlinq/backend/internal/middleware"
	"github.com/securedlinq/backend/internal/service"
)

// LoadHandler handles load HTTP requests
type LoadHandler struct {
	loadService *service.LoadService
}

// NewLoadHandler creates a new load handler
func NewLoadHandler(loadService *service.LoadService) *LoadHandler {
	return &LoadHandler{
		loadService: loadService,
	}
}

// CreateLoadRequest represents a request to create a load
type CreateLoadRequest struct {
	LoadNumber      string `json:"load_number" binding:"required"`
	Description     string `json:"description"`
	PickupAddress   string `json:"pickup_address"`
	DeliveryAddress string `json:"delivery_address"`
}

// CreateLoad creates a new load (admin only)
func (h *LoadHandler) CreateLoad(c *gin.Context) {
	var req CreateLoadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Load number is required"})
		return
	}

	// Get admin user from context
	user, ok := middleware.GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	load, err := h.loadService.CreateLoad(&service.CreateLoadRequest{
		LoadNumber:      req.LoadNumber,
		Description:     req.Description,
		PickupAddress:   req.PickupAddress,
		DeliveryAddress: req.DeliveryAddress,
		CreatedByID:     user.UserID,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"load":    load,
	})
}

// GetAllLoads gets all loads with pagination (admin only)
func (h *LoadHandler) GetAllLoads(c *gin.Context) {
	page := 1
	pageSize := 20

	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if ps := c.Query("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 {
			pageSize = parsed
		}
	}

	loads, total, err := h.loadService.GetAllLoads(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"loads":   loads,
		"total":   total,
		"page":    page,
		"pageSize": pageSize,
	})
}

// GetLoadByID gets a load by ID
func (h *LoadHandler) GetLoadByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid load ID"})
		return
	}

	load, err := h.loadService.GetLoadByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Load not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"load":    load,
	})
}

// AssignDriverRequest represents a request to assign a driver to a load
type AssignDriverRequest struct {
	DriverID uint `json:"driver_id" binding:"required"`
}

// AssignDriver assigns a driver to a load (admin only)
func (h *LoadHandler) AssignDriver(c *gin.Context) {
	loadIDStr := c.Param("id")
	loadID, err := strconv.ParseUint(loadIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid load ID"})
		return
	}

	var req AssignDriverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Driver ID is required"})
		return
	}

	if err := h.loadService.AssignDriverToLoad(uint(loadID), req.DriverID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get updated load
	load, _ := h.loadService.GetLoadByID(uint(loadID))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Driver assigned successfully",
		"load":    load,
	})
}

// GetDriverLoads gets loads assigned to the current driver
func (h *LoadHandler) GetDriverLoads(c *gin.Context) {
	user, ok := middleware.GetCurrentUser(c)
	if !ok || user.UserType != "driver" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	page := 1
	pageSize := 20

	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if ps := c.Query("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 {
			pageSize = parsed
		}
	}

	loads, total, err := h.loadService.GetLoadsByDriverID(uint(user.UserID), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"loads":    loads,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// UpdateLoadStatusRequest represents a request to update load status
type UpdateLoadStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// UpdateLoadStatus updates a load's status (driver only - Assigned → Completed)
func (h *LoadHandler) UpdateLoadStatus(c *gin.Context) {
	user, ok := middleware.GetCurrentUser(c)
	if !ok || user.UserType != "driver" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	loadIDStr := c.Param("id")
	loadID, err := strconv.ParseUint(loadIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid load ID"})
		return
	}

	var req UpdateLoadStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status is required"})
		return
	}

	if err := h.loadService.UpdateLoadStatus(uint(loadID), req.Status, uint(user.UserID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get updated load
	load, _ := h.loadService.GetLoadByID(uint(loadID))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Load status updated",
		"load":    load,
	})
}

// MarkCompleted marks a load as completed (driver only)
func (h *LoadHandler) MarkCompleted(c *gin.Context) {
	user, ok := middleware.GetCurrentUser(c)
	if !ok || user.UserType != "driver" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	loadIDStr := c.Param("id")
	loadID, err := strconv.ParseUint(loadIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid load ID"})
		return
	}

	if err := h.loadService.MarkLoadCompleted(uint(loadID), uint(user.UserID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get updated load
	load, _ := h.loadService.GetLoadByID(uint(loadID))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Load marked as completed",
		"load":    load,
	})
}

// StartMeeting starts a meeting for a completed load (admin only)
func (h *LoadHandler) StartMeeting(c *gin.Context) {
	loadIDStr := c.Param("id")
	loadID, err := strconv.ParseUint(loadIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid load ID"})
		return
	}

	load, err := h.loadService.StartMeeting(uint(loadID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Meeting can now be started",
		"load":    load,
	})
}

// GetLoadsByStatus gets loads by status (admin only)
func (h *LoadHandler) GetLoadsByStatus(c *gin.Context) {
	status := c.Query("status")
	if status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status is required"})
		return
	}

	page := 1
	pageSize := 20

	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if ps := c.Query("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 {
			pageSize = parsed
		}
	}

	loads, total, err := h.loadService.GetLoadsByStatus(status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"loads":    loads,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// DeleteLoad deletes a load (admin only)
func (h *LoadHandler) DeleteLoad(c *gin.Context) {
	loadIDStr := c.Param("id")
	loadID, err := strconv.ParseUint(loadIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid load ID"})
		return
	}

	if err := h.loadService.DeleteLoad(uint(loadID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Load deleted successfully",
	})
}

