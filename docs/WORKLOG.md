# Work Log

## Current status

The backend implements accounts/authentication, Driver onboarding/review and operating
context, ride requests, service-scoped geographic marketplace discovery and offers,
Rider-selected assignment, Trip execution, cancellation, history, Driver location,
minimal cash settlement/receipt data, and the service catalog for Driver onboarding. Marketplace and assignment preserve the selected approved
vehicle/service and agreed fare snapshot rather than following later profile changes.

The shared Flutter Android client implements account entry, Rider-first capability
selection, Rider request creation with Economy/Comfort and proposed fare, Rider offer
comparison/selection/rejection, Driver marketplace responses and counteroffers, Trip
start/completion/cancellation, foreground recovery/location publication, cash
collection confirmation and receipt/history presentation, Driver onboarding/review
state, and the ADR-0008 capability-shell dashboard. Android background Driver
presence is implemented, but Google Maps route preview and dynamic Rider catalog
rendering are not yet implemented.

PR #70 introduced the ADR-0009 new-Driver onboarding application model. A Driver
chooses one service for the first vehicle, reviews the application, and submits it
for review. Submitted data remains separate from approved Driver/vehicle data and
can be restored as pending, approved, or rejected state. The public legacy
`PUT /v1/driver` write path is removed so Driver accounts cannot bypass review.

The reviewer slice implements the narrow administration dependency defined by
ADR-0010. Reviewer routes are registered only when internal reviewer credentials
are configured. A reviewer can list pending applications, inspect one application,
approve it, or reject it with a required reason through a small server-rendered page
or the protected reviewer API.

Approval transactionally promotes the submitted snapshot into an approved Driver
profile, vehicle, and explicit selected-service enrollment while retaining the
application decision/history. Rejection creates no Driver/vehicle/service records.

Only Rider selection assigns a Trip. Accepting the Rider's proposed fare and
counteroffering both create pending offers; neither reserves the ride or Driver.
Rejected offers leave the Rider's active comparison but remain persisted, and a
fresh response from the same eligible Driver may reopen the offer as pending.

PR #80 completed the Rider ↔ Driver marketplace-to-trip mobile milestone. Automated
backend/PostgreSQL and Flutter validation passed before merge, and the complete
physical-device acceptance matrix was confirmed by the end of the milestone on
2026-09-12. Device coverage included exact fares, Driver counteroffers, offer
rejection/resubmission, Rider selection, assignment, start/completion, Rider and
Driver cancellation, network failure/recovery, and reopening during
assigned/in-progress trips.

---

## Completed milestones

- [x] Minimal cash settlement/receipt API and Flutter confirmation/history — PRs #83, #84, #86, #87 and #88; full physical-device cash checklist still requires a recorded pass
- [x] Cash ride pilot acceptance **checklist document** — PR #89 (not evidence of a successful acceptance run)
- [x] Android Driver background presence and guarded Driver-to-Rider switch — PRs #113–#118

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
- [x] Rider ↔ Driver Marketplace and Trip Milestone — PR #80

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

- MVP operating context is exactly one selected verified vehicle. Marketplace
  eligibility is derived from **all** active approved service enrollments on it.
- The current operating selection has no independent global service selector.
- Changing the selected operating vehicle requires the Driver to go offline first.
- Going online publishes current location before the online transition.
- Going offline never requires location permission.
- Marketplace eligibility additionally requires Driver capability/approval, online
  state, fresh location, and no conflicting active Trip or other applicable commitment.
- Marketplace/offers/assignment/Trip history preserve the vehicle/service context
  used at the time rather than following later Driver selections.

### Marketplace and assignment

- A new ride request requires pickup, destination, service, and proposed fare/currency.
- There is no Rider booking-mode choice.
- Exact-fare responses and counteroffers both create/update pending offers.
- Drivers may offer on multiple rides; multiple Drivers may offer on one ride.
- Rider active comparison returns pending offers; rejected offers remain persisted.
- A rejected Driver may submit a fresh response while the request remains open and
  eligible, which reopens that Driver's offer as pending with the current fare/context.
- Rider selection checks the displayed offer revision, locks and revalidates the
  request, offer, and Driver, and requires current operating eligibility.
- At most one Trip exists per ride, and a Driver has at most one active Trip.
- Competing offers on the selected ride close.
- Historical rides without service/fare context remain readable but cannot enter
  the current marketplace flow.

ADR-0007 remains authoritative for the Ride Request marketplace.

### Trip lifecycle

