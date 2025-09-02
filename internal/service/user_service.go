package service

import (
	"errors"
	"xquant-default-management/internal/config"
	"xquant-default-management/internal/core"
	"xquant-default-management/internal/repository"
	"xquant-default-management/internal/utils"

	"gorm.io/gorm"
)

// UserService 定义了用户相关的业务逻辑接口
type UserService interface {
	Register(username, password, role string) (*core.User, error)
	Login(username, password string) (string, error) // 新增

}

type userService struct {
	userRepo repository.UserRepository
	cfg      config.Config // 新增，用于访问 JWT Secret 和 TTL
}

// NewUserService 修改构造函数以接收配置
func NewUserService(userRepo repository.UserRepository, cfg config.Config) UserService {
	return &userService{userRepo: userRepo, cfg: cfg}
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

// Login 验证用户凭据并返回 JWT
func (s *userService) Login(username, password string) (string, error) {
	// 1. 根据用户名查找用户
	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("invalid username or password")
		}
		return "", err
	}

	// 2. 检查密码是否匹配
	if !utils.CheckPasswordHash(password, user.Password) {
		return "", errors.New("invalid username or password")
	}

	// 3. 生成 JWT
	token, err := utils.GenerateToken(user.ID, user.Role, s.cfg)
	if err != nil {
		return "", err
	}

	return token, nil
}
