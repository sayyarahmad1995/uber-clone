# Post-Merge Architecture Conformance Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Repair the post-merge client/server contract and documentation drift found on `main` while preserving HiGO's current modular-monolith, PostgreSQL-authoritative, shared-Flutter MVP architecture.

**Architecture:** Keep the Go backend as a modular monolith with PostgreSQL owning transactional correctness, keep application-owned authentication with the Kratos adapter, and keep polling for the pilot. Repair the Flutter transport/lifecycle seams first, then remove obsolete client APIs, reconcile authoritative documentation, consolidate only the lifecycle types that are currently duplicated, and finally extract polling mechanics without merging Rider/Driver business state.

**Tech Stack:** Go, PostgreSQL, Flutter 3.47.2, Dart 3.13.2, Riverpod, Dio, Freezed/json_serializable, Docker Compose, Ory Kratos.

**Spec:** `docs/architecture-decisions.md`, `docs/mvp-scope.md`, `docs/ADR-0007-ride-request-marketplace-model.md`, `docs/ADR-0009-driver-service-vehicle-eligibility.md`, `docs/ADR-0011-minimal-cash-settlement-and-receipt.md`, `docs/flutter-client-architecture.md`, `docs/pilot-deployment.md`.

## Global Constraints

- Preserve the Go modular monolith; do not introduce microservices, Redis, queues, WebSockets, Kubernetes, a second application database, or an event bus.
- PostgreSQL remains authoritative for marketplace deadlines, locking, assignment uniqueness, Trip state, settlement state, and historical operation snapshots.
- Driver responses create offers only; Rider selection remains the only marketplace assignment boundary.
- Trip lifecycle and settlement lifecycle remain independent. Cash collection must not synthesize a Trip status named `settled`.
- Preserve existing public HTTP contracts unless this plan explicitly repairs a client call to the already-existing backend contract.
- Flutter ride-flow repositories expose semantic operations; concrete URL paths remain private to API repository implementations.
- Rider and Driver ride-flow controllers remain separate state machines.
- Application authentication remains provider-neutral with Kratos behind backend adapters; the Flutter client never calls Ory directly.
- Polling remains the pilot update mechanism.
- Local development and pilot deployment remain separate Compose configurations.
- Use TDD for every behavior change. Each task ends with an independently reviewable commit.
- Do not stage or commit local environment artifacts such as `backend/.gocache/`, `backend/.gomodcache/`, `mobile/config/`, or `mobile/devtools_options.yaml`.

---

### Task 1: Repair semantic ride-flow HTTP contract mapping

**Files:**
- Modify: `mobile/lib/features/ride_flow/ride_flow_repository.dart`
- Modify: `mobile/test/ride_flow_repository_test.dart`

**Interfaces:**
- Consumes: backend route `POST /v1/driver/ride-requests/{ride_request_id}/accept`.
- Produces: `RideFlowRepository.acceptProposedFare(String rideRequestId)` mapped to the existing backend route; adapter-level tests covering every semantic ride-flow method's HTTP method/path/payload.

- [ ] **Step 1: Replace the fake-only repository test with an HTTP recording adapter harness**

Create a `RecordingRideFlowAdapter implements HttpClientAdapter` inside `mobile/test/ride_flow_repository_test.dart` that captures the last `RequestOptions`, returns minimal valid JSON for each GET path, and returns `{}` for command responses.

The test fixture must construct:

```dart
final dio = Dio(BaseOptions(baseUrl: 'http://application.test'))
  ..httpClientAdapter = adapter;
final repository = ApiRideFlowRepository(dio, StaticSessionStore());
```

Reuse the existing authenticated `StaticSessionStore` test double where available.

- [ ] **Step 2: Add a failing exact-fare acceptance contract test**

Add:

```dart
test('accept proposed fare uses existing backend accept route', () async {
  await repository.acceptProposedFare('ride-1');
  expect(adapter.request!.method, 'POST');
  expect(adapter.request!.path, '/v1/driver/ride-requests/ride-1/accept');
});
```

