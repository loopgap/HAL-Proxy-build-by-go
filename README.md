# BridgeOS

BridgeOS is a local hardware capability control plane for humans, agents, CLIs, IDEs, and plugins.

Current repository state:

- product name: `BridgeOS`
- CLI binary: `bridge`
- daemon binary: `bridgeosd`
- default version line: pre-v1 (`0.4.3`)
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

## Security Configuration

### JWT Authentication

**Required**: Set a secure JWT secret (minimum 32 characters):

```bash
export BRIDGEOS_JWT_SECRET="your-secure-secret-at-least-32-characters-long"
```

**Important**: The default placeholder secret is rejected at startup to prevent security misconfiguration.

### Authentication Methods

| Method | Use Case | Configuration |
|--------|----------|---------------|
| JWT Bearer Token | API clients, CI/CD | `Authorization: Bearer <token>` |
| API Key | Service-to-service | `X-API-Key: <key>` |
| Local Trusted Mode | Local development only | `BRIDGEOS_LOCAL_TRUSTED=true` |

### Local Trusted Mode (Development Only)

When `BRIDGEOS_LOCAL_TRUSTED=true`, requests from localhost/loopback bypass JWT authentication. This is intended for local development only.

**Security Warning**: Never enable `BRIDGEOS_LOCAL_TRUSTED` in production or on exposed networks.

### Rate Limiting

Rate limiting is enabled by default (60 requests/minute, burst 10). Configure via:

```bash
export BRIDGEOS_RATE_LIMIT_ENABLED=true
export BRIDGEOS_RATE_LIMIT_RPM=60
export BRIDGEOS_RATE_LIMIT_BURST=10
```

### Security Best Practices

1. **Use HTTPS** in production environments
2. **Rotate JWT secrets** periodically
3. **Disable local trusted mode** in production (`BRIDGEOS_LOCAL_TRUSTED=false`)
4. **Use API keys** for service-to-service authentication
5. **Monitor logs** for authentication failures

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

## Security & Robustness (Advanced Enhancements)

BridgeOS features industrial-grade security controls and daemon robustness protections:

### 1. Timing Attack Protections (API Key Hashing)
All configured API keys are pre-hashed using SHA-256 during middleware initialization. Request keys are hashed and matched using `crypto/subtle.ConstantTimeCompare` against a dummy value fallback, completely preventing network-level timing side-channel attacks from brute-forcing valid API keys.

### 2. Strong JWT Cryptography
The token validator strictly enforces HMAC-SHA256 signature algorithm checks (blocking `none` algorithm bypass attempts) and enforces a minimum JWT secret length of **at least 32 characters** (256-bit entropy). Short or empty credentials trigger immediate startup configuration blocks.

### 3. Local File Inclusion (LFI) Sandbox Container
The report content delivery service enforces directory boundary confinement. All paths are resolved and standardized using `filepath.Clean` to strip `..` relative segments, and strictly validated to have the authorized `artifacts/` folder prefix (`strings.HasPrefix`). Traversal attempts outside the directory are automatically blocked with `403 Forbidden`.

### 4. Reverse Proxy & Rate Limiter DoS Mitigations
To support standard container reverse proxy setups (Nginx, HAProxy, Ingress), `isTrustedProxy` extracts the real client IP from `X-Forwarded-For` only when the loopback connection is explicitly whitelisted in `BRIDGEOS_TRUSTED_PROXIES` (e.g. `export BRIDGEOS_TRUSTED_PROXIES="127.0.0.1"`). Otherwise, direct loopback trusted mode handles direct client traffic safely.

### 5. Daemon Auto-Recovery (`SafeGo`)
All critical background tickers (Rate-limiter cleaner, Prometheus telemetry updates) run encapsulated inside the self-recovering `logging.SafeGo` executor. Panics inside background goroutines are gracefully caught, structural logs with detailed stack traces are saved, and the master daemon daemon process is protected from crash collapses.

