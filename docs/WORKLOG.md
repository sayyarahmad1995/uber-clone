# Work Log

## Purpose

This file tracks meaningful project progress and the current stopping point so work can resume without reconstructing previous decisions.

Update this file after every meaningful work session.

---

## Current Status

**Current engineering milestone:** Trip Execution Foundation — implemented and verified through PR #16.

**Current stopping point:** An accepted Driver candidate is now atomically promoted into an application-owned Trip. The assigned Driver can start and complete that Trip, completion releases the Driver for future matching, and accepted candidate history remains immutable. Static tests, migration/backfill behavior, authorization, transition/idempotency behavior, persistence, and Driver release/rematch have all been verified. The next business slice should be selected from fresh `main` after PR #16 merges.

**Current MVP planning approach:** Define the current milestone precisely, keep the next business milestone reasonably clear, and intentionally leave later slices flexible until completed work provides new information.

---

## Completed Milestones

- [x] Deployment Foundation — PR #1
- [x] User Entry and Rider Foundation — PR #2
- [x] Replaceable Authentication Provider Boundaries — PR #3
- [x] Shared-account Driver Capability Foundation — PR #4
- [x] Minimal Driver Operational Foundation — PR #5
- [x] Ride Request Foundation — PR #6
- [x] Basic Driver Matching Foundation — PR #7
- [x] API Composition Cleanup — PR #8
- [x] Ride Request required-location hardening — PR #9
- [x] Driver Candidate Accept/Reject Foundation — PR #10
- [x] Verified Identity Authentication Enforcement — PR #11
- [x] Session Extension Contract Correction — PR #12
- [x] Candidate Reselection After Driver Rejection — PR #14
- [x] Driver Active-Candidate Exclusivity Across Rides — PR #15
- [x] Trip Execution Foundation — PR #16

---

## Ride Request Required-Location Hardening

Merged through PR #9.

### Contract

Ride request pickup/destination objects and each latitude/longitude field are required at the HTTP transport boundary. Presence is distinguished from numeric zero using pointer-backed request DTOs; the Ride domain remains transport-neutral.

### Verification completed

- `go test ./...` passes.
- `go vet ./...` passes.
- Docker image builds successfully.
- Docker Compose starts successfully.
- PostgreSQL is healthy.
- API starts successfully.
- Kratos migration exits successfully with code 0.
- Valid ride request returns `201 Created`.
- Missing pickup returns `400 Bad Request`.
- Missing destination returns `400 Bad Request`.
- Missing pickup latitude/longitude returns `400 Bad Request`.
- Missing destination latitude/longitude returns `400 Bad Request`.
- Explicitly supplied `(0,0)` coordinates remain valid and return `201 Created`.

---

## Driver Candidate Accept/Reject Foundation

Merged through PR #10.

### Business flow

`Matched candidate → matched Driver accepts or rejects → application-owned candidate decision persists`

### Implemented contract

- Candidate lifecycle states: `pending`, `accepted`, `rejected`.
- Candidate decisions persist `decided_at`.
- Candidate decision rows are serialized with `FOR UPDATE`.
- Repeating the same decision is idempotent and preserves the original `decided_at`.
- Attempting the opposite decision after resolution returns conflict.
- Candidate ownership is derived from the authenticated application User with Driver capability; no client-supplied Driver ID is accepted.
- A Driver cannot act on another Driver's candidate; the API returns `404 Not Found` to avoid assignment leakage.
- Driver-scoped endpoints:
  - `POST /v1/driver/ride-requests/{ride_request_id}/accept`
  - `POST /v1/driver/ride-requests/{ride_request_id}/reject`

### Verification completed

- `go test ./...` passes.
- `go vet ./...` passes.
- Docker image builds successfully.
- Docker Compose starts successfully.
- PostgreSQL is healthy.
- API starts successfully.
- Kratos migration exits successfully with code 0.
- Matched Driver accept returns `200 OK`, `status=accepted`, non-null `decided_at`.
- Repeated accept returns `200 OK` with unchanged `decided_at`.
- Reject after accept returns `409 Conflict`.
- Another Driver attempting the decision returns `404 Not Found`.
- Matched Driver reject returns `200 OK`, `status=rejected`, non-null `decided_at`.
- Repeated reject returns `200 OK` with unchanged `decided_at`.
- Accept after reject returns `409 Conflict`.
- Unauthenticated decision returns `401 Unauthorized`.
- Rider-only account returns `403 Forbidden`.
- After PR #9 merged, PR #10 was refreshed onto the new `main`; the combined tree again passed `go test ./...`, `go vet ./...`, Docker build, Compose startup, PostgreSQL health, API startup, and Kratos migration.

