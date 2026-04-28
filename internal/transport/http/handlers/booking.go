package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"room-service/internal/domain"
	"room-service/internal/service"

	"github.com/gin-gonic/gin"
)

type BookingHandler struct {
	bookingService *service.BookingService
}

func NewBookingHandler(bookingService *service.BookingService) *BookingHandler {
	return &BookingHandler{bookingService: bookingService}
}

type CreateBookingRequest struct {
	SlotID string `json:"slotId" binding:"required"`
}

func (h *BookingHandler) Create(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req CreateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_REQUEST", "message": "Invalid request body"},
		})
		return
	}

	booking, err := h.bookingService.Create(c.Request.Context(), userID.(string), req.SlotID)

	if err != nil {
		if errors.Is(err, domain.ErrSlotNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "SLOT_NOT_FOUND", "message": "slot not found"}})
			return
		}
		if errors.Is(err, domain.ErrSlotInPast) {
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": "cannot book past slot"}})
			return
		}
		if errors.Is(err, domain.ErrSlotAlreadyBooked) {
			c.JSON(http.StatusConflict, gin.H{"error": gin.H{"code": "SLOT_ALREADY_BOOKED", "message": "slot already booked"}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "server error"}})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"booking": booking})
}

func (h *BookingHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	if pageSize > 100 {
		pageSize = 100
	}

	bookings, pagination, err := h.bookingService.ListAll(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "server error"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"bookings":   bookings,
		"pagination": pagination,
	})
}

func (h *BookingHandler) My(c *gin.Context) {
	userID, _ := c.Get("user_id")

	bookings, err := h.bookingService.ListFutureByUser(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "server error"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"bookings": bookings})
}

func (h *BookingHandler) Cancel(c *gin.Context) {
	userID, _ := c.Get("user_id")
	bookingID := c.Param("bookingId")

	booking, err := h.bookingService.Cancel(c.Request.Context(), bookingID, userID.(string))
	if err != nil {
		if errors.Is(err, domain.ErrBookingNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "BOOKING_NOT_FOUND", "message": "booking not found"}})
			return
		}
		if errors.Is(err, domain.ErrForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": gin.H{"code": "FORBIDDEN", "message": "cannot cancel another user's booking"}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "server error"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"booking": booking})
}
