# Work Log

## Purpose

This file is the repository-level source of truth for meaningful project progress, the current stopping point, established architecture, and the next likely business slice.

Keep it concise. Detailed implementation and verification history belongs in the corresponding pull requests and commits.

---

## Current Status

**Current engineering milestone:** Rider Active-Trip Driver Location Read Foundation — implemented and verified in PR #30; pending merge.

**Current `main`:** `de1e92c48c9397917bd98150545d5cf77bb5b9a6` (worklog alignment through Driver Location Foundation).

**Current feature head before this worklog commit:** `0595a1e8c19a15581da0e580684f43a1b5b10b69`.

**Current stopping point:** Drivers publish one application-owned current location through `PUT /v1/driver/location`. An authenticated Rider can now read the latest location of the Driver assigned to a specific owned active Ride Request through `GET /v1/ride-requests/{ride_request_id}/driver-location`. Authorization is derived from Ride Request ownership and active Trip assignment; no Driver identity is accepted from or exposed to the Rider. The read is valid only while the Trip is `assigned` or `in_progress` and returns coordinates plus server-owned `updated_at`.

The Rider location read deliberately does not assume Rider-side active-Trip singularity. Ride Request ID is part of the resource path, so a Rider with multiple active Trips remains deterministic. Missing/non-owned/non-active Trips produce the same non-leaking result. No map, routing, geocoding, PostGIS, location history, streaming, or external provider concept has entered the public API or business model.

**Next architectural decision:** geographic marketplace ranking is now the leading product slice because Driver location has both a write foundation and a narrow authorized Rider consumer. Before implementation, define the minimum application-owned eligibility policy: online state, location freshness threshold, distance calculation, deterministic tie-breaking, and any MVP search-radius/fallback behavior.

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
- [x] Ride Request Required-Location Hardening — PR #9
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
- [x] Driver Current-Trip Read Foundation — PR #24
- [x] HTTP API Structure Refactor — PR #25
- [x] Rider Ride-Request List Foundation — PR #26
- [x] Driver Trip History Foundation — PR #27
- [x] Driver Location Foundation — PR #28
- [ ] Rider Active-Trip Driver Location Read Foundation — PR #30 (implemented and verified; pending merge)

Worklog-only alignment PRs are intentionally omitted from the business-milestone list.

---

## Recent Milestones

### Driver Trip History Foundation — PR #27

Merged as `3acf605e8759e9571214792546961b98080764fd`.

- Endpoint: `GET /v1/driver/trips`.
- Requires Driver capability and returns only the authenticated Driver's historical Trips.
- Historical statuses are `completed` and `cancelled`; active `assigned`/`in_progress` remains the responsibility of `GET /v1/driver/trip`.
- Results are newest-first and bounded to 50.
- Rider identity and redundant Driver identity are not exposed.
- No migration was required.

### Driver Location Foundation — PR #28

Merged as `e053add8c461008c8661846ff84d53f848659988`.

- Endpoint: `PUT /v1/driver/location`.
- Requires authentication, Driver capability, and an active Driver profile.
- Request requires latitude and longitude; valid ranges are `[-90, 90]` and `[-180, 180]`.
- Explicit `(0,0)` is valid.
- `driver_locations` stores one current row per Driver and updates it in place.
- `updated_at` is server-owned and represents the latest accepted update time.
- Driver identity is authentication-derived, never client-supplied, and is not echoed in the response.
- Migration `013_driver_location_foundation.sql` owns persistence constraints.
- Runtime verification confirmed valid upsert/re-upsert, timestamp advancement, zero coordinates, range validation, unauthenticated rejection, migration application, and one-row current-location persistence.

### Rider Active-Trip Driver Location Read Foundation — PR #30

Implemented and verified; pending merge.

- Endpoint: `GET /v1/ride-requests/{ride_request_id}/driver-location`.
- Requires Rider capability and scopes the read to a Ride Request owned by the authenticated Rider.
- The Ride Request must have an active Trip in `assigned` or `in_progress` state.
- Authorization and assigned-Driver resolution happen in one application-owned query using authenticated Rider ID plus Ride Request ID.
- Success returns `ride_request_id`, latitude, longitude, and server-owned `updated_at`; Rider and Driver identity are not exposed.
- Nonexistent, non-owned, completed, cancelled, or otherwise non-active Trips return the same non-leaking `active trip not found` result.
- An active Trip without a published Driver location maps to the application-owned `driver location not available` result.
- No migration or duplicate location state was introduced.
- Static verification passed: `gofmt`, `go test ./...`, and `go vet ./...`.
- Runtime verification confirmed assigned and in-progress reads, immediate latest-location propagation after Driver re-publish, completed-Trip cutoff, cross-Rider isolation, invalid-ID handling, unauthenticated rejection, and response privacy.
- The no-location branch was not destructively forced at runtime because the fixture Driver already had current location state; repository/application tests cover the missing-row contract.

Deliberately not included: location history, background-location policy, Rider location publishing, push/stream subscriptions, route/ETA/maps integration, geocoding, PostGIS, and geographic marketplace ranking.

