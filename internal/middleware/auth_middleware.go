// package middleware 存放所有 Gin 的中间件。
// 中间件是一种在请求处理链中，执行于实际的业务处理器之前或之后的函数。
// 它们通常用于处理认证、日志、跨域、panic 恢复等横切关注点。
package middleware

import (
	"net/http"
	"strings"
	"xquant-default-management/internal/config"
	"xquant-default-management/internal/utils"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware 是一个创建认证中间件的工厂函数。
// 它接收一个 config.Config 依赖，以便在验证 JWT 时能够获取到 JWTSecret。
// 这种返回 gin.HandlerFunc 的模式是 Gin 中间件的标准写法，允许我们向中间件传递依赖。
func AuthMiddleware(cfg config.Config) gin.HandlerFunc {
	// 返回的这个匿名函数才是真正的中间件处理器。
	return func(c *gin.Context) {
		// 1. 从请求头中获取 Authorization 字段。
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// 如果请求头中没有 Authorization 字段，说明请求未携带令牌，是未授权的。
			// c.AbortWithStatusJSON 会立即中断请求处理链，并返回一个 JSON 响应。
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}

		// 2. 校验 Authorization 请求头的格式。
		// 正确的格式应该是 "Bearer {token}"，由空格分隔成两部分。
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			// 如果格式不正确，同样视为未授权。
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer {token}"})
			return
		}

		// 3. 提取并验证 JWT。
		tokenString := parts[1]
		claims, err := utils.ValidateToken(tokenString, cfg)
		if err != nil {
			// 如果 ValidateToken 函数返回错误（例如 token 过期、签名无效），则视为令牌无效。
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		// 4. 将解析出的用户信息存入 Gin 的上下文中。
		// 这是中间件之间以及中间件与最终处理器之间传递数据的关键方式。
		// 将 userID 和 role 存入后，后续的处理器 (handler) 就可以通过 c.Get("userID") 来获取当前登录用户的信息，
		// 无需重复解析和验证 JWT。
		c.Set("userID", claims.UserID)
		c.Set("role", claims.Role)

		// 5. 请求有效，继续处理。
		// c.Next() 会将请求的控制权交还给处理链中的下一个中间件或最终的处理器。
		// 如果前面的步骤中调用了 c.Abort...，那么 c.Next() 将不会被执行。
		c.Next()
	}
}
