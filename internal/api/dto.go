package api

import "time"

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

// CreateApplicationRequest 代表创建违约申请的请求体
type CreateApplicationRequest struct {
	CustomerName string `json:"customer_name" binding:"required"`
	Severity     string `json:"severity" binding:"required,oneof=High Medium Low"`
	Reason       string `json:"reason" binding:"required"`
	Remarks      string `json:"remarks"`
}

// ApplicationResponse 代表返回给客户端的申请信息
type ApplicationResponse struct {
	ID              string    `json:"id"`
	CustomerName    string    `json:"customer_name"`
	Status          string    `json:"status"`
	Severity        string    `json:"severity"`
	ApplicantID     string    `json:"applicant_id"`
	ApplicationTime time.Time `json:"application_time"`
}
