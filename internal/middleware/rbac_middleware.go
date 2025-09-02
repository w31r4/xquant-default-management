package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RBACMiddleware 检查用户是否具有所需角色
func RBACMiddleware(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// AuthMiddleware 必须先运行，它会将 role 存入上下文
		role, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Role not found in token, access denied"})
			return
		}

		userRole, ok := role.(string)
		if !ok || userRole != requiredRole {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
			return
		}

		c.Next()
	}
}
