# Driver Background Presence Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Keep explicitly online Android Drivers available through backgrounding and screen lock by renewing the existing server location lease independently of Flutter UI lifetime.

**Architecture:** A session-scoped `DriverPresenceService` owns exactly one Android `location` foreground task and is injected into the auto-disposed Driver controller. That task reconstructs the existing secure-session/Driver-repository stack and publishes fresh location; the controller never runs a parallel online presence heartbeat. The server's PostgreSQL-backed two-minute freshness check remains authoritative.

**Tech Stack:** Flutter/Dart, Riverpod, `flutter_foreground_task` pinned to 11.0.3 (verify compatibility before dependency write), geolocator, flutter_secure_storage, Dio, Android manifest, Go/PostgreSQL integration tests.

**Spec:** `docs/superpowers/specs/2026-09-19-driver-background-presence-design.md`

## Global Constraints

- Start at `fix/driver-background-presence` commit `7e6ae5d08d4a912ad9c43dd6418e14f1ccef6990`; preserve all prior vehicle eligibility, Decline, and trip-context fixes; no merge or force push.
- No change to marketplace matching, two-minute freshness, ten-second cleanup, trip lifecycle, or foreground marketplace polling; no Redis, queue, WebSocket, microservice, Kubernetes, or database addition.
- Only Android background execution is in scope; there is no iOS project. No boot auto-start, background service launch, or guaranteed immediate offline on process kill.
- Service is the **sole online marketplace-presence publisher even in foreground**. A distinct active-trip location path can remain when offline, but must not concurrently publish online presence.
- Read the spec and current repository APIs before editing; write and observe each focused failing test before its production code; small verified commits; preserve concurrent changes.

## File structure

- `mobile/lib/features/driver_workspace/application/driver_presence_service.dart`: small interface plus session/running-state reconciliation and publisher contract; no Android plugin calls in controller.
- `mobile/lib/features/driver_workspace/data/android_driver_presence_service.dart`: plugin start/stop/notification and task entry point. Separate `driver_presence_publisher.dart` if needed so the task's auth/location logic is unit-testable without plugin channels.
- `mobile/lib/features/driver_workspace/application/driver_controller.dart`: explicit online/offline and resume coordination; remove implicit offline paths and online timer writer.
- `mobile/lib/core/providers.dart`, `mobile/lib/features/authentication/application/session_controller.dart`, `mobile/lib/features/driver_workspace/presentation/driver_workspace_screen.dart`: app/session ownership, sign-out cleanup, lifecycle routing.
- `mobile/android/app/src/main/AndroidManifest.xml`, `mobile/pubspec.yaml`, `mobile/pubspec.lock`: service registration and pinned dependency.
- `mobile/test/driver_controller_test.dart`, `mobile/test/driver_presence_service_test.dart`, `mobile/test/driver_workspace_test.dart`, `mobile/test/test_doubles.dart`: behavior, adapter, lifecycle, and manifest tests.
- `backend/internal/driver/postgres_repository_integration_test.go`: existing freshness/expiry and trip independence strengthened; no production Go changes anticipated.
- `docs/driver-operational-readiness.md`: revised lease and Android-only operational contract.

## Review Focus

1. Token expires during background publishing: task stops using credentials; never writes with expired token (Task 2).
2. User signs out then another account signs in: stale task cannot renew old or new account as the old session (Task 4).
3. Network fails during Go Offline: do not stop a still-online publisher and falsely claim offline (Task 3).
4. Service startup succeeds at Android but callback cannot access secure storage/geolocator: adapter test and physical-device check must catch it, with rollback and no durable-online claim (Tasks 1–3, 6).
5. Online Driver also has active-trip-specific location handling: never create a second online location writer or alter assigned trip (Tasks 3, 5).

---

### Task 1: Service boundary and Android prerequisites

**Files:** Create `mobile/lib/features/driver_workspace/application/driver_presence_service.dart`; create `mobile/test/driver_presence_service_test.dart`; modify `mobile/android/app/src/main/AndroidManifest.xml`, `mobile/pubspec.yaml`, `mobile/pubspec.lock`.

