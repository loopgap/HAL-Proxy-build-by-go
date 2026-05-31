# BridgeOS API

## Base

- base path: `/v1`
- content type: `application/json`
- health endpoints are unauthenticated
- other endpoints require JWT/API key unless the request is loopback and local trusted mode is enabled
- no username/password login endpoint is currently exposed; use local trusted mode on loopback or an existing Bearer token/API key

## Health

### `GET /v1/health`

Returns:

```json
{
  "status": "healthy",
  "name": "BridgeOS",
  "version": "0.4.4"
}
```

### `GET /v1/health/ready`

Checks daemon readiness and database reachability.

### `GET /v1/health/live`

Checks daemon liveness.

## Cases

### `POST /v1/cases`

Create a case.

### `GET /v1/cases`

List cases. Response is paginated:

```json
{
  "items": [],
  "next_cursor": "",
  "has_more": false
}
```

### `GET /v1/cases/{id}`

Get a single case.

### `POST /v1/cases/{id}/run`

Run or resume a case.

### `GET /v1/cases/{id}/events`

Get paginated case events. Query parameters:

- `limit`
- `offset`

Returns:

```json
{
  "items": [],
  "total": 0,
  "limit": 100,
  "offset": 0
}
```

## Approvals

### `GET /v1/approvals`

Optional query:

- `case_id`

### `POST /v1/approvals/{id}/approve`

Approve a pending approval.

### `POST /v1/approvals/{id}/reject`

Reject a pending approval.

## Reports

### `GET /v1/reports`

List reports. Optional query:

- `case_id`

### `GET /v1/reports/{id}`

Get report metadata by report id.

### `GET /v1/reports/{id}/content`

Return the generated report body as `text/markdown`.

### `POST /v1/reports/{case_id}/build`

Generate a report for a case.

## Devices And Sessions

### `GET /v1/devices`

List devices. Current implementation is mock/read-only.

### `GET /v1/sessions`

List sessions. Current implementation is mock/read-only.

## Error Shape

HTTP and CLI both converge on structured errors:

```json
{
  "error": "resource_not_found",
  "message": "Resource not found",
  "code": 2001
}
```

`code` is present for application errors and omitted for generic transport/runtime failures.

## Error Codes

| Code | Error | Description |
|------|-------|-------------|
| 1000 | `general_error` | General error (used for base errors) |
| 1001 | `internal_server_error` | Unexpected server error (details logged server-side) |
| 1002 | `invalid_input` | Request validation failed |
| 1003 | `not_found` | Requested resource does not exist |
| 1004 | `unauthorized` | Missing or invalid authentication |
| 1005 | `forbidden` | Insufficient permissions |
| 1006 | `conflict` | Resource conflict (e.g., concurrent modification) |
| 1007 | `timeout` | Request timed out |
| 2000 | `case_base` | Case error base code (for range detection) |
| 2001 | `case_not_found` | Case ID not found |
| 2002 | `case_invalid_status` | Case cannot transition to requested status |
| 2003 | `case_already_exists` | Case with this ID already exists |
| 2004 | `case_not_runnable` | Case is in terminal state (completed/rejected) |
| 3000 | `approval_base` | Approval error base code (for range detection) |
| 3001 | `approval_not_found` | Approval ID not found |
| 3002 | `approval_invalid` | Approval validation failed |
| 3003 | `approval_expired` | Approval has expired |
| 4000 | `report_base` | Report error base code (for range detection) |
| 4001 | `report_not_found` | Report ID not found |
| 4002 | `report_generation_failed` | Failed to generate report |
| 4003 | `report_content_missing` | Report file not found on disk |
| 5000 | `store_base` | Store error base code (for range detection) |
| 5001 | `store_init_failed` | Database initialization failed |
| 5002 | `store_operation_failed` | Database operation failed |
| 5003 | `store_not_found` | Record not found in database |

## HTTP Status Codes

| Status | Meaning |
|--------|---------|
| 200 | Success |
| 201 | Created |
| 400 | Bad Request (invalid input) |
| 401 | Unauthorized (missing/invalid auth) |
| 403 | Forbidden (insufficient permissions) |
| 404 | Not Found |
| 409 | Conflict (concurrent modification) |
| 413 | Request Entity Too Large |
| 429 | Too Many Requests (rate limited) |
| 500 | Internal Server Error |
| 503 | Service Unavailable |