### Deliberately deferred

- Candidate reselection after rejection.
- Driver reservation/exclusivity across different rides.
- Ride/trip execution state.
- Driver location and proximity matching.
- Pricing and payments.
- Live tracking.

---

## Verified Identity Authentication Enforcement

Merged through PR #11.

### Authentication contract

`Register → unverified → login denied → verify identity → login succeeds → authenticated APIs succeed`

### Implemented behavior

- Unverified login is denied with `403 Forbidden` and application-owned `verification_required`.
- A provider session created during an unverified login attempt is revoked best-effort before any token can be returned.
- Session extension also requires a verified identity.
- Protected application APIs reject stale/unverified sessions with `401 Unauthorized`.
- Kratos verification state remains inside the authentication/identity adapters; it does not become part of the public User model.
- OIDC replacement remains provider-neutral through the application-owned verification contract.
- Verification recovery uses the existing `POST /v1/auth/verify` endpoint to initiate a fresh verification challenge for the same registered email.

### Verification completed

- Unverified login returns `403 verification_required` and does not leak an access token.
- Restarting verification with the registered email returns a fresh application-owned `verification_id`.
- Verification completion, subsequent login, and verified protected access succeed.
- Unverified/stale sessions are rejected on protected APIs.

### Deliberately deferred

- Application-owned resend cooldowns.
- Verification resend rate limiting.
- Provider-specific resend timers or throttling details in the public API.

For MVP, OTP/verification-flow expiry and provider-side throttling remain provider responsibilities. If application-level abuse protection is added later, it should remain provider-neutral and use application-owned errors such as `429 Too Many Requests` rather than exposing Kratos-specific mechanics.

---

## Session Extension Contract Correction

Merged through PR #12.

### Defect corrected

`POST /v1/auth/session/extend` previously could return `200 OK` even when the provider had not advanced the persisted session expiry, causing `expires_in` to continue decreasing while the API implied that extension succeeded.

### Application contract

- Valid, verified session and provider advances `expires_at` → `200 OK` with refreshed `expires_in`.
- Valid, verified session but provider does not advance `expires_at` → `409 Conflict` with application-owned `session_not_extendable`.
- Expired/invalid session → `401 Unauthorized` with `invalid_credentials`.
- Provider-specific HTTP status behavior does not define the public application contract.

### Adapter invariant

After a successful provider extension response, the adapter re-reads the provider session and requires the new `expires_at` to be strictly later than the pre-extension value. A successful provider HTTP status with unchanged expiry is treated as not extendable rather than as application success.

### Verification completed

- `go test ./...` passes.
- `go vet ./...` passes.
- Runtime verification with `SESSION_LIFESPAN=2m` and `SESSION_EARLIEST_POSSIBLE_EXTEND=60s`:
  - Before 60 seconds: repeated extension attempts return `409 session_not_extendable`.
  - After the eligibility interval: extension returns `200 OK` and resets `expires_in` to approximately 119 seconds.
  - After the final extended session expires: extension returns `401 invalid_credentials`.

---

## Candidate Reselection After Driver Rejection

Merged through PR #14.

### Business flow

`Requested ride → candidate Driver rejects → rejected attempt retained → matching excludes prior Drivers → next eligible Driver selected`

### Implemented contract

- Candidate history is retained per `(ride_request_id, driver_user_id)`.
- A ride has at most one active `pending` or `accepted` candidate.
- Repeated matching while a candidate is pending or accepted is idempotent and returns that candidate.
- After rejection, matching excludes every Driver previously attempted for that ride and selects the next eligible Driver using deterministic ordering.
- Rejected Drivers remain eligible for different rides; rejection history is ride-scoped.
- When all eligible Drivers have rejected a ride, matching returns `409 Conflict` with `no eligible driver available`.
- Ride-level `FOR UPDATE` locking remains the serialization point for concurrent match attempts.
- Existing HTTP and domain contracts remain unchanged.

