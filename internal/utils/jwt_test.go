package utils

import (
	"testing"
	"time"
	"xquant-default-management/internal/config"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGenerateAndValidateToken(t *testing.T) {
	cfg := config.Config{
		JWTSecret: "test-secret",
		TokenTTL:  1, // 1 hour
	}
	userID := uuid.New()
	role := "Applicant"

	tokenString, err := GenerateToken(userID, role, cfg)
	assert.NoError(t, err, "GenerateToken should not return an error")
	assert.NotEmpty(t, tokenString, "Token string should not be empty")

	claims, err := ValidateToken(tokenString, cfg)
	assert.NoError(t, err, "ValidateToken should not return an error for a valid token")
	assert.NotNil(t, claims, "Claims should not be nil for a valid token")
	assert.Equal(t, userID, claims.UserID, "UserID in claims should match the original UserID")
	assert.Equal(t, role, claims.Role, "Role in claims should match the original role")
}

func TestValidateTokenInvalidSignature(t *testing.T) {
	cfg1 := config.Config{JWTSecret: "secret-one", TokenTTL: 1}
	cfg2 := config.Config{JWTSecret: "secret-two", TokenTTL: 1}
	userID := uuid.New()
	role := "Approver"

	// Generate token with one secret
	tokenString, _ := GenerateToken(userID, role, cfg1)

	// Try to validate with another secret
	_, err := ValidateToken(tokenString, cfg2)
	assert.Error(t, err, "ValidateToken should return an error for an invalid signature")
}

func TestValidateTokenExpired(t *testing.T) {
	// Create a config with a very short TTL (negative to ensure it's expired)
	cfg := config.Config{
		JWTSecret: "expired-secret",
		TokenTTL:  -1, // Already expired
	}
	userID := uuid.New()
	role := "Admin"

	tokenString, _ := GenerateToken(userID, role, cfg)

	// It might take a moment for the token to be considered expired, so we wait briefly.
	time.Sleep(1 * time.Second)

	_, err := ValidateToken(tokenString, cfg)
	assert.Error(t, err, "ValidateToken should return an error for an expired token")
}
