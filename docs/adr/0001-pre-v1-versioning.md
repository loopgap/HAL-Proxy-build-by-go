# ADR 0001: Pre-v1 Versioning Rules

## Status

Accepted

## Decision

BridgeOS follows strict pre-v1 versioning:

- `0.0.x` for fixes, stability work, schema corrections, and documentation updates
- `0.x.0` only when a key capability loop is fully implemented
- `1.0.0` only after the control plane semantics are stable

Advancing `+0.1` requires:

- user-visible capability increment
- CLI command surface
- Core events and evidence persistence
- policy and approval handling where relevant
- at least one end-to-end test path
- documentation and examples

## Consequences

- Incomplete capability work does not count as a minor release
- Early implementation favors stable semantics over breadth
- The repository may contain later-phase interface placeholders, but only complete loops are treated as shipped milestones
