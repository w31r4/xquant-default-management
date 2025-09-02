package repository

import (
	"xquant-default-management/internal/core"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ApplicationRepository interface {
	Create(app *core.DefaultApplication) error
	FindPendingByCustomerID(customerID uuid.UUID) (*core.DefaultApplication, error)
}

type applicationRepository struct {
	db *gorm.DB
}

func NewApplicationRepository(db *gorm.DB) ApplicationRepository {
	return &applicationRepository{db: db}
}

func (r *applicationRepository) Create(app *core.DefaultApplication) error {
	return r.db.Create(app).Error
}

func (r *applicationRepository) FindPendingByCustomerID(customerID uuid.UUID) (*core.DefaultApplication, error) {
	var app core.DefaultApplication
	err := r.db.Where("customer_id = ? AND status = ?", customerID, "Pending").First(&app).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil // 找到不到不是一个应用错误，返回 nil, nil
	}
	return &app, err
}
