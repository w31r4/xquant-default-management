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
	Name           string `gorm:"size:255;not null;uniqueIndex"`
	Industry       string `gorm:"size:100;index"` // 行业
	Region         string `gorm:"size:100;index"` // 区域
	IsDefault      bool   `gorm:"default:false;index"`
	LatestExtGrade string `gorm:"size:50"` // 新增：最新外部等级

}

// DefaultApplication 代表一条客户违约认定的申请记录。
// 它是系统中的核心业务实体，记录了从申请提交到审批的全过程。
type DefaultApplication struct {
	// BaseModel 嵌入基础模型 (ID, CreatedAt, UpdatedAt, DeletedAt)。
	BaseModel

	// CustomerID 关联的客户 ID，明确指出此申请是针对哪个客户。
	CustomerID uuid.UUID `gorm:"type:uuid;not null;index"`
	// Customer 关联的客户实体 (用于 GORM 预加载客户的详细信息)。
	Customer Customer `gorm:"foreignKey:CustomerID"`

	// Status 申请的当前状态。
	// 可选值：Pending (待处理), Approved (已批准), Rejected (已拒绝)。
	Status string `gorm:"size:50;not null;index;default:'Pending'"`

	// Severity 违约事件的严重等级，用于风险评估。
	// 可选值：High (高), Medium (中), Low (低)。
	Severity string `gorm:"size:50;not null"`

	// DefaultReason 提交违约认定的主要原因，是审批的重要依据。
	DefaultReason   string `gorm:"type:text;not null"`
	RejectionReason string `gorm:"type:text"` // 新增：用于存储拒绝原因
	RebirthReason   string `gorm:"type:text"` // 新增：重生原因

	// Remarks 申请人填写的额外备注信息 (可选)。
	Remarks string `gorm:"type:text"`

	// ApplicantID 提交此申请的用户的 ID (即申请人)。
	ApplicantID uuid.UUID `gorm:"type:uuid;not null"`
	// Applicant 关联的申请人实体 (用于 GORM 预加载申请人的详细信息)。
	Applicant User `gorm:"foreignKey:ApplicantID"`
	// --- 新增字段 ---
	ApproverID   *uuid.UUID `gorm:"type:uuid"`             // 审核人 ID, 使用指针类型以允许 NULL 值
	Approver     *User      `gorm:"foreignKey:ApproverID"` // 关联到 User 模型
	ApprovalTime *time.Time // 审核时间，使用指针类型以允许 NULL 值

	// 新增：重生审批相关字段
	RebirthApproverID   *uuid.UUID `gorm:"type:uuid"`
	RebirthApprover     *User      `gorm:"foreignKey:RebirthApproverID"`
	RebirthApprovalTime *time.Time

	// ApplicationTime 申请被正式提交的时间戳。
	ApplicationTime time.Time `gorm:"not null"`
}
