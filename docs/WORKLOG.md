# Work Log

## Purpose

This file tracks meaningful project progress and the current stopping point so work can resume without reconstructing previous decisions.

Update this file after every meaningful work session.

---

## Current Status

**Current engineering milestone:** Ride Cancellation Foundation — implemented, runtime-verified, and merged through PR #22.

**Current stopping point:** Riders and assigned Drivers can now cancel open Ride Requests through application-owned cancellation semantics. Cancellation terminates the Ride Request, cancels any assigned or in-progress Trip, releases automatic matching commitments, closes pending marketplace offers, preserves historical decisions, and prevents cancelled rides from being matched or assigned again. Rider status reads expose authoritative cancellation metadata and optional Trip state. Completed Trips remain non-cancellable.

**Current MVP planning approach:** Define the current milestone precisely, keep the next business milestone reasonably clear, and intentionally leave later slices flexible until completed work provides new information. The next business slice should be selected from a concrete Rider/Driver client-flow requirement rather than introducing location, maps, notifications, or payment infrastructure speculatively.

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
- [x] Ride Offer Marketplace Foundation — PR #17
- [x] Rider Offer Selection Foundation — PR #18
- [x] Driver Marketplace Discovery Foundation — PR #19
- [x] Rider Ride-Request Status Foundation — PR #21
- [x] Ride Cancellation Foundation — PR #22

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

## Ride Offer Marketplace Foundation

Implemented and verified through PR #17.

### Business flow

`Rider creates offers-mode request with proposed fare → eligible Drivers submit/update bounded offers → Rider lists offers`

This slice intentionally stops before Rider selection and exclusive assignment.

### Implemented contract

- Ride Requests now have application-owned booking modes: `automatic` and `offers`.
- Empty/omitted booking mode defaults to `automatic`, preserving the previous contract.
- `offers` requests require a Rider proposed fare represented as integer minor units plus a three-letter currency.
- Money remains application-owned; no payment-provider concepts are present in Ride or Offer APIs.
- Driver offers are separate from `ride_driver_candidates` and do not reserve Drivers or create Trips.
- A Driver has at most one mutable offer row per Ride Request; updating an offer preserves `created_at` and advances `updated_at`.
- For MVP, an offer must be between 90% and 130% of the Rider proposed fare, with minimum rounding upward and maximum rounding downward.
- Offering Drivers must have Driver capability, an active Driver profile, a vehicle, be online, and have no unreleased pending/accepted candidate assignment.
- A Driver cannot offer on their own Ride Request.
- The automatic matching endpoint rejects `offers` Ride Requests rather than silently dispatching them.
- The owning Rider can list offers, ordered deterministically by amount, creation time, then Driver ID.
- Endpoints:
  - `PUT /v1/driver/ride-requests/{ride_request_id}/offer`
  - `GET /v1/ride-requests/{ride_request_id}/offers`

### Verification completed

- `go test ./...` passes on the final formatted PR head.
- `go vet ./...` passes on the final formatted PR head.
- Docker image builds successfully.
- Docker Compose starts successfully after a clean `docker compose down -v` reset.
- PostgreSQL becomes healthy, Kratos migrations apply successfully, and the API starts.
- Migration `010_ride_offer_marketplace_foundation.sql` is recorded in `schema_migrations`.
- Creating an offers-mode Ride Request with proposed fare `100000 PKR` returns `201 Created` and persists the offers mode and proposed fare.
- Calling automatic `/match` for an offers-mode request returns `409 Conflict` with `ride request uses the offers marketplace`.
- Eligible online Drivers can submit offers at the exact 90% and 130% boundaries (`90000` and `130000`) with `200 OK`.
- Offers immediately outside the bounds (`89999` and `130001`) return `400 Bad Request`.
- Updating a Driver offer from `90000` to `105000` preserves `created_at`, advances `updated_at`, and leaves one row for that Driver/Ride pair.
- Rider offer listing returns the expected two offers ordered by amount.
- Database inspection confirms the two expected `ride_offers` rows and confirms zero `ride_driver_candidates` for the offers-mode Ride Request.
- Creating a Ride Request without `booking_mode` returns `booking_mode=automatic` and no proposed fare.
- Existing automatic matching remains backward compatible and returns `201 Created` with a Driver candidate.
- A stale second-Rider token returned `401 Unauthorized`; the separate authenticated non-owning Rider `404` runtime assertion was not repeated after the clean database reset.

