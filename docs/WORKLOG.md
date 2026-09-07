# Work Log

## Current status

The backend implements accounts/authentication, legacy Driver operations, ride
requests, geographic marketplace discovery and offers, Rider-selected assignment,
Trip execution, cancellation, history, and Driver location storage/read access.

The shared Flutter Android client implements account entry, Rider-first capability
selection, Rider request creation/status/cancellation, the shared ADR-0008 map-first
dashboard, and the initial Driver readiness flow from PR #69.

The current slice replaces the new-Driver onboarding semantics from PR #69 with the
ADR-0009 service-application model. A Driver chooses one service for the first
vehicle, reviews the application, and submits it for review. The submitted data is
stored separately from approved operational Driver/vehicle data and can be restored
as pending, approved, or rejected state. The client no longer treats new onboarding
submission as immediate creation or editing of an approved Driver profile.

This slice does **not** provide a public approval/rejection endpoint. Reviewer
authorization and promotion of an approved application into operational
Driver/vehicle/service state are the next backend foundation. The legacy one-vehicle
operational path remains for existing Drivers until that transition is complete.

Only Rider selection assigns a Trip. Accepting the Rider's proposed fare and
counteroffering both create pending offers; neither reserves the ride or Driver.

---

## Completed milestones

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
- [x] Legacy Candidate and Booking-Mode Retirement — PR #56
- [x] Geographic Marketplace Discovery and Rider Comparison — PR #57
- [x] Driver Public Presentation — PR #58
- [x] Flutter Rider Entry Foundation — PR #59
- [x] Flutter Rider Ride-Request Creation — PR #60
- [x] Shared Flutter Dashboard Foundation and Interaction Contract — PRs #61–#68
- [x] Flutter Driver Setup and Readiness — PR #69; retained as a transitional
  operational path for existing Drivers while ADR-0009 onboarding replaces its
  immediate new-Driver setup semantics

Worklog-only alignment PRs are intentionally omitted from the business-milestone list.

---

## Current product model and invariants

### Identity, capability, service

- One account supports Rider and Driver capabilities; identity is shared.
- Rider remains the default capability.
- `Driver` is an account capability.
- Economy, Comfort, and similar ride services are not capabilities. They are
  service products associated with a Driver vehicle through eligibility/enrollment.

### Driver onboarding and vehicle/service applications

- Initial onboarding applies one vehicle to one Driver-selected service.
- Flow: Driver details → service → requirements → vehicle → deterministic precheck
  → application review → submit → approved/rejected with reason.
- Precheck can reject only deterministic rules represented by application policy;
  it must not invent final verification state.
- Submission creates a separate application snapshot. It does not immediately
  create or overwrite approved operational Driver/vehicle data.
- At most one initial onboarding application may be pending for a Driver.
- A Driver who already has an operational Driver profile cannot submit another
  initial onboarding application; later vehicles/services use dedicated flows.
- A Driver may ultimately own multiple vehicles.
- Vehicle verification and service enrollment are separate states.
- Initial onboarding chooses one service. After approval, the same vehicle can
  apply for additional services without being registered again.
- Service hierarchy may establish derived technical eligibility but never silent
  enrollment. For example, Comfort approval may make Economy eligible when the
  catalog explicitly defines that implication; the Driver still chooses to add it.

See [ADR-0009](ADR-0009-driver-service-vehicle-eligibility.md).

### Approved information and change governance

- Approved Driver and vehicle information remains authoritative until a submitted
  revision is approved.
- Local drafts are discardable.
- After submission, a server-owned cancellation window permits direct cancellation.
- After that deadline, withdrawal requires an appeal.
- Approval, rejection, cancellation, withdrawal, and historical versions are retained.
- The concrete revision tables, review UI, cancellation endpoint, appeal workflow,
  and administration tooling remain deferred. Current client code must not introduce
  immediate-overwrite semantics in the meantime.

### Driver online operating context

- Target MVP operating context is exactly one selected verified vehicle and one
  selected approved service at a time.
- If a vehicle has one approved service, the client may omit the redundant service selector.
- Changing vehicle/service requires the Driver to go offline first.
- Going online publishes current location before the online transition.
- Going offline never requires location permission.
- Marketplace eligibility additionally requires Driver capability/approval, online
  state, fresh location, and no conflicting active Trip or other applicable commitment.
- Marketplace/offers/assignment/Trip history must preserve the vehicle/service
  context used at the time rather than following later Driver selections.
- The current legacy operational Driver profile still stores one vehicle and does
  not yet satisfy this target. Marketplace client expansion must wait for the
  multi-vehicle/service operating foundation rather than deepen that assumption.

### Marketplace and assignment

- A new ride request requires pickup, destination, and proposed fare/currency.
- There is no Rider booking-mode choice.
- Exact-fare responses and counteroffers both create/update pending offers.
- Drivers may offer on multiple rides; multiple Drivers may offer on one ride.
- Rider selection locks and revalidates the request, offer, and Driver.
- At most one Trip exists per ride, and a Driver has at most one active Trip.
- Competing offers on the selected ride close.
- Legacy rides without fares remain readable but cannot enter the marketplace.

ADR-0007 remains authoritative for the Ride Request marketplace.

