# BridgeOS

<p align="center">
  <img src="docs/architecture-diagram.png" alt="BridgeOS Architecture" width="600">
</p>

<p align="center">
  本地硬件能力控制平面，用于人类、代理、CLI、IDE 和插件
</p>

<p align="center">
  <a href="https://github.com/your-org/bridgeos/actions">
    <img src="https://github.com/your-org/bridgeos/workflows/CI/CD Pipeline/badge.svg" alt="CI">
  </a>
  <a href="https://goreportcard.com/report/github.com/your-org/bridgeos">
    <img src="https://goreportcard.com/badge/github.com/your-org/bridgeos" alt="Go Report Card">
  </a>
  <a href="https://pkg.go.dev/github.com/your-org/bridgeos">
    <img src="https://pkg.go.dev/badge/github.com/your-org/bridgeos" alt="GoDoc">
  </a>
</p>

---

## 📋 目录

- [功能特性](#-功能特性)
- [快速开始](#-快速开始)
- [架构设计](#-架构设计)
- [API 文档](#-api-文档)
- [开发指南](#-开发指南)
- [部署](#-部署)
- [监控](#-监控)
- [贡献指南](#-贡献指南)

---

## ✨ 功能特性

### 已实现功能

| 功能 | 命令 | 说明 |
|------|------|------|
| 创建案例 | `bridge case new` | 从 JSON 规格创建案例 |
| 运行案例 | `bridge case run` | 执行案例中的命令 |
| 查看案例 | `bridge case show` | 查看案例详情 |
| 查看事件 | `bridge case events` | 查看案例事件日志 |
| 列出审批 | `bridge approval ls` | 列出待处理的审批 |
| 批准审批 | `bridge approval approve` | 批准高风险命令 |
| 拒绝审批 | `bridge approval reject` | 拒绝高风险命令 |
| 生成报告 | `bridge report build` | 生成执行报告 |
| 列出设备 | `bridge device ls` | 列出可用设备 |
| 列出会话 | `bridge session ls` | 列出活跃会话 |
| HTTP 服务 | `bridgeosd` | 本地 REST API 守护进程 |

### 安全特性

- ✅ **JWT 认证** - Token-based 身份验证
- ✅ **API Key 认证** - API Key 支持
- ✅ **速率限制** - 滑动窗口限流
- ✅ **CORS** - 跨域资源共享
- ✅ **安全头** - XSS、CSRF 防护
- ✅ **乐观锁** - 并发控制
- ✅ **事务支持** - 原子操作

### 待实现功能

- [ ] 真实串口适配器集成
- [ ] OpenOCD 集成
- [ ] MCP (Model Context Protocol) 接口
- [ ] 桌面 UI
- [ ] 设备会话租借

---

## 🚀 快速开始

### 前置要求

- Go 1.26+
- Node.js 20+ (前端开发)
- Docker (可选)

### 安装

```bash
# 克隆项目
git clone https://github.com/your-org/bridgeos.git
cd bridgeos

# 下载依赖
go mod download

# 安装前端依赖
cd ui && npm install && cd ..
```

### 运行

```bash
# 启动后端服务
make run-bridgeosd

# 新终端 - 创建案例
go run ./cmd/bridge case new --spec ./testdata/demo-case.json

# 运行案例
go run ./cmd/bridge case run --id <case-id>

# 查看审批
go run ./cmd/bridge approval ls

# 批准并继续
go run ./cmd/bridge approval approve --id <approval-id>
go run ./cmd/bridge case run --id <case-id>

# 生成报告
go run ./cmd/bridge report build --id <case-id>
```

### Docker 部署

```bash
# 构建镜像
make docker-build

# 启动服务
make docker-run

# 停止服务
make docker-stop
```

---

## 🏗 架构设计

### 系统架构

```
┌─────────────────────────────────────────────────────────────┐
│                        BridgeOS                              │
├─────────────────────────────────────────────────────────────┤
│  CLI (bridge)  │  HTTP API (bridgeosd)  │  UI (Web)       │
├─────────────────────────────────────────────────────────────┤
│                      Core Service Layer                      │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐      │
│  │  Cases   │ │Approval │ │ Reports │ │  Policy  │      │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘      │
├─────────────────────────────────────────────────────────────┤
│                      Middleware Layer                        │
│  ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐  │
│  │  Auth  │ │  CORS  │ │  Rate  │ │  Log   │ │Security│  │
│  └────────┘ └────────┘ └────────┘ └────────┘ └────────┘  │
├─────────────────────────────────────────────────────────────┤
│                      Storage Layer                           │
│                    SQLite Repository                         │
└─────────────────────────────────────────────────────────────┘
```

### 风险等级

| 等级 | 说明 | 需要审批 |
|------|------|----------|
| `observe` | 只读操作，无修改 | 否 |
| `mutate` | 修改状态的操作 | 是 |
| `destructive` | 可能导致数据丢失 | 是 |
| `exclusive` | 需要独占访问 | 是 |

---

## 📚 API 文档

### 基础信息

- **Base URL**: `http://localhost:8080/v1`
- **认证**: Bearer Token / API Key

### 健康检查

```
GET /v1/health
```

响应:
```json
{
  "status": "healthy",
  "version": "1.0.0"
}
```

### 案例管理

```
# 创建案例
POST /v1/cases
Content-Type: application/json

{
  "title": "Test Case",
  "commands": [
    {"name": "read", "action": "read_mem", "risk_class": "observe"}
  ]
}

# 列出案例
GET /v1/cases

# 获取案例详情
GET /v1/cases/{id}

# 运行案例
POST /v1/cases/{id}/run

# 查看案例事件
GET /v1/cases/{id}/events
```

### 审批管理

```
# 列出审批
GET /v1/approvals
GET /v1/approvals?case_id={case_id}

# 批准审批
POST /v1/approvals/{id}/approve

# 拒绝审批
POST /v1/approvals/{id}/reject
```

### 报告管理

```
# 生成报告
POST /v1/reports/{case_id}/build
```

详细 API 文档请查看 [docs/api.md](docs/api.md)

---

## 🔧 开发指南

### 项目结构

```
bridgeos/
├── cmd/                    # 命令行入口
│   ├── bridge/             # CLI 客户端
│   └── bridgeosd/          # HTTP 服务
├── internal/               # 内部包
│   ├── api/               # HTTP API
│   │   └── middleware/    # 中间件
│   ├── config/            # 配置管理
│   ├── core/              # 核心业务逻辑
│   ├── domain/            # 领域模型
│   ├── errors/           # 错误处理
│   ├── logging/          # 日志
│   ├── metrics/          # Prometheus 指标
│   ├── policy/           # 策略引擎
│   └── store/            # 数据存储
├── ui/                    # 前端 React 应用
├── docs/                  # 文档
└── testdata/              # 测试数据
```

### 构建

```bash
# 构建所有二进制
make build

# 构建 CLI
make build-bridge

# 构建守护进程
make build-bridgeosd

# Docker 构建
make docker-build
```

### 测试

```bash
# 运行所有测试
make test

# 带覆盖率报告
make test-coverage

# 单元测试
make test-unit

# 集成测试
make test-integration
```

### 代码规范

```bash
# 格式化代码
make fmt

# 运行 linter
make lint

# CI 检查
make ci
```

---

## 🚢 部署

### Docker Compose

```bash
# 开发环境
docker-compose up -d

# 生产环境
docker-compose -f docker-compose.prod.yml up -d
```

### Kubernetes

```bash
# 部署
kubectl apply -f k8s/

# 查看 pod
kubectl get pods -l app=bridgeosd

# 查看日志
kubectl logs -l app=bridgeosd
```

### 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `BRIDGEOS_DB` | `bridgeos.db` | SQLite 数据库路径 |
| `BRIDGEOS_ADDR` | `:8080` | HTTP API 监听地址 |
| `BRIDGEOS_ARTIFACTS` | `artifacts/` | 报告输出目录 |
| `BRIDGEOS_ENV` | `development` | 运行环境 |
| `BRIDGEOS_LOG_LEVEL` | `info` | 日志级别 |

---

## 📊 监控

### Prometheus 指标

访问 `http://localhost:8080/metrics` 获取 Prometheus 格式的指标。

**关键指标**:

- `bridgeos_http_requests_total` - HTTP 请求总数
- `bridgeos_http_request_duration_seconds` - 请求延迟
- `bridgeos_cases_created_total` - 创建的案例数
- `bridgeos_cases_running` - 运行中的案例数
- `bridgeos_approvals_requested_total` - 请求的审批数
- `bridgeos_commands_executed_total` - 执行的命令数

### 健康检查

```
GET /v1/health
GET /v1/health/ready  # 就绪检查
GET /v1/health/live   # 存活检查
```

---

## 🤝 贡献指南

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

### 开发流程

```bash
# 1. 安装依赖
make install

# 2. 开发
make dev  # 等同于: make install fmt test

# 3. 确保所有测试通过
make test

# 4. 提交前运行 linter
make lint
```

---

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解更多详情。

---

## 🙏 致谢

- [Go](https://golang.org/) - 编程语言
- [React](https://reactjs.org/) - UI 框架
- [Tailwind CSS](https://tailwindcss.com/) - CSS 框架
- [SQLite](https://www.sqlite.org/) - 数据库
- [Prometheus](https://prometheus.io/) - 监控系统
