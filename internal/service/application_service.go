package service

import (
	"errors"
	"time"
	"xquant-default-management/internal/core"
	"xquant-default-management/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ApplicationService interface {
	CreateApplication(customerName, severity, reason, remarks string, applicantID uuid.UUID) (*core.DefaultApplication, error)
}

type applicationService struct {
	appRepo      repository.ApplicationRepository
	customerRepo repository.CustomerRepository
}

func NewApplicationService(appRepo repository.ApplicationRepository, customerRepo repository.CustomerRepository) ApplicationService {
	return &applicationService{appRepo: appRepo, customerRepo: customerRepo}
}

func (s *applicationService) CreateApplication(customerName, severity, reason, remarks string, applicantID uuid.UUID) (*core.DefaultApplication, error) {
	// 1. 查找客户
	customer, err := s.customerRepo.GetByName(customerName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("customer not found")
		}
		return nil, err
	}

	// 2. 检查客户是否已经是违约状态
	if customer.IsDefault {
		return nil, errors.New("customer is already in default status")
	}

	// 3. 检查是否已有待处理的申请
	existingApp, err := s.appRepo.FindPendingByCustomerID(customer.ID)
	if err != nil {
		return nil, err
	}
	if existingApp != nil {
		return nil, errors.New("there is already a pending application for this customer")
	}

	// 4. 创建新的申请实体
	app := &core.DefaultApplication{
		CustomerID:      customer.ID,
		Status:          "Pending",
		Severity:        severity,
		DefaultReason:   reason,
		Remarks:         remarks,
		ApplicantID:     applicantID,
		ApplicationTime: time.Now(),
	}

	if err := s.appRepo.Create(app); err != nil {
		return nil, err
	}

	return app, nil
}
