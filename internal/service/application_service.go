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
	ApproveApplication(appID, approverID uuid.UUID) error                     // ApplicationService 接口增加 ApproveApplication 方法
	RejectApplication(appID, approverID uuid.UUID, reason string) error       // 新增
	GetPendingApplications() ([]core.DefaultApplication, error)               // 新增
	ApplyForRebirth(appID, applicantID uuid.UUID, rebirthReason string) error // 新增
	ApproveRebirth(appID, approverID uuid.UUID) error                         // 新增
}

// applicationService 是 ApplicationService 接口的具体实现。
// 它持有所有需要的 Repository 依赖，以便与数据库交互。
type applicationService struct {
	appRepo      repository.ApplicationRepository
	customerRepo repository.CustomerRepository
	db           *gorm.DB // 新增一个 db 字段用于事务
}

// NewApplicationService 是 applicationService 的构造函数。
// 通过依赖注入的方式，传入所需的 Repository 实例。
func NewApplicationService(db *gorm.DB, appRepo repository.ApplicationRepository, customerRepo repository.CustomerRepository) ApplicationService {
	return &applicationService{db: db, appRepo: appRepo, customerRepo: customerRepo}
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

	return app, nil
}

// ApproveApplication 批准一个违约申请
//批准一个违约申请需要修改客户的状态 isDefault 和修改申请单的状态
//申请单需要修改"status", "approver_id", "approval_time"

func (s *applicationService) ApproveApplication(appID, approverID uuid.UUID) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		txAppRepo := repository.NewApplicationRepository(tx)
		//建立新的申请处理仓管
		txCustomerRepo := repository.NewCustomerRepository(tx)
		//建立新的顾客仓管

		// 1. 获取申请单，它已经包含了 Customer 信息
		app, err := txAppRepo.GetByID(appID)
		//再 GetByID 中，*customer 容器被填充
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("application not found")
			}
			return err
		}

		// 2. 检查申请单状态
		if app.Status != "Pending" {
			return errors.New("application is not in pending state")
		}

		// --- 关键修改在这里 ---
		// 3. 直接使用预加载的 Customer 对象
		// 我们需要先检查一下 Customer 是否真的被加载了，这是一个好的防御性编程习惯
		if app.Customer.ID == uuid.Nil {
			return errors.New("customer data is missing in the application")
		}
		customer := &app.Customer // 直接获取指针

		// 4. 先更新客户状态
		customer.IsDefault = true
		if err := txCustomerRepo.Update(customer, "IsDefault"); err != nil {
			return err // 如果这里失败，事务回滚
		}
		app.Customer = core.Customer{} // 或者 app.Customer = *new(core.Customer)
		//这里将客户容器清空，防止后续 GORM 误操作

		// 5. 再更新申请单状态（使用结构体方式）
		now := time.Now()
		app.Status = "Approved"
		app.ApproverID = &approverID
		app.ApprovalTime = &now

		// 使用 Select 明确指定要更新的字段
		if err := txAppRepo.Update(app, "status", "approver_id", "approval_time"); err != nil {
			return err
		}

		return nil

	})
}

// RejectApplication 拒绝一个违约申请
func (s *applicationService) RejectApplication(appID, approverID uuid.UUID, reason string) error {
	// 即使只更新一张表，使用事务也是一个好习惯，可以保持代码风格一致性
	return s.db.Transaction(func(tx *gorm.DB) error {
		txAppRepo := repository.NewApplicationRepository(tx)

		// 1. 获取申请单
		app, err := txAppRepo.GetByID(appID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("application not found")
			}
			return err
		}

		// 2. 检查状态
		if app.Status != "Pending" {
			return errors.New("application is not in pending state")
		}

		// 3. 更新申请单状态
		now := time.Now()
		app.Status = "Rejected"
		app.ApproverID = &approverID
		app.ApprovalTime = &now
		app.RejectionReason = reason // 记录拒绝原因
		// 需要更新的数据库列名列表
		updateFields := []string{
			"Status",
			"ApproverID",
			"ApprovalTime",
			"RejectionReason",
		}
		// 使用我们之前优化的、更安全的 Update 方法
		if err := txAppRepo.Update(app, updateFields...); err != nil {
			return err // 事务将回滚
		}

		return nil // 事务将提交
	})
}

// GetPendingApplications 获取所有待处理的申请
func (s *applicationService) GetPendingApplications() ([]core.DefaultApplication, error) {
	return s.appRepo.FindAllByStatus("Pending")
}

// ...

// ApplyForRebirth 为一个已违约的申请发起重生
func (s *applicationService) ApplyForRebirth(appID, applicantID uuid.UUID, rebirthReason string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		txAppRepo := repository.NewApplicationRepository(tx)

		app, err := txAppRepo.GetByID(appID)
		if err != nil {
			// 如果是 GORM 经典的 "没找到" 错误，我们就把它翻译成一个更业务化的错误
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("application not found")
			}
			// 如果是其他数据库错误，直接返回
			return err
		}

		// 业务规则：只有 "Approved" 状态的申请才能发起重生
		if app.Status != "Approved" {
			return errors.New("only approved applications can apply for rebirth")
		}

		app.Status = "RebirthPending"
		app.RebirthReason = rebirthReason

		return txAppRepo.Update(app, "Status", "RebirthReason")
	})
}

// ApproveRebirth 批准一个重生申请
func (s *applicationService) ApproveRebirth(appID, approverID uuid.UUID) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		txAppRepo := repository.NewApplicationRepository(tx)
		txCustomerRepo := repository.NewCustomerRepository(tx)

		app, err := txAppRepo.GetByID(appID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("application not found")
			}
			return err
		}

		// 业务规则：只有 "RebirthPending" 状态的申请才能被批准重生
		if app.Status != "RebirthPending" {
			return errors.New("application is not pending for rebirth approval")
		}

		// 1. 更新申请单状态
		now := time.Now()
		app.Status = "Reborn"
		app.RebirthApproverID = &approverID
		app.RebirthApprovalTime = &now
		updateAppFields := []string{"Status", "RebirthApproverID", "RebirthApprovalTime"}
		if err := txAppRepo.Update(app, updateAppFields...); err != nil {
			return err
		}

		// 2. 更新客户状态
		// 注意：txAppRepo.GetByID 已经 Preload 了 Customer，所以我们不需要重新查询
		if app.Customer.ID == uuid.Nil {
			// 这是一个防御性检查，防止 Preload 失败
			return errors.New("customer data is missing for this application")
		}
		customer := &app.Customer
		customer.IsDefault = false
		if err := txCustomerRepo.Update(customer, "IsDefault"); err != nil {
			return err
		}

		return nil
	})
}