**Interfaces:** Produce `DriverPresenceService` with `Future<bool> isRunningFor(String driverUserId)`, `Future<void> start(String driverUserId)`, `Future<void> stop()`. The non-secret account ID binds a running task to its initiating session; tests use a fake implementing this interface. Never pass a token to `start()`.

- [ ] **Step 1: Confirm baseline and dependency compatibility.** Verify clean HEAD, local `flutter --version`, current Flutter 3.47.2 CI, Android Kotlin 2.4.0/Gradle 9.3.1, `flutter_foreground_task` 11.0.3 platform requirements, and `flutter pub get` feasibility. If version incompatible, choose a specific compatible maintained release and record the reason before editing dependency constraints.
- [ ] **Step 2: Write failing interface/manifest tests.** In `driver_presence_service_test.dart`, require a fake `DriverPresenceService` with idempotent start/stop counters; read `android/app/src/main/AndroidManifest.xml` and assert `FOREGROUND_SERVICE`, `FOREGROUND_SERVICE_LOCATION`, `POST_NOTIFICATIONS`, non-exported `com.pravera.flutter_foreground_task.service.ForegroundService`, and `android:foregroundServiceType="location"`. The manifest test must fail before the manifest edit.

```dart
final manifest = File('android/app/src/main/AndroidManifest.xml').readAsStringSync();
expect(manifest, contains('android.permission.FOREGROUND_SERVICE_LOCATION'));
expect(manifest, contains('android:foregroundServiceType="location"'));
```

- [ ] **Step 3: Run `flutter test test/driver_presence_service_test.dart` in `mobile` and record the expected missing-interface/manifest failures.** Add the interface and manifest declarations, pin `flutter_foreground_task: 11.0.3`, run `flutter pub get`, and rerun the test. Avoid adding `ACCESS_BACKGROUND_LOCATION` or auto-start receivers.
- [ ] **Step 4: Verify `flutter analyze` and `git diff --check`; commit only this independently reviewable boundary and configuration.**

### Task 2: Authenticated foreground-task publisher

**Files:** Create `mobile/lib/features/driver_workspace/data/driver_presence_publisher.dart`, `mobile/lib/features/driver_workspace/data/android_driver_presence_service.dart`; expand `mobile/test/driver_presence_service_test.dart`; use `mobile/lib/core/session/session_store.dart`, `mobile/lib/features/driver_workspace/data/driver_repository.dart`, `mobile/lib/features/rider_request/data/device_location.dart` without changing their public contracts.

**Interfaces:** `DriverPresencePublisher(DriverRepository repository, DeviceLocation location, SessionStore sessions)` with `Future<bool> renew(String expectedDriverUserId)` (`false` on missing/expired token, changed account, offline profile, or permanent auth rejection; errors on transient failure); task callback constructs `SecureSessionStore(const FlutterSecureStorage())`, `ApiDriverRepository(Dio(BaseOptions(baseUrl: AppConfig.apiBaseUrl,...)), sessions)`, and a geolocator-backed location reader. Service adapter implements Task 1 interface and checks `isRunningFor(driverUserId)` before starting. It passes only the non-secret expected Driver user ID to the task.

- [ ] **Step 1: Write failing publisher tests.** Fake repo/session/location: `renew()` obtains a *new* location and makes exactly one repository `publishLocation(point)` call; expired/missing token yields no publish and signals stop; unauthorized publish signals stop; transient network or revoked location permission never publishes cached coordinates or fake freshness. Name the production change each test would detect.

```dart
final publisher = DriverPresencePublisher(repository: repo, location: location, sessions: sessions);
expect(await publisher.renew('user-1'), isTrue);
expect(repo.calls, ['location']);
sessions.token = null;
expect(await publisher.renew('user-1'), isFalse);
```

