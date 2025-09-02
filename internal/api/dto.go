package api

// RegisterRequest 代表用户注册的请求体
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=4"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role" binding:"required,oneof=Applicant Approver"`
}

// UserResponse 代表返回给客户端的用户信息 (隐藏了密码)
type UserResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

// LoginRequest 代表用户登录的请求体
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 代表登录成功后返回的响应
type LoginResponse struct {
	Token string `json:"token"`
}
