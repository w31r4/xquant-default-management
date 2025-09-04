package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"xquant-default-management/internal/api"
	"xquant-default-management/internal/core"
	"xquant-default-management/internal/mocks"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	return router
}

func TestUserHandler_Register(t *testing.T) {
	mockUserService := new(mocks.UserService)
	userHandler := NewUserHandler(mockUserService)

	t.Run("success", func(t *testing.T) {
		router := setupRouter()
		router.POST("/register", userHandler.Register)

		reqBody := api.RegisterRequest{Username: "test", Password: "password", Role: "Applicant"}
		user := &core.User{BaseModel: core.BaseModel{ID: uuid.New()}, Username: reqBody.Username, Role: reqBody.Role}

		mockUserService.On("Register", reqBody.Username, reqBody.Password, reqBody.Role).Return(user, nil).Once()

		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockUserService.AssertExpectations(t)
	})

	t.Run("invalid request body", func(t *testing.T) {
		router := setupRouter()
		router.POST("/register", userHandler.Register)

		req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer([]byte(`{"username": "test"`))) // Malformed JSON
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("username conflict", func(t *testing.T) {
		router := setupRouter()
		router.POST("/register", userHandler.Register)

		reqBody := api.RegisterRequest{Username: "existing", Password: "password", Role: "Applicant"}
		mockUserService.On("Register", reqBody.Username, reqBody.Password, reqBody.Role).Return(nil, errors.New("username already exists")).Once()

		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		mockUserService.AssertExpectations(t)
	})

	t.Run("internal server error", func(t *testing.T) {
		router := setupRouter()
		router.POST("/register", userHandler.Register)

		reqBody := api.RegisterRequest{Username: "test", Password: "password", Role: "Applicant"}
		mockUserService.On("Register", reqBody.Username, reqBody.Password, reqBody.Role).Return(nil, errors.New("some db error")).Once()

		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockUserService.AssertExpectations(t)
	})
}

func TestUserHandler_Login(t *testing.T) {
	mockUserService := new(mocks.UserService)
	userHandler := NewUserHandler(mockUserService)

	t.Run("success", func(t *testing.T) {
		router := setupRouter()
		router.POST("/login", userHandler.Login)

		reqBody := api.LoginRequest{Username: "test", Password: "password"}
		mockUserService.On("Login", reqBody.Username, reqBody.Password).Return("some-jwt-token", nil).Once()

		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var res api.LoginResponse
		json.Unmarshal(w.Body.Bytes(), &res)
		assert.Equal(t, "some-jwt-token", res.Token)
		mockUserService.AssertExpectations(t)
	})

	t.Run("invalid credentials", func(t *testing.T) {
		router := setupRouter()
		router.POST("/login", userHandler.Login)

		reqBody := api.LoginRequest{Username: "test", Password: "wrongpassword"}
		mockUserService.On("Login", reqBody.Username, reqBody.Password).Return("", errors.New("invalid username or password")).Once()

		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockUserService.AssertExpectations(t)
	})
}
