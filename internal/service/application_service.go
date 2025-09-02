// package service 包含了应用的核心业务逻辑。
// Service 层负责编排 Repository 层进行数据读写，并执行所有业务规则，
// 确保数据操作的有效性和一致性。它不关心 HTTP 请求或响应的具体格式。
package service

import (
	"errors"
	"time"
	"xquant-default-management/internal/core"
	"xquant-default-management/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ApplicationService 定义了与违约申请相关的业务操作接口。
// 接口化设计使得 Handler 层可以解耦具体的实现，方便进行单元测试。
type ApplicationService interface {
	// CreateApplication 定义了创建新违约申请的业务流程。
	CreateApplication(customerName, severity, reason, remarks string, applicantID uuid.UUID) (*core.DefaultApplication, error)
}

// applicationService 是 ApplicationService 接口的具体实现。
// 它持有所有需要的 Repository 依赖，以便与数据库交互。
type applicationService struct {
	appRepo      repository.ApplicationRepository
	customerRepo repository.CustomerRepository
}

// NewApplicationService 是 applicationService 的构造函数。
// 通过依赖注入的方式，传入所需的 Repository 实例。
func NewApplicationService(appRepo repository.ApplicationRepository, customerRepo repository.CustomerRepository) ApplicationService {
	return &applicationService{appRepo: appRepo, customerRepo: customerRepo}
}

// CreateApplication 实现了创建新违约申请的核心业务逻辑。
// 它按照业务规则进行一系列校验，全部通过后才会创建新的申请记录。
func (s *applicationService) CreateApplication(customerName, severity, reason, remarks string, applicantID uuid.UUID) (*core.DefaultApplication, error) {
	// 业务规则 1: 确认客户存在。
	// 在进行任何操作前，必须先通过客户名称查询，确保我们操作的目标客户是存在的。
	customer, err := s.customerRepo.GetByName(customerName)
	if err != nil {
		// 如果错误是 gorm.ErrRecordNotFound，说明数据库中没有这个客户。
		// 我们将其转换为一个对上层（Handler）更友好的、不暴露底层细节的业务错误。
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("customer not found")
		}
		// 对于其他类型的数据库错误，直接返回。
		return nil, err
	}

	// 业务规则 2: 检查客户是否已经是违约状态。
	// 如果客户的 IsDefault 标志已经为 true，则不允许为其重复提交新的违约申请。
	if customer.IsDefault {
		return nil, errors.New("customer is already in default status")
	}

	// 业务规则 3: 检查是否已有待处理 (Pending) 的申请。
	// 为防止重复劳动和流程冲突，系统不允许在已有申请待处理的情况下，为同一客户再次提交申请。
	existingApp, err := s.appRepo.FindPendingByCustomerID(customer.ID)
	if err != nil {
		// 这不是“未找到”错误，而是查询本身可能出了问题，应视为一个错误。
		return nil, err
	}
	if existingApp != nil {
		// 如果找到了待处理的申请，则返回一个业务冲突错误。
		return nil, errors.New("there is already a pending application for this customer")
	}

	// 4. 所有业务规则校验通过后，创建新的申请实体。
	// 用传入的参数和系统生成的值来填充 DefaultApplication 结构体。
	app := &core.DefaultApplication{
		CustomerID:      customer.ID,
		Status:          "Pending", // 新申请的状态默认为 "Pending"
		Severity:        severity,
		DefaultReason:   reason,
		Remarks:         remarks,
		ApplicantID:     applicantID,
		ApplicationTime: time.Now(), // 记录申请提交的精确时间
	}

	// 5. 将新创建的申请实体持久化到数据库。
	// 调用 Repository 层的 Create 方法来执行数据库插入操作。
	if err := s.appRepo.Create(app); err != nil {
		return nil, err
	}

	// 6. 成功创建后，返回新生成的申请实体指针和 nil 错误。
	// 返回的 app 对象将包含由数据库生成的 ID 和时间戳等信息。
	app.Customer = *customer // 把我们第一步查出来的 customer 对象，手动塞进 app 的 "容器" 里
	return app, nil
}
