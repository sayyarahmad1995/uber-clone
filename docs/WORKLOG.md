# Work Log

## Purpose

This file is the repository-level source of truth for meaningful project progress, the current stopping point, established architecture, and the next likely business slice.

Keep it concise. Detailed implementation and verification history belongs in the corresponding pull requests and commits.

---

## Current Status

**Current engineering milestone:** Core ride lifecycle and marketplace foundations are implemented: Rider ride requests, Driver discovery/offers, geographic eligibility/ranking, Driver commitment rules, Trip execution, cancellation, and DB-backed transactional coverage.

**Current architecture correction:** the product has one Rider Ride Request flow. The Rider supplies pickup, destination, and a proposed fare. The Rider does **not** choose between separate `automatic` and `offers` products. Eligible Drivers either accept the Rider proposed fare or submit a counteroffer. This is now authoritative in `ADR-0007-ride-request-marketplace-model.md`.

Earlier implementation introduced `booking_mode = automatic | offers` to add marketplace behavior without breaking the existing candidate flow. That state remains temporarily as implementation debt, but future work must not deepen it into a Rider-facing product concept.

**Current stopping point:** much of the existing backend logic remains reusable, but the next product work should reconcile Driver ride discovery/actionability with the unified marketplace model before expanding the legacy automatic-candidate abstraction.

Geographic matching remains valid as marketplace infrastructure: active/online Driver eligibility, fresh location, Haversine pickup distance, deterministic ordering, active-candidate/Trip exclusions, and transactional assignment constraints. Its future business role is eligibility, distribution, and ranking of Ride Requests to Drivers rather than a separate Rider booking mode.

Trip acceptance/assignment, Start, Complete, and cancellation are protected by PostgreSQL integration tests. Completion and cancellation release active Driver commitments while preserving history. Cancellation is idempotent and completed Trips cannot be cancelled.

A dedicated `*_test` PostgreSQL database pattern protects repository integration tests from development-data contamination. No PostGIS, Redis dispatch worker, background timeout scheduler, or external maps/routing provider is required by the current MVP backend.

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
- [x] Matching PostgreSQL Integration Coverage — merged replacement PR #37
- [x] Trip Acceptance PostgreSQL Integration Coverage — PR #38
- [x] Trip Execution PostgreSQL Integration Coverage — PR #40
- [x] Cancellation PostgreSQL Integration Coverage — PR #42
- [x] Automatic Candidate Reject Timeout Alignment — PR #44

Worklog-only alignment PRs are intentionally omitted from the business-milestone list.

---

## Current Product Model and Invariants

### Unified Ride Request marketplace

- The Rider creates one Ride Request with pickup, destination, and proposed fare/currency.
- The Rider is not asked to choose an automatic or offers booking mode.
- Eligible Drivers receive or discover actionable Ride Requests according to marketplace policy.
- A Driver may accept the Rider proposed fare or submit a counteroffer.
- Exact-fare acceptance may assign immediately if the Driver and Ride Request are still eligible.
- Counteroffers remain non-reserving until the Rider accepts one.
- Rider acceptance of a counteroffer assigns atomically.
- All assignment paths converge on the same Trip lifecycle and concurrency invariants.
- Existing `booking_mode` persistence is migration debt; new Rider-facing behavior must not depend on choosing it.

### Geographic marketplace policy

- Driver must have Driver capability, an active profile, be online, have a vehicle, and not be actively committed elsewhere.
- Driver location freshness is **2 minutes** for current geographic matching.
- Missing or stale Driver location makes that Driver ineligible where geographic policy is applied.
- Straight-line Haversine pickup distance is an acceptable MVP ordering mechanism.
- Driver UUID provides deterministic tie-breaking.
- A fixed pickup radius remains deferred until launch-area policy, dispatch data, or an ETA/acceptance target justifies it.
- No city/same-city assumption should be encoded into matching.

### Driver commitments and assignment

- A Driver may have at most one active candidate/assignment commitment.
- A Driver may have at most one active Trip with status `assigned` or `in_progress`.
- Assignment-time Driver eligibility is revalidated atomically.
- At most one Driver may win a Ride Request.
- Losing candidates/offers cannot create a second Trip.
- Released candidate history does not block future Driver participation.
- Trip completion or cancellation releases the active Driver commitment while preserving history.

### Legacy automatic-candidate timeout

