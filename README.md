# BridgeOS

BridgeOS 是一个本地硬件能力控制平面，用于人类、代理、CLI、IDE 和插件。

## 当前功能

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

### 待实现功能

- 真实串口适配器集成
- OpenOCD 集成
- MCP (Model Context Protocol) 接口
- 桌面 UI
- 设备会话租借

## 快速开始

### 1. 创建案例

```bash
go run ./cmd/bridge case new --spec ./testdata/demo-case.json
```

### 2. 运行案例

```bash
go run ./cmd/bridge case run --id <case-id>
```

### 3. 查看审批

```bash
go run ./cmd/bridge approval ls
```

### 4. 批准并继续

```bash
go run ./cmd/bridge approval approve --id <approval-id>
go run ./cmd/bridge case run --id <case-id>
```

### 5. 生成报告

```bash
go run ./cmd/bridge report build --id <case-id>
```

所有命令默认返回结构化 JSON，便于本地代理程序解析。

## 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `BRIDGEOS_DB` | `bridgeos.db` | SQLite 数据库路径 |
| `BRIDGEOS_ADDR` | `:8080` | HTTP API 监听地址 |
| `BRIDGEOS_ARTIFACTS` | `artifacts/` | 报告输出目录 |

## 风险等级

| 等级 | 说明 | 需要审批 |
|------|------|----------|
| `observe` | 只读操作，无修改 | 否 |
| `mutate` | 修改状态的操作 | 是 |
| `destructive` | 可能导致数据丢失 | 是 |
| `exclusive` | 需要独占访问 | 是 |

## 文档

- [API 文档](docs/api.md) - REST API 详细说明
- [架构文档](docs/architecture.md) - 系统架构设计

## 开发

### 构建

```bash
go build ./cmd/bridge
go build ./cmd/bridgeosd
```

### 测试

```bash
go test ./...
```

### 运行 Daemon

```bash
BRIDGEOS_DB=./data/bridgeos.db BRIDGEOS_ADDR=:8080 go run ./cmd/bridgeosd
```