---

## Current Business Model and Invariants

### Identity and capabilities

- One application account may hold multiple capabilities, including Rider and Driver.
- Authentication is shared; Driver does not have a separate identity system.
- Capability membership is separate from capability-specific profile and operational data.
- Verified identity is required for authenticated application use.

### Ride Requests and marketplace

- Ride Requests support application-owned `automatic` and `offers` booking modes.
- Offers-mode money uses integer minor units plus currency; payment-provider concepts are not part of the model.
- Driver offers do not reserve a Driver until assignment.
- Multiple Drivers may offer; only one assignment may win.
- Automatic matching and marketplace selection converge at the same Trip boundary.

### Driver commitments

- A Driver may have at most one active candidate/assignment commitment.
- Candidate activity is `pending`, or `accepted` while `released_at IS NULL`.
- A Driver may have at most one active Trip with status `assigned` or `in_progress`.
- Assignment paths revalidate Driver eligibility atomically.
- Trip completion or cancellation releases the active Driver commitment while preserving history.

### Trips and cancellation

- Trip states are `assigned`, `in_progress`, `completed`, and `cancelled`.
- Rider and assigned Driver cancellation converge on the same application-owned Ride Request/Trip state.
- Repeated cancellation is idempotent.
- Completed Trips cannot be cancelled.
- Current Driver Trip reads exclude completed/cancelled history; Driver Trip history excludes active Trips.

### Driver location

- Location is volatile operational telemetry, not durable Driver profile/vehicle state.
- Current location is stored separately from `driver_profiles`.
- One latest location row exists per Driver.
- Server-owned `updated_at` is the current freshness signal.
- Location publishing does not require the Driver to be online or to have an active Trip; downstream consumers decide whether a location is eligible/fresh enough for their use case.
- Rider reads resolve Driver location through an owned, explicitly identified active Ride Request; Rider-side active-Trip singularity is not assumed.
- Rider location responses expose location state, not Driver account identity.

---

## Architecture Rule: Replaceable External Providers

External systems implement application-defined ports. External-system concepts must never define business models or public APIs.

For authentication:

`Client → Application Authentication API → Authentication domain → Provider adapter → Identity infrastructure`

Kratos is the current identity implementation, not a domain dependency.

The same rule applies to maps, routing, geocoding, payments, notifications, storage, messaging, and future external services.

Provider-neutral application concepts come first; vendor adapters come later and remain replaceable.

---

## Next Business Vertical Slice

### Preferred: Marketplace Geographic Ranking Foundation

Driver location now has a write path and a narrow authorized Rider consumer. The next useful location consumer is automatic marketplace/matching selection.

Before implementation, define the smallest explicit policy needed by that consumer:

- only operationally eligible Drivers participate;
- current location must satisfy a concrete freshness threshold owned by the application;
- compute application-owned straight-line distance for MVP rather than introducing a maps/routing provider;
- define deterministic ordering for equal/near-equal candidates;
- define search-radius and fallback behavior only if the MVP flow requires them;
- keep the existing active-candidate and active-Trip exclusion invariants authoritative;
- do not introduce PostGIS until query scale or database-side geospatial behavior actually warrants it.

The matching contract should remain provider-neutral. Maps, road-network routing, ETA, geocoding, and vendor-specific location objects stay outside the matching domain.

### Correctness cleanup to track

Runtime fixture work for PR #30 exposed a pre-existing error-semantics issue: when the wrong Driver attempts to accept an automatic candidate, the acceptance path can fall through marketplace-offer handling and return a generic `500` instead of an application-owned candidate/authorization result. This predates PR #30 and should be corrected in a focused slice rather than scope-creeping the Rider location read.

---

## Deferred

- CI/CD and Kubernetes.
- iOS implementation.
- Courier and freight capabilities.
- Administrator operations.
- Enterprise-level workflow/process infrastructure.
- Advanced dispatch queues and optimization.
- PostGIS/geospatial indexing until scale/query behavior warrants it.
- Live location streaming and breadcrumb history.
- Route progress, ETA, and maps integration.
- Push notifications/subscriptions.
- Payments until concretely required by an MVP slice.
- Promotions and advanced analytics.
- Marketplace offer expiration and multi-round negotiation.
- Advanced cancellation fees/refunds/no-show policy.
- Authentication resend cooldown/rate-limit policy beyond provider defaults.

---

## Working Principles

- Build the MVP incrementally using business vertical slices.
- Finish and verify the current slice before expanding into the next one.
- Keep the repository worklog aligned with actual merged state.
- Keep `cmd/api` as composition root and business packages transport-neutral.
- Prefer clean application-owned boundaries over provider-specific abstractions.
- Do not introduce duplicated lifecycle state when an existing invariant is authoritative.
- Add infrastructure only when a concrete business flow consumes it.
- Preserve ownership/privacy boundaries and avoid identity leakage in read models.
- Use persistence constraints for important singularity/exclusivity invariants, with application transactions providing business serialization.
- Challenge premature abstractions, vendor lock-in, duplicated state, and speculative enterprise complexity.
