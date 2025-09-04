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
