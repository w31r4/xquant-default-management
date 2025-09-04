# 业务流程总览

本文档整合了系统的核心业务流程，包括用户认证、违约申请、重生申请以及数据查询与统计。

---

# 1. 用户认证业务流程

## 流程图 (Mermaid)

```mermaid
graph TD
    subgraph Public Access
        A[用户] --> B{API: /register};
        A --> C{API: /login};
    end

    subgraph Registration Process
        B --> B1[Handler: userHandler.Register];
        B1 --> B2[Service: userService.Register];
        B2 --> B3{检查用户名是否存在?};
        B3 -- Yes --> B4[返回错误: 用户已存在];
        B3 -- No --> B5[哈希加密密码];
        B5 --> B6[Repo: userRepository.Create];
        B6 --> B7[创建新用户记录];
        B7 --> B8[返回成功];
    end

    subgraph Login Process
        C --> C1[Handler: userHandler.Login];
        C1 --> C2[Service: userService.Login];
        C2 --> C3{验证用户名和密码};
        C3 -- Invalid --> C4[返回错误: 凭证无效];
        C3 -- Valid --> C5[Utils: 生成 JWT];
        C5 --> C6[返回 JWT Token];
    end

    subgraph Protected API Access
        D[用户 with JWT] --> E{受保护的 API};
        E --> F[Middleware: AuthMiddleware];
        F --> G{验证 JWT Token};
        G -- Invalid --> H[返回 401 Unauthorized];
        G -- Valid --> I{需要特定角色?};
        I -- No --> K[执行目标 Handler];
        I -- Yes --> J[Middleware: RBACMiddleware];
        J --> J1{检查用户角色};
        J1 -- Not Allowed --> L[返回 403 Forbidden];
        J1 -- Allowed --> K;
        K --> M[返回 API 响应];
    end

    style B4 fill:#f9f,stroke:#333,stroke-width:2px
    style C4 fill:#f9f,stroke:#333,stroke-width:2px
    style H fill:#f9f,stroke:#333,stroke-width:2px
    style L fill:#f9f,stroke:#333,stroke-width:2px
```

## 关键代码点

*   **路由定义**: [`cmd/server/main.go`](cmd/server/main.go:115)
*   **用户 Handler**: [`internal/handler/user_handler.go`](internal/handler/user_handler.go)
*   **用户 Service**: [`internal/service/user_service.go`](internal/service/user_service.go)
*   **认证中间件**: [`internal/middleware/auth_middleware.go`](internal/middleware/auth_middleware.go)
*   **授权中间件**: [`internal/middleware/rbac_middleware.go`](internal/middleware/rbac_middleware.go)
*   **JWT 工具**: [`internal/utils/jwt.go`](internal/utils/jwt.go)
*   **密码加密**: [`internal/utils/crypto.go`](internal/utils/crypto.go)

---

# 2. 违约申请业务流程

## 流程图 (Mermaid)

```mermaid
graph TD
    subgraph Applicant Role
        A[申请人 Applicant] --> B{API: POST /applications};
        B[提交违約申请<br>customerName, severity, reason] --> B1[Middleware: RBAC 'Applicant'];
    end

    subgraph Application Creation
        B1 --> C[Handler: appHandler.CreateApplication];
        C --> D[Service: appService.CreateApplication];
        D --> D1{客户是否存在?};
        D1 -- No --> E1[返回 404 Not Found];
        D1 -- Yes --> D2{客户是否已违约?};
        D2 -- Yes --> E2[返回 409 Conflict];
        D2 -- No --> D3{是否存在待处理申请?};
        D3 -- Yes --> E3[返回 409 Conflict];
        D3 -- No --> D4[Repo: appRepo.Create];
        D4 --> D5[创建申请记录<br>status='Pending'];
        D5 --> E4[返回 201 Created];
    end

    subgraph Approver Role
        F[审批人 Approver] --> G{API: GET /applications/pending};
        G[查询待审批列表] --> G1[Middleware: RBAC 'Approver'];

        F --> H{API: POST /applications/review/...};
        H[审批操作] --> H1[Middleware: RBAC 'Approver'];
    end

    subgraph Approval Process
        G1 --> I[Handler: appHandler.GetPendingApplications];
        I --> J[Service: appService.GetPendingApplications];
        J --> K[Repo: appRepo.FindAllByStatus 'Pending'];
        K --> L[返回待审批列表];

        H1 --> M{Approve or Reject?};
        M -- Approve --> N[Handler: appHandler.ApproveApplication];
        N --> O[Service: appService.ApproveApplication];
        O --> P{申请状态是否为 'Pending'?};
        P -- No --> Q1[返回 409 Conflict];
        P -- Yes --> R[**事务开始**];
        R --> S[Repo: customerRepo.Update<br>设置 isDefault=true];
        S --> T[Repo: appRepo.Update<br>设置 status='Approved', approverID];
        T --> U[**事务提交**];
        U --> Q2[返回 200 OK];

        M -- Reject --> V[Handler: appHandler.RejectApplication];
        V --> W[Service: appService.RejectApplication];
        W --> X{申请状态是否为 'Pending'?};
        X -- No --> Q1;
        X -- Yes --> Y[Repo: appRepo.Update<br>设置 status='Rejected', approverID, rejectionReason];
        Y --> Q2;
    end

    style E1 fill:#f9f,stroke:#333,stroke-width:2px
    style E2 fill:#f9f,stroke:#333,stroke-width:2px
    style E3 fill:#f9f,stroke:#333,stroke-width:2px
    style Q1 fill:#f9f,stroke:#333,stroke-width:2px
```