Run:

```text
flutter test test/ride_flow_repository_test.dart
```

Expected before the implementation change: FAIL because the client calls `/accept-proposed-fare`.

- [ ] **Step 3: Repair the client route without changing the backend**

Change only `ApiRideFlowRepository.acceptProposedFare`:

```dart
await _request(
  '/v1/driver/ride-requests/$rideRequestId/accept',
  'POST',
  null,
);
```

- [ ] **Step 4: Add adapter-level tests for every remaining semantic operation**

Cover exact method/path/payload for:

```text
listMarketplaceRequests       GET  /v1/driver/marketplace/ride-requests
getCurrentDriverTrip          GET  /v1/driver/trip
listDriverTrips               GET  /v1/driver/trips
listRiderRides                GET  /v1/ride-requests
getRiderRide                  GET  /v1/ride-requests/{id}
listRiderOffers               GET  /v1/ride-requests/{id}/offers
getDriverLocation             GET  /v1/ride-requests/{id}/driver-location
submitOffer                   PUT  /v1/driver/ride-requests/{id}/offer
acceptProposedFare            POST /v1/driver/ride-requests/{id}/accept
selectOffer                   POST /v1/ride-requests/{id}/offers/{driver}/accept
startTrip                     POST /v1/driver/ride-requests/{id}/start
completeTrip                  POST /v1/driver/ride-requests/{id}/complete
confirmCashCollected          POST /v1/driver/ride-requests/{id}/cash-collected
cancelDriverTrip              POST /v1/driver/ride-requests/{id}/cancel
```

Assert `submitOffer` sends `{'amount_minor': amountMinor}` and `selectOffer` sends UTC `updated_at`.

- [ ] **Step 5: Run focused verification**

```text
flutter format lib/features/ride_flow/ride_flow_repository.dart test/ride_flow_repository_test.dart
flutter analyze
flutter test test/ride_flow_repository_test.dart
```

Expected: analyzer clean; repository tests pass.

- [ ] **Step 6: Commit**

```text
git add mobile/lib/features/ride_flow/ride_flow_repository.dart mobile/test/ride_flow_repository_test.dart
git commit -m "fix: align ride flow client routes with backend"
```

---

### Task 2: Correct Rider terminal-state semantics after cash settlement

**Files:**
- Modify: `mobile/lib/features/rider_request/application/rider_request_controller.dart`
- Modify: `mobile/test/ride_request_controller_test.dart`

**Interfaces:**
- Consumes: `RideRequest.trip.status` and `RideRequest.trip.settlement.status`.
- Produces: a Rider active-request selector that treats completed/cash-collected rides as terminal without inventing a `settled` Trip state.

- [ ] **Step 1: Add failing lifecycle tests**

Add focused cases proving:

```text
requested + no Trip                         => active
accepted + assigned Trip                    => active
accepted + in_progress Trip                 => active
accepted + completed + unsettled            => active
accepted + completed + cash_collected       => terminal
cancelled request                            => terminal
expired request                              => terminal
cancelled Trip                               => terminal
```

The key assertion should be equivalent to:

```dart
expect(
  stateWith(
    tripStatus: 'completed',
    settlementStatus: 'cash_collected',
  ).active,
  isNull,
);
```

Run:

```text
flutter test test/ride_request_controller_test.dart
```

Expected before the fix: the completed/cash-collected case remains active.

- [ ] **Step 2: Replace the retired synthetic `settled` check**

Refactor `RiderRequestState.active` so terminal Trip logic is explicit:

```dart
final trip = request.trip;
final tripTerminal = trip?.status == 'cancelled' ||
    (trip?.status == 'completed' &&
        trip?.settlement?.status == 'cash_collected');

if (request.status == 'cancelled' ||
    request.status == 'expired' ||
    tripTerminal) {
  continue;
}
```

