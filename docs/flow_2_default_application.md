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
