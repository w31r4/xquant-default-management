package middleware

import (
	"net/http"
	"strings"
	"xquant-default-management/internal/config"
	"xquant-default-management/internal/utils"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware 创建一个认证中间件
func AuthMiddleware(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer {token}"})
			return
		}

		tokenString := parts[1]
		claims, err := utils.ValidateToken(tokenString, cfg)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		// 将解析出的用户信息存入 Gin 的上下文中，以便后续的处理器使用
		c.Set("userID", claims.UserID)
		c.Set("role", claims.Role)

		c.Next() // 请求有效，继续处理
	}
}
