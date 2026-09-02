# Work Log

## Purpose

This file is the repository-level source of truth for meaningful project progress, the current stopping point, established architecture, and the next likely business slice.

Keep it concise. Detailed implementation and verification history belongs in the corresponding pull requests and commits.

---

## Current Status

**Current engineering milestone:** Automatic Ride Dispatch Lifecycle — geographic matching, candidate lifecycle/timeout semantics, Trip execution, cancellation, and DB-backed transactional coverage are implemented and merged through PR #44.

**Current `main`:** `0ecbd93febf8b75efaeb1df1db25c80488e3b2ee`.

**Current stopping point:** Automatic matching selects the nearest operationally eligible Driver using fresh application-owned Driver location, straight-line Haversine distance, deterministic UUID tie-breaking, active-candidate/Trip exclusion, and a 30-second Driver response window. The response window is enforced consistently across matching, acceptance, and rejection. Candidate release semantics align with database uniqueness constraints, and released history no longer blocks future Driver participation.

Trip acceptance, Start, Complete, and cancellation are protected by PostgreSQL integration tests. Completion and cancellation release active Driver commitments while preserving history. Rider/Driver cancellation is idempotent, completed Trips cannot be cancelled, and cancellation also closes pending marketplace offers for the cancelled ride.

A dedicated `*_test` PostgreSQL database pattern protects repository integration tests from development-data contamination. No new runtime infrastructure, exposed database port, PostGIS dependency, background timeout worker, Redis, scheduler, or external maps/routing provider has been introduced.

**Next product slice:** expose the authenticated Driver's current automatic candidate as an explicit read model. Today Drivers can accept/reject a known Ride Request, but there is no narrow Driver-facing read endpoint for the currently assigned pending automatic candidate. Add that read before introducing additional matching policy.

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
- [x] Rider Active-Trip Driver Location Read Foundation — PR #30
- [x] Automatic Candidate Acceptance Error Semantics Fix — PR #31
- [x] Geographic Automatic Matching — PR #32
- [x] Released Candidate Exclusivity Alignment — PR #33
- [x] Automatic Candidate Response Timeout — PR #35
- [x] Matching PostgreSQL Integration Coverage — PR #36
- [x] Trip Acceptance PostgreSQL Integration Coverage — PR #38
- [x] Trip Execution PostgreSQL Integration Coverage — PR #40
- [x] Cancellation PostgreSQL Integration Coverage — PR #42
- [x] Automatic Candidate Reject Timeout Alignment — PR #44

Replacement PRs created only to work around the GitHub Draft→Ready connector issue are listed by their merged PR number. Worklog-only alignment PRs are intentionally omitted from the business-milestone list.

---

## Recent Architecture and Behavior

### Geographic automatic matching

- Automatic matching requires an active, online Driver with Driver capability, a vehicle, and a fresh current location.
- Driver location freshness is **2 minutes**.
- Missing or stale Driver location makes that Driver ineligible for new candidate selection.
- Eligible Drivers are ordered by application-owned Haversine pickup distance, then Driver UUID for deterministic tie-breaking.
- Existing pending/accepted candidates for a ride are returned idempotently instead of being re-ranked.
- Drivers previously considered for the same ride are not retried after timeout/rejection.
- Active candidate and active Trip exclusivity remain authoritative.
- No PostGIS, routing/ETA provider, destination-distance ranking, city model, service-area model, or fixed pickup radius is part of the current matching policy.

### Automatic candidate lifecycle

- A pending automatic candidate has a **30-second response window**.
- Timeout is lazy and transactional; no background worker is required.
- Matching releases expired pending candidates before reselection.
- Late acceptance releases the pending candidate and returns the resolved-assignment result without creating a Trip.
- Late rejection releases the pending candidate and returns the resolved-candidate result without recording misleading rejection history.
- Accepted candidates do not expire; the Trip lifecycle owns the commitment after acceptance.
- `released_at` is the authoritative release marker; no separate `expired` status exists.

### Trip and cancellation lifecycle

- Trip states are `assigned`, `in_progress`, `completed`, and `cancelled`.
- Acceptance creates at most one assigned Trip and is idempotent for an already accepted candidate.
- Start transitions `assigned → in_progress` and is idempotent while already in progress.
- Complete requires an in-progress Trip, transitions to `completed`, and releases the accepted candidate at the completion timestamp.
- Repeated Complete preserves the original completion timestamp.
- Rider and assigned Driver cancellation converge on the same transactional Ride Request/Trip state.
- Cancellation releases pending/accepted candidates and closes pending marketplace offers for the cancelled ride.
- Repeated cancellation preserves the original cancellation metadata.
- Completed Trips cannot be cancelled.

