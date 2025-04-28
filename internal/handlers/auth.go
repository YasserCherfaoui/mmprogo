package handlers

import (
	"marketprogo/internal/models"
	"marketprogo/pkg/auth"
	"marketprogo/pkg/database"
	"marketprogo/pkg/logger"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Phone     string `json:"phone"`
	UserType  string `json:"user_type" binding:"required,oneof=B2C B2B"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user already exists
	var existingUser models.User
	if err := database.GetDB().Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error.Printf("Failed to hash password: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process registration"})
		return
	}

	// Create user
	user := models.User{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Phone:        req.Phone,
		UserType:     models.UserType(req.UserType),
		IsActive:     true,
		LastLogin:    time.Now(),
	}

	if err := database.GetDB().Create(&user).Error; err != nil {
		logger.Error.Printf("Failed to create user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	accessToken, err := auth.GenerateToken(user.ID, string(user.UserType), user.CompanyID)
	if err != nil {
		logger.Error.Printf("Failed to generate access token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
		return
	}
	refreshToken, err := auth.GenerateToken(user.ID, string(user.UserType), user.CompanyID)
	if err != nil {
		logger.Error.Printf("Failed to generate refresh token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}
	tokens := TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    3600,
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"tokens":  tokens,
		"user":    user,
	})
}

func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find user
	var user models.User
	if err := database.GetDB().Where("email = ?", req.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}
		logger.Error.Printf("Failed to find user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process login"})
		return
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Update last login
	user.LastLogin = time.Now()
	if err := database.GetDB().Save(&user).Error; err != nil {
		logger.Error.Printf("Failed to update last login: %v", err)
	}

	// TODO: Generate JWT tokens
	accessToken, err := auth.GenerateToken(user.ID, string(user.UserType), user.CompanyID)
	if err != nil {
		logger.Error.Printf("Failed to generate access token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
		return
	}
	refreshToken, err := auth.GenerateToken(user.ID, string(user.UserType), user.CompanyID)
	if err != nil {
		logger.Error.Printf("Failed to generate refresh token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	tokens := TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    3600,
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"tokens":  tokens,
		"user":    user,
	})
}

func RefreshToken(c *gin.Context) {
	// TODO: Implement token refresh logic
	refreshToken := c.PostForm("refresh_token")
	if refreshToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Refresh token is required"})
		return
	}

	claims, err := auth.ValidateToken(refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	if claims.ExpiresAt.Before(time.Now()) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token has expired"})
		return
	}

	accessToken, err := auth.GenerateToken(claims.UserID, claims.UserType, claims.CompanyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
		return
	}

	tokens := TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    3600,
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Token refreshed successfully",
		"tokens":  tokens,
	})
}
