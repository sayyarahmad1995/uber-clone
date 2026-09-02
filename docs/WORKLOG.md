# Work Log

## Purpose

This file is the repository-level source of truth for meaningful project progress, the current stopping point, established architecture, and the next likely business slice.

Keep it concise. Detailed implementation and verification history belongs in the corresponding pull requests and commits.

---

## Current Status

**Current engineering milestone:** Driver Location Foundation — implemented, statically verified, runtime-verified, and merged through PR #28.

**Current `main`:** `e053add8c461008c8661846ff84d53f848659988`.

**Current stopping point:** Drivers can publish application-owned current coordinates through `PUT /v1/driver/location`. The application stores one latest location row per active Driver profile, validates coordinate ranges, accepts explicit `(0,0)`, owns freshness through server-generated `updated_at`, and does not expose Driver identity in the response. No map, routing, geocoding, PostGIS, location history, or external provider concept has entered the public API or business model.

The recovery/read surface is now reasonably balanced: Riders can read/list their Ride Requests; Drivers can read their current Trip and historical Trips; cancellation and marketplace assignment converge on the shared Trip model; current Driver location is available as an application-owned operational signal for the next concrete consumer.

**Next architectural decision:** consume Driver location in a narrow business flow before adding broader geospatial infrastructure. The preferred next slice is Rider Active-Trip Driver Location Read Foundation unless marketplace ranking becomes the higher product priority.

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

Worklog-only alignment PRs are intentionally omitted from the business-milestone list.

---

## Recent Milestones

### HTTP API Structure Refactor — PR #25

Merged as `43c5cdc233c753707cdac23b5ca6290c8b51960b`.

- `cmd/api` is now the composition root rather than the HTTP transport layer.
- HTTP transport lives under `internal/httpapi`.
- Business packages remain HTTP-neutral.
- The refactor intentionally introduced no API, business-lifecycle, or migration changes.
- A single `internal/httpapi` package is deliberate for the current scale; transport subpackages should appear only when concrete complexity justifies them.

### Rider Ride-Request List Foundation — PR #26

Merged as `b78c30e86c26c8970bf3aeab74d3c9f26e09460b`.

- Endpoint: `GET /v1/ride-requests`.
- Returns only the authenticated Rider's Ride Requests.
- Automatic and offers-mode requests share one strategy-neutral read model.
- Ride Request status and optional Trip status remain separate.
- Results are newest-first and bounded to 50.
- No Rider identity is echoed.
- No migration was required.

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

Deliberately not included: location history, background-location policy, Rider location tracking, push/stream subscriptions, route/ETA/maps integration, geocoding, PostGIS, and geographic marketplace ranking.

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

### Preferred: Rider Active-Trip Driver Location Read Foundation

Use the newly established Driver location state in the narrowest concrete consumer flow.

Proposed contract direction:

- An authenticated Rider may read the latest location of the Driver assigned to their own active Trip.
- Authorization is derived from Ride Request ownership plus the active Trip assignment; no Driver ID is accepted from the Rider.
- Only active `assigned`/`in_progress` Trips participate.
- The response should expose coordinates and `updated_at`, not Driver account identity.
- Missing current Driver location should have an application-owned non-leaking result.
- Freshness policy should be explicit if the product flow needs one; do not invent a threshold before the client behavior requires it.
- No streaming, WebSocket, push, maps SDK, routing provider, or breadcrumb history in this first consumer slice.

Why this comes before geographic marketplace ranking: it exercises location ownership, authorization, missing-location behavior, and freshness semantics with a narrow deterministic consumer. Marketplace geographic ranking requires additional policy choices—online eligibility, freshness threshold, distance computation, ranking/tie-breaking, and search radius—and should be built after those primitives are proven.

### Following candidate: Marketplace Geographic Ranking Foundation

When product priority requires it:

- consider only operationally eligible Drivers with sufficiently fresh current locations;
- compute application-owned distance for MVP without introducing PostGIS unless query scale justifies it;
- define deterministic proximity ordering and fallback behavior;
- keep maps/routing/geocoding providers outside matching-domain contracts.

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
