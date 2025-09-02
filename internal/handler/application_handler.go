package handler

import (
	"net/http"
	"xquant-default-management/internal/api"
	"xquant-default-management/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ApplicationHandler struct {
	appService service.ApplicationService
}

func NewApplicationHandler(appService service.ApplicationService) *ApplicationHandler {
	return &ApplicationHandler{appService: appService}
}

func (h *ApplicationHandler) CreateApplication(c *gin.Context) {
	var req api.CreateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 从认证中间件设置的上下文中获取申请人 ID
	applicantIDVal, _ := c.Get("userID")
	applicantID, ok := applicantIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID in context"})
		return
	}

	app, err := h.appService.CreateApplication(req.CustomerName, req.Severity, req.Reason, req.Remarks, applicantID)
	if err != nil {
		// 根据错误类型返回不同状态码
		switch err.Error() {
		case "customer not found":
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case "customer is already in default status", "there is already a pending application for this customer":
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create application"})
		}
		return
	}

	// 将核心模型映射到响应 DTO (此处为简化，实际项目中可能有专用 mapper)
	res := api.ApplicationResponse{
		ID:              app.ID.String(),
		CustomerName:    app.Customer.Name, // GORM 不会自动填充这个，我们需要在 service/repo 中 Preload
		Status:          app.Status,
		Severity:        app.Severity,
		ApplicantID:     app.ApplicantID.String(),
		ApplicationTime: app.ApplicationTime,
	}
	c.JSON(http.StatusCreated, res)
}
