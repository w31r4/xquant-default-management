package service

import (
	"errors"
	"testing"
	"xquant-default-management/internal/config"
	"xquant-default-management/internal/core"
	"xquant-default-management/internal/mocks"
	"xquant-default-management/internal/utils"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestUserService_Register(t *testing.T) {
	mockUserRepo := new(mocks.UserRepository)
	cfg := config.Config{} // Not needed for register
	userService := NewUserService(mockUserRepo, cfg)

	username := "newuser"
	password := "password123"
	role := "Applicant"

	t.Run("success", func(t *testing.T) {
		// Expect GetByUsername to be called and return ErrRecordNotFound
		mockUserRepo.On("GetByUsername", username).Return(nil, gorm.ErrRecordNotFound).Once()
		// Expect Create to be called with any user object
		mockUserRepo.On("Create", mock.AnythingOfType("*core.User")).Return(nil).Once()

		user, err := userService.Register(username, password, role)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, username, user.Username)
		assert.True(t, utils.CheckPasswordHash(password, user.Password))
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("username already exists", func(t *testing.T) {
		existingUser := &core.User{Username: username}
		// Expect GetByUsername to be called and return an existing user
		mockUserRepo.On("GetByUsername", username).Return(existingUser, nil).Once()

		_, err := userService.Register(username, password, role)

		assert.Error(t, err)
		assert.Equal(t, "username already exists", err.Error())
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("database error on get", func(t *testing.T) {
		dbErr := errors.New("db find error")
		mockUserRepo.On("GetByUsername", username).Return(nil, dbErr).Once()

		_, err := userService.Register(username, password, role)

		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("database error on create", func(t *testing.T) {
		dbErr := errors.New("db create error")
		mockUserRepo.On("GetByUsername", username).Return(nil, gorm.ErrRecordNotFound).Once()
		mockUserRepo.On("Create", mock.AnythingOfType("*core.User")).Return(dbErr).Once()

		_, err := userService.Register(username, password, role)

		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestUserService_Login(t *testing.T) {
	mockUserRepo := new(mocks.UserRepository)
	cfg := config.Config{JWTSecret: "test-secret", TokenTTL: 1}
	userService := NewUserService(mockUserRepo, cfg)

	username := "testuser"
	password := "password123"
	hashedPassword, _ := utils.HashPassword(password)
	user := &core.User{
		BaseModel: core.BaseModel{ID: uuid.New()},
		Username:  username,
		Password:  hashedPassword,
		Role:      "Approver",
	}

	t.Run("success", func(t *testing.T) {
		mockUserRepo.On("GetByUsername", username).Return(user, nil).Once()

		token, err := userService.Login(username, password)

		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		mockUserRepo.On("GetByUsername", username).Return(nil, gorm.ErrRecordNotFound).Once()

		_, err := userService.Login(username, password)

		assert.Error(t, err)
		assert.Equal(t, "invalid username or password", err.Error())
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("incorrect password", func(t *testing.T) {
		mockUserRepo.On("GetByUsername", username).Return(user, nil).Once()

		_, err := userService.Login(username, "wrongpassword")

		assert.Error(t, err)
		assert.Equal(t, "invalid username or password", err.Error())
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("database error", func(t *testing.T) {
		dbErr := errors.New("db login error")
		mockUserRepo.On("GetByUsername", username).Return(nil, dbErr).Once()

		_, err := userService.Login(username, password)

		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
		mockUserRepo.AssertExpectations(t)
	})
}
