package handlers

import (
	"log"
	"net/http"

	"room-service/internal/service"

	"github.com/gin-gonic/gin"
)

type RoomHandler struct {
	roomService *service.RoomService
}

func NewRoomHandler(roomService *service.RoomService) *RoomHandler {
	return &RoomHandler{roomService: roomService}
}

type CreateRoomRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
	Capacity    *int    `json:"capacity"`
}

func (h *RoomHandler) Create(c *gin.Context) {
	var req CreateRoomRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": "Invalid request body (name is required)",
			},
		})
		return
	}

	room, err := h.roomService.CreateRoom(c.Request.Context(), req.Name, req.Description, req.Capacity)
	if err != nil {
		log.Printf("Failed to create room: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to create room",
			},
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"room": room})
}

func (h *RoomHandler) List(c *gin.Context) {
	rooms, err := h.roomService.ListRooms(c.Request.Context())
	if err != nil {
		log.Printf("Failed to fetch rooms: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to fetch rooms",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"rooms": rooms})
}
