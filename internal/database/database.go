package database

// package database

import (
	"fmt"
	"log"
	"xquant-default-management/internal/config"
	"xquant-default-management/internal/core"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// Connect 初始化数据库连接
func Connect(cfg config.Config) {
	var err error
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort)

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Database connection established")
	// 自动迁移模型
	err = DB.AutoMigrate(&core.User{}, &core.Customer{}, &core.DefaultApplication{}) // 新增
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	log.Println("Database migrated")
}
