package handler

import (
	"net/http"
	"xquant-default-management/internal/api"
	"xquant-default-management/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ApplicationHandler 封装了与违约申请相关的 HTTP 请求处理器。
// 它依赖于 ApplicationService 来执行核心的业务逻辑。
type ApplicationHandler struct {
	appService service.ApplicationService
}

// NewApplicationHandler 是 ApplicationHandler 的构造函数。
// 通过依赖注入的方式，将 ApplicationService 传入。
func NewApplicationHandler(appService service.ApplicationService) *ApplicationHandler {
	return &ApplicationHandler{appService: appService}
}

// CreateApplication 处理创建新违约申请的 HTTP POST 请求。
// 它负责解析请求、调用业务逻辑、处理错误并返回标准化的响应。
func (h *ApplicationHandler) CreateApplication(c *gin.Context) {
	// 声明一个用于接收请求体的 DTO (Data Transfer Object) 变量。
	var req api.CreateApplicationRequest

	// 1. 解析和验证请求体
	// ShouldBindJSON 会将请求的 JSON Body 解析到 req 结构体中。
	// 如果 JSON 格式错误或不符合 DTO 中定义的 `binding` 验证规则，将返回错误。
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 2. 从 Gin 上下文中获取用户信息
	// 此处的 "userID" 是由之前的 AuthMiddleware (认证中间件) 在验证 JWT 成功后存入上下文的。
	// 这确保了我们知道是哪个已登录的用户在提交申请。
	applicantIDVal, _ := c.Get("userID")

	// 对获取到的值进行类型断言，确保它是一个有效的 uuid.UUID。
	applicantID, ok := applicantIDVal.(uuid.UUID)
	if !ok {
		// 如果类型断言失败，说明上下文中存储的用户 ID 有问题，这是一个服务端内部错误。
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID in context"})
		return
	}

	// 3. 调用核心业务逻辑
	// 将通过验证的请求数据和申请人 ID 传递给 Service 层进行处理。
	// 所有的业务规则（如检查客户是否存在、是否已有待处理申请等）都在 Service 层中执行。
	app, err := h.appService.CreateApplication(req.CustomerName, req.Severity, req.Reason, req.Remarks, applicantID)
	if err != nil {
		// 4. 精细化错误处理
		// 根据 Service 层返回的不同错误类型，映射到不同的 HTTP 状态码，为前端提供更明确的反馈。
		switch err.Error() {
		case "customer not found":
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()}) // 404 Not Found
		case "customer is already in default status", "there is already a pending application for this customer":
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()}) // 409 Conflict，表示请求与当前服务器状态冲突
		default:
			// 对于其他未知错误，返回通用的 500 服务器内部错误。
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create application"})
		}
		return
	}

	// 5. 格式化并返回成功响应
	// 将 Service 层返回的核心业务模型 (core.DefaultApplication) 映射到用于响应的 DTO (api.ApplicationResponse)。
	// 这样做可以避免直接暴露数据库模型结构，并能灵活控制返回给客户端的字段。
	res := api.ApplicationResponse{
		ID: app.ID.String(),
		// 注意：GORM 在 Create 后默认不会自动填充关联的 Customer 实体。
		// CustomerName 可能为空。如果需要显示，我们应在 Service/Repository 层使用 Preload 预加载。
		CustomerName:    app.Customer.Name,
		Status:          app.Status,
		Severity:        app.Severity,
		ApplicantName:   app.Applicant.Username, // 同上
		ApplicationTime: app.ApplicationTime,
	}

	// 返回 201 Created 状态码，表示资源创建成功，并在响应体中包含新创建的申请信息。
	c.JSON(http.StatusCreated, res)
}

// ApproveApplication 处理批准申请的请求
func (h *ApplicationHandler) ApproveApplication(c *gin.Context) {
	var req api.ApproveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 从请求体解析 ApplicationID
	appID, err := uuid.Parse(req.ApplicationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid application ID format"})
		return
	}

	// 从认证中间件的上下文中获取审核人 ID
	approverIDVal, _ := c.Get("userID")
	approverID, ok := approverIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID in context"})
		return
	}

	err = h.appService.ApproveApplication(appID, approverID)
	if err != nil {
		switch err.Error() {
		case "application not found":
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case "application is not in pending state":
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve application"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Application approved successfully"})
}

// RejectApplication 处理拒绝申请的请求
func (h *ApplicationHandler) RejectApplication(c *gin.Context) {
	var req api.RejectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	appID, _ := uuid.Parse(req.ApplicationID)
	approverIDVal, _ := c.Get("userID")
	approverID := approverIDVal.(uuid.UUID)

	err := h.appService.RejectApplication(appID, approverID, req.RejectionReason)
	if err != nil {
		// 错误处理逻辑与 Approve 类似
		switch err.Error() {
		case "application not found":
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case "application is not in pending state":
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject application"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Application rejected successfully"})
}

// GetPendingApplications 处理查询待审批列表的请求
func (h *ApplicationHandler) GetPendingApplications(c *gin.Context) {
	apps, err := h.appService.GetPendingApplications()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve pending applications"})
		return
	}

	// 将数据库模型列表映射到 API DTO 列表
	var res []api.ApplicationResponse
	for _, app := range apps {
		res = append(res, api.ApplicationResponse{
			ID:              app.ID.String(),
			CustomerName:    app.Customer.Name, // 因为 Service->Repo 预加载了，这里可以直接用
			Status:          app.Status,
			Severity:        app.Severity,
			ApplicantName:   app.Applicant.Username, // 同上
			ApplicationTime: app.ApplicationTime,
		})
	}

	c.JSON(http.StatusOK, res)
}
