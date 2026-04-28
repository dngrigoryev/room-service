package handlers

import (
	"errors"
	"net/http"

	"room-service/internal/domain"
	"room-service/internal/service"

	"github.com/gin-gonic/gin"
)

type SlotHandler struct {
	slotService *service.SlotService
}

func NewSlotHandler(slotService *service.SlotService) *SlotHandler {
	return &SlotHandler{slotService: slotService}
}

func (h *SlotHandler) List(c *gin.Context) {
	roomID := c.Param("roomId")
	if roomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_REQUEST", "message": "roomId is required"},
		})
		return
	}

	dateStr := c.Query("date")
	if dateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_REQUEST", "message": "date query parameter is required"},
		})
		return
	}

	slots, err := h.slotService.GenerateAvailableSlots(c.Request.Context(), roomID, dateStr)
	if err != nil {
		if errors.Is(err, domain.ErrRoomNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{"code": "ROOM_NOT_FOUND", "message": "room not found"},
			})
			return
		}

		if err.Error() == "invalid date structure. Expected YYYY-MM-DD" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error()},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"slots": slots})
}
