// package repository 包含数据访问层 (Data Access Layer, DAL) 的所有代码。
// 它的职责是封装所有与数据库的直接交互（CRUD 操作），
// 为 Service 层提供清晰、简洁的数据操作接口，并隐藏底层的数据库实现细节（如 GORM）。
package repository

import (
	"xquant-default-management/internal/core"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ApplicationRepository 定义了与 DefaultApplication 模型相关的数据操作接口。
// Service 层将依赖此接口，而不是具体的实现，以实现解耦。
type ApplicationRepository interface {
	// Create 插入一个新的违约申请记录。
	Create(app *core.DefaultApplication) error

	// FindPendingByCustomerID 根据客户 ID 查找一个状态为 "Pending" 的申请。
	// 如果没有找到，它会返回 (nil, nil)，表示“未找到”是一个正常的业务场景，而非错误。
	FindPendingByCustomerID(customerID uuid.UUID) (*core.DefaultApplication, error)
	GetByID(id uuid.UUID) (*core.DefaultApplication, error) // 新增
	// Update(app *core.DefaultApplication, updates map[string]interface{}) error // 修改接口
	Update(app *core.DefaultApplication, fields ...string) error
	FindAllByStatus(status string) ([]core.DefaultApplication, error) // 新增

}

// applicationRepository 是 ApplicationRepository 接口的具体实现。
// 它内部持有 *gorm.DB 数据库连接实例。
type applicationRepository struct {
	db *gorm.DB
}

// NewApplicationRepository 是 applicationRepository 的构造函数。
// 通过依赖注入的方式传入数据库连接。
func NewApplicationRepository(db *gorm.DB) ApplicationRepository {
	return &applicationRepository{db: db}
}

// Create 将一个新的违约申请记录插入到数据库中。
// 它直接使用 GORM 的 Create 方法，并返回可能发生的任何数据库错误。
func (r *applicationRepository) Create(app *core.DefaultApplication) error {
	return r.db.Create(app).Error
}

// FindPendingByCustomerID 在数据库中查找特定客户的、状态为 "Pending" 的违约申请。
func (r *applicationRepository) FindPendingByCustomerID(customerID uuid.UUID) (*core.DefaultApplication, error) {
	var app core.DefaultApplication

	// 使用 GORM 构建查询，条件为 customer_id 匹配且 status 为 "Pending"。
	// First() 方法会查找第一条匹配的记录。
	err := r.db.Where("customer_id = ? AND status = ?", customerID, "Pending").First(&app).Error

	// 关键的错误处理逻辑：
	// 在业务上，“找不到一个待处理的申请”是一个非常正常的、预期内的结果，而不是一个需要上报的“系统错误”。
	// 因此，当 GORM 返回 ErrRecordNotFound 时，我们在这里将其“消化”掉，
	// 返回 (nil, nil) 来清晰地告诉上层 Service：“没有找到，并且查询过程没有出错”。
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	// 如果 err 不是 gorm.ErrRecordNotFound，那它可能是一个真实的数据库连接错误或查询语法错误，
	// 这种情况下，我们需要将错误和找到的记录（可能不完整）一起返回给上层处理。
	return &app, err
}

// GetByID 根据 ID 获取一个申请单，并预加载关联的客户信息
func (r *applicationRepository) GetByID(id uuid.UUID) (*core.DefaultApplication, error) {
	var app core.DefaultApplication
	// Preload("Customer") 会自动执行一次额外的查询来填充 Customer 字段
	err := r.db.Preload("Customer").First(&app, id).Error
	return &app, err
}

// Update 方法现在只更新传入的 map 中指定的字段
// func (r *applicationRepository) Update(app *core.DefaultApplication, updates map[string]interface{}) error {
// 	return r.db.Model(app).Updates(updates).Error
// }

// Update 只更新指定的字段
func (r *applicationRepository) Update(app *core.DefaultApplication, fields ...string) error {
	return r.db.Model(app).Select(fields).Updates(app).Error
}

// FindAllByStatus 根据状态查找所有申请单
func (r *applicationRepository) FindAllByStatus(status string) ([]core.DefaultApplication, error) {
	var apps []core.DefaultApplication
	// 为了在列表中显示客户和申请人信息，我们必须在这里预加载它们
	err := r.db.Preload("Customer").Preload("Applicant").Where("status = ?", status).Find(&apps).Error
	return apps, err
}