Do not add a `settled` Trip status anywhere.

- [ ] **Step 3: Add a submission regression test**

Prove a Rider can submit a new ride after the previous ride is `completed + cash_collected`, while `completed + unsettled` still blocks a second active ride.

- [ ] **Step 4: Run focused verification**

```text
flutter format lib/features/rider_request/application/rider_request_controller.dart test/ride_request_controller_test.dart
flutter analyze
flutter test test/ride_request_controller_test.dart
```

Expected: analyzer clean; all Rider request lifecycle tests pass.

- [ ] **Step 5: Commit**

```text
git add mobile/lib/features/rider_request/application/rider_request_controller.dart mobile/test/ride_request_controller_test.dart
git commit -m "fix: close rider ride after cash settlement"
```

---

### Task 3: Make Flutter historical and unavailable marketplace data compatible with backend contracts

**Files:**
- Modify: `mobile/lib/features/ride_flow/domain/ride_offer.dart`
- Modify: `mobile/lib/features/ride_flow/domain/ride_snapshot.dart`
- Modify: `mobile/lib/features/rider_request/domain/ride_request.dart`
- Modify: `mobile/lib/features/ride_flow/ride_flow_panels.dart`
- Modify: `mobile/test/ride_flow_models_test.dart`
- Modify: `mobile/test/ride_request_repository_test.dart`
- Regenerate: affected `*.freezed.dart` and `*.g.dart`

**Interfaces:**
- Consumes: backend `pickup_distance_meters: null` for temporarily unavailable Driver offers; backend omission/nullability of `proposed_fare` for historical ride rows.
- Produces: nullable client fields and safe UI rendering without weakening new-ride creation requirements.

- [ ] **Step 1: Add failing parsing coverage for unavailable Driver distance**

Add a `RiderOfferComparison.fromJson` test with:

```dart
'pickup_distance_meters': null,
'selectable': false,
```

Assert parsing succeeds and `pickupDistanceMeters == null`.

- [ ] **Step 2: Change Rider offer distance to nullable**

Change:

```dart
@JsonKey(name: 'pickup_distance_meters') int? pickupDistanceMeters,
```

Keep marketplace discovery `MarketplaceRequest.pickupDistanceMeters` required because Driver discovery eligibility already requires fresh location.

- [ ] **Step 3: Render nullable Rider offer distance explicitly**

In `ride_flow_panels.dart`, replace unconditional division by `offer.pickupDistanceMeters` with:

```dart
final distance = offer.pickupDistanceMeters;
Text(
  distance == null
      ? 'Pickup distance unavailable'
      : '${(distance / 1000).toStringAsFixed(1)} km to pickup · straight-line distance',
),
```

- [ ] **Step 4: Add failing historical ride parsing tests**

Add cases for both Rider request models where `proposed_fare` is absent or null and assert parsing succeeds with `proposedFare == null`.

- [ ] **Step 5: Make read-side proposed fare nullable while preserving create-side fare requirement**

Change read models to:

```dart
@JsonKey(name: 'proposed_fare') Money? proposedFare,
```

in `features/rider_request/domain/ride_request.dart`, and:

```dart
@JsonKey(name: 'proposed_fare') RideFare? proposedFare,
```

in `features/ride_flow/domain/ride_snapshot.dart`.

Do not change `RideRequestRepository.create`; creation must still require a concrete `Money proposedFare`.

- [ ] **Step 6: Update UI/helper code for nullable historical fare**

Any history/list presentation that reads a Rider request proposed fare must display `Fare unavailable` when null rather than asserting non-null.

- [ ] **Step 7: Regenerate and verify generated code**

```text
dart run build_runner build --delete-conflicting-outputs
flutter format lib test
flutter analyze
flutter test test/ride_flow_models_test.dart test/ride_request_repository_test.dart
```

Expected: nullable JSON parses successfully; analyzer clean.

- [ ] **Step 8: Commit**

