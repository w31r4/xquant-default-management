package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RBACMiddleware 是一个基于角色的访问控制 (Role-Based Access Control) 中间件的工厂函数。
// 它接收一个 requiredRole 参数，并返回一个 Gin 中间件处理器。
// 这个返回的处理器会检查当前登录用户的角色是否与 requiredRole 匹配。
// 这种工厂模式使得中间件可以被灵活地复用，例如：
// - applications.POST("", middleware.RBACMiddleware("Applicant"), ...)
// - admin.DELETE("/users/:id", middleware.RBACMiddleware("Admin"), ...)
func RBACMiddleware(requiredRole string) gin.HandlerFunc {
	// 返回的这个匿名函数才是真正的中间件处理器。
	return func(c *gin.Context) {
		// 1. 从 Gin 上下文中获取用户角色。
		// !!! 重要前提 !!!
		// 这个中间件必须在 AuthMiddleware 之后运行。
		// AuthMiddleware 负责验证 JWT 并将解析出的用户信息（包括 role）存入 Gin 的上下文中。
		role, exists := c.Get("role")
		if !exists {
			// 如果上下文中不存在 "role"，说明 AuthMiddleware 没有成功运行或被错误地遗漏了。
			// 这是一种潜在的安全配置问题，因此我们拒绝访问。
			// 使用 403 Forbidden 表示服务器理解请求，但拒绝授权。
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Role not found in token, access denied"})
			return
		}

		// 2. 校验用户角色是否匹配所需的角色。
		// c.Get("role") 返回的是一个 interface{} 类型，需要先进行类型断言，将其转换为 string。
		userRole, ok := role.(string)
		
		// 检查类型断言是否成功，以及用户的角色是否与此中间件实例要求的角色 (requiredRole) 相符。
		if !ok || userRole != requiredRole {
			// 如果用户的角色不匹配，则没有权限访问此路由。
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
			return
		}

		// 3. 授权成功，继续处理请求。
		// 如果代码能执行到这里，说明用户的角色已通过验证。
		// c.Next() 将请求的控制权传递给处理链中的下一个处理器（可能是另一个中间件或最终的业务 Handler）。
		c.Next()
	}
}