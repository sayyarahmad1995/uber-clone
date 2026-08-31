# Work Log

## Purpose

This file tracks meaningful project progress and the current stopping point so work can resume without reconstructing previous decisions.

Update this file after every meaningful work session.

---

## Current Status

**Current engineering milestone:** Authentication hardening and session-extension correction — merged and verified through PR #12.

**Current stopping point:** PR #11 (verified-identity authentication enforcement) and PR #12 (session-extension contract correction) are merged and runtime-verified. The immediate next business slice remains candidate reselection after Driver rejection so a rejected ride can progress to another eligible Driver instead of remaining stranded.

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
