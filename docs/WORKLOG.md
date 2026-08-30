# Work Log

## Purpose

This file tracks meaningful project progress and the current stopping point so work can resume without reconstructing previous decisions.

Update this file after every meaningful work session.

---

## Current Status

**Current engineering milestone:** Minimal Driver Operational Foundation — In Progress

**Current work item:** Make a Driver-capable account operational with one MVP vehicle, application-owned active status, and online/offline availability. Ride Request Foundation remains next after this slice is complete and runtime-verified.

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
- [x] Existing `user_capabilities` persistence reused; no Driver-specific identity is required
- [x] Idempotent authenticated `PUT /v1/me/capabilities/driver` endpoint implemented
- [x] Driver capability activation reuses the existing shared user/account and authentication session
- [x] `/v1/me` exposes the complete capability set for the account
- [x] Driver capability flow fully runtime-verified
- [x] Driver Capability Foundation merged through PR #4

---

## Driver Capability Foundation

PR #4 — Add shared-account Driver capability foundation — merged into `main`.

Driver remains a capability of the existing application user, not a separate authentication identity or authentication system.

`Authenticated User → PUT /v1/me/capabilities/driver → Same User has rider + driver → GET /v1/me exposes both`

Authentication remains shared across Rider and Driver.

---

## Current Minimal Driver Operational Foundation

Active branch: `feature/minimal-driver-operational-foundation`.

### Why this slice exists

Capability membership alone is enough to establish that an account may act as a Driver, but basic matching will soon need a small amount of application-owned operational state.

This slice intentionally adds only the data the near-term ride flow can consume:

- one Driver operational profile per application user,
- one MVP vehicle per Driver,
- application-owned Driver status,
- online/offline availability.

### Current contract

A user must already have the `driver` capability before using Driver operational APIs.

`Authenticated Driver-capable User → PUT /v1/driver → Active Driver profile + one vehicle → GET /v1/driver → PUT /v1/driver/availability`

Newly onboarded Drivers start offline. For the MVP, valid onboarding immediately creates an `active` Driver because no manual/compliance approval workflow currently exists.

### MVP vehicle data

The current required vehicle fields are deliberately narrow:

- `make`
- `model`
- `color`
- `license_plate`

The onboarding operation is idempotent for the user: repeating it updates the existing MVP vehicle/profile rather than creating duplicate Driver profiles.

### Explicitly deferred

Do not add these until a concrete workflow requires them:

- license/document verification,
- background checks,
- insurance verification,
- manual approval queues,
- approval-provider integrations,
- multiple vehicles,
- vehicle classes/categories,
- earnings/tax onboarding,
- compliance expiration workflows,
- detailed Driver biography/profile information.

### Completion gate

Before this milestone is complete:

1. `go test ./...` passes.
2. `go vet ./...` passes.
3. Docker Compose applies the Driver operational migration successfully and required services are healthy.
4. A Rider-only account is forbidden from Driver operational APIs until Driver capability is enabled.
5. A Driver-capable account can onboard with the required vehicle data.
6. Onboarding returns the same application user ID, `active` status, one vehicle, and `is_online=false`.
7. Repeating onboarding updates/reuses the same Driver profile rather than creating duplicates.
8. `GET /v1/driver` returns the persisted Driver operational state.
9. Availability can be changed online and offline after onboarding.
10. Availability changes before onboarding are rejected.
11. Unauthenticated Driver operational requests are rejected.

---

## Architecture Rule: Replaceable External Providers

External systems implement application-defined ports. Provider-specific concepts must not define business models or public APIs.

For authentication:

`Client → Application Authentication API → Authentication domain → Provider adapter → Identity infrastructure`

The concrete identity provider is selected at the composition root. Kratos is the current implementation, not a business-domain dependency.

Driver operational state is application-owned and stored in the application database. Future document, background-check, vehicle-data, or compliance providers must implement application-defined boundaries rather than define the Driver business model.

The same boundary rule applies to future maps, routing, payments, notifications, storage, messaging, and other external services.

---

## Next Business Vertical Slice

**Ride Request Foundation**

Entry condition: Rider capability, Driver capability, and Minimal Driver Operational Foundation are established and runtime-verified.

Expected end-to-end outcome:

`Authenticated Rider → Set pickup → Set destination → Create persisted ride request`

The slice should stay deliberately narrow. Driver matching, live location, pricing, payments, and other later workflow concerns should not enter unless they become concrete dependencies of this slice.

---

## Rough MVP Direction After Ride Request

- Driver availability consumed by matching
- Basic matching
- Driver accept/reject
- Trip execution
- Live location/status updates
- Trip completion
- Trip history

Exact later boundaries remain intentionally flexible.

---

## Deferred

- Full production Driver onboarding/compliance
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
- Represent capability membership separately from capability-specific operational data.
- Add Driver operational data only when near-term business flows consume it.
- Keep the MVP simple and avoid speculative enterprise complexity.
- Organize code around business domains with infrastructure behind application-defined boundaries.
- Maintain future extensibility primarily through clean ports/adapters, not unused implementations.
- Challenge architectural decisions that create vendor lock-in or premature complexity.
- Update this worklog after meaningful work so the repository remains the source of truth for the stopping point.