### Verification completed

- `go test ./...` passes.
- `go vet ./...` passes.
- Docker/Compose runtime and migration startup succeed after using the repository's valid Go-style Kratos duration configuration.
- Rejection followed by rematch selects a different Driver.
- Multiple sequential rejections progress through different eligible Drivers without reselection.
- Pending candidate matching is idempotent.
- Accepted candidate matching is idempotent.
- Exhausting all four eligible Drivers returns `409 no eligible driver available`.
- A Driver rejected on one ride remains eligible on a different ride.
- While D1 is the pending candidate, D2/D3/D4 cannot accept the ride and receive `404 ride request candidate not found`; database inspection confirms only one pending candidate row.
- Ten concurrent match calls against an existing pending candidate all return the same Driver and original `created_at`.
- Ten concurrent first-match calls against a ride with zero candidates produce exactly one `201 Created` and nine `200 OK` responses, all returning the same Driver and `created_at`; database inspection confirms exactly one pending candidate row.
- Concurrent first matches for two different riders/rides can select the same Driver, which motivated PR #15.

---

## Driver Active-Candidate Exclusivity Across Rides

Merged and verified through PR #15.

### Business flow

`Concurrent ride requests → matching reserves different eligible Drivers → one Driver cannot hold active candidates for multiple rides`

### Implemented contract

- Drivers with an active candidate are excluded from matching for other rides.
- Before Trip Execution Foundation, `pending` and `accepted` candidates were both indefinitely active; PR #16 refines this so `pending`, and `accepted` with `released_at IS NULL`, are active.
- Candidate/assignment activity remains the source of truth; no speculative global Driver busy-state was introduced.
- Driver profile rows are selected with `FOR UPDATE ... SKIP LOCKED` to serialize cross-ride matching without blocking on a Driver another matching transaction is reserving.
- A partial unique index on `driver_user_id` enforces the persistence invariant that one Driver can hold at most one active ride assignment.
- Rejected candidate history remains ride-scoped and does not globally exclude the Driver.
- Rejection releases the Driver for matching on another ride.
- Trip completion now releases an accepted Driver while preserving the accepted candidate as immutable assignment history.
- Existing matching HTTP/domain contracts remain unchanged except that acceptance is now owned atomically by the Trip application boundary.

### Verification completed

- `go test ./...` passes.
- `go vet ./...` passes.
- Migration `008_driver_active_candidate_exclusivity.sql` applies successfully on invariant-clean data.
- The migration refuses pre-existing duplicate active candidates rather than silently rewriting them as rejected decisions.
- Docker image builds and Compose startup succeed with migration `008` recorded in `schema_migrations`.
- Two concurrent first-match requests for different riders/rides return `201 Created` with different Drivers.
- Database inspection confirms no Driver has more than one active candidate.
- Rejecting D1 releases D1; a subsequent ride can select D1 again.
- Accepting D2 keeps D2 reserved; a later ride skips D2 and selects another available Driver.
- With all four Drivers actively reserved, concurrent matching returns `409 no eligible driver available` rather than double-assigning a Driver.
- After freeing exactly one Driver, ten concurrent initial match calls for the same ride produce exactly one `201 Created` and nine `200 OK` responses, all returning the same Driver and identical `created_at`; database inspection confirms exactly one pending row.

### Deliberately deferred at PR #15

- Candidate timeout/expiry.
- Trip execution state.
- Live Driver/rider location tracking.
- Geographic/proximity matching.
- Dispatch queues.
- Pricing/payments.

---

## Trip Execution Foundation

Implemented and verified through PR #16.

### Business flow

`Matched Driver accepts → assigned Trip exists → assigned Driver starts → Trip becomes in_progress → assigned Driver completes → Trip becomes completed → Driver becomes eligible for matching again`

### Implemented contract