- Trip states: `assigned`, `in_progress`, `completed`, and `cancelled`.
- Start and completion are idempotent. Completion requires an in-progress Trip.
- Rider and assigned Driver cancellation share transactional behavior.
- Cancellation closes pending offers for that ride and is idempotent.
- Completed Trips cannot be cancelled.
- Assigned/in-progress Trip state is server-owned and recovers after app reopening
  or transient network failure.

### Settlement and receipt

- ADR-0011 selects minimal cash settlement and receipt display as the next MVP slice.
- The selected offer's agreed fare snapshot remains the authoritative commercial
  amount for settlement.
- Settlement is scoped to completed ride-hailing Trips and cash handling only.
- Receipt/history display should use immutable Trip context: agreed fare, service,
  Driver/vehicle context, completion time, Trip status, and settlement status.
- This slice must not introduce card payments, wallets, refunds, cancellation fees,
  no-show policy, payout reconciliation, commissions, promo credits, or broad
  administrator payment operations.

See [ADR-0011](ADR-0011-minimal-cash-settlement-and-receipt.md).

### Location and geographic policy

- One latest location row is stored per Driver, separate from profile/vehicle.
- Server-owned timestamps support freshness decisions.
- Marketplace locations must be no older than two minutes and not in the future.
- Discovery ranks eligible requests by Haversine pickup distance before the feed limit.
- Raw Driver coordinates and license plates are not exposed before assignment.
- No arbitrary pickup radius, service boundary, Driver pickup ETA, or PostGIS requirement
  has been introduced. Selected Rider trip routing and advisory fares are the next
  milestone under ADR-0013; they are not yet implemented.

### Flutter client

- The client calls only application-owned APIs and does not expose Ory concepts.
- Session secrets use secure storage; capability preference uses simple preferences.
- Rider and Driver use the shared ADR-0008 dashboard interaction contract.
- New Driver onboarding loads the service catalog, restores the latest application,
  runs server precheck, displays a review confirmation, and submits for review.
- Pending/rejected/approved application state is shown from backend data.
- The capability shell owns Rider/Driver switching, Driver enablement, real secondary
  destinations, and logout; operational controls remain on the dashboard.
- Rider ride flow supports service/fare request creation, active offer comparison,
  offer rejection/selection, assigned/in-progress recovery, cancellation, and history.
- Driver ride flow supports service-scoped discovery, exact-fare response,
  counteroffer, assignment recovery, start/completion/cancellation, and history.
- Foreground polling is serialized with user commands so an active refresh cannot
  silently drop a marketplace action; failed commands reload authoritative state.
- Existing operational Driver accounts retain the PR #69 readiness path temporarily.

---

## Earlier onboarding implementation context

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

### Flutter onboarding context

The Driver onboarding screen distinguishes:

- operational Driver profile;
- onboarding loading/error;
- service/vehicle application form;
- application pending/rejected/approved status.

The form selects one service, collects Driver/vehicle information, performs the
server precheck, shows a final review dialog, and submits a pending application.
Backend `error` text is preserved when a separate `message` field is absent.

### Capability shell and read-only Driver surfaces

PR #72 moved secondary navigation into a standard Flutter Drawer, removed the bottom
capability switch, and made committed dashboard extent persist across workflow and
Rider/Driver transitions while keeping ADR-0008 gesture semantics intact.

`Driver details` and `Vehicles` routes use real operational-profile or onboarding
application data and do not expose fake edit controls when the corresponding reviewed
change workflow is not implemented.

---

## Verification

Backend verification to run from `backend/`:

```text
go test -p 1 ./...
go vet ./...
```

Database integration tests require the existing dedicated `_test` database setup.
Marketplace coverage includes service/selection eligibility, immutable offer/trip
snapshots, revised fares, concurrent assignment, rejected-offer visibility and
resubmission, cancellation, and location rules.

Flutter verification to run from `mobile/`:

```text
dart format --output=none --set-exit-if-changed lib test
dart run build_runner build
flutter analyze
flutter test
```

PR #80 readiness validation ran the backend/PostgreSQL suite plus `flutter analyze`
and `flutter test`. Physical-device acceptance then verified the marketplace/trip
lifecycle, both cancellation paths, network recovery, and reopening recovery.

PR #81 hardened readiness validation before new product work by adding CI checks for
changed-file Go/Dart formatting, backend vet, backend PostgreSQL tests, explicit
Flutter dependency setup, build-runner execution, generated-code drift, Flutter
analysis, and Flutter tests.

---

## Next implementation order

The owner-selected next feature milestone is **dynamic Rider ride services and Google
Maps booking preview with advisory service-specific fare estimation**. ADR-0012
and ADR-0013 are the implementation contracts. The precise PR order, service
catalog/API design, Google provider boundaries, pricing versioning, and acceptance
gates are in [the implementation plan](superpowers/plans/2026-09-29-google-maps-rider-booking.md).

