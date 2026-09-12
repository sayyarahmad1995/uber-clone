# Work Log

## Current status

The backend implements accounts/authentication, legacy Driver operations, ride
requests, geographic marketplace discovery and offers, Rider-selected assignment,
Trip execution, cancellation, history, and Driver location storage/read access.

The shared Flutter Android client implements account entry, Rider-first capability
selection, Rider request creation/status/cancellation, the shared ADR-0008 map-first
dashboard, Driver onboarding/review state, and the capability-shell Drawer merged in
PR #72. The current slice adds read-only Driver details and Vehicles destinations
backed only by real operational-profile or onboarding-application data.

PR #70 introduces the ADR-0009 new-Driver onboarding application model. A Driver
chooses one service for the first vehicle, reviews the application, and submits it
for review. Submitted data remains separate from approved Driver/vehicle data and
can be restored as pending, approved, or rejected state. The public legacy
`PUT /v1/driver` write path is removed so Driver accounts cannot bypass review.

The current reviewer slice implements the narrow administration dependency defined
by ADR-0010. Reviewer routes are registered only when internal reviewer credentials
are configured. A reviewer can list pending applications, inspect one application,
approve it, or reject it with a required reason through a small server-rendered page
or the protected reviewer API.

Approval transactionally promotes the submitted snapshot into an approved Driver
profile, vehicle, and explicit selected-service enrollment while retaining the
application decision/history. Rejection creates no Driver/vehicle/service records.
Newly approved profiles use status `approved`, not `active`, so the legacy online
and marketplace paths cannot treat them as operational before the vehicle/service
operating-context slice is implemented.

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
- [x] Driver Service Onboarding Foundation — PR #70
- [x] Minimal Driver Onboarding Reviewer — PR #71
- [x] Capability Drawer Shell and Dashboard State Persistence — PR #72

Worklog-only alignment PRs are intentionally omitted from the business-milestone list.

---

## Current product model and invariants

### Identity, capability, service

- One account supports Rider and Driver capabilities; identity is shared.
- Rider remains the default capability.
- `Driver` is an account capability.
- Economy, Comfort, and similar ride services are not capabilities. They are
  service products associated with a Driver vehicle through eligibility/enrollment.
- Administrator review is an operator responsibility, not an account capability.

### Driver onboarding and vehicle/service applications

- Initial onboarding applies one vehicle to one Driver-selected service.
- Flow: Driver details → service → requirements → vehicle → deterministic precheck
  → application review → submit → approved/rejected with reason.
- Precheck can reject only deterministic rules represented by application policy;
  it must not invent final verification state.
- Submission creates a separate application snapshot. It does not immediately
  create or overwrite approved Driver/vehicle data.
- At most one initial onboarding application may be pending for a Driver.
- A Driver who already has approved Driver records cannot submit another initial
  onboarding application; later vehicles/services use dedicated flows.
- A Driver may ultimately own multiple vehicles.
- Vehicle verification and service enrollment are separate states.
- Initial onboarding chooses one service. After approval, the same vehicle can
  apply for additional services without being registered again.
- Service hierarchy may establish derived technical eligibility but never silent
  enrollment. For example, Comfort approval may make Economy eligible when the
  catalog explicitly defines that implication; the Driver still chooses to add it.

See [ADR-0009](ADR-0009-driver-service-vehicle-eligibility.md).

### Initial onboarding review

- Initial onboarding review follows ADR-0010.
- Reviewer routes are disabled unless both internal reviewer credentials are configured.
- Approval/rejection is backend-authoritative and transactional.
- Approval promotes exactly the submitted Driver/vehicle/service snapshot.
- Approval creates one explicit enrollment for the selected service; implied lower
  services are not silently enrolled.
- Rejection requires a non-blank reason and creates no approved records.
- A decided application cannot be decided again through the reviewer surface.
- Reviewer identity and decision time are retained on the application.
- Newly approved profiles are `approved` and offline, not `active`.
- Approved-only profiles remain outside the legacy operational `/v1/driver`, online,
  and marketplace path until the new operating-context slice is implemented.

See [ADR-0010](ADR-0010-minimal-driver-onboarding-reviewer.md).

### Approved information and change governance

- Approved Driver and vehicle information remains authoritative until a submitted
  revision is approved.
- Local drafts are discardable.
- After submission, a server-owned cancellation window permits direct cancellation.
- After that deadline, withdrawal requires an appeal.
- Approval, rejection, cancellation, withdrawal, and historical versions are retained.
- The current reviewer implements **initial onboarding decisions only**. The future
  Driver/vehicle revision tables, cancellation endpoint, withdrawal appeal workflow,
  and versioned-change reviewer UI remain deferred.

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
- Existing legacy `active` Drivers still use the one-vehicle operational path.
- Newly approved ADR-0009 Drivers intentionally cannot enter that legacy path.
- Marketplace client expansion must wait for the multi-vehicle/service operating
  foundation rather than deepen the legacy assumption.

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
- The capability shell owns Rider/Driver switching, Driver enablement, real secondary
  destinations, and logout; operational controls remain on the dashboard.
- Read-only Driver details and Vehicles surfaces use the operational Driver profile
  when available and otherwise fall back to the latest onboarding application.
- The immediate `Edit Driver details` action was removed from the operational
  dashboard because approved information must not be overwritten directly.
- Existing operational Driver accounts retain the PR #69 readiness path temporarily.

---

## Current slice implementation

### Driver onboarding submission foundation

