# Work Log

## Purpose

This file tracks meaningful project progress and the current stopping point so work can resume without reconstructing previous decisions.

Update this file after every meaningful work session.

---

## Current Status

**Current engineering milestone:** User Entry and Rider Foundation — Completed; provider-boundary hardening fully verified and ready for merge

**Current work item:** Merge PR #3, then start Ride Request Foundation.

**Current MVP planning approach:** Define the current milestone precisely, keep the next business milestone reasonably clear, and intentionally leave later slices flexible until we learn from completed work.

---

## Completed

- [x] Repository created: `sayyarahmad1995/uber-clone`
- [x] MVP-first development approach established
- [x] Business-domain-oriented modular monolith selected
- [x] Shared Flutter mobile application direction established
- [x] Android selected as the initial client platform
- [x] Go selected for the backend
- [x] PostgreSQL selected as the primary database
- [x] Redis reserved for concrete fast/transient-data needs
- [x] Docker Compose selected for the current deployment mechanism
- [x] Deployment Foundation implemented, runtime-verified, and merged
- [x] Application-owned authentication API established
- [x] Ory Kratos isolated behind internal authentication/identity adapters
- [x] Direct client-to-provider authentication flow superseded
- [x] PostgreSQL-backed User domain implemented
- [x] External identities separated from application users
- [x] Rider capability created by default and idempotently
- [x] Registration, verification, login, authenticated `/v1/me`, logout, and session extension implemented and locally verified
- [x] User Entry and Rider Foundation merged through PR #2
- [x] Replaceable-provider architecture defined in ADR-0006
- [x] Application-owned identity namespace selected (`primary-identity-v1`)
- [x] Public verification terminology changed from provider `flow` language to application-owned `verification_id`
- [x] Provider errors translated to application-owned authentication errors
- [x] Provider selection made explicit at the composition root through `AUTH_PROVIDER`
- [x] Kratos session lifespan and extension window made configurable through deployment configuration
- [x] Go module metadata completed with committed `backend/go.sum`
- [x] Full backend test suite passed locally with `go test ./...`
- [x] Static analysis passed locally with `go vet ./...`
- [x] Complete Docker Compose stack starts successfully
- [x] Registration runtime-verified with application-owned `verification_id`
- [x] Breached-password runtime-verified as `password_rejected` with useful user-facing message
- [x] Invalid-email runtime-verified as `registration_invalid`
- [x] Duplicate identifier runtime-verified as `identifier_already_exists`
- [x] Verification completion runtime-verified through `POST /v1/auth/verify/complete`
- [x] Login runtime-verified with stable application session contract
- [x] Authenticated `GET /v1/me` runtime-verified and returns Rider capability
- [x] Session extension runtime-verified through `POST /v1/auth/session/extend`; the current token remains valid afterward
- [x] Logout runtime-verified through `POST /v1/auth/logout`
- [x] Logged-out token runtime-verified as invalid through `GET /v1/me` returning `401 Unauthorized`

---

## Current Boundary-Hardening Branch

Active branch: `refactor/auth-provider-boundaries`.

PR: #3 — Harden replaceable authentication provider boundaries.

This branch combines the provider-boundary cleanup with the work previously developed on `feature/configurable-session-lifecycle`.

Session lifecycle defaults:

- `SESSION_LIFESPAN=336h`
- `SESSION_EARLIEST_POSSIBLE_EXTEND=24h`

The values are deployment configuration for the current Kratos adapter and do not become application-domain concepts.

Verification gate: **complete**. The branch is ready for merge.

---

## Architecture Rule: Replaceable External Providers

External systems implement application-defined ports. Provider-specific concepts must not define business models or public APIs.

For authentication:

`Client → Application Authentication API → Authentication domain → Provider adapter → Identity infrastructure`

The concrete identity provider is selected at the composition root. Kratos is the current implementation, not a business-domain dependency.

An external identity maps to an application user using an application-owned identity source plus the provider subject. Replacing the provider may still require explicit subject migration/account linking if the new provider assigns different identifiers.

The same boundary rule applies to future maps, routing, payments, notifications, storage, messaging, and other external services.

---

## Next Business Vertical Slice

**Ride Request Foundation**

Expected end-to-end outcome:

`Authenticated Rider → Set pickup → Set destination → Create persisted ride request`

The slice should stay deliberately narrow. Driver matching, live location, pricing, payments, and other later workflow concerns should not enter unless they become concrete dependencies of this slice.

---

## Rough MVP Direction After Ride Request

- Minimum Driver participation required for the ride flow
- Basic matching
- Driver accept/reject
- Trip execution
- Live location/status updates
- Trip completion
- Trip history

Exact later boundaries remain intentionally flexible.

---

## Deferred

- CI/CD
- Kubernetes
- iOS implementation
- Courier capability
- Freight capability
- Administrator operations
- Enterprise-level business logic
- Advanced matching/dispatch
- Payments unless concretely required by an MVP slice
- Promotions
- Advanced analytics

---

## Important Working Principles

- Build the MVP incrementally using business vertical slices.
- Keep the current milestone precise and the next milestone reasonably clear.
- Do not design later slices in detail prematurely.
- Keep the MVP simple and avoid speculative enterprise complexity.
- Organize code around business domains with infrastructure behind application-defined boundaries.
- Maintain future extensibility primarily through clean ports/adapters, not unused implementations.
- Challenge architectural decisions that create vendor lock-in or premature complexity.
- Update this worklog after meaningful work so the repository remains the source of truth for the stopping point.