```text
git add mobile/lib/features/ride_flow/domain mobile/lib/features/rider_request/domain mobile/lib/features/ride_flow/ride_flow_panels.dart mobile/test/ride_flow_models_test.dart mobile/test/ride_request_repository_test.dart
git commit -m "fix: support nullable ride history contracts"
```

---

### Task 4: Remove the obsolete direct Driver onboarding client seam

**Files:**
- Modify: `mobile/lib/features/driver_workspace/data/driver_repository.dart`
- Modify: `mobile/lib/features/driver_workspace/application/driver_controller.dart`
- Modify: `mobile/test/driver_repository_test.dart`
- Modify: `mobile/test/driver_controller_test.dart`

**Interfaces:**
- Consumes: review-based `DriverOnboardingRepository` flow already used by `DriverWorkspaceScreen`.
- Produces: `DriverRepository` limited to operational Driver reads/actions; no client call to removed `PUT /v1/driver`.

- [ ] **Step 1: Add a static regression assertion before deletion**

Add or update a test asserting the operational repository surface contains only:

```text
get
operatingState
selectOperation
listVehicles
setOnline
publishLocation
```

Do not preserve an onboarding method for compatibility.

- [ ] **Step 2: Delete `DriverRepository.onboard` and `ApiDriverRepository.onboard`**

Remove:

```dart
Future<DriverProfile> onboard(String displayName, DriverVehicle vehicle);
```

and the implementation that calls `PUT /v1/driver`.

- [ ] **Step 3: Delete `DriverController.onboard`**

Remove the controller method delegating to the obsolete repository operation.

- [ ] **Step 4: Remove obsolete tests**

Delete the `repo.onboard(...)` expectation from `driver_repository_test.dart` and any `DriverController.onboard` test coverage. Preserve tests for reviewed onboarding in `driver_onboarding_repository_test.dart` and Driver workspace application flow tests.

- [ ] **Step 5: Run forbidden-reference checks**

Run repository searches and require zero production/test matches for:

```text
repo.onboard(
DriverController.onboard
Future<DriverProfile> onboard(
'/v1/driver' used as a PUT onboarding write
```

The legitimate `GET /v1/driver` profile read must remain.

- [ ] **Step 6: Run focused verification**

```text
flutter format lib/features/driver_workspace test/driver_repository_test.dart test/driver_controller_test.dart
flutter analyze
flutter test test/driver_repository_test.dart test/driver_controller_test.dart test/driver_onboarding_repository_test.dart test/driver_workspace_test.dart
```

Expected: analyzer clean; onboarding remains available only through review-based APIs.

- [ ] **Step 7: Commit**

```text
git add mobile/lib/features/driver_workspace mobile/test/driver_repository_test.dart mobile/test/driver_controller_test.dart
git commit -m "refactor: remove legacy mobile driver onboarding path"
```

---

### Task 5: Make marketplace command scope explicit

**Files:**
- Modify: `docs/flutter-client-architecture.md`
- Modify: `docs/ADR-0007-ride-request-marketplace-model.md`
- Test: architecture/documentation grep checks only

**Interfaces:**
- Consumes: backend support for Driver `skip` and Rider offer `reject`.
- Produces: an explicit MVP statement that the current mobile client exposes accept/counteroffer/select while backend skip/reject remain supported API operations but are not required mobile controls in this slice.

- [ ] **Step 1: Add explicit scope text to ADR-0007**

Add a short implementation-status paragraph stating:

```text
The backend retains explicit Driver skip and Rider offer rejection operations.
The current mobile MVP is not required to expose dedicated controls for those
operations; omission of a mobile button is not a different assignment model.
If either control is added later, it must use semantic repository operations and
must not change Rider-selected assignment ownership.
```

- [ ] **Step 2: Reflect the same boundary in Flutter architecture docs**

State that the implemented mobile marketplace commands are exact-fare response, counteroffer, Rider selection, Trip transitions, cancellation, and cash confirmation; skip/reject UI is intentionally deferred.

