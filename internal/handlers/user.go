package handlers

import (
	"marketprogo/internal/models"
	"marketprogo/pkg/database"
	"marketprogo/pkg/logger"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UpdateProfileRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
}

func GetUserProfile(c *gin.Context) {
	// TODO: Get user ID from JWT token
	userID := uint(1) // Temporary hardcoded user ID

	var user models.User
	if err := database.GetDB().Preload("Company").Preload("Addresses").First(&user, userID).Error; err != nil {
		logger.Error.Printf("Failed to get user profile: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func UpdateUserProfile(c *gin.Context) {
	// TODO: Get user ID from JWT token
	userID := uint(1) // Temporary hardcoded user ID

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := database.GetDB().First(&user, userID).Error; err != nil {
		logger.Error.Printf("Failed to find user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	// Update fields
	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}

	if err := database.GetDB().Save(&user).Error; err != nil {
		logger.Error.Printf("Failed to update user profile: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"user":    user,
	})
}