1. Document and reconcile ADR-0012/ADR-0013 with the current code and worklog.
2. Implement the Rider catalog from existing `driver_service_catalog`; remove
   hardcoded Economy/Comfort checks in Go and Flutter, and prove that a third
   compatible service is visible on an unchanged updated app after refresh.
3. Migrate both shared dashboards to Google Maps while preserving ADR-0008.
4. Add Google Places search/selection and route preview through app-owned Go APIs;
   show road distance, estimated duration and polyline for the Rider's chosen trip.
5. Add versioned service/currency-specific fare policies with owner-approved base
   fare, distance rate, duration rate, minimum fare and rounding; compute an
   advisory suggestion in Go, preserve Rider proposals and Driver counteroffers.
6. Finish Rider booking UX and run the complete [Google booking acceptance matrix](pilot-google-booking-acceptance.md), plus all existing cash-ride regressions.

**Release prerequisite:** run and record the existing [pilot cash ride acceptance
checklist](pilot-cash-ride-acceptance.md) on the current application. Its presence
in the repository does not prove a complete physical-device pass. Any blocking
failure takes priority before expansion into the pilot. Production Economy/Comfort
pricing parameters require explicit owner approval; fixtures are not tariffs.

The future approved-information revision/cancellation-window/withdrawal-appeal
slice remains deferred until profile/vehicle change management becomes a concrete
dependency.

---

## Architecture authority

Accepted ADRs, `architecture-decisions.md`, `product-and-capability-model.md`,
`mvp-scope.md`, and this worklog describe the intended MVP. ADR-0007 is the Ride
Request marketplace authority, ADR-0008 is the dashboard interaction authority,
ADR-0009 is the Driver service/vehicle/onboarding authority, ADR-0010 is the
initial Driver onboarding reviewer authority, ADR-0011 is the minimal cash
settlement and Trip receipt authority, ADR-0012 is the next Rider catalog authority,
and ADR-0013 is the next Google route preview/advisory pricing authority.

Implementation must not silently redefine these product rules.

## Deferred

- Fixed pickup/search radius and service areas until justified by launch policy.
- PostGIS, large-scale Route Matrix, Driver-to-pickup ETA/rerouting and full turn-by-turn navigation. Selected-trip routing and Places geocoding are in the next milestone under ADR-0013.
- Live location streaming, breadcrumbs, and push notifications.
- Redis, background dispatch workers, and advanced dispatch optimization.
- Sophisticated payment infrastructure beyond the minimal cash settlement and
  receipt slice defined by ADR-0011; this includes card processing, wallets,
  payouts, refunds, cancellation fees, no-show policy, commissions, and promo credits.
- Full administrator operations beyond the minimal review boundary needed by the
  current Driver application flow.
- Proper administrator identity/role management beyond the temporary internal
  reviewer credentials.
- Complete versioned Driver/vehicle change review, cancellation-window, and appeal UI.
- Courier, Freight, promotions, analytics platforms, multi-round/chat negotiation,
  full production continuous deployment, Kubernetes, and iOS implementation.

## Working principles

- Build small business slices and verify each before expanding scope.
- Keep business domains transport-neutral and `cmd/api` as the composition root.
- Use application-owned interfaces for external providers.
- Enforce assignment and approval invariants transactionally where implemented.
- Preserve ownership/privacy/history boundaries.
- Add infrastructure only when a concrete flow needs it.
- Keep this worklog aligned with implementation and distinguish target behavior
  from completed behavior.

## Rider ↔ Driver ride milestone (completed)

- Completed in PR #80 and merged into `main` on 2026-09-12 at merge commit `a6fb64975629b595346dfea47448441b8604ef47`.
- Added service-scoped marketplace operations and captured vehicle/service/fare context for offers and trips, with exact offer-revision acceptance.
- Connected Rider offer choice/rejection and Driver exact-fare/counteroffer responses to assignment, Trip controls, foreground recovery/location behavior, and terminal history.
- The complete physical-device acceptance matrix was confirmed by the end of the milestone for full request → offer → Rider choice → start → complete flow, Rider cancellation, Driver cancellation, transient network recovery, and reopening during assigned/in-progress trips.
- Acceptance follow-ups fixed the foreground-poll/counteroffer race and rejected-offer visibility while preserving rejected lifecycle history and Driver resubmission behavior.
- See [completed acceptance and rollout record](rider-driver-ride-milestone.md). No acceptance item remains open for this milestone.
