package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
	"xquant-default-management/internal/service"

	"github.com/gin-gonic/gin"
)

// StatisticsHandler 封装了所有与统计相关的 HTTP 处理器
type StatisticsHandler struct {
	statsService service.StatisticsService
}

// NewStatisticsHandler 创建一个新的 StatisticsHandler 实例
func NewStatisticsHandler(statsService service.StatisticsService) *StatisticsHandler {
	return &StatisticsHandler{statsService: statsService}
}

// getStatistics 是一个私有的、通用的处理函数，用于处理所有统计请求。
// 它负责解析和验证通用的查询参数（如 'year'），然后调用相应的服务方法。
func (h *StatisticsHandler) getStatistics(c *gin.Context, dimension, status string) {
	// 1. 从查询参数中获取 'year'
	yearStr := c.Query("year")
	if yearStr == "" {
		// 如果 'year' 参数缺失，返回 400 Bad Request
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'year' is required"})
		return
	}

	// 2. 将 'year' 字符串转换为整数，并进行严格的范围验证
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		// 如果 'year' 不是一个有效的数字，返回 400 Bad Request
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'year' must be a valid number"})
		return
	}

	// 进行业务逻辑上的范围检查，防止无效的年份查询
	currentYear := time.Now().Year()
	if year < 2000 || year > currentYear {
		errorMsg := fmt.Sprintf("Invalid 'year' value. Please provide a year between 2000 and %d.", currentYear)
		c.JSON(http.StatusBadRequest, gin.H{"error": errorMsg})
		return
	}
	// 2. 新增：解析 'include_historical' 参数
	// c.Query() 返回字符串，我们将其转换为布尔值。
	// 如果参数不存在或值为 "false"、"0" 等，strconv.ParseBool 会返回 false。
	includeHistorical, _ := strconv.ParseBool(c.Query("include_historical"))
	// 3. 调用 Service 层获取经过计算的统计数据
	// 3. 根据参数选择调用哪个 Service 方法
	var stats interface{} // 使用 interface{} 来接收不同方法返回的相同 DTO 类型
	if includeHistorical {
		stats, err = h.statsService.GetStatisticsByDimensionIncludeHistorical(year, dimension, status)
	} else {
		stats, err = h.statsService.GetStatisticsByDimension(year, dimension, status)
	}
	if err != nil {
		// 如果 Service 层返回错误，这通常是服务器内部问题（例如数据库连接失败），
		// 记录日志（在实际项目中）并返回 500 Internal Server Error。
		// 注意：不要将底层的数据库错误直接暴露给客户端。
		c.JSON(http.StatusInternalServerError, gin.H{"error": "An internal error occurred while retrieving statistics"})
		return
	}

	// 4. 成功获取数据，返回 200 OK 和统计结果
	c.JSON(http.StatusOK, stats)
}

// --- 按行业 (Industry) 统计 ---

// GetDefaultsByIndustry 处理 GET /.../defaults/by-industry 的请求
// 统计指定年份按行业划分的【违约】数量、占比和趋势
func (h *StatisticsHandler) GetDefaultsByIndustry(c *gin.Context) {
	h.getStatistics(c, "industry", "Approved")
}

// GetRebirthsByIndustry 处理 GET /.../rebirths/by-industry 的请求
// 统计指定年份按行业划分的【重生】数量、占比和趋势
func (h *StatisticsHandler) GetRebirthsByIndustry(c *gin.Context) {
	h.getStatistics(c, "industry", "Reborn")
}

// --- 按区域 (Region) 统计 ---

// GetDefaultsByRegion 处理 GET /.../defaults/by-region 的请求
// 统计指定年份按区域划分的【违约】数量、占比和趋势
func (h *StatisticsHandler) GetDefaultsByRegion(c *gin.Context) {
	h.getStatistics(c, "region", "Approved")
}

// GetRebirthsByRegion 处理 GET /.../rebirths/by-region 的请求
// 统计指定年份按区域划分的【重生】数量、占比和趋势
func (h *StatisticsHandler) GetRebirthsByRegion(c *gin.Context) {
	h.getStatistics(c, "region", "Reborn")
}
