package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRBACMiddleware(t *testing.T) {
	t.Run("success - correct role", func(t *testing.T) {
		router := gin.Default()
		router.Use(func(c *gin.Context) { // Mocking the AuthMiddleware part
			c.Set("role", "Admin")
		})
		router.Use(RBACMiddleware("Admin"))
		router.GET("/test-rbac", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req, _ := http.NewRequest(http.MethodGet, "/test-rbac", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("failure - incorrect role", func(t *testing.T) {
		router := gin.Default()
		router.Use(func(c *gin.Context) {
			c.Set("role", "Applicant")
		})
		router.Use(RBACMiddleware("Admin")) // Requires Admin
		router.GET("/test-rbac", func(c *gin.Context) {
			c.Status(http.StatusOK) // This should not be reached
		})

		req, _ := http.NewRequest(http.MethodGet, "/test-rbac", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.JSONEq(t, `{"error": "Permission denied"}`, w.Body.String())
	})

	t.Run("failure - role not in context", func(t *testing.T) {
		router := gin.Default()
		// No role set in context
		router.Use(RBACMiddleware("Admin"))
		router.GET("/test-rbac", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req, _ := http.NewRequest(http.MethodGet, "/test-rbac", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.JSONEq(t, `{"error": "Role not found in token, access denied"}`, w.Body.String())
	})

	t.Run("failure - role is not a string", func(t *testing.T) {
		router := gin.Default()
		router.Use(func(c *gin.Context) {
			c.Set("role", 123) // Role is an int, not a string
		})
		router.Use(RBACMiddleware("Admin"))
		router.GET("/test-rbac", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req, _ := http.NewRequest(http.MethodGet, "/test-rbac", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.JSONEq(t, `{"error": "Permission denied"}`, w.Body.String())
	})
}