### Deliberately deferred

- Rider selecting an offer.
- Marketplace-driven Trip assignment.
- Driver discovery/feed and geospatial offer targeting.
- Offer expiration/cancellation.
- Back-and-forth negotiation history.
- Platform fare estimation or surge pricing.
- Payments.
- Real-time location.

---

## Rider Offer Selection Foundation

Implemented and verified through PR #18.

### Business flow

`Rider creates offers-mode request → Driver either accepts proposed fare or submits counteroffer → exact proposed-fare acceptance assigns immediately → Rider may accept/reject counteroffers → accepted counteroffer assigns atomically`

### Implemented contract

- Offer decision states are `pending`, `accepted`, `rejected`, and `closed`, with `decided_at` recorded for resolved offers.
- Exact proposed-fare acceptance by an eligible Driver creates an accepted offer and assigned Trip atomically.
- Repeating exact proposed-fare acceptance by the winning Driver is idempotent and returns the original assignment timestamps.
- Another Driver attempting to accept after assignment receives marketplace conflict rather than an automatic-candidate error.
- Counteroffers remain non-reserving until Rider selection.
- The owning Rider can accept or reject a specific pending Driver offer.
- Selecting a counteroffer revalidates Driver capability, active profile, online state, vehicle, active candidate state, and active Trip state while serialized under the marketplace transaction.
- Successful marketplace assignment closes competing pending offers while preserving offer history.
- Marketplace assignment creates no `ride_driver_candidates`; automatic dispatch and marketplace selection converge at the same application-owned Trip boundary.
- Automatic matching excludes Drivers with active Trips, including Trips created through the marketplace.
- A partial unique index permits at most one accepted offer per Ride Request.
- A partial unique index permits at most one active Trip (`assigned` or `in_progress`) per Driver.
- PostgreSQL active-Trip uniqueness races are mapped narrowly to application-owned Driver-unavailable conflict instead of leaking `500 Internal Server Error`.
- Endpoints:
  - existing `POST /v1/driver/ride-requests/{ride_request_id}/accept` supports automatic candidate acceptance and offers-mode proposed-fare acceptance.
  - existing `PUT /v1/driver/ride-requests/{ride_request_id}/offer` assigns immediately when the submitted amount equals the Rider proposed fare.
  - `POST /v1/ride-requests/{ride_request_id}/offers/{driver_user_id}/accept`
  - `POST /v1/ride-requests/{ride_request_id}/offers/{driver_user_id}/reject`

### Verification completed

- `go test ./...` passes on the final implementation head.
- `go vet ./...` passes on the final implementation head.
- Clean Docker startup applies migrations through `011_rider_offer_selection_foundation.sql`.
- Exact proposed-fare Driver acceptance returns `200 OK`, creates exactly one accepted offer and one assigned Trip, and repeated acceptance is idempotent.
- Losing exact-fare acceptance after another Driver wins returns `409 Conflict` with marketplace-not-open semantics, not automatic candidate `404`.
- Rider rejection persists `rejected`; Rider acceptance persists the selected offer as `accepted`, closes the competing pending offer, creates exactly one Trip, and creates zero candidate rows.
- Taking a Driver offline after offer submission causes Rider acceptance to return `409 Conflict`; the pending offer remains unchanged and no Trip is created.
- Automatic matching skips a Driver who already has an active marketplace Trip.
- Marketplace assignment rejects a Driver who already has an active automatic candidate.
- Two Drivers exact-accepting the same offers ride concurrently produce exactly one assignment and no `500`; the losing Driver receives controlled conflict.
- Two Rider selections against different pending offers concurrently produce one `200`, one `409`, exactly one accepted offer, one closed competing offer, and one Trip.
- The same Driver exact-accepting two independent marketplace rides concurrently produces one `200`, one `409`, one active Trip, and no partial loser state.
- Marketplace acceptance racing with automatic matching preserves one active commitment per Driver; the automatic matcher selects another eligible Driver when appropriate and no cross-strategy double commitment occurs.
- Marketplace-created Trips start and complete through the existing Trip execution endpoints, and completion releases the Driver for future work.

### Deliberately deferred

- Driver discovery/feed and geospatial targeting.
- Offer expiration/cancellation.
- Rider cancellation policy.
- Back-and-forth negotiation history.
- Platform fare estimation/surge.
- Payments.
- Live location.

