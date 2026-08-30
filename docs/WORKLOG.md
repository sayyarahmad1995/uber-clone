# Work Log

## Purpose

This file tracks meaningful project progress and the current stopping point so work can resume without reconstructing previous decisions.

Update this file after every meaningful work session.

---

## Current Status

**Current engineering milestone:** Driver Capability Foundation — In Progress

**Current work item:** Establish the Driver capability and driver entry/onboarding flow on top of the existing shared authentication and User foundation. Ride Request Foundation remains blocked until both Rider and Driver capabilities are established.

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

---

## Current Driver Capability Branch

Active branch: `feature/driver-capability-foundation`.

### Architectural direction

Driver is a capability of the existing application user, not a separate authentication identity or separate authentication system.

The authentication flow remains shared:

`Client → Application Authentication API → Authentication domain → Provider adapter → Identity infrastructure`

After authentication, the application determines which business capabilities the user has. One account may hold multiple capabilities. Rider remains the default capability; Driver is added through application-owned Driver onboarding/activation.

We should not create driver-specific copies of registration, verification, login, session extension, or logout endpoints unless a concrete business requirement later proves they are necessary.

### Expected end-to-end outcome

`Authenticated User → Enter/enable Driver capability → Create/load Driver profile → Authenticated identity exposes Driver capability`

The exact Driver onboarding data should stay MVP-minimal and be defined from concrete requirements before implementation. Licensing, vehicle details, approval workflows, document verification, background checks, availability, and dispatch behavior should not be added merely because a production ride-hailing system may eventually need them.

### Completion gate

Before this milestone is considered complete:

1. Driver capability semantics are application-owned and coexist cleanly with Rider on one user account.
2. Driver persistence/profile boundaries are defined without coupling business models to an external provider.
3. An authenticated user can obtain/activate the Driver capability through an application API.
4. Driver capability creation is idempotent and appropriately authorized.
5. `/v1/me` (or the application-owned identity contract that replaces it) exposes the resulting capabilities consistently.
6. Provider-independent tests cover the Driver entry/capability contract.
7. `go test ./...` and `go vet ./...` pass.
8. The complete Driver entry flow is runtime-verified with Docker Compose.

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
- Keep the MVP simple and avoid speculative enterprise complexity.
- Organize code around business domains with infrastructure behind application-defined boundaries.
- Maintain future extensibility primarily through clean ports/adapters, not unused implementations.
- Challenge architectural decisions that create vendor lock-in or premature complexity.
- Update this worklog after meaningful work so the repository remains the source of truth for the stopping point.
