# Work Log

## Purpose

This file tracks meaningful project progress and the current stopping point so work can resume without reconstructing previous decisions.

Update this file after every meaningful work session.

---

## Current Status

**Current engineering milestone:** Trip Execution Foundation — implemented and verified through PR #16.

**Current stopping point:** An accepted Driver candidate is now atomically promoted into an application-owned Trip. The assigned Driver can start and complete that Trip, completion releases the Driver for future matching, and accepted candidate history remains immutable. Static tests, migration/backfill behavior, authorization, transition/idempotency behavior, persistence, and Driver release/rematch have all been verified.

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
- Docker image builds successfully and Compose starts successfully.
- Migration `009_trip_execution_foundation.sql` is recorded in `schema_migrations`.
- Two pre-existing accepted candidates were backfilled as `assigned` Trips with null `started_at`, `completed_at`, and `released_at`.
- The active-Driver partial unique index now requires `released_at IS NULL` for active `pending`/`accepted` candidates.
- Fresh matching created a candidate with `201 Created`.
- First accept returned `200 OK`; repeated accept returned the same candidate `created_at` and `decided_at`.
- Complete-before-start returned `409 Conflict`.
- Another Driver attempting to start the Trip returned `404 Not Found`.
- First start returned `in_progress`; repeated start preserved the same `started_at`.
- First complete returned `completed`; repeated complete preserved the same `completed_at`.
- Start after completion returned `409 Conflict`.
- Database inspection confirmed the completed Trip retained candidate status `accepted` and set `released_at` equal to `completed_at`.
- A fresh ride immediately matched the completed Trip's Driver again with `201 Created`, directly proving completion releases the Driver for future matching.

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

## Recent Foundation Invariants

### Driver candidate decisions

- Candidate lifecycle states are `pending`, `accepted`, and `rejected`.
- Candidate decisions persist `decided_at`.
- Repeating the same decision is idempotent and preserves the original timestamp.
- Opposite decisions after resolution conflict.
- Candidate ownership derives from the authenticated Driver capability; clients never supply Driver identity.
- Wrong/unassigned Drivers receive `404 Not Found`.

### Candidate reselection

- Candidate history is retained per `(ride_request_id, driver_user_id)`.
- A ride has at most one active candidate.
- Matching while a candidate is pending or accepted is idempotent.
- After rejection, matching excludes Drivers previously attempted for that ride.
- Rejected Drivers remain eligible on other rides.
- Exhaustion returns `409 no eligible driver available`.

### Driver active-assignment exclusivity

- A Driver cannot hold more than one active assignment across rides.
- Active means `pending`, or `accepted` with `released_at IS NULL`.
- Driver profile rows are selected with `FOR UPDATE ... SKIP LOCKED` for cross-ride matching concurrency.
- Persistence enforces the active-Driver invariant with a partial unique index.
- Rejection releases a pending Driver immediately.
- Trip completion releases an accepted Driver while preserving acceptance history.

### Authentication

- External identity systems implement application-defined auth/identity ports.
- Unverified login/session extension is denied by application-owned verification rules.
- Session extension succeeds only when the provider actually advances persisted expiry.
- Provider-specific concepts remain outside the public User/business model.

---

## Architecture Rule: Replaceable External Providers

External systems implement application-defined ports. Provider-specific concepts must not define business models or public APIs.

For authentication:

`Client → Application Authentication API → Authentication domain → Provider adapter → Identity infrastructure`

The concrete identity provider is selected at the composition root. Kratos is the current implementation, not a business-domain dependency.

The same boundary rule applies to future maps, routing, payments, notifications, storage, messaging, and other external services.

---

## Next Business Vertical Slice

The next slice should be selected from the first concrete user-visible capability that consumes the completed Trip lifecycle. Likely candidates are Rider/Driver trip status retrieval, minimal trip history, or live operational location/status updates. The exact boundary remains intentionally undecided until PR #16 is merged and the resulting model is reviewed from fresh `main`.

Do not introduce pricing, payments, route optimization, cancellation policy, ETA calculation, or dispatch sophistication until a concrete MVP slice requires them.

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
