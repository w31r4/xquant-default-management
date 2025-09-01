package main

import (
	"log"
	"net/http"
	"xquant-default-management/internal/config"
	"xquant-default-management/internal/database"
	"xquant-default-management/internal/handler"    // 新增
	"xquant-default-management/internal/repository" // 新增
	"xquant-default-management/internal/service"    // 新增

	"github.com/gin-gonic/gin"
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

	// 3. 依赖注入：将所有组件连接起来
	// Repository -> Service -> Handler
	// 将组件一步步链接
	userRepository := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepository)
	userHandler := handler.NewUserHandler(userService)

	// 4. 初始化 Gin 引擎
	router := gin.Default()

	// 5. 注册路由
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	// 新增用户注册路由
	router.POST("/api/v1/register", userHandler.Register)

	// 6. 启动服务器
	log.Printf("Server is listening on port %s", cfg.ServerPort)
	if err := router.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
