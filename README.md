# BridgeOS

BridgeOS is a local hardware capability control plane for humans, agents, CLIs, IDEs, and plugins.

Current repository state:

- product name: `BridgeOS`
- CLI binary: `bridge`
- daemon binary: `bridgeosd`
- default version line: pre-v1 (`0.2.3`)
- API shape: local-first HTTP API with structured JSON responses

## Current Capabilities

- `bridge case new`
- `bridge case run`
- `bridge case show`
- `bridge case events`
- `bridge approval ls`
- `bridge approval approve`
- `bridge approval reject`
- `bridge report build`
- `bridge device ls`
- `bridge session ls`
- `bridge version`
- `bridgeosd` local daemon with `/v1/*` endpoints

High-risk commands pause for approval. Reports are generated into `artifacts/`. CLI and HTTP share the same core service semantics.

## Versioning

BridgeOS follows strict pre-v1 versioning:

- `0.0.x`: fixes, hardening, docs, compatibility
- `0.x.0`: a fully closed new capability loop
- `1.0.0`: only after core semantics, agent-safe CLI, API, approvals, evidence, and replay are stable

The current repository is intentionally **not** `1.0.0`.

## Quickstart

```powershell
go run ./cmd/bridge case new --spec .\testdata\demo-case.json
go run ./cmd/bridge case run --id <case-id>
go run ./cmd/bridge approval ls --case-id <case-id>
go run ./cmd/bridge approval approve --id <approval-id>
go run ./cmd/bridge case run --id <case-id>
go run ./cmd/bridge report build --id <case-id>
go run ./cmd/bridge version
```

Run the daemon:

```powershell
go run ./cmd/bridgeosd
```

The daemon defaults to local-first trusted access for loopback requests when `BRIDGEOS_LOCAL_TRUSTED=true`. Remote access should use JWT or API key based auth.

The current pre-v1 build does not expose a username/password login endpoint. For UI access, use loopback trusted mode locally or provide an existing Bearer token.

## Environment

Preferred environment variables:

- `BRIDGEOS_CONFIG`
- `BRIDGEOS_ADDR`
- `BRIDGEOS_DB`
- `BRIDGEOS_ARTIFACTS`
- `BRIDGEOS_ENV`
- `BRIDGEOS_LOG_LEVEL`
- `BRIDGEOS_JWT_SECRET`
- `BRIDGEOS_JWT_ISSUER`
- `BRIDGEOS_API_KEYS`
- `BRIDGEOS_LOCAL_TRUSTED`
- `BRIDGEOS_LOCAL_TRUSTED_USER_ID`
- `BRIDGEOS_LOCAL_TRUSTED_ROLES`

Legacy `HAL_PROXY_*` variables are still accepted for compatibility.

## Build And Test

```bash
make build
make test
make frontend-build
```

Go checks are limited to `./cmd/...` and `./internal/...` so local UI dependency folders do not pollute backend builds.

## Docs

- API: [docs/api.md](/D:/Destop/test_ui/BridgeOS/docs/api.md)
- Architecture: [docs/architecture.md](/D:/Destop/test_ui/BridgeOS/docs/architecture.md)
- ADR: [docs/adr/0001-pre-v1-versioning.md](/D:/Destop/test_ui/BridgeOS/docs/adr/0001-pre-v1-versioning.md)
