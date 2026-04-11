# BridgeOS API 文档

## 概述

BridgeOS 提供 RESTful API 接口，默认监听地址 `http://localhost:8080`。

## 基础信息

- **Base URL**: `/v1`
- **Content-Type**: `application/json`
- **超时时间**: 30 秒

## 错误响应

所有错误响应遵循统一格式：

```json
{
  "error": "错误信息"
}
```

### 错误码

| HTTP 状态码 | 含义 |
|-------------|------|
| 400 | 请求格式错误或参数无效 |
| 404 | 资源不存在 |
| 413 | 请求体过大（最大 1MB）|
| 500 | 服务器内部错误 |

## API 端点

### 案例管理

#### 创建案例
```
POST /v1/cases
```

**请求体**:
```json
{
  "title": "my-case",
  "commands": [
    {
      "name": "read-memory",
      "action": "read_mem",
      "risk_class": "observe",
      "parameters": {
        "address": "0x20000000",
        "length": 16
      }
    }
  ]
}
```

**响应** (201 Created):
```json
{
  "id": "case-abc123",
  "title": "my-case",
  "status": "ready",
  "spec": {...},
  "next_command": 0,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

#### 获取案例
```
GET /v1/cases/{case_id}
```

**响应** (200 OK):
```json
{
  "id": "case-abc123",
  "title": "my-case",
  "status": "ready",
  "spec": {...},
  "next_command": 0,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

#### 运行案例
```
POST /v1/cases/{case_id}:run
```

**响应** (200 OK):
```json
{
  "case": {...},
  "status": "completed",
  "pending_approval": null
}
```

可能的状态值：
- `already_completed`: 案例已完成
- `awaiting_approval`: 等待审批
- `rejected`: 被拒绝
- `completed`: 已完成

#### 获取案例事件
```
GET /v1/cases/{case_id}/events
```

**响应** (200 OK):
```json
[
  {
    "sequence": 1,
    "case_id": "case-abc123",
    "type": "bridge.case.created",
    "payload": {},
    "created_at": "2024-01-01T00:00:00Z"
  }
]
```

### 审批管理

#### 列出审批
```
GET /v1/approvals
GET /v1/approvals?case_id={case_id}
```

**响应** (200 OK):
```json
[
  {
    "id": "approval-xyz",
    "case_id": "case-abc123",
    "command_index": 1,
    "command_name": "reset-device",
    "risk_class": "destructive",
    "status": "pending",
    "created_at": "2024-01-01T00:00:00Z"
  }
]
```

#### 批准审批
```
POST /v1/approvals/{approval_id}:approve
```

**响应** (200 OK):
```json
{
  "id": "approval-xyz",
  "case_id": "case-abc123",
  "command_index": 1,
  "command_name": "reset-device",
  "risk_class": "destructive",
  "status": "approved",
  "decided_by": "daemon",
  "decided_at": "2024-01-01T00:00:00Z",
  "reason": "",
  "created_at": "2024-01-01T00:00:00Z"
}
```

#### 拒绝审批
```
POST /v1/approvals/{approval_id}:reject
```

**响应** (200 OK): 与批准响应格式相同，`status` 为 `rejected`

### 报告管理

#### 生成报告
```
POST /v1/reports/{case_id}:build
```

**响应** (200 OK):
```json
{
  "id": "report-123",
  "case_id": "case-abc123",
  "path": "artifacts/case-abc123-report.md",
  "command_count": 3,
  "event_count": 10,
  "created_at": "2024-01-01T00:00:00Z"
}
```

### 其他端点

#### 列出设备
```
GET /v1/devices
```

#### 列出会话
```
GET /v1/sessions
```

## 风险等级

| 等级 | 说明 | 需要审批 |
|------|------|----------|
| `observe` | 只读操作，无修改 | 否 |
| `mutate` | 修改状态的操作 | 是 |
| `destructive` | 可能导致数据丢失 | 是 |
| `exclusive` | 需要独占访问 | 是 |

## 事件类型

案例执行过程中会产生以下事件：

| 事件类型 | 说明 |
|----------|------|
| `bridge.case.created` | 案例创建 |
| `bridge.case.run_requested` | 运行请求 |
| `bridge.case.completed` | 案例完成 |
| `bridge.step.started` | 步骤开始 |
| `bridge.command.dispatched` | 命令分发 |
| `bridge.observation.recorded` | 观察记录 |
| `bridge.approval.requested` | 审批请求 |
| `bridge.approval.resolved` | 审批解决 |
| `bridge.approval.accepted` | 审批通过 |
| `bridge.report.generated` | 报告生成 |
