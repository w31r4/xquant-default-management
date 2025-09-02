package main

import (
	"log"
	"net/http"
	"strings"
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
	// Repositories

	userRepository := repository.NewUserRepository(db)
	customerRepository := repository.NewCustomerRepository(db) // 新增
	appRepository := repository.NewApplicationRepository(db)   // 新增

	// Services

	userService := service.NewUserService(userRepository, cfg)
	// 将 cfg 传递给 UserService
	appService := service.NewApplicationService(appRepository, customerRepository) // 新增

	// Handlers
	userHandler := handler.NewUserHandler(userService)
	appHandler := handler.NewApplicationHandler(appService) // 新增

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

			// 新增一个彩蛋路由
			protected.GET("/easter-egg", func(c *gin.Context) {
				// 我们可以从上下文中获取用户信息，让彩蛋更个性化
				userID, _ := c.Get("userID")

				// 把 UUID 转换成字符串，只取第一部分，让它看起来像个代号
				userIDStr := userID.(uuid.UUID).String()
				agentCode := strings.Split(userIDStr, "-")[0]

				c.JSON(http.StatusOK, gin.H{
					"message":        "Congratulations, Agent " + agentCode + "!",
					"secret_mission": "Find the hidden rubber duck in the repository.",
				})
			})
			// 申请相关路由
			applications := protected.Group("/applications")
			{
				// 只有 'Applicant' 角色的用户才能提交申请
				applications.POST("", middleware.RBACMiddleware("Applicant"), appHandler.CreateApplication)
			}

		}
	}

	// 6. 启动服务器
	log.Printf("Server is listening on port %s", cfg.ServerPort)
	if err := router.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
