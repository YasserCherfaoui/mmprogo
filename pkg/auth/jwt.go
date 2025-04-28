package auth

import (
	"errors"
	"time"

	"marketprogo/config"

	"github.com/golang-jwt/jwt/v4"
)

var (
	// Change this to a secure key in production, preferably from environment variables
	secretKey = []byte(config.LoadConfig().JWTSecret)

	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

type Claims struct {
	UserID    uint   `json:"user_id"`
	UserType  string `json:"user_type"`
	CompanyID *uint  `json:"company_id,omitempty"`
	jwt.RegisteredClaims
}

// GenerateToken creates a new JWT token for a user
func GenerateToken(userID uint, userType string, companyID *uint) (string, error) {
	claims := Claims{
		UserID:    userID,
		UserType:  userType,
		CompanyID: companyID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // Token expires in 24 hours
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken validates the JWT token and returns the claims
func ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// RefreshToken generates a new token with extended expiration
func RefreshToken(oldTokenString string) (string, error) {
	claims, err := ValidateToken(oldTokenString)
	if err != nil && !errors.Is(err, ErrExpiredToken) {
		return "", err
	}

	// Generate new token with same claims but new expiration
	return GenerateToken(claims.UserID, claims.UserType, claims.CompanyID)
}