### 6. Fail-Closed Policy Engine
The policy engine enforces a fail-closed paradigm for unknown risk classes. In contrast to legacy behaviors that treated unknown risk classes as low-risk `observe` operations, unknown risks now default to requiring explicit human approvals, highest priority (`999`), and return a configuration warning, mitigating bypass vectors utilizing spelling errors (e.g., `"destrutive"`).

### 7. Enforced Transactional Integrity & TOCTOU Mitigations
All case executions (`RunCase`), report building (`BuildReport`), and complex reads (`GetCaseWithRelations`) bind their database accesses tightly inside unified transactions (`GetCaseInTx`, `ListEventsInTx`, etc.). This eliminates Time-of-Check to Time-of-Use (TOCTOU) race conditions in concurrent cases and ensures robust relational consistency.

### 8. Mounted Global HTTP Security Headers Middleware
The `middleware.Security` engine is actively mounted at the outermost layer of the HTTP handler chain. All API responses now enforce critical hardening headers, including Strict-Transport-Security (HSTS), Content-Security-Policy (CSP), X-Content-Type-Options (nosniff), X-Frame-Options (DENY), and X-XSS-Protection (1; mode=block).

### 9. Secure API Key Hashing & Verification
API keys are pre-hashed using SHA-256 during daemon startup inside `NewServer` to prevent exposure of raw keys. The `withAPIKeyClaims` pipeline performs verification using SHA-256 and constant-time comparisons (`crypto/subtle.ConstantTimeCompare`) against a dummy hash fallback, shutting down timing attack channels completely.

### 10. Deep Input Validation & Denial-of-Service (DoS) Defenses
The daemon enforces strict bounds on all incoming structures. Case creation restricts titles to 500 characters, command arrays to 100 entries, and verifies every command against `policy.ValidateRisk`. Request body reads are bounded using `http.MaxBytesReader` (1MB), and query parameters clamp output sizes (`limit` is capped at 100, `offset` is validated, and `case_id` query parameters undergo strict path traversal checks).

## Build And Test

### 1. Build Binaries
```bash
# Build backend CLI ('bridge') and daemon ('bridgeosd')
make build

# Build frontend asset bundle
make frontend-build
```

### 2. Run Tests
BridgeOS includes a highly comprehensive testing suite covering backend Go logic, frontend React unit tests, and Playwright End-to-End (E2E) browser automation.

```bash
# Run all backend Go unit & integration tests
make test

# Run UI frontend unit tests (using Vitest)
make frontend-test

# Run TypeScript compilation checks on the frontend
make frontend-typecheck

# Run full local pre-release CI gate (runs Go unit/race/integration tests, UI unit tests, UI typechecks, UI linting, and Playwright E2E browser tests)
make local-ci
```

### 3. Testing Coverage
- **Go Tests**: Runs unit tests, race condition detectors, integration tests, and aggregates coverage.
- **UI Vitest Tests**: Unit tests checking React components, styling, routing, authentication state management, and toast notifications.
- **Playwright E2E Tests**: Complete headless Chrome browser simulation verifying login modes (trusted loopback, bearer token, API key), dashboard metrics, case lists, case creation forms, approvals approve/reject dialogues, markdown report builds, 404 fallbacks, and API error recovery. Under automation, the API key client sets retries to 0 to enable immediate deterministic feedback.

Go checks are limited to `./cmd/...` and `./internal/...` so local UI dependency folders do not pollute backend builds.

## Docs

- API: [docs/api.md](/D:/Destop/test_ui/BridgeOS/docs/api.md)
- Architecture: [docs/architecture.md](/D:/Destop/test_ui/BridgeOS/docs/architecture.md)
- ADR: [docs/adr/0001-pre-v1-versioning.md](/D:/Destop/test_ui/BridgeOS/docs/adr/0001-pre-v1-versioning.md)
