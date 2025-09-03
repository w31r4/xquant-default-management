// package repository 包含了数据访问层 (Data Access Layer, DAL) 的所有代码。
// 它的职责是封装所有与数据库的直接交互（CRUD 操作），
// 为 Service 层提供清晰、简洁的数据操作接口，并隐藏底层的数据库实现细节。
package repository

import (
	"xquant-default-management/internal/core"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CustomerRepository 定义了与 Customer 模型相关的数据操作接口。
// 通过定义接口，Service 层可以依赖于此抽象，而不是具体的 GORM 实现，
// 这有助于实现分层解耦和方便进行单元测试。
type CustomerRepository interface {
	// Create 插入一个新的客户记录。
	Create(customer *core.Customer) error
	// GetByName 根据客户名称查找一个客户记录。客户名称被假定为唯一的。
	GetByName(name string) (*core.Customer, error)
	// GetByID 根据客户的 UUID 主键查找一个客户记录。
	GetByID(id uuid.UUID) (*core.Customer, error)
	Update(app *core.Customer, fields ...string) error
}

// customerRepository 是 CustomerRepository 接口的具体实现。
// 它内部持有 *gorm.DB 数据库连接实例，用于执行实际的数据库操作。
type customerRepository struct {
	db *gorm.DB
}

// NewCustomerRepository 是 customerRepository 的构造函数。
// 它接收一个数据库连接实例，并通过依赖注入的方式创建一个新的 Repository。
func NewCustomerRepository(db *gorm.DB) CustomerRepository {
	return &customerRepository{db: db}
}

// Create 使用 GORM 的 Create 方法将一个新的客户实体持久化到数据库中。
func (r *customerRepository) Create(customer *core.Customer) error {
	// r.db.Create(customer) 会生成 INSERT SQL 语句。
	// .Error 字段会返回执行过程中发生的任何错误。
	return r.db.Create(customer).Error
}

// GetByName 在数据库中通过客户名称查询客户。
// GORM 的 First 方法会查找第一条匹配的记录，如果没有找到，会返回 gorm.ErrRecordNotFound 错误。
func (r *customerRepository) GetByName(name string) (*core.Customer, error) {
	var customer core.Customer
	// 构建 WHERE name = ? 查询条件，并将结果填充到 customer 变量中。
	err := r.db.Where("name = ?", name).First(&customer).Error
	return &customer, err
}

// GetByID 在数据库中通过主键 (ID) 查询客户。
// 当 First 方法的第二个参数不是 struct 而是主键值时，GORM 会自动构建 WHERE id = ? 的查询。
func (r *customerRepository) GetByID(id uuid.UUID) (*core.Customer, error) {
	var customer core.Customer
	// 这是 GORM 按主键查询的便捷写法。
	err := r.db.First(&customer, id).Error
	return &customer, err
}

// func (r *customerRepository) Update(customer *core.Customer) error {
// 	// 同样使用 Updates 来避免 Save 的副作用
// 	return r.db.Model(&customer).Updates(customer).Error
// }

// Update 只更新指定的字段
func (r *customerRepository) Update(app *core.Customer, fields ...string) error {
	return r.db.Model(app).Select(fields).Updates(app).Error
}