- [ ] **Step 3: Verify no code is added solely to expose deferred controls**

No production Dart or Go file should change in this task.

- [ ] **Step 4: Commit**

```text
git add docs/flutter-client-architecture.md docs/ADR-0007-ride-request-marketplace-model.md
git commit -m "docs: clarify marketplace mobile command scope"
```

---

### Task 6: Reconcile authoritative architecture documentation with the implemented system

**Files:**
- Modify: `docs/technology-stack.md`
- Modify: `docs/mvp-scope.md`
- Modify: `docs/flutter-client-architecture.md`
- Modify: `docs/architecture-decisions.md`
- Review: `docs/authentication.md`
- Review: `identity/README.md`

**Interfaces:**
- Consumes: current implementation on `main`.
- Produces: authoritative docs that describe the actual MVP architecture without resurrecting Hydra/OIDC, Redis, WebSockets, or obsolete Driver setup assumptions.

- [ ] **Step 1: Correct the technology stack table**

Set current MVP decisions to:

```text
Real-time/update transport: HTTP polling for ride-flow pilot
Maps: flutter_map + OpenStreetMap tile adapter
Device location: geolocator behind DeviceLocation
Cache / fast ephemeral data: none currently required; Redis deferred until justified
Authentication: application-owned HTTP/session contracts with Ory Kratos adapter
```

Keep future directions provider-neutral and avoid promising WebSockets/Redis.

- [ ] **Step 2: Correct MVP identity wording**

Replace `External OIDC authentication` in `docs/mvp-scope.md` with language matching the current decision:

```text
Application-owned authentication/session APIs backed by provider-neutral backend interfaces, with Ory Kratos as the current adapter.
```

Hydra/OIDC + PKCE remain historical/deferred only.

- [ ] **Step 3: Bring Flutter architecture status forward**

Update `docs/flutter-client-architecture.md` so it states that marketplace discovery, typed offers, Rider selection, Trip execution, settlement display, and split Rider/Driver ride-flow controllers are implemented.

Remove statements that marketplace/offers are the next client slice. Replace obsolete Driver profile setup/edit wording with review-based onboarding, approved vehicle/service operating selection, and read-only approved information.

- [ ] **Step 4: Remove stale transitional language from the living architecture decisions**

In `docs/architecture-decisions.md`, update sections that still say operating-context/multi-vehicle work is not implemented. Preserve ADR-0009's deferred revision-governance work; do not claim approved-information editing exists.

- [ ] **Step 5: Cross-check identity docs**

Re-read `docs/authentication.md` and `identity/README.md`. Keep them unchanged if they already consistently state application-owned APIs + Kratos adapter and Hydra/OIDC deferred.

- [ ] **Step 6: Run documentation conformance searches**

Current architecture docs must not describe these as active MVP choices:

```text
External OIDC authentication
Maps | TBD
Device location | TBD
Redis as an active dependency
WebSocket as the implemented ride-flow transport
marketplace discovery and offers remain the next client slice
Hydra/PKCE as active runtime architecture
```

Historical ADR context may retain superseded wording only when explicitly labeled historical/superseded/deferred.

- [ ] **Step 7: Commit**

```text
git add docs/technology-stack.md docs/mvp-scope.md docs/flutter-client-architecture.md docs/architecture-decisions.md
git commit -m "docs: reconcile architecture with current implementation"
```

---

### Task 7: Consolidate the shared Flutter Trip and settlement lifecycle contract

**Files:**
- Create: `mobile/lib/features/ride_flow/domain/ride_execution.dart`
- Modify: `mobile/lib/features/ride_flow/domain/trip.dart`
- Modify: `mobile/lib/features/rider_request/domain/ride_request.dart`
- Modify: consumers/tests importing the duplicated lifecycle types
- Regenerate: affected Freezed/json_serializable output

