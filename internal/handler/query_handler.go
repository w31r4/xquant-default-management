package handler

import (
	"net/http"
	"strconv"
	"xquant-default-management/internal/api"
	"xquant-default-management/internal/repository"
	"xquant-default-management/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type QueryHandler struct {
	queryService service.QueryService
}

func NewQueryHandler(queryService service.QueryService) *QueryHandler {
	return &QueryHandler{queryService: queryService}
}

// FindApplications godoc
// @Summary      Find applications
// @Description  Find applications with optional filters for customer name and status, with pagination support.
// @Tags         Applications
// @Produce      json
// @Param        customer_name  query     string  false  "Customer Name"
// @Param        status         query     string  false  "Status"  Enums(Pending, Approved, Rejected, Reborn)
// @Param        page           query     int     false  "Page number"      default(1)
// @Param        pageSize       query     int     false  "Page size"        default(10)
// @Success      200            {object}  api.PaginatedApplicationsResponse
// @Failure      500            {object}  api.ErrorResponse
// @Security     ApiKeyAuth
// @Router       /applications [get]
func (h *QueryHandler) FindApplications(c *gin.Context) {
	// 1. 解析查询参数 (Query Parameters)
	var params repository.QueryParams

	// 使用 c.Query() 获取字符串参数。如果参数不存在，它会返回空字符串 ""。
	// 我们在 Repository 层使用指针，就是为了区分 "用户没传这个参数" 和 "用户传了一个空字符串"。
	customerName := c.Query("customer_name")
	if customerName != "" {
		params.CustomerName = &customerName
	}
	status := c.Query("status")
	if status != "" {
		params.Status = &status
	}

	// 使用 c.DefaultQuery() 获取分页参数，如果用户没传，就使用默认值。
	// strconv.Atoi 用于将字符串转换为整数。在这里我们暂时忽略了可能的转换错误。
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	params.Page = page
	params.PageSize = pageSize

	// 2. 调用 Service 层执行查询
	apps, total, err := h.queryService.FindApplications(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query applications"})
		return
	}

	// 3. 将核心模型列表 (apps) 映射到响应 DTO 列表 (data)
	// 这是“海关”步骤，确保我们只暴露安全和必要的信息。
	var data []api.ApplicationDetailResponse
	for _, app := range apps {
		// --- 安全的指针处理 ---
		// 在访问指针字段之前，必须检查它是否为 nil。

		var approverName *string
		// 如果 app.Approver 不是 nil (即 ApproverID 存在且已 Preload 成功)
		if app.Approver != nil {
			approverName = &app.Approver.Username
		}

		// 注意：DTO 中的字段也应该是指针类型，才能正确地表示 null。
		// 我们需要修改一下 ApplicationDetailResponse DTO。

		// --- 构建单个 DTO 对象 ---
		detail := api.ApplicationDetailResponse{
			ID:              app.ID.String(),
			CustomerName:    app.Customer.Name,
			LatestExtGrade:  app.Customer.LatestExtGrade,
			Status:          app.Status,
			DefaultReason:   app.DefaultReason,
			Severity:        app.Severity,
			ApplicationTime: app.ApplicationTime,
			RebirthReason:   app.RebirthReason, // 修正：应该是 RebirthReason

			// 安全地赋值
			ApprovalTime: app.ApprovalTime,
			ApproverName: approverName,
		}

		// Applicant 也可能由于某些原因（如用户被删除）加载失败，做个保护是好习惯
		if app.Applicant.ID != uuid.Nil {
			detail.ApplicantName = app.Applicant.Username
		}

		data = append(data, detail)
	}

	// 4. 返回包含分页信息的最终响应
	c.JSON(http.StatusOK, api.PaginatedApplicationsResponse{
		Total: total,
		Page:  page,
		Data:  data,
	})
}