- Trip lifecycle states are `assigned`, `in_progress`, and `completed`.
- Trip identity reuses the application-owned `ride_request_id` for this MVP.
- Candidate acceptance and Trip creation happen atomically; matching no longer exposes an independent acceptance path that could bypass Trip creation.
- Accepted candidate history remains `accepted`; completion does not rewrite acceptance as another decision.
- `released_at` on the accepted candidate marks when that assignment stops reserving the Driver.
- Driver active-candidate exclusivity covers `pending` candidates and `accepted` candidates whose `released_at` is null.
- Migration `009_trip_execution_foundation.sql` backfills existing accepted candidates into `assigned` Trips because assignment is derivable from the recorded acceptance; it does not fabricate start or completion events.
- Only the assigned authenticated Driver can start or complete a Trip.
- Repeated accept, start, and complete operations are idempotent and preserve their original timestamps.
- Complete-before-start and start-after-completion return conflict.
- Wrong-Driver Trip mutation returns not found to avoid assignment leakage.
- Driver-scoped execution endpoints:
  - `POST /v1/driver/ride-requests/{ride_request_id}/accept`
  - `POST /v1/driver/ride-requests/{ride_request_id}/start`
  - `POST /v1/driver/ride-requests/{ride_request_id}/complete`

### Verification completed

- `go test ./...` passes.
- `go vet ./...` passes.
- Docker image builds successfully and Docker Compose starts successfully.
- PostgreSQL is healthy, Kratos migration succeeds, and the API starts on port 8080.
- Migration `009_trip_execution_foundation.sql` is recorded in `schema_migrations`.
- Existing accepted candidates are backfilled as `assigned` Trips with null `started_at`, `completed_at`, and `released_at`.
- The active-Driver partial unique index now constrains `pending`/`accepted` candidates only while `released_at IS NULL`.
- Fresh matching creates a candidate with `201 Created`.
- First accept returns `200 OK`; repeated accept returns the same candidate `created_at` and `decided_at`.
- Complete-before-start returns `409 Conflict`.
- Another Driver attempting to start the Trip returns `404 Not Found`.
- First start returns `in_progress`; repeated start preserves the same `started_at`.
- First complete returns `completed`; repeated complete preserves the same `completed_at`.
- Start after completion returns `409 Conflict`.
- Database inspection confirms the completed Trip retains candidate status `accepted` and sets `released_at` equal to `completed_at`.
- A fresh ride immediately matches the completed Trip's Driver again with `201 Created`, directly proving completion releases the Driver for future matching.

### Deliberately deferred

- Live Driver/rider location updates.
- Route progress and ETA.
- Pricing and fare calculation.
- Payments.
- Cancellations.
- Ratings and receipts.
- Maps integration.
- Rider-facing trip history.

---

## Next Business Vertical Slice

Select the next slice from fresh `main` after PR #16 merges. The first concrete user-visible capability that consumes the completed Trip lifecycle is likely one of:

- Rider/Driver trip status retrieval.
- Minimal trip history.
- Live operational location/status updates.

The exact boundary remains intentionally undecided until the merged model is reviewed. Do not introduce pricing, payments, route optimization, cancellation policy, ETA calculation, or dispatch sophistication until a concrete MVP slice requires them.

---

## Architecture Rule: Replaceable External Providers

External systems implement application-defined ports. Provider-specific concepts must not define business models or public APIs.

For authentication:

`Client → Application Authentication API → Authentication domain → Provider adapter → Identity infrastructure`

The concrete identity provider is selected at the composition root. Kratos is the current implementation, not a business-domain dependency.

The same boundary rule applies to future maps, routing, payments, notifications, storage, messaging, and other external services.

---

## Rough MVP Direction After Trip Execution Foundation

- Trip status/read models and history.
- Live location/status updates.
- Trip lifecycle hardening where concrete product requirements demand it.

Exact later boundaries remain intentionally flexible.

---

## Deferred

- CI/CD.
- Kubernetes.
- iOS implementation.
- Courier capability.
- Freight capability.
- Administrator operations.
- Enterprise-level business logic.
- Advanced matching/dispatch.
- Payments unless concretely required by an MVP slice.
- Promotions.
- Advanced analytics.
- Authentication resend cooldown/rate-limit policy beyond provider defaults.

---

## Important Working Principles

- Build the MVP incrementally using business vertical slices.
- Finish the current slice before expanding into the next one.
- Keep authentication shared across capabilities; do not duplicate identity systems per business role.
- Represent capability membership separately from capability-specific profile/operational data.
- Organize code around business domains with infrastructure behind application-defined boundaries.
- Maintain future extensibility primarily through clean ports/adapters, not speculative implementations.
- Challenge architectural decisions that create vendor lock-in or premature complexity.
- Keep the repository worklog aligned with the actual merged state.
