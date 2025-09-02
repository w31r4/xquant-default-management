package core

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
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

// Customer 客户信息
type Customer struct {
	BaseModel
	Name      string `gorm:"size:255;not null;uniqueIndex"`
	Industry  string `gorm:"size:100;index"` // 行业
	Region    string `gorm:"size:100;index"` // 区域
	IsDefault bool   `gorm:"default:false;index"`
}

// DefaultApplication 违约认定申请
type DefaultApplication struct {
	BaseModel
	CustomerID uuid.UUID `gorm:"type:uuid;not null;index"`
	Customer   Customer  `gorm:"foreignKey:CustomerID"`

	Status        string `gorm:"size:50;not null;index;default:'Pending'"` // Pending, Approved, Rejected
	Severity      string `gorm:"size:50;not null"`                         // High, Medium, Low
	DefaultReason string `gorm:"type:text;not null"`
	Remarks       string `gorm:"type:text"`

	ApplicantID uuid.UUID `gorm:"type:uuid;not null"`
	Applicant   User      `gorm:"foreignKey:ApplicantID"`

	ApplicationTime time.Time `gorm:"not null"`
}
