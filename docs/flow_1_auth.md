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
