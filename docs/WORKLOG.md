# Work Log

## Purpose

This file tracks meaningful project progress and the current stopping point so work can resume without reconstructing previous decisions.

Update this file after every meaningful work session.

---

## Current Status

**Current engineering milestone:** Minimal Driver Operational Foundation — In Progress

**Current work item:** Runtime-verify the corrected Driver onboarding and availability JSON contracts, then complete the remaining authorization and persistence checks before merge.

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
- [x] Provider-boundary hardening merged through PR #3
- [x] Merged `main` post-merge smoke-tested for registration, verification, login, `/v1/me`, session extension, logout, and invalidated-session behavior
- [x] Driver capability added as an application-owned User capability alongside Rider
- [x] Existing `user_capabilities` persistence reused; no Driver-specific identity or migration is required
- [x] Idempotent authenticated `PUT /v1/me/capabilities/driver` endpoint implemented
- [x] Driver capability activation reuses the existing shared user/account and authentication session
- [x] `/v1/me` continues to expose the complete capability set for the account
- [x] Provider-independent User service test added for Driver capability enablement
- [x] Driver runtime flow verified: Rider-only account becomes Rider + Driver on the same user ID
- [x] Repeated Driver activation runtime-verified as idempotent
- [x] Unauthenticated Driver activation runtime-verified as `401 Unauthorized`
- [x] Driver branch backend test suite passed with `go test ./...`
- [x] Driver branch static analysis passed with `go vet ./...`
- [x] Driver branch Docker Compose stack verified healthy; Kratos migration exits successfully with code 0
- [x] Driver Capability Foundation merged through PR #4
- [x] Minimal Driver domain added separately from authentication and User capability membership
- [x] One operational Driver profile and one MVP vehicle modeled per Driver user
- [x] Driver onboarding, retrieval, and online/offline availability APIs implemented
- [x] Driver operational persistence added through migration `003_driver_operations.sql`
- [x] Provider-independent Driver service tests added
- [x] Minimal Driver branch passes `go test ./...` and `go vet ./...`
- [x] Minimal Driver Docker image builds and Compose stack starts with healthy PostgreSQL/Mailpit and successful Kratos migration
- [x] Runtime testing caught HTTP JSON contract mismatch for `license_plate` and `is_online`
- [x] Driver HTTP request DTOs corrected to preserve application-owned nested snake_case contracts
- [x] API-level regression tests added for onboarding `vehicle.license_plate` and availability `is_online`
- [x] Unauthenticated Driver operational endpoints runtime-verified as `401 Unauthorized`
- [x] Availability before onboarding runtime-verified as rejected with `409 Conflict`

---

## Minimal Driver Operational Foundation

Active branch: `feature/minimal-driver-operational-foundation`.

Draft PR: #5 — Add minimal Driver operational foundation.

### Architectural direction

Driver capability membership and Driver operational state are distinct application concepts.

- Authentication remains shared for the account.
- `driver` capability means the user may enter Driver workflows.
- The Driver domain owns operational profile, vehicle, status, and availability data.
- External authentication/provider concepts do not define Driver business models or public APIs.

### Current minimal operational contract

`Authenticated User with Driver capability → provide one MVP vehicle → active Driver starts offline → GET own Driver state → go online/offline`

The MVP vehicle contains only:

- make
- model
- color
- license plate

A valid MVP onboarding becomes `active` immediately. There is no manual or external compliance approval workflow in this slice.

Public request contracts use application-owned JSON:

- onboarding: `{"vehicle":{"make":"...","model":"...","color":"...","license_plate":"..."}}`
- availability: `{"is_online":true|false}`

### Remaining completion gate

Before this milestone is considered complete:

1. Pull the corrected branch and rerun `go test ./...` plus `go vet ./...`.
2. Rebuild/restart the Docker Compose stack with the corrected API binary.
3. A genuinely Rider-only account receives `403 Forbidden` from Driver operational APIs before Driver capability activation.
4. A Driver-capable account can onboard one vehicle and receives `status=active`, `is_online=false`.
5. `GET /v1/driver` returns the persisted operational state.
6. Availability can move online and offline using the public `is_online` contract.
7. Repeating onboarding reuses/updates the same Driver profile rather than duplicating it.
8. Unauthenticated Driver operational requests remain rejected.

Note: `test1@example.com` already had the Driver capability from the previous milestone because Docker Compose retained the PostgreSQL volume, so its pre-activation `GET /v1/driver` correctly reached the Driver domain and returned `404 profile not found`; that account cannot prove the Rider-only `403` case.

### Deliberately deferred

- Driver license/document verification
- Background checks
- Insurance verification
- Manual approval queues
- External compliance providers
- Multiple vehicles
- Earnings/tax onboarding
- Production-grade regulatory onboarding

---

## Architecture Rule: Replaceable External Providers

External systems implement application-defined ports. Provider-specific concepts must not define business models or public APIs.

For authentication:

`Client → Application Authentication API → Authentication domain → Provider adapter → Identity infrastructure`

The concrete identity provider is selected at the composition root. Kratos is the current implementation, not a business-domain dependency.

An external identity maps to an application user using an application-owned identity source plus the provider subject. Replacing the provider may still require explicit subject migration/account linking if the new provider assigns different identifiers.

The same boundary rule applies to future maps, routing, payments, notifications, storage, messaging, and other external services.

---

## Next Business Vertical Slice After Minimal Driver Operational Foundation

**Ride Request Foundation**

Expected end-to-end outcome:

`Authenticated Rider → Set pickup → Set destination → Create persisted ride request`

The slice should stay deliberately narrow. Driver matching, live location, pricing, payments, and other later workflow concerns should not enter unless they become concrete dependencies of this slice.

---

## Rough MVP Direction After Ride Request

- Basic matching using active online Drivers
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
- Keep authentication shared across capabilities; do not duplicate identity systems per business role.
- Represent capability membership separately from capability-specific profile/operational data.
- Keep the MVP simple and avoid speculative enterprise complexity.
- Organize code around business domains with infrastructure behind application-defined boundaries.
- Maintain future extensibility primarily through clean ports/adapters, not unused implementations.
- Challenge architectural decisions that create vendor lock-in or premature complexity.
- Update this worklog after meaningful work so the repository remains the source of truth for the stopping point.
