# Work Log

## Current status

The backend implements accounts/authentication, Driver operations, ride requests,
marketplace discovery and offers, Rider-selected assignment, Trip execution,
cancellation, history, and Driver location storage/read access.

PR #55 removed candidate influence from Rider selection and read models. This
slice completes candidate retirement in the runtime, adds migration 016, and
replaces candidate-based integration fixtures with Rider-selected offers.

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
- Driver capability, active profile, online state, vehicle, and active-Trip
  exclusions govern operational eligibility.
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
- The retired matching implementation used fresh locations and Haversine distance.
  The current marketplace discovery SQL orders requests by creation time; it does
  not yet apply geographic ranking or location freshness. Do not describe it as
  nearby-Driver matching until that integration is implemented and verified.
- No arbitrary pickup radius, service boundary, same-city restriction, routing
  ETA, or PostGIS requirement has been introduced.

## Verification for this slice

PostgreSQL integration tests cover fresh and existing-schema migrations,
reapplication, preserved Trips/offers, fare constraints, and rollback when an
accepted candidate is missing a matching Trip. Runtime tests cover concurrent
selection, Driver availability checks, completion/cancellation, and the
marketplace journey without candidate tables.

Use a dedicated database ending in `_test` and run `go test -p 1 ./...` and
`go vet ./...` from `backend`. Without `TEST_DATABASE_URL`, database tests skip.

## Architecture authority

Accepted ADRs, `architecture-decisions.md`, `product-and-capability-model.md`,
`mvp-scope.md`, and this worklog describe the intended MVP. ADR-0007 is the
marketplace authority. Explicitly update the relevant decision before changing
product behavior; implementation details must not redefine the product.

## Next proposed business slice

Bring geographic eligibility and pickup-distance ranking into the unified
Driver discovery and Rider offer views. First settle the minimal API/read-model
scope against ADR-0007, then reuse existing Driver location storage and ownership
boundaries. Driver/vehicle presentation should expose only the fields needed
for Rider choice. Routing ETA remains deferred.

This is a proposed follow-up, not work implemented by the retirement slice.

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
