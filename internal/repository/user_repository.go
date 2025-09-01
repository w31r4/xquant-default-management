package repository

import (
	"xquant-default-management/internal/core"

	"gorm.io/gorm"
)

// UserRepository 定义了用户数据操作的接口
type UserRepository interface {
	Create(user *core.User) error
	GetByUsername(username string) (*core.User, error)
}

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建一个新的 UserRepository 实例
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// Create 在数据库中创建一个新用户
func (r *userRepository) Create(user *core.User) error {
	return r.db.Create(user).Error
}

// GetByUsername 根据用户名查找用户
func (r *userRepository) GetByUsername(username string) (*core.User, error) {
	var user core.User
	err := r.db.Where("username = ?", username).First(&user).Error
	return &user, err
}
