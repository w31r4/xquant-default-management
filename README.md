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

## 如何查看 API 文档 (For Frontend Engineers)

作为一名前端工程师，你可能只关心如何运行后端服务并获取 API 接口文档。请遵循以下极简步骤：

**前提**: 你只需要安装 [Docker](https://www.docker.com/get-started) 和 [Go](https://golang.org/doc/install) 即可。

**第一步：克隆并进入项目**
```bash
git clone [你的项目仓库地址]
cd [项目目录]
```

**第二步：复制配置文件**
```bash
cp configs/config.example.yaml configs/config.yaml
```

**第三步：一键启动服务 (数据库 + 后端)**
```bash
docker-compose up -d && go run cmd/server/main.go
```

**第四步：查看文档**

服务启动后，在浏览器中打开以下地址即可查看 API 文档：

[http://localhost:8080/api/v1/swagger/index.html](http://localhost:8080/api/v1/swagger/index.html)


---

## 后端开发指南 (For Backend Developers)

### 1. 环境准备

- 确保你已经安装了 [Go](https://golang.org/doc/install) (版本 >= 1.21)
- 确保你已经安装了 [Docker](https://www.docker.com/get-started) 和 [Docker Compose](https://docs.docker.com/compose/install/)

### 2. 克隆项目

```bash
git clone [你的项目仓库地址]
cd [项目目录]
```

### 3. 配置项目

将 `configs/config.example.yaml` 复制一份并重命名为 `configs/config.yaml`。

```bash
cp configs/config.example.yaml configs/config.yaml
```
根据你的本地环境修改 `configs/config.yaml` 中的配置，特别是数据库连接信息。

### 4. 启动依赖服务

项目使用 Docker Compose 来管理数据库服务。在项目根目录下运行以下命令来启动 PostgreSQL 数据库：

```bash
docker-compose up -d
```

### 5. 安装项目依赖

此命令将会下载 `go.mod` 文件中定义的所有必需的模块。

```bash
go mod tidy
```

### 6. 运行后端服务

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