---

## Driver Marketplace Discovery Foundation

Implemented, verified, and merged through PR #19.

### Business flow

`Eligible online Driver opens marketplace → application lists open offers-mode Ride Requests → Driver chooses a Ride Request → Driver accepts proposed fare or submits/updates a counteroffer through the existing marketplace contract`

### Implemented contract

- Driver marketplace discovery endpoint: `GET /v1/driver/marketplace/ride-requests`.
- Discovery is read-only and does not reserve Drivers, mutate Ride Requests, create candidates, create offers, or create Trips.
- Only Ride Requests with `booking_mode=offers`, `status=requested`, and no Trip are discoverable.
- The authenticated Driver's own Ride Requests are excluded, preserving shared-account Rider/Driver semantics.
- Drivers who are not operationally eligible receive an empty feed. Eligibility requires Driver capability, an active Driver profile, online availability, a vehicle, no unreleased active candidate, and no active Trip.
- Discovery returns pickup, destination, proposed fare, Ride Request creation time, and the authenticated Driver's existing offer when one exists.
- Rider identity is not exposed in the Driver marketplace response.
- Results are deterministic, newest-first, and bounded to 50 Ride Requests for the MVP.
- Discovery is a snapshot only; offer submission and Trip assignment remain authoritative and revalidate eligibility/availability atomically.

### Verification completed

- `go test ./...` passes on the PR head.
- `go vet ./...` passes on the PR head.
- `gofmt` produced no changes.
- Docker image builds successfully and Docker Compose starts the rebuilt API successfully.
- A free eligible Driver sees open offers-mode Ride Requests in newest-first order.
- Automatic-mode Ride Requests are absent from the feed.
- A Driver with an active Trip receives an empty feed and cannot submit another offer; completing the Trip restores discovery eligibility.
- An offline Driver receives an empty feed; returning online restores eligibility when no other commitment exists.
- Submitting a counteroffer keeps the Ride Request discoverable and projects the Driver's pending `own_offer` into the feed.
- Historical pending offers are projected as `own_offer` while their Ride Requests remain open.
- Once a Ride Request is assigned and a Trip exists, it disappears from another eligible Driver's feed.
- A shared account with both Rider and Driver capability does not see its own Rider-created marketplace request in its Driver feed.
- The same self-created request is visible to another free eligible Driver, confirming that self-exclusion is owner-specific rather than a global visibility defect.

### Deliberately deferred

- Live Driver location.
- Geographic filtering and proximity ranking.
- Maps/routing provider integration.
- Push notifications or marketplace subscriptions.
- Cursor pagination.
- Offer expiration/cancellation.
- Pricing/surge.
- Payments.

---

## Rider Ride-Request Status Foundation

Implemented, verified, and merged through PR #21.

### Business flow

`Owning Rider creates Ride Request → Rider reads current Ride Request state → optional Trip assignment/execution state is projected without strategy-specific branching`

### Implemented contract

- Rider status endpoint: `GET /v1/ride-requests/{ride_request_id}`.
- Only the owning authenticated Rider can read the Ride Request; nonexistent and non-owned requests both return `404 Not Found`.
- Invalid Ride Request UUIDs return `400 Bad Request`.
- The read model composes application-owned Ride Request state with an optional Trip.
- Automatic and offers-mode Ride Requests use the same status contract.
- Ride Request status and Trip status remain separate concepts. A Ride Request can remain `requested` while its Trip progresses through `assigned`, `in_progress`, and `completed`.
- Trip projection includes assigned Driver identity and assignment/execution timestamps without exposing Rider identity.
- No new persistence lifecycle or migration was introduced for this read model.

### Verification completed

- `go test ./...` passes.
- `go vet ./...` passes.
- Runtime verification covered automatic and offers-mode Ride Requests through assignment and Trip execution.
- Ownership isolation returns non-leaking `404 Not Found` for another authenticated Rider.
- The same endpoint projects strategy-neutral Ride Request state and optional Trip state for both automatic matching and marketplace assignment.

---

## Ride Cancellation Foundation

Implemented, verified, and merged through PR #22.

### Business flow

`Open Ride Request → owning Rider or assigned Driver cancels → Ride Request becomes cancelled → optional assigned/in-progress Trip becomes cancelled → active Driver commitment is released → status read model reflects cancellation`

### Implemented contract

