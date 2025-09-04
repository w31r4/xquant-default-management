# 违约管理系统后端

这是一个基于 Go 语言开发的违约管理系统后端服务。

## ✨ 功能特性

- 完整的用户认证系统（注册、登录）
- 基于角色的访问控制 (RBAC)
- 违约申请的创建、审批、拒绝
- 违约申请的重生（Rebirth）流程
- 按行业、地区统计违约和重生数据
- 提供全面的 API 文档

## 🚀 技术栈

- **语言**: [Go](https://golang.org/)
- **Web 框架**: [Gin](https://gin-gonic.com/)
- **数据库**: [PostgreSQL](https://www.postgresql.org/)
- **ORM**: [GORM](https://gorm.io/)
- **API 文档**: [Swagger (go-swagger)](https://github.com/go-swagger/go-swagger)

## 本地开发

### 1. 环境准备

- 安装 [Go](https://golang.org/doc/install) (版本 >= 1.21)
- 安装 [Docker](https://www.docker.com/get-started) 和 [Docker Compose](https://docs.docker.com/compose/install/)
- 安装 [swag CLI](https://github.com/swaggo/swag)

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

### 2. 配置文件

将 `configs/config.example.yaml` 复制一份并重命名为 `configs/config.yaml`。

```bash
cp configs/config.example.yaml configs/config.yaml
```

根据你的本地环境修改 `configs/config.yaml` 中的配置，特别是数据库连接信息。

### 3. 启动数据库

项目使用 Docker Compose 来管理数据库服务。在项目根目录下运行以下命令来启动 PostgreSQL 数据库：

```bash
docker-compose up -d
```

### 4. 运行后端服务

```bash
go run cmd/server/main.go
```

服务将默认在 `8080` 端口启动。

## API 文档

项目使用 Swagger 来生成 API 文档。

### 生成文档

在修改了代码中的 API 注解后，需要重新生成 Swagger 文档：

```bash
swag init -g cmd/server/main.go
```

### 访问文档

启动后端服务后，在浏览器中访问以下地址即可查看 API 文档：

- **中文界面**: [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)

## 数据库迁移

项目使用 GORM 的 `AutoMigrate` 功能在服务启动时自动同步数据库表结构。这意味着你只需要在 `internal/core/model.go` 中定义好你的数据模型，GORM 会自动处理数据库的变更。

**注意**: 在生产环境中，建议使用更专业的数据库迁移工具（如 `golang-migrate/migrate`）来管理数据库变更。