**Interfaces:**
- Consumes: one authoritative client meaning for Trip status and settlement status.
- Produces: shared `TripLifecycleSnapshot` / `SettlementSnapshot` definitions used by both Rider request list state and full ride-flow state, while keeping feature-specific wrappers separate.

- [ ] **Step 1: Add a cross-feature lifecycle contract test**

Create a focused model test proving both Rider request parsing and ride-flow parsing accept the same canonical payload:

```json
{
  "status": "completed",
  "settlement": {
    "status": "cash_collected",
    "method": "cash",
    "cash_collected_at": "2026-09-17T00:00:00Z"
  }
}
```

The test must fail if either feature reintroduces a synthetic `settled` Trip status or interprets settlement differently.

- [ ] **Step 2: Introduce the shared lifecycle file**

Define only the lifecycle concepts that are genuinely shared:

```dart
@freezed
abstract class SettlementSnapshot with _$SettlementSnapshot {
  const factory SettlementSnapshot({
    required String status,
    String? method,
    @JsonKey(name: 'cash_collected_at') DateTime? cashCollectedAt,
  }) = _SettlementSnapshot;

  factory SettlementSnapshot.fromJson(Map<String, dynamic> json) =>
      _$SettlementSnapshotFromJson(json);
}
```

and a small shared Trip lifecycle representation containing `status`, `startedAt`, `completedAt`, `cancelledAt`, and `settlement` only if both features consume those fields. Do not move location, operation-context, Rider identity, or Driver history fields into a global model.

- [ ] **Step 3: Replace duplicate settlement/lifecycle declarations**

Import the shared lifecycle types from both `ride_request.dart` and `trip.dart`. Preserve feature-specific `TripSnapshot` wrappers where their payload shapes differ.

- [ ] **Step 4: Regenerate and run all model/controller tests**

```text
dart run build_runner build --delete-conflicting-outputs
flutter format lib test
flutter analyze
flutter test test/ride_flow_models_test.dart test/ride_request_controller_test.dart test/ride_request_repository_test.dart
```

Expected: no duplicate lifecycle semantics, no generated drift.

- [ ] **Step 5: Commit**

```text
git add mobile/lib/features/ride_flow/domain mobile/lib/features/rider_request/domain mobile/test
git commit -m "refactor: share trip settlement lifecycle contract"
```

---

### Task 8: Extract polling mechanics without merging Rider/Driver state

**Files:**
- Create: `mobile/lib/features/ride_flow/application/polling_loop.dart`
- Create: `mobile/test/ride_flow_polling_loop_test.dart`
- Modify: `mobile/lib/features/ride_flow/application/rider_active_ride_controller.dart`
- Modify: `mobile/lib/features/ride_flow/application/driver_marketplace_controller.dart`
- Modify: `mobile/lib/features/ride_flow/application/driver_trip_controller.dart`
- Modify: controller tests as needed

**Interfaces:**
- Consumes: existing five-second polling behavior, foreground suspension, command serialization, authoritative reload after ambiguous command failure.
- Produces: a mechanics-only polling helper; Rider/Driver domain state remains owned by the three separate controllers.

- [ ] **Step 1: Write polling helper behavior tests**

Cover:

```text
start schedules next refresh only after current refresh completes
background/suspend cancels pending timer
resume triggers one immediate refresh
concurrent passive refresh is ignored
command waits for an in-flight refresh before executing
failed command can execute an authoritative reload before the error is surfaced
stop/dispose prevents future scheduling
```

Use fake asynchronous callbacks and a short injected interval; do not use real five-second waits.

- [ ] **Step 2: Implement a mechanics-only helper**

Expose a small API such as:

```dart
class PollingLoop {
  PollingLoop({required Duration interval});

  Future<void> runRefresh(Future<void> Function() task);
  Future<void> runCommand(Future<void> Function() task);
  void setForeground(bool value, Future<void> Function() refresh);
  void dispose();
  bool get busy;
}
```

