package main

import (
	"log"
	"net/http"
	"xquant-default-management/internal/config"
	"xquant-default-management/internal/database"

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

	// 3. 初始化 Gin 引擎
	router := gin.Default()

	// 4. 创建一个简单的 "心跳" 路由
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	// 5. 启动服务器
	log.Printf("Server is listening on port %s", cfg.ServerPort)
	if err := router.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