### Trip lifecycle

- Trip states: `assigned`, `in_progress`, `completed`, and `cancelled`.
- Start and completion are idempotent. Completion requires an in-progress Trip.
- Rider and assigned Driver cancellation share transactional behavior.
- Cancellation closes pending offers for that ride and is idempotent.
- Completed Trips cannot be cancelled.

### Location and geographic policy

- One latest location row is stored per Driver, separate from profile/vehicle.
- Server-owned timestamps support freshness decisions.
- Marketplace locations must be no older than two minutes and not in the future.
- Discovery ranks eligible requests by Haversine pickup distance before the feed limit.
- Raw Driver coordinates and license plates are not exposed before assignment.
- No arbitrary pickup radius, service boundary, routing ETA, or PostGIS requirement
  has been introduced.

### Flutter client

- The client calls only application-owned APIs and does not expose Ory concepts.
- Session secrets use secure storage; capability preference uses simple preferences.
- Rider and Driver use the shared ADR-0008 dashboard interaction contract.
- New Driver onboarding loads the service catalog, restores the latest application,
  runs server precheck, displays a review confirmation, and submits for review.
- Pending/rejected/approved application state is shown from backend data.
- The immediate `Edit Driver details` action was removed from the operational
  dashboard because approved information must not be overwritten directly.
- Existing operational Driver accounts retain the PR #69 readiness path temporarily.

---

## Current slice implementation

### Backend

Migration 018 introduces:

- an application-owned Driver service catalog;
- initial Economy and Comfort catalog entries without invented model-year thresholds;
- optional explicit service implication (`Comfort` may imply Economy eligibility);
- Driver onboarding application snapshots with pending/approved/rejected status;
- one-pending-initial-application protection.

New application APIs:

```text
GET  /v1/driver/services
GET  /v1/driver/onboarding
POST /v1/driver/onboarding/precheck
POST /v1/driver/onboarding
```

No endpoint allows a Driver to approve or reject their own application.

### Flutter

The Driver screen now distinguishes:

- operational legacy Driver profile;
- onboarding loading/error;
- service/vehicle application form;
- application pending/rejected/approved status.

The form selects one service, collects Driver/vehicle information, performs the
server precheck, shows a final review dialog, and submits a pending application.
Backend `error` text is now preserved when a separate `message` field is absent.

---

## Verification

Backend verification to run from `backend/`:

```text
go test -p 1 ./...
go vet ./...
```

Database integration tests require the existing dedicated `_test` database setup.

Flutter verification to run from `mobile/`:

```text
dart format --output=none --set-exit-if-changed lib test
dart run build_runner build
flutter analyze
flutter test
```

Physical-device verification remains required. The current execution environment
used for this branch does not provide Flutter/Dart tooling or a reachable clone of
the GitHub repository, so runtime/test success must not be claimed until those
commands are run locally or in CI.

---

## Next implementation order

1. Define the minimal authorized review boundary for onboarding decisions; do not
   expose self-approval through Driver APIs.
2. On approval, create/promote the approved Driver, vehicle, and first service
   enrollment transactionally while preserving the submitted application history.
3. Replace the database one-vehicle assumption with Driver → multiple vehicles.
4. Add verified vehicle + approved service enrollment reads and operational selection.
5. Require selected verified vehicle/service when going online and lock switching
   until offline.
6. Add capability-shell sidebar and read-only Driver/Vehicle management surfaces
   without changing ADR-0008 dashboard gestures.
7. Run the physical-device gate.
8. Then continue Driver marketplace/offers client work, Rider offer selection, and
   Trip controls/history.

The future approved-information revision/cancellation-window/withdrawal-appeal
slice follows when profile/vehicle change management becomes a concrete dependency.

---

## Architecture authority

Accepted ADRs, `architecture-decisions.md`, `product-and-capability-model.md`,
`mvp-scope.md`, and this worklog describe the intended MVP. ADR-0007 is the Ride
Request marketplace authority, ADR-0008 is the dashboard interaction authority,
and ADR-0009 is the Driver service/vehicle/onboarding authority.

Implementation must not silently redefine these product rules.

## Deferred

- Fixed pickup/search radius and service areas until justified by launch policy.
- Routing, ETA, geocoding, and PostGIS.
- Live location streaming, breadcrumbs, and push notifications.
- Redis, background dispatch workers, and advanced dispatch optimization.
- Sophisticated payments, cancellation fees, refunds, and no-show policy.
- Full administrator operations beyond the minimal review boundary needed by the
  current Driver application flow.
- Complete versioned Driver/vehicle change review, cancellation-window, and appeal UI.
- Courier, Freight, promotions, analytics platforms, multi-round/chat negotiation,
  CI/CD, Kubernetes, and iOS implementation.

## Working principles

- Build small business slices and verify each before expanding scope.
- Keep business domains transport-neutral and `cmd/api` as the composition root.
- Use application-owned interfaces for external providers.
- Enforce assignment and approval invariants transactionally where implemented.
- Preserve ownership/privacy/history boundaries.
- Add infrastructure only when a concrete flow needs it.
- Keep this worklog aligned with implementation and distinguish target behavior
  from completed behavior.
