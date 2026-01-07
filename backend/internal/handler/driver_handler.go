package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/securedlinq/backend/internal/service"
)

// DriverHandler handles driver HTTP requests
type DriverHandler struct {
	driverService *service.DriverService
}

// NewDriverHandler creates a new driver handler
func NewDriverHandler(driverService *service.DriverService) *DriverHandler {
	return &DriverHandler{
		driverService: driverService,
	}
}

// GetAllDrivers gets all drivers with pagination (admin only)
func (h *DriverHandler) GetAllDrivers(c *gin.Context) {
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

	drivers, total, err := h.driverService.GetAllDrivers(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"drivers":  drivers,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// GetDriverByID gets a driver by ID (admin only)
func (h *DriverHandler) GetDriverByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid driver ID"})
		return
	}

	driver, err := h.driverService.GetDriverByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Driver not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"driver":  driver,
	})
}

// DeactivateDriver deactivates a driver account (admin only)
func (h *DriverHandler) DeactivateDriver(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid driver ID"})
		return
	}

	if err := h.driverService.DeactivateDriver(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Driver deactivated successfully",
	})
}

// ActivateDriver activates a driver account (admin only)
func (h *DriverHandler) ActivateDriver(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid driver ID"})
		return
	}

	if err := h.driverService.ActivateDriver(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Driver activated successfully",
	})
}