If the exact signature needs a callback for post-command reload, keep that callback explicit. Do not let `PollingLoop` know about rides, offers, Trips, repository types, or UI errors.

- [ ] **Step 3: Migrate RiderActiveRideController**

Keep Rider fields (`riderRide`, `offers`, `location`, `error`) and Rider load sequencing in the controller. Delegate only timer/serialization/foreground mechanics.

- [ ] **Step 4: Migrate DriverMarketplaceController**

Keep marketplace requests and offer commands in the controller. Delegate only mechanics.

- [ ] **Step 5: Migrate DriverTripController**

Keep current Trip, history, and Trip/settlement commands in the controller. Delegate only mechanics.

- [ ] **Step 6: Verify state-machine separation remains intact**

Run searches that require no new combined `RideFlowController`, no cross-controller mutable state object, and no raw transport paths in controllers.

- [ ] **Step 7: Run focused and full Flutter verification**

```text
flutter format lib test
flutter analyze
flutter test test/ride_flow_polling_loop_test.dart test/ride_flow_controller_split_test.dart test/ride_flow_test.dart
flutter test
```

Expected: full suite green; polling behavior unchanged.

- [ ] **Step 8: Commit**

```text
git add mobile/lib/features/ride_flow/application mobile/test/ride_flow_polling_loop_test.dart mobile/test/ride_flow_controller_split_test.dart mobile/test/ride_flow_test.dart
git commit -m "refactor: centralize ride flow polling mechanics"
```

---

### Task 9: Make pilot backup/restore cover application and identity data

**Files:**
- Modify: `docs/pilot-deployment.md`
- Modify: `README.md` only if it links to incomplete backup guidance

**Interfaces:**
- Consumes: one PostgreSQL container hosting both `$POSTGRES_DB` and `$ORY_KRATOS_DB`.
- Produces: a pilot runbook that backs up and restore-verifies both application and Kratos databases.

- [ ] **Step 1: Replace the single-database backup example with two explicit dumps**

Document:

```bash
docker compose -f docker-compose.pilot.yml exec -T postgres \
  pg_dump -U "$POSTGRES_USER" "$POSTGRES_DB" > higo-app-backup.sql

docker compose -f docker-compose.pilot.yml exec -T postgres \
  pg_dump -U "$POSTGRES_USER" "$ORY_KRATOS_DB" > higo-kratos-backup.sql
```

- [ ] **Step 2: Add restore-verification requirements**

State that both dumps must be restored into isolated databases and checked before being considered valid backups. Include verification that the application database migrations/schema load and that Kratos can start against the restored identity database.

- [ ] **Step 3: Keep deployment scope unchanged**

Do not add a second PostgreSQL server, backup service, object store, or orchestration platform in this task.

- [ ] **Step 4: Commit**

```text
git add docs/pilot-deployment.md README.md
git commit -m "docs: cover identity database in pilot backups"
```

---

### Task 10: Add final conformance gates and verify the whole repository

**Files:**
- Modify: `.github/workflows/readiness.yml`
- Modify/add focused tests only where required by the checks below

**Interfaces:**
- Consumes: all preceding tasks.
- Produces: CI guards against the exact post-merge regressions found by the architecture audit.

- [ ] **Step 1: Preserve existing readiness checks**

Keep the Flutter 3.47.2 pin, Dart constraint, generated-code drift check, Go formatting/vet/tests, Flutter formatting/analyze/tests, and HTTP raw-database boundary guard.

- [ ] **Step 2: Add a legacy Driver onboarding boundary guard**

Add a CI search over production Dart that fails on the removed immediate onboarding contract, specifically the combination of a Driver repository write to `/v1/driver` or a production `onboard(` method in `driver_repository.dart` / `driver_controller.dart`.

Do not reject the legitimate `GET /v1/driver` profile read.

- [ ] **Step 3: Add a retired Trip status guard**

