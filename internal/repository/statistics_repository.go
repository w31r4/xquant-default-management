package repository

import (
	"fmt"

	"gorm.io/gorm"
)

// StatResult 用于存储按维度统计的聚合结果
type StatResult struct {
	Dimension string `json:"dimension"` // 可以是行业，也可以是区域
	Count     int64  `json:"count"`
}

// StatisticsRepository 定义了统计查询的接口
type StatisticsRepository interface {
	// 新方法：按维度、年份和状态进行统计
	GetCountsByDimension(year int, dimension string, status string) ([]StatResult, error)
}

type statisticsRepository struct {
	db *gorm.DB
}

func NewStatisticsRepository(db *gorm.DB) StatisticsRepository {
	return &statisticsRepository{db: db}
}

// GetCountsByDimension 是一个通用的聚合查询函数
func (r *statisticsRepository) GetCountsByDimension(year int, dimension string, status string) ([]StatResult, error) {
	var results []StatResult

	// 基础查询，从申请表开始
	query := r.db.Table("default_applications as da").
		Joins("join customers as c on c.id = da.customer_id")

	// 1. 动态选择维度和分组依据
	// dimension 参数必须是 'industry' 或 'region'，由 Service 层保证，防止 SQL 注入
	selectClause := fmt.Sprintf("c.%s as dimension, count(da.id) as count", dimension)
	query = query.Select(selectClause).Group("c." + dimension)

	// 2. 动态选择时间和状态过滤条件
	switch status {
	case "Approved":
		query = query.Where("da.status = ? AND EXTRACT(YEAR FROM da.approval_time) = ?", "Approved", year)
	case "Reborn":
		query = query.Where("da.status = ? AND EXTRACT(YEAR FROM da.rebirth_approval_time) = ?", "Reborn", year)
	default:
		// 如果状态无效，返回空结果，避免查询全表
		return results, nil
	}

	// 3. 执行查询并将结果扫描到结构体中
	err := query.Order("count desc").Scan(&results).Error
	return results, err
}
