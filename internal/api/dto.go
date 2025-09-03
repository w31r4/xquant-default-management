package api

import "time"

// RegisterRequest 代表用户注册时客户端需要发送的请求体。
// binding 标签用于 Gin 框架进行输入验证。
type RegisterRequest struct {
	// Username 是用户注册时期望的用户名。
	// 验证规则：必填 (required)，最小长度为 4 个字符 (min=4)。
	Username string `json:"username" binding:"required,min=4"`

	// Password 是用户注册时设置的密码。
	// 验证规则：必填 (required)，最小长度为 6 个字符 (min=6)。
	Password string `json:"password" binding:"required,min=6"`

	// Role 是用户注册时指定的角色。
	// 验证规则：必填 (required)，且值必须是 "Applicant" 或 "Approver" 两者之一 (oneof=Applicant Approver)。
	Role string `json:"role" binding:"required,oneof=Applicant Approver"`
}

// UserResponse 代表成功创建用户或获取用户信息后，返回给客户端的数据结构。
// 这个结构体特意隐藏了密码等敏感字段，确保不会泄露给前端。
type UserResponse struct {
	// ID 是用户的唯一标识符 (UUID 格式的字符串)。
	ID string `json:"id"`
	// Username 是用户的名称。
	Username string `json:"username"`
	// Role 是用户的角色。
	Role string `json:"role"`
}

// LoginRequest 代表用户登录时客户端需要发送的请求体。
type LoginRequest struct {
	// Username 是用于登录的用户名。
	// 验证规则：必填 (required)。
	Username string `json:"username" binding:"required"`
	// Password 是用于登录的密码。
	// 验证规则：必填 (required)。
	Password string `json:"password" binding:"required"`
}

// LoginResponse 代表用户成功登录后，服务器返回的响应。
type LoginResponse struct {
	// Token 是一个 JWT (JSON Web Token)，客户端后续需要用它来进行身份认证。
	Token string `json:"token"`
}

// CreateApplicationRequest 代表创建违约申请时客户端需要发送的请求体。
type CreateApplicationRequest struct {
	// CustomerName 是申请所针对的客户名称。
	// 验证规则：必填 (required)。
	CustomerName string `json:"customer_name" binding:"required"`

	// Severity 是违约事件的严重等级。
	// 验证规则：必填 (required)，且值必须是 "High", "Medium", "Low" 三者之一。
	Severity string `json:"severity" binding:"required,oneof=High Medium Low"`

	// Reason 是提交违约申请的主要原因。
	// 验证规则：必填 (required)。
	Reason string `json:"reason" binding:"required"`

	// Remarks 是申请的附加备注信息 (可选字段)。
	Remarks string `json:"remarks"`
}

// ApplicationResponse 代表返回给客户端的单个违约申请的详细信息。
type ApplicationResponse struct {
	// ID 是违约申请的唯一标识符。
	ID string `json:"id"`
	// CustomerName 是该申请关联的客户名称。
	CustomerName string `json:"customer_name"`
	// Status 是申请的当前状态 (例如："Pending")。
	Status string `json:"status"`
	// Severity 是违约事件的严重等级。
	Severity string `json:"severity"`
	// ApplicantID 是提交此申请的用户的 ID。
	ApplicantID string `json:"applicant_id"`
	// ApplicationTime 是申请被提交的时间。
	ApplicationTime time.Time `json:"application_time"`
}

// ApproveRequest 代表审核操作的请求体
type ApproveRequest struct {
	ApplicationID string `json:"application_id" binding:"required,uuid"`
}
