# BridgeOS 架构文档

## 系统概览

BridgeOS 是一个本地硬件能力控制平面，用于管理案例（Case）的执行、审批和报告生成。

## 核心组件

```
┌─────────────────────────────────────────────────────────────┐
│                      BridgeOS                                │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐     │
│  │   bridge    │    │  bridgeosd  │    │   HTTP API  │     │
│  │   (CLI)     │    │  (Daemon)   │    │  /v1/*      │     │
│  └──────┬──────┘    └──────┬──────┘    └──────┬──────┘     │
│         │                   │                   │            │
│         └───────────────────┼───────────────────┘            │
│                             │                                │
│                    ┌────────▼────────┐                      │
│                    │    Service     │                       │
│                    │  (Core Logic)  │                       │
│                    └────────┬────────┘                      │
│                             │                                │
│         ┌───────────────────┼───────────────────┐            │
│         │                   │                   │            │
│  ┌──────▼──────┐    ┌──────▼──────┐    ┌──────▼──────┐     │
│  │   Domain    │    │   Policy    │    │   Store     │     │
│  │  (Types)    │    │  (Risk)     │    │  (SQLite)   │     │
│  └─────────────┘    └─────────────┘    └─────────────┘     │
└─────────────────────────────────────────────────────────────┘
```

## 模块说明

### 1. API 层 (`internal/api/`)

处理 HTTP 请求路由和响应格式化。

- **http.go**: HTTP 服务器实现
  - 请求体大小限制（1MB）
  - 请求超时保护（30秒）
  - 统一的错误响应

### 2. 核心服务层 (`internal/core/`)

实现核心业务逻辑。

- **service.go**: 
  - `CreateCase`: 创建新案例
  - `RunCase`: 执行案例（包含审批流程）
  - `ResolveApproval`: 处理审批决策
  - `BuildReport`: 生成执行报告

### 3. 领域模型 (`internal/domain/`)

定义核心数据结构和业务规则。

- **types.go**:
  - `CaseRecord`: 案例实体
  - `CaseStatus`: 案例状态机
  - `Approval`: 审批实体
  - `EventEnvelope`: 事件封装

### 4. 策略层 (`internal/policy/`)

管理风险评估和审批规则。

- **policy.go**:
  - 风险等级配置
  - 审批要求判断
  - 优先级定义

### 5. 存储层 (`internal/store/`)

数据持久化实现。

- **sqlite.go**: SQLite 数据库操作
- **store.go**: Repository 接口定义

## 数据流

### 案例执行流程

```
1. 创建案例 (POST /v1/cases)
   └─→ 存储到 SQLite
   └─→ 记录事件 bridge.case.created

2. 运行案例 (POST /v1/cases/:id:run)
   ├─→ 检查案例状态
   ├─→ 遍历命令列表
   │   ├─→ 低风险命令 → 直接执行
   │   └─→ 高风险命令 → 创建审批请求
   │       ├─→ 等待审批
   │       ├─→ 批准 → 继续执行
   │       └─→ 拒绝 → 停止执行
   └─→ 完成时记录事件

3. 生成报告 (POST /v1/reports/:id:build)
   └─→ 汇总事件和审批记录
   └─→ 生成 Markdown 报告
```

## 状态机

### 案例状态

```
     ┌──────────┐
     │  draft   │
     └────┬─────┘
          │ (自动转为 ready)
          ▼
     ┌──────────┐     ┌──────────┐
────►│  ready   │────►│ running  │
     └──────────┘     └────┬─────┘
          ▲                │
          │                ├────►┌──────────┐
          │                │     │ paused  │────► (返回 ready)
          │                │     └──────────┘
          │                │
          │                ├────►┌──────────┐
          │                │     │completed│
          │                │     └──────────┘
          │                │
          │                └────►┌──────────┐
          │                      │rejected │
          │                      └──────────┘
          │                            ▲
          └────────────────────────────┘
                 (拒绝时)
```

### 审批状态

```
┌─────────┐     ┌─────────┐     ┌──────────┐
│ pending │────►│ approved│     │ rejected │
└─────────┘     └─────────┘     └──────────┘
```

## 数据库表结构

### cases
| 字段 | 类型 | 说明 |
|------|------|------|
| id | TEXT | 主键 |
| title | TEXT | 标题 |
| status | TEXT | 状态 |
| spec_json | TEXT | 案例规格 (JSON) |
| next_command | INTEGER | 下一个执行的命令索引 |
| created_at | TEXT | 创建时间 |
| updated_at | TEXT | 更新时间 |

### events
| 字段 | 类型 | 说明 |
|------|------|------|
| sequence | INTEGER | 主键，自增 |
| case_id | TEXT | 关联案例ID |
| type | TEXT | 事件类型 |
| payload_json | TEXT | 事件数据 (JSON) |
| created_at | TEXT | 创建时间 |

### approvals
| 字段 | 类型 | 说明 |
|------|------|------|
| id | TEXT | 主键 |
| case_id | TEXT | 关联案例ID |
| command_index | INTEGER | 命令索引 |
| command_name | TEXT | 命令名称 |
| risk_class | TEXT | 风险等级 |
| status | TEXT | 审批状态 |
| reason | TEXT | 审批理由 |
| decided_by | TEXT | 审批人 |
| decided_at | TEXT | 审批时间 |
| created_at | TEXT | 创建时间 |

### reports
| 字段 | 类型 | 说明 |
|------|------|------|
| id | TEXT | 主键 |
| case_id | TEXT | 关联案例ID |
| path | TEXT | 报告文件路径 |
| command_count | INTEGER | 命令数量 |
| event_count | INTEGER | 事件数量 |
| created_at | TEXT | 创建时间 |

## 索引优化

为提升查询性能，创建了以下索引：

- `idx_events_case_id`: 事件按案例ID查询
- `idx_approvals_case_id`: 审批按案例ID查询
- `idx_approvals_case_command`: 审批按案例ID+命令索引查询
- `idx_cases_status`: 案例按状态查询
- `idx_reports_case_id`: 报告按案例ID查询

## 扩展性考虑

### 未来可能的扩展

1. **设备管理**: 集成真实的串口适配器
2. **会话管理**: 实现设备会话租借
3. **策略配置**: 支持 YAML/JSON 配置文件动态加载策略
4. **缓存层**: 添加 LRU 缓存提升读取性能
5. **MCP 集成**: 实现 Model Context Protocol 接口
