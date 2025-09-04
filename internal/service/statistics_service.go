package service

import (
	"math"
	"sort"
	"xquant-default-management/internal/api"
	"xquant-default-management/internal/repository"
)

// StatisticsService 定义了统计相关的业务逻辑接口
type StatisticsService interface {
	// 返回一个包含完整计算结果的 DTO 列表
	GetStatisticsByDimension(year int, dimension string, status string) ([]api.StatisticsResponse, error)
	GetStatisticsByDimensionIncludeHistorical(year int, dimension string, status string) ([]api.StatisticsResponse, error)
}

type statisticsService struct {
	statsRepo repository.StatisticsRepository
}

func NewStatisticsService(statsRepo repository.StatisticsRepository) StatisticsService {
	return &statisticsService{statsRepo: statsRepo}
}

// GetStatisticsByDimension 获取按维度统计的数据
func (s *statisticsService) GetStatisticsByDimension(year int, dimension string, status string) ([]api.StatisticsResponse, error) {
	return s.getStatisticsByDimensionWithOptions(year, dimension, status, false)
}

// GetStatisticsByDimensionIncludeHistorical 获取按维度统计的数据，包含历史维度
func (s *statisticsService) GetStatisticsByDimensionIncludeHistorical(year int, dimension string, status string) ([]api.StatisticsResponse, error) {
	return s.getStatisticsByDimensionWithOptions(year, dimension, status, true)
}

// getStatisticsByDimensionWithOptions 是核心计算函数
func (s *statisticsService) getStatisticsByDimensionWithOptions(year int, dimension string, status string, includeHistorical bool) ([]api.StatisticsResponse, error) {
	// 1. 获取当年的数据
	currentYearStats, err := s.statsRepo.GetCountsByDimension(year, dimension, status)
	if err != nil {
		return nil, err
	}

	// 2. 获取去年的数据，用于计算同比增长
	previousYearStats, err := s.statsRepo.GetCountsByDimension(year-1, dimension, status)
	if err != nil {
		return nil, err
	}

	// 3. 构建数据映射
	currentYearMap := make(map[string]int64)
	for _, stat := range currentYearStats {
		currentYearMap[stat.Dimension] = stat.Count
	}

	previousYearMap := make(map[string]int64)
	for _, stat := range previousYearStats {
		previousYearMap[stat.Dimension] = stat.Count
	}

	// 4. 确定要处理的维度集合
	var dimensionsToProcess []string
	if includeHistorical {
		// 包含所有历史维度
		allDimensions := make(map[string]bool)
		for _, stat := range currentYearStats {
			allDimensions[stat.Dimension] = true
		}
		for _, stat := range previousYearStats {
			allDimensions[stat.Dimension] = true
		}
		for dim := range allDimensions {
			dimensionsToProcess = append(dimensionsToProcess, dim)
		}
	} else {
		// 只处理当年有数据的维度
		for _, stat := range currentYearStats {
			dimensionsToProcess = append(dimensionsToProcess, stat.Dimension)
		}
	}

	// 5. 计算总数（只计算当年的数据）
	var totalCount int64
	for _, count := range currentYearMap {
		totalCount += count
	}

	// 6. 构建响应数据
	var response []api.StatisticsResponse
	for _, dim := range dimensionsToProcess {
		currentCount := currentYearMap[dim]   // 如果不存在，默认为 0
		previousCount := previousYearMap[dim] // 如果不存在，默认为 0

		// 计算占比（基于当年总数）
		var percentage float64
		if totalCount > 0 {
			percentage = float64(currentCount) / float64(totalCount)
		}

		// 计算同比增长率
		var growthRate *float64
		if previousCount > 0 {
			// 正常情况：去年有数据，计算增长率
			rate := (float64(currentCount) - float64(previousCount)) / float64(previousCount)
			rate = math.Round(rate*10000) / 10000
			growthRate = &rate
		} else if currentCount > 0 && previousCount == 0 {
			// 特殊情况：去年为 0，今年有数据，返回 currentCount * 100%
			// 例如：去年 0 个，今年 5 个，增长率为 5 * 100% = 500%
			rate := float64(currentCount)
			rate = math.Round(rate*10000) / 10000
			growthRate = &rate
		}
		// 如果 currentCount == 0 && previousCount == 0，growthRate 保持为 nil

		response = append(response, api.StatisticsResponse{
			Dimension:  dim,
			Count:      currentCount,
			Percentage: percentage,
			GrowthRate: growthRate,
		})
	}

	// 7. 按计数降序排列（可选）
	sort.Slice(response, func(i, j int) bool {
		return response[i].Count > response[j].Count
	})

	return response, nil
}

// // GetStatisticsByDimension 是核心计算函数
// func (s *statisticsService) GetStatisticsByDimension(year int, dimension string, status string) ([]api.StatisticsResponse, error) {
// 	// 1. 获取当年的数据
// 	currentYearStats, err := s.statsRepo.GetCountsByDimension(year, dimension, status)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// 2. 获取去年的数据，用于计算同比增长
// 	previousYearStats, err := s.statsRepo.GetCountsByDimension(year-1, dimension, status)
// 	if err != nil {
// 		return nil, err
// 	}
// 	// 将去年数据存入 map 以便快速查找
// 	previousYearMap := make(map[string]int64)
// 	for _, stat := range previousYearStats {
// 		previousYearMap[stat.Dimension] = stat.Count
// 	}

// 	// 3. 计算总数，用于计算占比
// 	var totalCount int64
// 	for _, stat := range currentYearStats {
// 		totalCount += stat.Count
// 	}

// 	// 4. 组合所有数据，进行最终计算
// 	var response []api.StatisticsResponse
// 	for _, stat := range currentYearStats {
// 		// 计算占比
// 		var percentage float64
// 		if totalCount > 0 {
// 			percentage = float64(stat.Count) / float64(totalCount)
// 		}

// 		// 计算同比增长率
// 		var growthRate *float64
// 		previousCount, ok := previousYearMap[stat.Dimension]
// 		if ok && previousCount > 0 {
// 			rate := (float64(stat.Count) - float64(previousCount)) / float64(previousCount)
// 			// 为了可读性，可以四舍五入到小数点后 4 位
// 			rate = math.Round(rate*10000) / 10000
// 			growthRate = &rate
// 		}

// 		response = append(response, api.StatisticsResponse{
// 			Dimension:  stat.Dimension,
// 			Count:      stat.Count,
// 			Percentage: percentage,
// 			GrowthRate: growthRate,
// 		})
// 	}

// 	return response, nil
// }
