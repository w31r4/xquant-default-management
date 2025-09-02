package core

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

// BaseModel 包含所有模型共有的字段，如文档 3.1 节所定义
type BaseModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;"`
	CreatedAt time.Time `gorm:"index"`
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// BeforeCreate 是一个 GORM 钩子，在创建记录前自动生成 UUID
func (base *BaseModel) BeforeCreate(tx *gorm.DB) (err error) {
	base.ID = uuid.New()
	return
}

// User 系统用户模型
type User struct {
	BaseModel
	Username string `gorm:"size:100;not null;uniqueIndex"`
	Password string `gorm:"size:255;not null"`
	Role     string `gorm:"size:50;not null;index"` // e.g., 'Applicant', 'Approver'
}