## 关键代码点

*   **路由定义**: [`cmd/server/main.go`](cmd/server/main.go:162)
*   **申请 Handler**: [`internal/handler/application_handler.go`](internal/handler/application_handler.go)
*   **申请 Service**: [`internal/service/application_service.go`](internal/service/application_service.go)
*   **数据模型**: [`internal/core/model.go`](internal/core/model.go)

---

# 3. 重生申请业务流程

## 流程图 (Mermaid)

```mermaid
graph TD
    subgraph Applicant Role
        A[申请人 Applicant] --> B{API: POST /applications/rebirth/apply};
        B[提交重生申请<br>applicationID, rebirthReason] --> B1[Middleware: RBAC 'Applicant'];
    end

    subgraph Rebirth Application
        B1 --> C[Handler: appHandler.ApplyForRebirth];
        C --> D[Service: appService.ApplyForRebirth];
        D --> D1{原申请是否存在?};
        D1 -- No --> E1[返回 404 Not Found];
        D1 -- Yes --> D2{原申请状态是否为 'Approved'?};
        D2 -- No --> E2[返回 409 Conflict];
        D2 -- Yes --> D3[Repo: appRepo.Update];
        D3 --> D4[更新申请记录<br>status='RebirthPending'];
        D4 --> E3[返回 200 OK];
    end

    subgraph Approver Role
        F[审批人 Approver] --> G{API: POST /applications/rebirth/approve};
        G[批准重生申请<br>applicationID] --> G1[Middleware: RBAC 'Approver'];
    end

    subgraph Rebirth Approval
        G1 --> H[Handler: appHandler.ApproveRebirth];
        H --> I[Service: appService.ApproveRebirth];
        I --> J{申请是否存在?};
        J -- No --> K1[返回 404 Not Found];
        J -- Yes --> L{申请状态是否为 'RebirthPending'?};
        L -- No --> K2[返回 409 Conflict];
        L -- Yes --> M[**事务开始**];
        M --> N[Repo: appRepo.Update<br>设置 status='Reborn'];
        N --> O[Repo: customerRepo.Update<br>设置 isDefault=false];
        O --> P[**事务提交**];
        P --> K3[返回 200 OK];
    end

    style E1 fill:#f9f,stroke:#333,stroke-width:2px
    style E2 fill:#f9f,stroke:#333,stroke-width:2px
    style K1 fill:#f9f,stroke:#333,stroke-width:2px
    style K2 fill:#f9f,stroke:#333,stroke-width:2px
```

## 关键代码点

*   **路由定义**: [`cmd/server/main.go`](cmd/server/main.go:175)
*   **申请 Handler**: [`internal/handler/application_handler.go`](internal/handler/application_handler.go)
*   **申请 Service**: [`internal/service/application_service.go`](internal/service/application_service.go)

---

# 4. 数据查询与统计业务流程

## 流程图 (Mermaid)

```mermaid
graph TD
    subgraph User [Any Authenticated Role]
        A[认证用户] --> B{API: GET /applications};
        B --> B1;

        A --> C{API: GET /statistics/...};
        C --> B1;
    end

    subgraph Query Logic
        B1 --> D[Handler: queryHandler.FindApplications];
        D --> E[Service: queryService.FindApplications];
        E --> F[Repo: appRepo.FindAll];
        F --> G[返回申请列表和总数];
    end

    subgraph Statistics Logic
        B1 --> H[Handler: statsHandler.Get...];
        H --> I[Service: statsService.Get...];
        I --> J[Repo: statsRepo.Get...];
        J --> K[返回统计结果];
    end

    subgraph Response
        G --> L[Handler: 格式化为分页响应 DTO];
        L --> M[返回 200 OK with Paginated Data];
        K --> N[Handler: 格式化为统计结果 DTO];
        N --> O[返回 200 OK with Statistics Data];
    end
```

## 关键代码点

*   **查询路由**: [`cmd/server/main.go`](cmd/server/main.go:160)
*   **统计路由**: [`cmd/server/main.go`](cmd/server/main.go:184)
*   **查询 Handler**: [`internal/handler/query_handler.go`](internal/handler/query_handler.go)
*   **统计 Handler**: [`internal/handler/statistics_handler.go`](internal/handler/statistics_handler.go)
*   **查询 Service**: [`internal/service/query_service.go`](internal/service/query_service.go)
*   **统计 Service**: [`internal/service/statistics_service.go`](internal/service/statistics_service.go)