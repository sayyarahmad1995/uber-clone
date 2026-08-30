# Work Log

## Purpose

This file tracks meaningful project progress and the current stopping point so work can resume without reconstructing previous decisions.

Update this file after every meaningful work session.

---

## Current Status

**Completed engineering milestone:** Deployment Foundation

**Completed business slice:** User Entry and Rider Foundation

**Current hardening task:** Make authentication and identity infrastructure explicitly replaceable before starting the next business slice.

**Next business slice:** Ride Request Foundation

Expected next outcome:

`Authenticated Rider → Set pickup → Set destination → Create persisted ride request`

Driver matching, live trip execution, pricing complexity, maps-provider coupling, and other later concerns remain outside that slice until refinement proves they are needed.

---

## Completed

- [x] Repository and MVP-first development approach established.
- [x] Modular monolith selected for the Go backend.
- [x] PostgreSQL selected as primary persistence.
- [x] Docker Compose deployment foundation implemented and runtime-verified.
- [x] `GET /health` and readiness support established.
- [x] Deployment Foundation merged through PR #1.
- [x] One application user may have multiple capabilities.
- [x] Rider is the default capability.
- [x] User domain is independent from external identity records.
- [x] External identities map to application users separately from the user domain.
- [x] Application-owned authentication API implemented.
- [x] Product-owned authentication UI boundary established.
- [x] Client-to-Kratos direct authentication flow rejected; providers remain internal.
- [x] Ory Kratos implemented behind authentication and identity adapters.
- [x] Registration, verification, login, authenticated `GET /v1/me`, logout, and session extension locally verified.
- [x] User Entry and Rider Foundation merged through PR #2.
- [x] MVP session strategy recorded in ADR-0005.

---

## Authentication Boundary Hardening

Active branch: `refactor/auth-provider-boundaries`

Goals before closing the current slice completely:

- [x] Replace provider flow terminology in the public contract with application-owned verification terminology.
- [x] Prevent provider-defined HTTP statuses/messages from leaking through the application API.
- [x] Make provider selection explicit at the composition root with `AUTH_PROVIDER`.
- [x] Make the durable identity-source label configuration-driven instead of hard-coded to a vendor name.
- [x] Record the replaceable-provider architecture and migration limitations in ADR-0006.
- [x] Add provider-independent HTTP contract tests using a fake authentication provider.
- [ ] Run the full Go test suite for the branch.
- [ ] Runtime-verify the Docker Compose authentication flow after these contract changes.

### Architectural rule

External systems implement application-defined ports. External-system concepts must not define business-domain models or public application APIs.

For authentication:

`Client → Application Authentication API → Authentication domain/ports → Provider adapters → Ory Kratos (current implementation)`

A future provider replacement should primarily require a new adapter and composition/configuration change.

Provider-issued subjects are not guaranteed to remain stable across providers. If a future provider assigns different subjects, account continuity requires an explicit identity migration/linking procedure. Active sessions may also be invalidated during such a migration; preserving them is not an MVP requirement.

---

## MVP Delivery Direction

The core MVP journey remains:

`Open app → Authenticate → Load user → Enter Rider → Set pickup → Set destination → Request ride → Basic matching → Driver accepts/rejects → Trip in progress → Live status/location → Trip completed → Trip history`

Later slice boundaries remain intentionally flexible and are refined only when the current slice is complete.

---

## Immediate Next Step

Finish verification of `refactor/auth-provider-boundaries`. Once merged, refine and start `Ride Request Foundation`.

---

## Deferred

- CI/CD
- Kubernetes
- iOS implementation
- Courier and Freight capabilities
- Full administrator operations
- Enterprise dispatch algorithms
- Advanced surge/pricing
- Promotions and incentives
- Advanced fraud/risk systems
- Sophisticated payment infrastructure until concretely required
- Multi-region architecture
- Microservices
- Advanced analytics/data pipelines

---

## Working Principles

- Build the MVP incrementally as vertical slices.
- Define the current milestone precisely and keep only the next milestone reasonably clear.
- Avoid speculative enterprise complexity.
- Organize code around business domains, not global technical layers.
- Keep infrastructure dependencies behind application-owned boundaries.
- Challenge weak or premature architectural decisions.
- Record meaningful progress and the current stopping point here.