- The current automatic-candidate path has a **30-second response window**.
- Timeout is lazy and transactional.
- Matching, acceptance, and rejection consistently release expired pending candidates.
- Accepted candidates do not expire; Trip lifecycle owns the commitment.
- `released_at` is the release marker; no separate expired status exists.

This behavior remains protected while we reconcile the implementation, but future product design should not assume a Rider-selected automatic mode.

### Trip and cancellation lifecycle

- Trip states are `assigned`, `in_progress`, `completed`, and `cancelled`.
- Start transitions `assigned → in_progress` and is idempotent.
- Complete requires an in-progress Trip and is idempotent once completed.
- Completion releases the accepted Driver commitment at the completion timestamp.
- Rider and assigned Driver cancellation converge on one transactional state change.
- Cancellation releases pending/accepted commitments and closes pending offers for the cancelled ride.
- Repeated cancellation is idempotent.
- Completed Trips cannot be cancelled.

### Identity and capabilities

- One application account may hold multiple capabilities, including Rider and Driver.
- Authentication is shared; Driver does not have a separate identity system.
- Capability membership is separate from capability-specific operational state.
- Rider/Driver capability switching must not be used to encode commercial ride strategy.

### Driver location

- Current location is volatile operational telemetry, stored separately from durable Driver profile/vehicle state.
- One latest location row exists per Driver.
- Server-owned `updated_at` is the freshness signal.
- Rider reads resolve Driver location through an owned active Ride Request/Trip relationship.
- Rider location responses expose location state, not Driver account identity.

---

## Architecture Authority

Accepted ADRs, `architecture-decisions.md`, `product-and-capability-model.md`, `mvp-scope.md`, and this worklog are the implementation contract for the current MVP.

Before designing a new slice, check it against those documents. Do not silently allow implementation details to redefine the product model.

If code and documentation conflict, determine whether the code drifted or the product requirement changed. A genuine requirement change must explicitly update or supersede the relevant architecture decision before deeper implementation depends on it.

The current marketplace authority is:

- `ADR-0007-ride-request-marketplace-model.md`

---

## Next Business Vertical Slice

### Preferred: Unified Driver Ride Request Actionability

Do **not** add a new Driver "current automatic candidate" product endpoint as previously planned. That would deepen the legacy split between automatic candidates and marketplace offers.

The next slice should instead reconcile Driver-facing Ride Request discovery/actionability around the unified marketplace model.

Target behavior:

```text
Rider creates Ride Request
(pickup + destination + proposed fare)
              |
              v
Eligible nearby Drivers receive/discover it
              |
      +-------+-------+
      |               |
      v               v
Accept Rider fare   Submit counteroffer
      |               |
      |               v
      |         Rider accepts/rejects
      |               |
      +-------+-------+
              v
             Trip
```

The implementation should reuse the good existing foundations rather than rewrite them unnecessarily:

- Driver operational eligibility;
- fresh-location/geographic ordering;
- marketplace discovery;
- exact-fare acceptance;
- counteroffer creation/update;
- Rider offer selection;
- active Driver exclusivity;
- atomic Trip assignment;
- Trip/cancellation lifecycle.

The slice should identify the smallest safe migration away from Rider-visible `booking_mode` semantics. Do not remove persistence compatibility recklessly; first establish the unified API/domain behavior, then migrate schema/legacy paths in focused steps.

### Pickup-radius policy remains deferred

Automatic/geographic matching currently has no fixed distance cutoff. A very distant but fresh Driver can technically remain eligible if no closer Driver exists.

Do not invent an arbitrary number. A radius is an operating policy that should be justified by launch geography, observed dispatch behavior, or an explicit pickup-time/acceptance target.

Service boundaries are a separate concern and must not become a hidden same-city rule. Future city-to-city or broader mobility products should remain possible without undoing premature schema assumptions.

---

## Deferred

- Fixed pickup/search radius until justified by operating policy/data.
- City/service-area and city-to-city product semantics.
- PostGIS/geospatial indexing until scale/query behavior warrants it.
- Routing/ETA/maps integration and geocoding.
- Live location streaming and breadcrumb history.
- Push notifications/subscriptions.
- Redis/background dispatch workers until a concrete flow requires them.
- Sophisticated payments until concretely required by an MVP slice.
- Courier and freight capabilities.
- Administrator operations.
- Advanced dispatch queues and optimization.
- Advanced cancellation fees/refunds/no-show policy.
- Multi-round/chat negotiation.
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
- Treat implementation-only fields as implementation details; do not let them silently become product concepts.
