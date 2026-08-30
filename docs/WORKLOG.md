# Work Log

## Purpose

This file tracks meaningful project progress and the current stopping point so work can resume without reconstructing previous decisions.

Update this file after every meaningful work session.

---

## Current Status

**Current engineering milestone:** Basic Driver Matching Foundation — Complete, pending PR merge

**Current work item:** PR #7 is fully runtime-verified and ready for review/merge.

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
- [x] Rider-only authorization runtime-verified as `403 Forbidden` for Driver operational APIs
- [x] Driver onboarding, persistence, availability, and repeated-onboarding identity behavior fully runtime-verified
- [x] Minimal Driver Operational Foundation merged through PR #5
- [x] Provider-neutral Ride domain added with application-defined repository port
- [x] Ride request persistence added through migration `004_ride_requests.sql`
- [x] Authenticated Rider ownership derived from the application User; no client-supplied rider ID
- [x] `POST /v1/ride-requests` implemented with nested pickup/destination coordinates and `201 Created`
- [x] New ride requests start with application-owned `requested` status
- [x] Latitude/longitude validation implemented in both service and database constraints
- [x] Provider-independent Ride service tests and HTTP JSON contract test added
- [x] Ride branch passes `go test ./...` and `go vet ./...`
- [x] Ride Docker image builds and Compose stack starts successfully
- [x] Authenticated Rider ride creation runtime-verified with persisted coordinates and matching Rider user ID
- [x] Invalid ride coordinates runtime-verified as `400 Bad Request`
- [x] Unauthenticated ride creation runtime-verified as `401 Unauthorized`
- [x] Ride Request Foundation merged through PR #6
- [x] Provider-neutral Matching domain added with application-defined repository port
- [x] Driver assignment candidate persistence added through migration `005_ride_driver_candidates.sql`
- [x] `POST /v1/ride-requests/{ride_request_id}/match` implemented for Rider-owned ride requests
- [x] Matching eligibility limited to Driver-capable profiles with an operational vehicle, `status=active`, and `is_online=true`
- [x] Dual-capability Rider self-matching explicitly excluded
- [x] Deterministic provider-free MVP selection implemented without claiming geographic proximity
- [x] At most one candidate persisted per ride request; repeated matching reuses the same candidate
- [x] Matching branch passes `go test ./...` and `go vet ./...`
- [x] Matching Docker image builds and Compose stack starts successfully
- [x] No eligible online Driver runtime-verified as `409 Conflict`
- [x] First eligible online Driver match runtime-verified as `201 Created`
- [x] Repeated matching runtime-verified as `200 OK` with the same candidate and creation time
- [x] Another Rider cannot match a ride request they do not own; runtime-verified as `404 Not Found`
- [x] Dual-capability Rider cannot match its own online Driver profile; runtime-verified as `409 Conflict` when no other Driver is eligible
- [x] Unauthenticated matching runtime-verified as `401 Unauthorized`

---

## Basic Driver Matching Foundation

Active branch: `feature/basic-driver-matching-foundation`.

Draft PR: #7 — Add Basic Driver Matching Foundation.

### Architectural direction

Matching is an application-owned business capability. It consumes existing Ride and Driver state but does not let provider-specific concepts define the model.

- The Rider can only initiate matching for a ride request they own.
- Eligible Drivers must retain Driver capability, have an operational vehicle, be `active`, and be online.
- A user with both Rider and Driver capabilities cannot match their own ride request to their own Driver profile.
- Driver location is not modeled yet, so this milestone does not claim nearest-Driver or geographic matching.
- Candidate persistence is idempotent: one ride request has at most one Driver candidate in this slice.

### Completed minimal contract

`Requested Rider-owned ride → find eligible active online Driver → persist one Driver assignment candidate`

Public endpoint:

`POST /v1/ride-requests/{ride_request_id}/match`

First successful matching returns `201 Created`; repeated matching for the same ride returns `200 OK` with the same candidate.

### Verification completed

1. `go test ./...` passes.
2. `go vet ./...` passes.
3. Docker image builds successfully and Compose starts successfully.
4. With no eligible online Driver, matching returns `409 Conflict`.
5. An eligible active Driver can be brought online and selected.
6. First successful match returns `201 Created`.
7. Returned `ride_request_id` matches the Rider-owned request.
8. Returned `driver_user_id` matches the eligible Driver.
9. Repeated matching returns `200 OK` with the same candidate and `created_at`.
10. Another Rider receives `404 Not Found` when attempting to match a ride they do not own.
11. A dual-capability Rider is excluded from matching its own Driver profile.
12. Unauthenticated matching returns `401 Unauthorized`.

### Deliberately deferred

- Driver location
- nearest/geographic Driver selection
- maps and routing providers
- Driver reservation/exclusivity across concurrent rides
- Driver accept/reject
- candidate rejection/reselection/history
- pricing and payments
- live tracking
- trip execution and completion

---

## Architecture Rule: Replaceable External Providers

External systems implement application-defined ports. Provider-specific concepts must not define business models or public APIs.

For authentication:

`Client → Application Authentication API → Authentication domain → Provider adapter → Identity infrastructure`

The concrete identity provider is selected at the composition root. Kratos is the current implementation, not a business-domain dependency.

An external identity maps to an application user using an application-owned identity source plus the provider subject. Replacing the provider may still require explicit subject migration/account linking if the new provider assigns different identifiers.

The same boundary rule applies to future maps, routing, payments, notifications, storage, messaging, and other external services.

---

## Next Business Vertical Slice After Basic Driver Matching Foundation

**Driver Accept/Reject Foundation**

Expected end-to-end outcome:

`Matched candidate → assigned Driver views candidate → Driver accepts or rejects → application-owned ride/matching state changes`

This next slice should introduce only the state transitions and ownership rules concretely required by Driver response. Candidate reselection after rejection may be included only if it is necessary to complete the end-to-end accept/reject flow.

---

## Rough MVP Direction After Driver Response

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
