// package database 封装了所有与数据库初始化和连接相关的逻辑。
// 这个包的目标是提供一个集中的地方来管理数据库连接和 GORM 的自动迁移功能，
// 使得应用启动时能够可靠地准备好数据层。
package database

import (
	"fmt"
	"log"
	"xquant-default-management/internal/config"
	"xquant-default-management/internal/core"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ==========================================================================================
// DB 是一个全局的数据库连接池实例。也就是数据库链接的抽象
// 它被声明为包级别的变量，并在 Connect 函数中被初始化。
// 因为它是导出的（首字母大写），所以项目的其他包（主要是 repository 包）可以直接访问这个实例来执行数据库操作。
// 在小型到中型应用中，这是一种简单方便的管理方式。
// ==========================================================================================
var DB *gorm.DB

// ==========================================================================================
// Connect 函数负责初始化到 PostgreSQL 数据库的连接，并执行自动化的数据模型迁移。
// 它在 main 函数中被调用，是应用启动流程的关键一步。
// ==========================================================================================
func Connect(cfg config.Config) {
	var err error

	// 1. 构建 DSN (Data Source Name) 字符串。
	// DSN 是一个标准格式的字符串，包含了连接数据库所需的所有信息。
	// 这里我们从配置对象 cfg 中动态读取主机、用户、密码等信息，实现了配置的外部化。
	// sslmode=disable 在本地开发中是常见的设置，表示不使用加密连接。
	// TimeZone=Asia/Shanghai 设置了连接的时区，确保时间数据的正确性。
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort)

	// 2. 使用 GORM 打开数据库连接。
	// gorm.Open 接收一个数据库驱动（这里是 postgres.Open(dsn)）和 GORM 的配置。
	// 如果连接失败，err 将不为 nil。
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		// 如果数据库连接失败，这是一个致命错误，整个应用无法启动。
		// log.Fatalf 会打印错误信息并立即退出程序。
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Database connection established")

	// 3. 自动迁移数据模型。
	// DB.AutoMigrate 是 GORM 的一个强大功能。它会检查数据库中是否存在与模型（User, Customer, DefaultApplication）
	// 对应的表。如果表不存在，它会自动创建。如果表存在但缺少字段，它会自动添加新字段。
	// 注意：AutoMigrate 不会删除不再需要的字段或修改字段类型，以防数据丢失。
	// 这对于开发阶段快速迭代模型非常方便。
	err = DB.AutoMigrate(&core.User{}, &core.Customer{}, &core.DefaultApplication{})
	if err != nil {
		// 如果迁移失败，同样是致命错误。
		log.Fatalf("Failed to migrate database: %v", err)
	}
	log.Println("Database migrated")
}
