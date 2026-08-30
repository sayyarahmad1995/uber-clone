# Work Log

## Purpose

This file tracks meaningful project progress and the current stopping point so work can resume without reconstructing previous decisions.

Update this file after every meaningful work session.

---

## Current Status

**Current engineering milestone:** Driver Capability Foundation — In Progress

**Current work item:** Verify the shared-account Driver capability implementation, then runtime-test capability activation and `/v1/me`. Ride Request Foundation remains blocked until Driver Capability Foundation is complete.

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

---

## Current Driver Capability Branch

Active branch: `feature/driver-capability-foundation`.

### Architectural direction

Driver is a capability of the existing application user, not a separate authentication identity or separate authentication system.

The authentication flow remains shared:

`Client → Application Authentication API → Authentication domain → Provider adapter → Identity infrastructure`

After authentication, the application determines which business capabilities the user has. One account may hold multiple capabilities. Rider remains the default capability; Driver is added through an application-owned capability activation endpoint.

We should not create driver-specific copies of registration, verification, login, session extension, or logout endpoints unless a concrete business requirement later proves they are necessary.

### Current minimal Driver contract

`Authenticated User → PUT /v1/me/capabilities/driver → Same User now has rider + driver → GET /v1/me exposes both`

The existing database already permits the `driver` capability, so this slice does not require another migration.

A separate Driver profile is intentionally **not** introduced merely to represent capability membership. Driver-specific data such as licensing, vehicle details, approval state, documents, availability, or operational status should be added only when a concrete Driver workflow requires those concepts.

### Remaining completion gate

Before this milestone is considered complete:

1. `go test ./...` passes on the branch.
2. `go vet ./...` passes on the branch.
3. Docker Compose starts successfully with the branch.
4. A Rider-only authenticated account returns only `rider` from `GET /v1/me` before activation.
5. `PUT /v1/me/capabilities/driver` returns the same user with both `driver` and `rider` capabilities.
6. Repeating Driver activation is idempotent and does not duplicate capability data.
7. A subsequent `GET /v1/me` exposes both capabilities consistently.
8. Unauthenticated Driver activation is rejected.

---

## Architecture Rule: Replaceable External Providers

External systems implement application-defined ports. Provider-specific concepts must not define business models or public APIs.

For authentication:

`Client → Application Authentication API → Authentication domain → Provider adapter → Identity infrastructure`

The concrete identity provider is selected at the composition root. Kratos is the current implementation, not a business-domain dependency.

An external identity maps to an application user using an application-owned identity source plus the provider subject. Replacing the provider may still require explicit subject migration/account linking if the new provider assigns different identifiers.

The same boundary rule applies to future maps, routing, payments, notifications, storage, messaging, and other external services.

---

## Next Business Vertical Slice After Driver Capability

**Ride Request Foundation**

Entry condition: Rider and Driver capabilities are both established and runtime-verified.

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

- Driver-specific profile data until a concrete workflow needs it
- Driver licensing and document verification
- Driver approval/background-check workflows
- Driver vehicle details until required by a later slice
- Driver availability/online status until matching requires it
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
- Represent capability membership separately from future capability-specific profile/operational data.
- Keep the MVP simple and avoid speculative enterprise complexity.
- Organize code around business domains with infrastructure behind application-defined boundaries.
- Maintain future extensibility primarily through clean ports/adapters, not unused implementations.
- Challenge architectural decisions that create vendor lock-in or premature complexity.
- Update this worklog after meaningful work so the repository remains the source of truth for the stopping point.
