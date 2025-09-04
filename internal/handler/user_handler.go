package handler

import (
	"net/http"
	"xquant-default-management/internal/api"
	"xquant-default-management/internal/service"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// Register godoc
// @Summary      User registration
// @Description  Register a new user with username, password and role
// @Tags         User
// @Accept       json
// @Produce      json
// @Param        user  body      api.RegisterRequest  true  "User registration info"
// @Success      201   {object}  api.UserResponse
// @Failure      400   {object}  api.ErrorResponse
// @Failure      409   {object}  api.ErrorResponse
// @Failure      500   {object}  api.ErrorResponse
// @Router       /register [post]
func (h *UserHandler) Register(c *gin.Context) {
	var req api.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.Register(req.Username, req.Password, req.Role)
	if err != nil {
		// 这里可以根据 service 返回的错误类型，返回更具体的 HTTP 状态码
		if err.Error() == "username already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}

	// 返回创建成功的用户信息，注意不要泄露密码
	res := api.UserResponse{
		ID:       user.ID.String(),
		Username: user.Username,
		Role:     user.Role,
	}
	c.JSON(http.StatusCreated, res)
}

// Login godoc
// @Summary      User login
// @Description  Login with username and password to get a token
// @Tags         User
// @Accept       json
// @Produce      json
// @Param        user  body      api.LoginRequest  true  "User login info"
// @Success      200   {object}  api.LoginResponse
// @Failure      400   {object}  api.ErrorResponse
// @Failure      401   {object}  api.ErrorResponse
// @Router       /login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req api.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.userService.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, api.LoginResponse{Token: token})
}
