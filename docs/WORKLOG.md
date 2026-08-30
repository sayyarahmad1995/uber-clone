# Work Log

## Purpose

This file tracks meaningful project progress and the current stopping point so work can resume without reconstructing previous decisions.

Update this file after every meaningful work session.

---

## Current Status

**Current engineering milestone:** Driver Candidate Accept/Reject Foundation — merged and verified.

**Current stopping point:** PR #9 (Ride Request required-location hardening) and PR #10 (Driver candidate accept/reject foundation) are merged. The immediate next business slice is candidate reselection after Driver rejection so a rejected ride can progress to another eligible Driver instead of remaining stranded.

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

## Next Business Vertical Slice

**Candidate Reselection After Driver Rejection**

Expected end-to-end outcome:

`Requested ride → candidate Driver rejects → rejected attempt is retained → matching excludes rejected Driver → next eligible Driver can be selected`

The current one-candidate-per-ride model is insufficient for this behavior. The next slice should introduce the minimum candidate-attempt history needed to preserve rejection history and select a subsequent eligible Driver. Avoid introducing dispatch queues, Driver reservations, geographic matching, pricing, trip execution, or other unrelated complexity unless a concrete invariant requires it.

---

## Architecture Rule: Replaceable External Providers

External systems implement application-defined ports. Provider-specific concepts must not define business models or public APIs.

For authentication:

`Client → Application Authentication API → Authentication domain → Provider adapter → Identity infrastructure`

The concrete identity provider is selected at the composition root. Kratos is the current implementation, not a business-domain dependency.

The same boundary rule applies to future maps, routing, payments, notifications, storage, messaging, and other external services.

---

## Rough MVP Direction After Reselection

- Trip execution.
- Live location/status updates.
- Trip completion.
- Trip history.

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