- [ ] **Step 2: Run `flutter test test/driver_presence_service_test.dart` and observe the expected missing publisher.** Implement the smallest adapter using the existing `readValidToken()` boundary, existing repository method, and a per-tick fresh geolocator position. A token may be checked first to stop on expiration; the repository must still read it itself for each request.
- [ ] **Step 3: Write failing task cadence/idempotence tests** using an injected repeat-event handler, not an actual 20-second wait: two overlapping repeat events cause at most one publish; stationary consecutive events each obtain fresh position and publish; account switch yields `get().userId != expectedDriverUserId` and stops without a location write; server-offline profile stops; auth rejection stops; network/location failure leaves server lease unrenewed. Run red, implement serial/in-flight guard and repeat cadence comfortably under two minutes (e.g., 20–30 seconds), run green. The profile read checks account identity and online truth, not marketplace offers.
- [ ] **Step 4: Wire plugin callback and manifest-linked Android service** with persistent Driver-online notification, `location` service type, notification permission handling, no boot restart, and plugin initialization in task engine. `start()` must check actual service result and task readiness, not just issue a fire-and-forget platform request; callback startup/registration failures must be observable by caller. Test result and timeout paths without a device where possible.
- [ ] **Step 5: Run focused tests, `flutter analyze`, and Android debug build**; explicitly record that Dart unit tests cannot prove secure-storage/geolocator plugin registration in the Android task engine, which remains a device gate. Commit publisher/service adapter.

### Task 3: Explicit availability and UI lifecycle transitions

**Files:** Modify `mobile/lib/features/driver_workspace/application/driver_controller.dart`, `mobile/test/driver_controller_test.dart`, `mobile/test/test_doubles.dart`, `mobile/test/driver_workspace_test.dart`, and any direct controller construction tests.

**Interfaces:** Inject Task 1 `DriverPresenceService` into `DriverController(DriverRepository, DeviceLocation, DriverPresenceService)`; no service ownership/disposal in controller. `load()` reconciles server `isOnline` and service running status while activity is foregrounded; `setOnline(bool)` keeps its public signature and uses `profile.userId` to bind the service.

- [ ] **Step 1: Change old tests to desired RED assertions.** Previously-online first load performs no `online=false`; background/hidden and `dispose()` do not send `online=false`; resume fetches server state; offline profile stays offline; pending lookup/background transition never claims online without a service. Run `flutter test test/driver_controller_test.dart` and record behavior failures.
- [ ] **Step 2: Implement minimal removal of `_firstLoad` and implicit-offline branches** (`setForeground`, `_run`, `load`, `dispose`); no online heartbeat in `_scheduleHeartbeat`, including when `hasActiveTrip` is true and profile is online. Preserve distinct offline active-trip location publishing. Run focused tests until green.
- [ ] **Step 3: Add RED ordering/failure tests:** explicit online = location → server `online=true` → service start; repeated online/resume = no second start; start failure = attempted server offline and stop partial service, reported error/server reload; rollback failure = error and no durable-online UI claim. Offline = server `online=false` → service stop; offline server failure leaves service running; stop failure after confirmed offline is reported and reconciled. Verify red before implementation, then implement and rerun.

```dart
await controller.setOnline(true);
expect(events, ['location', 'online=true', 'service.start:user-1']);
controller.setForeground(false);
expect(events, isNot(contains('online=false')));
```

- [ ] **Step 4: Add lifecycle/widget RED tests** for paused/hidden/detached and resume in `driver_workspace_test.dart`, overriding the service provider; update old UI test expectations for service start. Implement only required observer/guard changes. Ensure dispose cancels local timers/listeners and never mutates availability.
- [ ] **Step 5: Run controller/widget focused tests, `flutter analyze`, `git diff --check`; commit.**

### Task 4: App/session-scoped ownership and sign-out

**Files:** Modify `mobile/lib/core/providers.dart`, `mobile/lib/features/authentication/application/session_controller.dart`, `mobile/test/session_controller_test.dart` (or existing auth-session test), `mobile/test/driver_presence_service_test.dart`.

**Interfaces:** Non-auto-dispose `driverPresenceServiceProvider` supplies one Android adapter for the app session; inject into Driver controller and session cleanup. Sign-out stops publisher before clearing session. The service is bound to the `driverUserId` passed to `start`, and every task renewal verifies the authenticated `get().userId` and `isOnline` against that ID before publishing. Never persist a second token.