Fail CI if production Dart under Rider/ride-flow features compares a Trip status to `'settled'`.

Settlement status strings such as `cash_collected` remain valid.

- [ ] **Step 4: Add the exact semantic repository adapter test to required CI tests**

Ensure `mobile/test/ride_flow_repository_test.dart` is part of the normal full Flutter suite and does not rely only on interface fakes.

- [ ] **Step 5: Run complete backend verification**

From `backend/`:

```text
gofmt -w <changed-go-files-if-any>
go vet ./...
go test -p 1 ./...
```

Expected: all packages pass.

- [ ] **Step 6: Run complete Flutter verification**

From `mobile/` with Flutter 3.47.2 / Dart 3.13.2:

```text
dart run build_runner build --delete-conflicting-outputs
git diff --exit-code -- lib test
flutter analyze
flutter test
```

Expected: generated code clean, analyzer clean, full suite green.

- [ ] **Step 7: Run architecture searches**

Require all of these invariants:

```text
backend/internal/httpapi has no database/sql import or *sql.DB field
no production json.RawMessage operation-context boundary
no monolithic RideFlowController
no public ride-flow get(path)/act(path) repository methods
no production Driver direct-onboarding write seam
no production Trip status == 'settled'
Hydra/PKCE references are historical/deferred only
pilot Compose has no Kratos --dev and no PostgreSQL published port
```

- [ ] **Step 8: Render both Compose configurations**

Run:

```text
docker compose -f docker-compose.yml config
docker compose -f docker-compose.pilot.yml config
```

using required pilot environment variables. Expected: both render successfully.

- [ ] **Step 9: Run whitespace verification**

```text
git diff --check
```

Expected: no output.

- [ ] **Step 10: Commit**

```text
git add .github/workflows/readiness.yml mobile backend docs README.md
git commit -m "ci: guard post-merge architecture conformance"
```

---

## Final Acceptance Matrix

The plan is complete only when all of the following are true on the final branch:

- Driver exact-fare acceptance uses `POST /v1/driver/ride-requests/{id}/accept` and is protected by an adapter-level HTTP contract test.
- Completed/cash-collected Rider rides are terminal; completed/unsettled rides remain visible until cash collection.
- No Flutter production code treats `settled` as a Trip status.
- Rider offer pickup distance safely supports `null` when the Driver is temporarily unavailable.
- Historical rides without a proposed fare remain readable in Flutter.
- The operational `DriverRepository` contains no direct onboarding write path; reviewed onboarding is the only client onboarding write flow.
- Backend skip/reject operations have an explicit documented mobile-scope decision.
- Current architecture docs match application-owned Kratos-backed sessions, HTTP polling, implemented maps/location adapters, no active Redis requirement, implemented marketplace/Trip/settlement slices, and current Driver operating-selection behavior.
- Shared Trip/settlement lifecycle meaning is defined once on the Flutter side; feature-specific payload models remain separate where their shapes differ.
- Polling mechanics are shared without recombining Rider and Driver state machines.
- Pilot backup guidance covers both the HiGO application database and the Kratos identity database.
- `go vet ./...` passes.
- `go test -p 1 ./...` passes.
- `flutter analyze` passes on Flutter 3.47.2 / Dart 3.13.2.
- Full `flutter test` passes.
- Both local and pilot Compose configurations render successfully.
- Static architecture checks pass.
- `git diff --check` is clean.

## Self-Review Result

- Spec coverage: all runtime defects, stale onboarding seam, documentation drift, lifecycle duplication, polling duplication, and pilot backup gap from the architecture audit are assigned to explicit tasks.
- Placeholder scan: no implementation step depends on a TBD/TODO or an unspecified future decision.
- Type consistency: nullable Rider offer distance and historical proposed fare are introduced before UI/model consumers are updated; shared lifecycle extraction happens only after behavior is corrected and tested.
- Scope control: no new infrastructure, transport, service boundary, product mode, or payment platform is introduced.