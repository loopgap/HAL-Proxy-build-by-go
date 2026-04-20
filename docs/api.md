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
  "version": "0.2.3"
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