- [ ] **Step 1: Write RED tests** for navigating away and controller disposal without stopping service, logout stopping service even when remote logout fails, session/account replacement not renewing with new account, and expired credentials stopping task on next tick. Use provider overrides and fake service/session store, not Android platform calls; run focused tests and observe failure.
- [ ] **Step 2: Inject session-scoped service, hook sign-out/account change**, and check auth before each background publish. Keep cleanup idempotent, cancel in-flight task safely, do not call offline from `dispose()`. Run focused tests to green.
- [ ] **Step 3: Run session/controller/widget tests and analyze; commit.**

### Task 5: PostgreSQL lease and trip independence evidence

**Files:** Modify `backend/internal/driver/postgres_repository_integration_test.go` and, only if a RED test proves needed, directly related backend production file. Inspect existing `TestPresenceExpiryPreservesFreshAndClearsStale`, marketplace eligibility tests, and trip fixtures before editing.

**Interfaces:** Existing `SetOnline`, `ExpirePresence`, location publishing, marketplace discovery, and trip snapshot contracts remain unchanged.

- [ ] **Step 1: Add independent integration cases:** a renewed location before expiry remains online and marketplace eligible; a stale selected-vehicle Driver becomes marketplace-ineligible before cleanup and offline after cleanup; expiry does not mutate an assigned trip. Use isolated `_test` PostgreSQL only; never dev DB.

```go
if err := repo.ExpirePresence(context.Background()); err != nil { t.Fatal(err) }
if got := readTripStatus(t, db, rideID); got != "assigned" { t.Fatalf("presence changed trip: %s", got) }
```

- [ ] **Step 2: Run named integration tests with `TEST_DATABASE_URL` and `go test -count=1 -p 1 ./internal/driver -run 'TestPresence|TestMarketplace'` from `backend`.** A test proving pre-existing behavior may pass immediately; report that honestly rather than calling it a TDD red. If one fails for the expected missing behavior, make only minimal backend fix and re-run. Do not alter two-minute policy.
- [ ] **Step 3: Run gofmt for changed Go tests, `go vet ./...`, and serial full backend tests; commit test evidence (and narrowly justified fix, if any).**

### Task 6: Documentation and final acceptance

**Files:** Modify `docs/driver-operational-readiness.md`; check `mobile/android/app/src/main/AndroidManifest.xml`, `mobile/pubspec.lock`, `mobile/test/driver_presence_service_test.dart` and CI workflow guards. No unrelated UI, backend lifecycle, or marketplace changes.

**Interfaces:** The spec's explicit online-session/renewable-location-lease contract is public operational guidance.

- [ ] **Step 1: Update documentation**: minimize/lock/navigation behavior; sole service-owned online publisher; user-driven offline; notification and permissions; no kill callback guarantee; two-minute freshness and ten-second cleanup; trip independence; Android-only and iOS unsupported/unverified. Keep marketplace polling foreground-only.
- [ ] **Step 2: Run focused Flutter tests plus full `flutter test`, `flutter analyze`, `dart run build_runner build --delete-conflicting-outputs`, `git diff --exit-code --ignore-space-at-eol` (account for intended changes before comparison), Dart formatting, manifest guard, architecture/encoding grep guards from `.github/workflows/readiness.yml`. Run Android `flutter build apk --debug` and `flutter build apk --profile` with relevant `--dart-define-from-file=config/development.json` only if local Android toolchain permits; record exact block otherwise.
- [ ] **Step 3: Run `gofmt -l` on changed Go files, `go vet ./...`, `TEST_DATABASE_URL=<isolated _test db> go test -count=1 -p 1 ./...`; report database unavailable if tests skip rather than claiming PostgreSQL evidence. Inspect `git diff --check` and branch status. Commit docs/guards if changed.
- [ ] **Step 4: Push branch normally, wait for GitHub Readiness at exact pushed HEAD, inspect both job logs and results.** Do not merge. If network/auth/CI unavailable, report it rather than claiming success. Give physical-device steps for >20–30-second minimize/lock, visible notification, online/server freshness, resume, explicit offline, force-stop/lease expiry, and active trip. Device acceptance is not proven by automated CI; report its status separately.
