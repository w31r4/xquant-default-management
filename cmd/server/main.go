package main

import (
	"log"
	"net/http"
	"strings"
	"xquant-default-management/internal/config"
	"xquant-default-management/internal/database"
	"xquant-default-management/internal/handler"
	"xquant-default-management/internal/middleware"
	"xquant-default-management/internal/repository"
	"xquant-default-management/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func main() {
	// =========================================================================
	// 1. 初始化配置 (Initialization)
	// =========================================================================
	// 应用启动的第一步：加载所有外部配置。
	// cfg 对象是整个应用的“配置中心”，包含了数据库连接信息、JWT 密钥等所有非代码的配置项。
	// 这种方式让配置与代码分离，便于在不同环境（开发、测试、生产）中切换。
	cfg, err := config.LoadConfig("./configs")
	if err != nil {
		log.Fatalf("无法加载配置: %v", err)
	}

	// =========================================================================
	// 2. 初始化数据库 (Database)
	// =========================================================================
	// 基于加载的配置，建立与 PostgreSQL 数据库的连接池。
	// 同时，GORM 的 AutoMigrate 功能会检查 Go 代码中的数据模型 (User, Customer, etc.)
	// 并自动在数据库中创建或更新对应的表结构。
	// 注意：AutoMigrate 在开发阶段非常方便，但在生产环境中需要谨慎使用，通常会用更专业的迁移工具。
	database.Connect(cfg)
	db := database.DB // 获取全局的数据库连接实例 (*gorm.DB)

	// =========================================================================
	// 3. 依赖注入 (Dependency Injection)
	// =========================================================================
	// 这是整个应用架构的核心所在。
	// 我们按照“由内向外”、“自底向上”的顺序，手动创建和“注入”所有组件的依赖关系。
	// 最终形成一个清晰的依赖链：Handler -> Service -> Repository -> DB

	// --- 数据访问层 (Repositories) ---
	// Repositories 是最底层的组件，它们直接与数据库对话，是数据的“仓库管理员”。
	// 它们只依赖于数据库连接 (db)，不依赖于任何其他业务组件，所以最先被创建。
	userRepository := repository.NewUserRepository(db)
	customerRepository := repository.NewCustomerRepository(db)
	appRepository := repository.NewApplicationRepository(db)

	// --- 业务逻辑层 (Services) ---
	// Services 是应用的核心，负责编排业务流程，是决策的“项目经理”。
	// 它们依赖于 Repositories 来获取和存储数据。
	// 注意，userService 还需要 cfg 来读取 JWT 相关的配置（密钥和过期时间）。
	userService := service.NewUserService(userRepository, cfg)
	appService := service.NewApplicationService(db, appRepository, customerRepository)

	// --- API 接口层 (Handlers) ---
	// Handlers 是最外层的组件，负责处理 HTTP 请求和响应，是应用的“前台接待”。
	// 它们依赖于 Services 来执行具体的业务操作。
	userHandler := handler.NewUserHandler(userService)
	appHandler := handler.NewApplicationHandler(appService)

	// =========================================================================
	// 4. 初始化 Web 引擎和注册路由 (Routing)
	// =========================================================================
	// 使用 Gin 框架作为我们的 HTTP 服务器。gin.Default() 会创建一个带有基本中间件（如日志、崩溃恢复）的路由引擎。
	router := gin.Default()

	// --- 健康检查路由 ---
	// 一个简单的 /ping 接口，用于检查服务是否正在运行。
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	// --- API 路由分组 ---
	// 使用路由分组来组织我们的 API，这有两个主要好处：
	// 1. 共享路径前缀：所有组内的路由都会自动加上 "/api/v1" 前缀，便于 API 版本管理。
	// 2. 共享中间件：可以对整个分组应用中间件。
	apiV1 := router.Group("/api/v1")
	{
		// --- 公开路由 (Public Routes) ---
		// 这些路由不需要用户登录即可访问。
		apiV1.POST("/register", userHandler.Register)
		apiV1.POST("/login", userHandler.Login)

		// --- 受保护的路由组 (Protected Routes) ---
		// 创建一个子分组，用于存放所有需要用户认证才能访问的接口。
		protected := apiV1.Group("/")
		// .Use() 方法会给这个分组下的所有路由都应用上指定的中间件。
		// AuthMiddleware 是我们的“保安 A”，负责检查请求是否携带了有效的 JWT (认证)。
		protected.Use(middleware.AuthMiddleware(cfg))
		{
			// 一个简单的个人资料接口，用于测试认证是否成功。
			protected.GET("/profile", func(c *gin.Context) {
				// 中间件成功验证 token 后，会将用户信息存入 gin.Context。
				// 后续的 Handler 就可以从中取出这些信息使用。
				userID, _ := c.Get("userID")
				role, _ := c.Get("role")

				c.JSON(http.StatusOK, gin.H{
					"message": "欢迎来到个人资料页！",
					"user_id": userID.(uuid.UUID).String(),
					"role":    role.(string),
				})
			})

			// 彩蛋路由，同样受 AuthMiddleware 保护。
			protected.GET("/easter-egg", func(c *gin.Context) {
				userID, _ := c.Get("userID")
				userIDStr := userID.(uuid.UUID).String()
				agentCode := strings.Split(userIDStr, "-")[0]

				c.JSON(http.StatusOK, gin.H{
					"message":        "恭喜你，" + agentCode + " 号特工！",
					"secret_mission": "在仓库里找到隐藏的橡皮鸭。",
				})
			})

			// --- 申请相关的路由 ---
			// 再次使用分组来组织所有与“申请 (applications)”相关的接口。
			applications := protected.Group("/applications")
			{
				// 这个接口需要双重保护：
				// 1. 继承自父分组 `protected` 的 AuthMiddleware (保安 A - 认证)，确保用户已登录。
				// 2. 针对此路由单独应用的 RBACMiddleware (保安 B - 授权)，确保用户角色是 'Applicant'。
				// 中间件会按照定义的顺序依次执行。
				applications.POST("", middleware.RBACMiddleware("Applicant"), appHandler.CreateApplication)
				// --- 新增审批路由 ---
				// 将审批相关的路由分组到 /review 下，更符合 RESTful 风格
				review := applications.Group("/review")
				review.Use(middleware.RBACMiddleware("Approver")) // 只有 Approver 角色能访问
				{
					review.POST("/approve", appHandler.ApproveApplication)
				}
			}
		}
	}

	// =========================================================================
	// 6. 启动服务器 (Start Server)
	// =========================================================================
	// 启动 HTTP 服务器，并监听在配置文件中指定的端口。
	// router.Run() 是一个阻塞操作，它会一直运行，直到程序被中断。
	log.Printf("服务器正在端口 %s 上监听...", cfg.ServerPort)
	if err := router.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}