### PostgreSQL integration testing

- Repository integration tests are gated by `TEST_DATABASE_URL`.
- Integration tests refuse to run against a database whose name does not end in `_test`.
- Tests apply the real embedded migrations and use generated UUID fixtures with cleanup.
- PostgreSQL remains internal to the Docker Compose network; tests run from a temporary Go container attached to that network.
- No Testcontainers dependency, extra Compose service, or host PostgreSQL port was added.
- Matching, automatic acceptance, Trip execution, and cancellation critical transactional invariants now have DB-backed coverage.

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
- Candidate activity is pending or accepted while `released_at IS NULL`.
- A Driver may have at most one active Trip with status `assigned` or `in_progress`.
- Assignment paths revalidate Driver eligibility atomically.
- Released candidate history does not block future participation.
- Trip completion or cancellation releases the active Driver commitment while preserving history.

### Driver location

- Location is volatile operational telemetry, not durable Driver profile/vehicle state.
- Current location is stored separately from `driver_profiles`.
- One latest location row exists per Driver.
- Server-owned `updated_at` is the current freshness signal.
- Location publishing does not require the Driver to be online or to have an active Trip; consumers own eligibility/freshness policy.
- Rider reads resolve Driver location through an owned, explicitly identified active Ride Request.
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

### Preferred: Driver Current Automatic Candidate Read

Automatic dispatch can create a pending Driver candidate, and Drivers can already accept or reject a known Ride Request. The missing product boundary is a narrow authenticated Driver read for the current pending automatic candidate.

The slice should:

- derive Driver identity from authentication;
- return at most the authenticated Driver's one current unreleased pending automatic candidate;
- expose only Rider/ride information required for the Driver to decide whether to accept or reject;
- include candidate timing information needed by the client to represent the existing 30-second response window;
- return a non-leaking not-found/empty result when no actionable candidate exists;
- preserve current candidate/Trip exclusivity and timeout semantics;
- avoid polling infrastructure, push notifications, WebSockets, Redis, or background workers in this slice.

### Pickup-radius policy remains deferred

Automatic matching currently chooses the nearest fresh eligible Driver with no fixed distance cutoff. A very distant but fresh Driver can therefore still be selected if no closer eligible Driver exists.

Do **not** introduce an arbitrary radius yet. A fixed pickup radius is a marketplace operating policy, not merely a technical safeguard. Define it when we have an explicit launch-area/service-boundary requirement, observed dispatch data, or an ETA/acceptance target that justifies the number.

Keep service boundaries separate from nearest-Driver ranking. Future city-to-city or broader mobility products must not be constrained by a prematurely hard-coded same-city assumption.

---

## Deferred

- Fixed pickup/search radius until justified by operating policy/data.
- City/service-area and city-to-city product semantics.
- PostGIS/geospatial indexing until scale/query behavior warrants it.
- Routing/ETA/maps integration and geocoding.
- Live location streaming and breadcrumb history.
- Push notifications/subscriptions.
- Redis/background dispatch workers until a concrete flow requires them.
- Payments until concretely required by an MVP slice.
- Courier and freight capabilities.
- Administrator operations.
- Advanced dispatch queues and optimization.
- Advanced cancellation fees/refunds/no-show policy.
- Marketplace multi-round negotiation.
- CI/CD and Kubernetes.
- iOS implementation.
- Enterprise-level workflow/process infrastructure.

---

## Working Principles

- Build the MVP incrementally using business vertical slices.
- Finish and verify the current slice before expanding into the next one.
- Keep this worklog aligned with actual merged state.
- Keep `cmd/api` as composition root and business packages transport-neutral.
- Prefer clean application-owned boundaries over provider-specific abstractions.
- Do not introduce duplicated lifecycle state when an existing invariant is authoritative.
- Add infrastructure only when a concrete business flow consumes it.
- Preserve ownership/privacy boundaries and avoid identity leakage in read models.
- Use persistence constraints for important singularity/exclusivity invariants, with application transactions providing business serialization.
- Challenge premature abstractions, vendor lock-in, duplicated state, speculative enterprise complexity, and unvalidated business constants.
