# HAL-Proxy Agent Guide

## 项目概述

**HAL-Proxy** - 本地硬件能力控制平面，用于人类、代理、CLI、IDE 和插件

- **语言**: Go 1.26+ (后端), React + TypeScript (前端)
- **数据库**: SQLite
- **架构**: Go API 服务 + React UI

## 技术栈

### 后端 (Go)
- `github.com/golang-jwt/jwt/v5` - JWT 认证
- `go.opentelemetry.io/otel` - 追踪
- `modernc.org/sqlite` - SQLite 驱动
- `github.com/prometheus/client_golang` - 监控指标

### 前端 (React)
- Vite + TypeScript
- Tailwind CSS
- React Router

## 目录结构

```
hal-proxy/
├── cmd/
│   ├── bridge/              # CLI 客户端
│   └── hal-proxyd/           # HTTP 守护进程
├── internal/
│   ├── api/                 # HTTP API handlers
│   ├── domain/              # 领域模型 (types.go)
│   ├── errors/              # 错误定义
│   ├── logging/             # 日志
│   ├── metrics/             # Prometheus 指标
│   ├── policy/              # 策略引擎
│   └── store/               # SQLite 存储
├── ui/                      # React 前端
├── docs/                    # 文档
└── testdata/                # 测试数据
```

## 核心概念

### 案例 (Case)
- 包含一组命令的执行单元
- 状态: pending → running → completed/failed
- 命令风险等级: `observe` | `mutate` | `destructive` | `exclusive`

### 审批 (Approval)
- 高风险命令需要审批才能执行
- 操作: approve | reject

### 报告 (Report)
- 案例执行结果汇总

## 常用命令

```bash
# 启动服务
make run-hal-proxyd

# 构建
make build

# 测试
make test

# 代码格式
make fmt
```

## 代码规范

1. Go 代码使用 `gofmt` 格式化
2. 错误处理使用 `internal/errors` 包
3. 所有导出函数需要有文档注释
4. 公开接口在 `internal/domain/types.go` 定义

## 注意事项

1. **数据库**: SQLite 文件 `hal-proxy.db` 在项目根目录
2. **API 端口**: 默认 `localhost:8080`
3. **认证**: Bearer Token 或 API Key
4. **高风险操作**: `mutate`, `destructive`, `exclusive` 级别需要审批
