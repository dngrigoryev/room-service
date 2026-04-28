package handlers

import (
	"errors"
	"net/http"

	"room-service/internal/domain"
	"room-service/internal/service"

	"github.com/gin-gonic/gin"
)

type ScheduleHandler struct {
	scheduleService *service.ScheduleService
}

func NewScheduleHandler(scheduleService *service.ScheduleService) *ScheduleHandler {
	return &ScheduleHandler{scheduleService: scheduleService}
}

type CreateScheduleRequest struct {
	DaysOfWeek []int  `json:"daysOfWeek" binding:"required"`
	StartTime  string `json:"startTime" binding:"required"`
	EndTime    string `json:"endTime" binding:"required"`
}

func (h *ScheduleHandler) Create(c *gin.Context) {
	roomID := c.Param("roomId")
	if roomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": "Room ID is missing in path parameters",
			},
		})
		return
	}

	var req CreateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": "Invalid request parameters (JSON field missing or wrong type)",
			},
		})
		return
	}

	schedule, err := h.scheduleService.Create(c.Request.Context(), roomID, req.DaysOfWeek, req.StartTime, req.EndTime)

	if err != nil {
		if errors.Is(err, domain.ErrScheduleExists) {
			c.JSON(http.StatusConflict, gin.H{
				"error": gin.H{
					"code":    "SCHEDULE_EXISTS",
					"message": "schedule for this room already exists and cannot be changed",
				},
			})
			return
		}

		if errors.Is(err, domain.ErrRoomNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "ROOM_NOT_FOUND",
					"message": "room to apply schedule not found",
				},
			})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"schedule": schedule,
	})
}
