# BridgeOS Architecture

## Summary

BridgeOS is a local-first control plane. The core domain owns case execution, approval, evidence, and reporting. CLI and HTTP are thin entry points over the same service layer.

## Main Layers

```text
bridge / bridgeosd / local agents / UI
            |
         HTTP/CLI
            |
         Core Service
            |
   Domain / Policy / Store
            |
          SQLite
```

## Core Boundaries

- `internal/domain`: case, approval, event, report, session, device types
- `internal/policy`: risk normalization and approval requirement rules
- `internal/core`: orchestration, status transitions, event emission, report generation
- `internal/store`: SQLite persistence, optimistic locking, transactions
- `internal/api`: HTTP routing, auth layering, transport validation, error mapping

## Execution Model

1. Create case
2. Persist case and `bridge.case.created`
3. Run case command-by-command
4. Low-risk commands continue directly
5. High-risk commands create pending approvals and pause execution
6. Approval resolution transitions case back to ready or rejected
7. Report generation composes persisted case state and events into artifacts

## Security Model

- loopback trusted mode is available for local-only workflows
- remote or non-loopback requests should use JWT or API key auth
- approval resolution still enforces role checks (`admin` or `approver`)
- request IDs and actor identity are carried through the request path for auditability

## Storage Model

- SQLite is the system of record
- `cases` and `approvals` use optimistic locking
- case run uses a transaction for status and event coupling
- reports are generated as files plus persisted metadata

## Pre-v1 Position

The repository is still pre-v1. The current focus is convergence:

- stable naming
- stable version semantics
- route and error contract correctness
- local agent usability without weakening remote auth posture
- build and repository hygiene

## Current Placeholders

- `devices` and `sessions` are currently mock/read-only responses
- report generation is persisted in SQLite and emitted as artifact files