- Rider cancellation endpoint: `POST /v1/ride-requests/{ride_request_id}/cancel`.
- Driver cancellation endpoint: `POST /v1/driver/ride-requests/{ride_request_id}/cancel`.
- Rider may cancel their own Ride Request before Trip completion.
- Driver may cancel only a Trip assigned to their authenticated Driver account.
- Repeated cancellation is idempotent and preserves the original cancellation timestamp.
- Ride Request cancellation records application-owned `cancelled_at` and `cancelled_by` (`rider` or `driver`).
- Assigned/in-progress Trips transition to `cancelled` with the same cancellation timestamp.
- Completed Trips cannot be cancelled and return conflict without mutating completed state.
- Active automatic candidates are released while preserving candidate decision history.
- Pending marketplace offers are closed with a decision timestamp; accepted/rejected/closed offer history remains intact.
- Cancelled automatic Ride Requests cannot be matched again.
- Automatic acceptance locks and revalidates the Ride Request before locking the candidate so cancellation and acceptance serialize without resurrecting cancelled work.
- Start/complete on cancelled Trips return conflict.
- The Rider status read model exposes Ride Request cancellation metadata and optional Trip cancellation state.

### Verification completed

- `gofmt` produced no changes and the working tree remained clean.
- `go test ./...` passes.
- `go vet ./...` passes.
- Runtime verification used a fresh database and fresh Rider/Driver accounts.
- Rider cancellation before assignment returns `cancelled`, `cancelled_by=rider`, and `trip=null`; repeated cancellation preserves the original `cancelled_at`.
- Matching a cancelled automatic Ride Request returns `409 Conflict`.
- Rider cancellation after automatic assignment cancels both Ride Request and Trip and releases the Driver for immediate rematching.
- Rider cancellation during `in_progress` preserves `started_at`, leaves `completed_at` null, and prevents subsequent completion.
- Driver-initiated cancellation records `cancelled_by=driver`; a non-assigned Driver receives non-leaking `404 Not Found`.
- Cancelling an offers-mode Ride Request closes pending marketplace offers; further offer submission returns `409 Conflict`.
- Cancelling an assigned marketplace Trip releases the Driver so they can immediately accept another marketplace Ride Request.
- Rider and Driver cancellation attempts against a completed Trip both return `409 Conflict` and leave the completed Trip unchanged.
- A concurrent Rider-cancel versus automatic Driver-accept test serialized correctly: cancellation won, acceptance returned `409 Conflict`, and final state remained cancelled with no Trip.

### Deliberately deferred

- Cancellation fees or refunds.
- No-show semantics.
- Driver compensation.
- Cancellation reason taxonomy.
- Abuse/rate-limit policy.
- Notifications.
- Location/geofence policy.
- Payments.

---

## Next Business Vertical Slice

Not fixed yet. Select the next slice from the next concrete Rider/Driver MVP client flow rather than introducing infrastructure speculatively.

Strong candidates, only when a consuming flow requires them:

- Live Driver Location Foundation if marketplace ranking, dispatch quality, or active-trip tracking needs current coordinates.
- Marketplace geographic filtering/ranking once location ownership and freshness semantics are defined.
- Rider/Driver current-trip or trip-history read models if the client needs retrieval beyond the existing Rider Ride-Request status view.
- Offer expiration or marketplace lifecycle hardening if stale marketplace inventory becomes a concrete client problem.

The next slice must preserve the established boundary: external maps, routing, notification, or payment systems implement application-defined ports and must not define business models or public APIs.

---

## Architecture Rule: Replaceable External Providers

External systems implement application-defined ports. Provider-specific concepts must not define business models or public APIs.

For authentication:

`Client → Application Authentication API → Authentication domain → Provider adapter → Identity infrastructure`

The concrete identity provider is selected at the composition root. Kratos is the current implementation, not a business-domain dependency.

The same boundary rule applies to future maps, routing, payments, notifications, storage, messaging, and other external services.

---

## Rough MVP Direction After Ride Cancellation Foundation

- Choose the next concrete Rider/Driver client flow before adding infrastructure.
- Add live location/status only when marketplace ranking, dispatch quality, or active-trip UX consumes it.
- Add geographic marketplace filtering/ranking only after application-owned location and freshness semantics exist.
- Add Rider/Driver current-trip or history read models only when concrete client flows require them.
- Harden marketplace offer lifecycle, cancellation policy, or pricing only when an end-to-end flow exposes the need.

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
