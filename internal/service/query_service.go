package service

import (
	"xquant-default-management/internal/core"
	"xquant-default-management/internal/repository"
)

type QueryService interface {
	FindApplications(params repository.QueryParams) ([]core.DefaultApplication, int64, error)
}

type queryService struct {
	appRepo repository.ApplicationRepository
}

func NewQueryService(appRepo repository.ApplicationRepository) QueryService {
	return &queryService{appRepo: appRepo}
}

func (s *queryService) FindApplications(params repository.QueryParams) ([]core.DefaultApplication, int64, error) {
	// 可以在此层添加缓存等逻辑
	return s.appRepo.FindAll(params)
}