Migration 018 introduces:

- an application-owned Driver service catalog;
- initial Economy and Comfort catalog entries without invented model-year thresholds;
- optional explicit service implication (`Comfort` may imply Economy eligibility);
- Driver onboarding application snapshots with pending/approved/rejected status;
- one-pending-initial-application protection.

Application APIs:

```text
GET  /v1/driver/services
GET  /v1/driver/onboarding
POST /v1/driver/onboarding/precheck
POST /v1/driver/onboarding
```

No endpoint allows a Driver to approve or reject their own application.

### Minimal reviewer

Migration 019:

- adds `approved` as a non-operational Driver profile status;
- records the reviewer identity on decided onboarding applications;
- adds explicit approved vehicle/service enrollment persistence.

Protected reviewer API:

```text
GET  /v1/admin/driver-onboarding-applications
GET  /v1/admin/driver-onboarding-applications/{application_id}
POST /v1/admin/driver-onboarding-applications/{application_id}/approve
POST /v1/admin/driver-onboarding-applications/{application_id}/reject
```

The server-rendered reviewer is available at `/admin/driver-onboarding` only when
`ADMIN_REVIEW_USERNAME` and `ADMIN_REVIEW_PASSWORD` are both configured. All reviewer
responses are protected by HTTP Basic auth; mutation routes also reject explicit
cross-origin requests. The current credentials are an internal MVP mechanism, not the
long-term administrator identity architecture.

See [Driver onboarding reviewer](driver-onboarding-reviewer.md).

### Flutter

The Driver screen distinguishes:

- operational legacy Driver profile;
- onboarding loading/error;
- service/vehicle application form;
- application pending/rejected/approved status.

The form selects one service, collects Driver/vehicle information, performs the
server precheck, shows a final review dialog, and submits a pending application.
Backend `error` text is preserved when a separate `message` field is absent.

An approved-only profile remains hidden from the legacy operational Driver endpoint,
so the mobile client continues to show the onboarding application as approved rather
than exposing a non-functional legacy Go online control.

### Capability shell and read-only Driver surfaces

PR #72 moved secondary navigation into a standard Flutter Drawer, removed the bottom
capability switch, and made committed dashboard extent persist across workflow and
Rider/Driver transitions while keeping ADR-0008 gesture semantics intact.

The current slice adds `Driver details` and `Vehicles` routes for Driver-capable
accounts. These screens are read-only. They prefer real `DriverProfile` data for
legacy operational Drivers and otherwise use the latest `DriverOnboardingApplication`
snapshot, including approved-only Drivers intentionally hidden from `/v1/driver`.
`Vehicles` remains plural to match the accepted target model but exposes no fake
add/remove/switch/edit controls before the multi-vehicle backend exists.

---

## Verification

Backend verification to run from `backend/`:

```text
go test -p 1 ./...
go vet ./...
```

Database integration tests require the existing dedicated `_test` database setup.
Reviewer integration coverage includes approval promotion, explicit service enrollment,
no implied lower-service enrollment, rejection without operational records, and the
active-only availability guard.

Flutter verification to run from `mobile/`:

```text
dart format --output=none --set-exit-if-changed lib test
dart run build_runner build
flutter analyze
flutter test
```

Flutter format, analysis, and test validation runs on Flutter 3.47.2 before
physical-device validation. Device verification remains required for shell navigation,
readability, back navigation, and confirming the shared dashboard behavior is unchanged.

---

## Next implementation order

1. Validate and merge the current read-only Driver details and Vehicles shell slice.
2. Replace the database one-vehicle assumption with Driver → multiple vehicles while
   preserving historical vehicle associations.
3. Add approved vehicle + approved service enrollment reads and operational selection.
4. Require selected approved vehicle/service when transitioning the Driver to
   operational `active`/online state and lock switching until offline.
5. Run the broader physical-device gate for vehicle/service operation.
6. Then continue Driver marketplace/offers client work, Rider offer selection, and
   Trip controls/history.

The future approved-information revision/cancellation-window/withdrawal-appeal
slice follows when profile/vehicle change management becomes a concrete dependency.

---

## Architecture authority

Accepted ADRs, `architecture-decisions.md`, `product-and-capability-model.md`,
`mvp-scope.md`, and this worklog describe the intended MVP. ADR-0007 is the Ride
Request marketplace authority, ADR-0008 is the dashboard interaction authority,
ADR-0009 is the Driver service/vehicle/onboarding authority, and ADR-0010 is the
initial Driver onboarding reviewer authority.

Implementation must not silently redefine these product rules.

## Deferred

- Fixed pickup/search radius and service areas until justified by launch policy.
- Routing, ETA, geocoding, and PostGIS.
- Live location streaming, breadcrumbs, and push notifications.
- Redis, background dispatch workers, and advanced dispatch optimization.
- Sophisticated payments, cancellation fees, refunds, and no-show policy.
- Full administrator operations beyond the minimal review boundary needed by the
  current Driver application flow.
- Proper administrator identity/role management beyond the temporary internal
  reviewer credentials.
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

## Rider ↔ Driver ride milestone (in review)

- Added service-scoped marketplace operations and captured vehicle/service/fare context for offers and trips, with exact offer-revision acceptance.
- Connected Rider offer choice and Driver marketplace/trip controls, foreground refresh/location recovery, and terminal history.
- See [acceptance and rollout](rider-driver-ride-milestone.md). Physical-device acceptance is pending; this milestone is not yet complete.
