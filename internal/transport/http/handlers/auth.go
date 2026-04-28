package handlers

import (
	"log"
	"net/http"
	"time"

	"room-service/internal/service"
	"room-service/pkg/token"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	tokenManager *token.Manager
	authService  *service.AuthService
}

func NewAuthHandler(tm *token.Manager, authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		tokenManager: tm,
		authService:  authService,
	}
}

type DummyLoginRequest struct {
	Role string `json:"role" binding:"required"`
}

type TokenResponse struct {
	Token string `json:"token"`
}

func (h *AuthHandler) DummyLogin(c *gin.Context) {
	var req DummyLoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("DummyLogin validation failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": "Invalid request body or role missing",
			},
		})
		return
	}

	if req.Role != "admin" && req.Role != "user" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": "Role must be 'admin' or 'user'",
			},
		})
		return
	}

	const adminUUID = "00000000-0000-0000-0000-000000000001"
	const userUUID = "00000000-0000-0000-0000-000000000002"

	var userID string
	if req.Role == "admin" {
		userID = adminUUID
	} else {
		userID = userUUID
	}

	t, err := h.tokenManager.Generate(userID, req.Role, 24*time.Hour)
	if err != nil {
		log.Printf("Token generation failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Could not generate token",
			},
		})
		return
	}

	c.JSON(http.StatusOK, TokenResponse{Token: t})
}

type AuthRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role" binding:"required"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_REQUEST", "message": "Invalid request parameters"},
		})
		return
	}

	user, err := h.authService.Register(c.Request.Context(), req.Email, req.Password, req.Role)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "REGISTRATION_FAILED", "message": err.Error()},
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": user.ID, "email": user.Email, "role": user.Role})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_REQUEST", "message": "Invalid request parameters"},
		})
		return
	}

	token, err := h.authService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{"code": "UNAUTHORIZED", "message": "Invalid credentials"},
		})
		return
	}

	c.JSON(http.StatusOK, TokenResponse{Token: token})
}
