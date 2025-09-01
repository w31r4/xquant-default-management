package service

import (
	"errors"
	"xquant-default-management/internal/core"
	"xquant-default-management/internal/repository"
	"xquant-default-management/internal/utils"

	"gorm.io/gorm"
)

// UserService 定义了用户相关的业务逻辑接口
type UserService interface {
	Register(username, password, role string) (*core.User, error)
}

type userService struct {
	userRepo repository.UserRepository
}

// NewUserService 创建一个新的 UserService 实例
func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

func (s *userService) Register(username, password, role string) (*core.User, error) {
	// 1. 检查用户名是否已存在
	_, err := s.userRepo.GetByUsername(username)
	if err == nil {
		return nil, errors.New("username already exists")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err // 其他数据库错误
	}

	// 2. 哈希密码
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	// 3. 创建用户实体
	user := &core.User{
		Username: username,
		Password: hashedPassword,
		Role:     role,
	}

	// 4. 保存到数据库
	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}
