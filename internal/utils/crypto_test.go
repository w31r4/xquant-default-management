package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashPassword(t *testing.T) {
	password := "plain-text-password"
	hash, err := HashPassword(password)

	assert.NoError(t, err, "HashPassword should not return an error")
	assert.NotEmpty(t, hash, "Hash should not be empty")
	assert.NotEqual(t, password, hash, "Hash should be different from the password")
}

func TestCheckPasswordHash(t *testing.T) {
	password := "my-secret-password"
	hash, _ := HashPassword(password)

	// Test with correct password
	assert.True(t, CheckPasswordHash(password, hash), "CheckPasswordHash should return true for correct password")

	// Test with incorrect password
	assert.False(t, CheckPasswordHash("wrong-password", hash), "CheckPasswordHash should return false for incorrect password")
}
