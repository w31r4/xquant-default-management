package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"xquant-default-management/internal/config"
	"xquant-default-management/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
	cfg := config.Config{JWTSecret: "test-secret-key", TokenTTL: 1}
	userID := uuid.New()
	role := "Applicant"

	// Create a test router with the middleware and a test handler
	router := gin.Default()
	router.Use(AuthMiddleware(cfg))
	router.GET("/test", func(c *gin.Context) {
		uid, uidExists := c.Get("userID")
		r, rExists := c.Get("role")

		assert.True(t, uidExists)
		assert.True(t, rExists)
		assert.Equal(t, userID, uid)
		assert.Equal(t, role, r)

		c.Status(http.StatusOK)
	})

	t.Run("success - valid token", func(t *testing.T) {
		token, _ := utils.GenerateToken(userID, role, cfg)
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("failure - no auth header", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.JSONEq(t, `{"error": "Authorization header is required"}`, w.Body.String())
	})

	t.Run("failure - invalid auth header format", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "invalid-token-format")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.JSONEq(t, `{"error": "Authorization header format must be Bearer {token}"}`, w.Body.String())
	})

	t.Run("failure - invalid token", func(t *testing.T) {
		invalidCfg := config.Config{JWTSecret: "wrong-secret", TokenTTL: 1}
		token, _ := utils.GenerateToken(userID, role, invalidCfg) // Token signed with a different key
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.JSONEq(t, `{"error": "Invalid token"}`, w.Body.String())
	})

	t.Run("failure - expired token", func(t *testing.T) {
		expiredCfg := config.Config{JWTSecret: cfg.JWTSecret, TokenTTL: -1} // Expired TTL
		token, _ := utils.GenerateToken(userID, role, expiredCfg)
		time.Sleep(1 * time.Second) // Ensure it's expired

		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.JSONEq(t, `{"error": "Invalid token"}`, w.Body.String())
	})
}
