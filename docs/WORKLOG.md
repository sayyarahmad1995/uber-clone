# Work Log

## Current status

The backend implements accounts/authentication, Driver operations, ride requests,
marketplace discovery and offers, Rider-selected assignment, Trip execution,
cancellation, history, and Driver location storage/read access.

PR #56 completed candidate retirement and migration 016. This slice adds
location-aware discovery, Rider offer comparison, and consistent geographic
eligibility through offer submission and selection. See the
[API contract](geographic-marketplace-api.md).

Only Rider selection assigns a Trip. Accepting the Rider's proposed fare and
counteroffering both create pending offers. A pending offer reserves neither
the ride nor the Driver.

Migration 016 preserves existing Trips, offers, and legacy rides without fares.
It stops if an accepted unreleased candidate has no matching Trip. See the
[rollout guide](candidate-retirement-rollout.md) before upgrading an existing
installation. Old API instances must be stopped for the schema removal.

The Flutter application is planned but is not yet present in this repository.

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

## Current product model and invariants

### Marketplace and assignment

- One account supports Rider and Driver capabilities; identity is shared.
- A new ride request requires pickup, destination, and proposed fare/currency.
- There is no Rider booking-mode choice.
- Exact-fare responses and counteroffers both create/update pending offers.
- Drivers may offer on multiple rides; multiple Drivers may offer on one ride.
- Rider selection locks and revalidates the request, offer, and Driver.
- Driver capability, active profile, online state, vehicle, fresh location, and
  active-Trip exclusions govern marketplace eligibility. The Driver domain owns
  the shared transactional check used by offers and Trip assignment.
- At most one Trip exists per ride, and a Driver has at most one active Trip.
- Competing offers on the selected ride close. The winning Driver's other offers
  cannot create a second active Trip; availability is rechecked on selection.
- There are no runtime candidate reservations, timeout rules, or release markers.
- Legacy rides without fares remain readable but cannot enter the marketplace.

### Trip lifecycle

- Trip states: `assigned`, `in_progress`, `completed`, and `cancelled`.
- Start and completion are idempotent. Completion requires an in-progress Trip.
- Rider and assigned Driver cancellation share transactional behavior.
- Cancellation closes pending offers for that ride and is idempotent.
- Completed Trips cannot be cancelled.
- Completion/cancellation free the Driver through Trip status while preserving
  history. No separate commitment lifecycle is needed.

### Location and geographic policy

- One latest location row is stored per Driver, separate from profile/vehicle.
- Server-owned update timestamps support freshness decisions.
- Rider location reads require ownership of the associated active ride/Trip.
- Marketplace locations must be no older than two minutes and not in the future.
- Discovery ranks all eligible requests by Haversine pickup distance before the
  feed limit, with deterministic creation-time/UUID ties.
- Rider comparison includes vehicle make/model/color, nullable pickup distance,
  fare-match indication, and current selectability. Raw Driver coordinates and
  license plates are not exposed before assignment.
- Selection rechecks eligibility after acquiring locks. Offer views are snapshots;
  temporary unavailability does not change pending offer status.
- No arbitrary pickup radius, service boundary, same-city restriction, routing
  ETA, or PostGIS requirement has been introduced.

## Verification for this slice

PostgreSQL tests retain migration, assignment, completion, and cancellation
coverage. This slice adds ranking-before-limit checks, geographic edge cases,
consistent eligibility across discovery/responses/selection, Rider comparison,
ownership, location refresh, and a concurrent location-update check. HTTP
contract tests verify nullable distance and restricted pre-assignment fields.

Use a dedicated database ending in `_test` and run `go test -p 1 ./...` and
`go vet ./...` from `backend`. Without `TEST_DATABASE_URL`, database tests skip.

## Architecture authority

Accepted ADRs, `architecture-decisions.md`, `product-and-capability-model.md`,
`mvp-scope.md`, and this worklog describe the intended MVP. ADR-0007 is the
marketplace authority. Explicitly update the relevant decision before changing
product behavior; implementation details must not redefine the product.

## Follow-up scope

The geographic marketplace API is ready for client integration. The Flutter
client and application-owned Driver names/photos remain separate slices; choose
the next concrete user journey before adding providers or infrastructure.

## Deferred

- Fixed pickup/search radius and service areas until justified by launch policy.
- Routing, ETA, maps, geocoding, and PostGIS.
- Live location streaming, breadcrumbs, and push notifications.
- Redis, background dispatch workers, and advanced dispatch optimization.
- Sophisticated payments, cancellation fees, refunds, and no-show policy.
- Courier, Freight, administrator operations, promotions, and analytics platforms.
- Multi-round/chat negotiation, CI/CD, Kubernetes, and iOS implementation.

## Working principles

- Build small business slices and verify each before expanding scope.
- Keep business domains transport-neutral and `cmd/api` as the composition root.
- Use application-owned interfaces for external providers.
- Enforce assignment invariants through transactions and database constraints.
- Preserve ownership/privacy boundaries and avoid unnecessary identity exposure.
- Add infrastructure only when a concrete flow needs it.
- Keep this worklog aligned with the implementation and distinguish proposals
  from completed behavior.
