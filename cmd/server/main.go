package main

import (
	"log"
	"net/http"
	"xquant-default-management/internal/config"
	"xquant-default-management/internal/database"
	"xquant-default-management/internal/handler"
	"xquant-default-management/internal/middleware" // 新增
	"xquant-default-management/internal/repository"
	"xquant-default-management/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid" // 新增
)

func main() {
	// 1. 初始化配置
	cfg, err := config.LoadConfig("./configs")
	if err != nil {
		log.Fatalf("could not load config: %v", err)
	}

	// 2. 初始化数据库连接
	database.Connect(cfg)
	db := database.DB // 获取数据库实例
	//database 库根据已有模板向数据库中创建对应表格

	// 3. 依赖注入：将所有组件连接起来
	// Repository -> Service -> Handler
	// 将组件一步步链接
	userRepository := repository.NewUserRepository(db)
	// 将 cfg 传递给 UserService

	userService := service.NewUserService(userRepository, cfg)
	userHandler := handler.NewUserHandler(userService)

	// 4. 初始化 Gin 引擎
	router := gin.Default()

	// 5. 注册路由
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})

	})

	// 路由分组
	apiV1 := router.Group("/api/v1")
	{
		// 公开路由：注册和登录
		apiV1.POST("/register", userHandler.Register)
		apiV1.POST("/login", userHandler.Login)

		// 受保护的路由组
		protected := apiV1.Group("/")
		protected.Use(middleware.AuthMiddleware(cfg)) // 应用认证中间件
		{
			// 一个简单的测试端点，用于检查认证是否有效
			protected.GET("/profile", func(c *gin.Context) {
				// 从上下文中获取 userID 和 role
				userID, _ := c.Get("userID")
				role, _ := c.Get("role")

				c.JSON(http.StatusOK, gin.H{
					"message": "Welcome to your profile!",
					"user_id": userID.(uuid.UUID).String(),
					"role":    role.(string),
				})
			})
		}
	}

	// 6. 启动服务器
	log.Printf("Server is listening on port %s", cfg.ServerPort)
	if err := router.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
