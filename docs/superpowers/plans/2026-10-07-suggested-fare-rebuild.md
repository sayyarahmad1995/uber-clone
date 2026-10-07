# Suggested Fare Rebuild Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Deliver backend-calculated editable PKR suggestions, versioned admin-published rates, and atomic immutable request pricing snapshots.

**Architecture:** PostgreSQL stores immutable policy versions and a current pointer. Go combines one road route with exact policy calculation; request creation validates the preview version and copies rates in its existing write transaction. Flutter owns editable proposal state and never reprices an existing request or Trip.

**Tech Stack:** Go, PostgreSQL/database/sql, existing HTTP admin authentication, Flutter/Riverpod/Dio, existing Google routing adapter.

**Spec:** ../specs/2026-10-07-suggested-fare-rebuild-design.md — owner approved 2026-10-07.

## Global Constraints

- This is a fresh implementation. Do not restore, cherry-pick or copy the implementation from closed PR #124.
- Keep the original service dropdown, shared dashboard interactions, Rider and Driver route fixes, Rider-only assignment, Driver eligibility, polling, and existing cash settlement.
- The suggestion is advisory; the Rider may edit it.
- Real Economy/Comfort tariffs require explicit owner approval; fixture numbers are never production tariffs.
- No Google geometry, distance or duration is persisted as a side effect.
- Snapshot the submitted amount as the request's price. Do not replace it with the suggestion if the Rider edited it.
- Publication never scans or updates requests, offers, Trips or settlements.
- No persistent quote record or quote token is introduced.
- No surge, tolls, taxes, waiting charges, discounts, cancellation fees or measured-trip repricing.
- Do not backfill invented snapshots for legacy requests.
- Use one deployment-level suggested-fares enable setting, disabled by default.
- Do not seed real tariffs or auto-enable pricing after migration/publication.

## Review Focus

1. Publication between preview and submission must return pricing_policy_changed, leave zero request/snapshot rows and preserve Rider edits — Tasks 3 and 6.
2. An absent current-policy row and simultaneous first publication must obey the same lock order, without bypass or deadlock — Tasks 2 and 3.
3. A backend restart or rollout transition must not make an old client silently create an unsnapshotted priced request — Tasks 4 and 6.
4. Account change, catalog removal or same-coordinate Place-ID replacement must invalidate price ownership and stale replies — Task 6.
5. Cancellation/recovery and completed-unsettled cash collection after tariff changes must retain the request proposal and selected offer amount — Task 7.

---

## Execution and file map

Branch from plan/suggested-fare-rebuild-2026-10-07 (spec commit fa468af30c80e58f45c2adfdaeb702ac8e1d32dd), preserving the stacked Rider/Driver fixes. Verify the exact head and migration inventory before execution. Migration 031 is unused at this baseline; reassign only if intervening work takes it.

New pricing package: pricing.go (types/validation), calculate.go (exact math), amount.go (admin decimal parsing), postgres_repository.go (publication/current lookup/locking), with corresponding unit/integration tests. New migration 031_suggested_fare_pricing.sql includes version/current/snapshot tables and immutable-history protections.

Modify existing ride request/domain/repository and HTTP contract files, app config/composition and rideservice visibility. Add focused admin_pricing_policy.go and tests; link it from current Operations. Flutter changes stay in existing Rider repositories/model/controller/screen plus a small rider_fare_controller.dart; core/providers.dart wires it. Generated request models change only if their source schema changes.

Use existing _test PostgreSQL helpers; never run destructive integration fixtures on a production database. Focused commands run from backend/ or mobile/ as indicated. If local Go/Flutter cannot execute, use Actions on the working branch and inspect actual logs; any diagnostic workflow is temporary and removed before final exact-head Readiness.

### Task 1: Exact pricing types and calculation

**Files:** Create backend/internal/pricing/{pricing.go,calculate.go,amount.go,pricing_test.go,calculate_test.go,amount_test.go}.

**Interfaces:** Policy contains ID uuid.UUID, ServiceCode/Currency/CalculationRule/PublishedBy string, Version int64, BaseFareMinor/RateMinorPerKm/RateMinorPerMinute/MinimumFareMinor/RoundingIncrementMinor int64, EffectiveFrom/PublishedAt time.Time. Draft contains service/currency and the five money fields. Define Validate(Draft) error; Calculate(Policy, distanceMeters, durationSeconds int64) (int64,error); ParsePKR(string) (int64,error). Errors: ErrInvalidPolicy, ErrInvalidRoute, ErrFareOutOfRange, ErrUnavailable, ErrPolicyChanged, ErrInvalidService. Constant MaxAmountMinor=1_000_000_000_000 and rule half_up_once_v1.

- [ ] Write TestCalculateSingleRounding, TestCalculateMinimum, TestCalculateInvalidAndOverflow, TestValidatePolicy, TestParsePKR. Fixture assertions: base=10000, perKm=1000, perMinute=100, d=2500, t=90, increment=100, minimum=10000 gives 12700; base=0, perKm=1, perMinute=1, d=500, t=30, increment=1, minimum=1 gives 1 (rounding each contribution would incorrectly give 2). Assert minimum=125, increment=100 yields 125 when rounded raw is below minimum. ParsePKR("12.34")==1234; reject negative, exponent, NaN, >2 decimals and overflow.
- [ ] Run go test ./internal/pricing -run 'Test(Calculate|Validate|Parse)' -count=1; confirm red from missing implementation.
- [ ] Implement the interfaces using math/big rational/integer intermediates, final half-up increment rounding and range checking before int64 conversion. PKR only; base/rates >=0, minimum/increment >0, all <=MaxAmountMinor. Reject nonpositive route metrics.
- [ ] Run the same command; expect PASS. Include Economy/Comfort/third-service fixture calculations, exact ties and very large intermediate products.
- [ ] Commit: feat: add exact advisory PKR fare calculation.

### Task 2: Immutable policy storage and serialized publication

**Files:** Create migration 031_suggested_fare_pricing.sql; pricing/postgres_repository.go and postgres_repository_integration_test.go.

**Interfaces:** NewPostgresRepository(*sql.DB) *PostgresRepository; Current(ctx,service,currency string)(Policy,error); Publish(ctx,Draft,reviewer string)(Policy,error); Disable(ctx,service,currency string,expectedPolicyID uuid.UUID,reviewer string) error; List(ctx)([]Policy,error). Export LockCurrent(ctx,*sql.Tx,service,currency string)(Policy,error) for request creation.

Schema: ride_pricing_policy_versions (UUID PK, service/currency/version unique, complete immutable policy fields); ride_pricing_policy_current (service/currency PK, nullable current_policy_id, updated_by/time, FK matching version service/currency); ride_request_pricing_snapshots (ride_request_id PK/FK, policy identity/version/service/currency, copied five money fields, calculation rule, snapshot time). Snapshot columns do not contain routing metrics or a client-supplied suggestion. Use NOT NULL/check constraints and immutable update/delete guards for historical versions and snapshots; no tariff seeds.

- [ ] Write TestPricingPublishHistory, TestPricingDisable, TestPricingConcurrentPublish and TestPricingMissingPointerRace. Assert version increments, unchanged old rows, recorded reviewer/database time, rejected inactive/hidden service, one current pointer and rejected history/snapshot updates.
- [ ] Run go test ./internal/pricing -run TestPricing -count=1 with the dedicated test DB; confirm red.
- [ ] Implement migration and methods. All publish/disable/create locking starts by locking driver_service_catalog: publisher FOR UPDATE, creator FOR SHARE; then current pointer. Separate post-lock SQL statements re-read the latest pointer under READ COMMITTED. The catalog row also serializes a missing pointer/first publication. Publish inserts history and updates the pointer in one transaction; no future scheduling.
- [ ] Run focused tests; assert two concurrent publishers create distinct monotonic versions. Use held database locks/barriers and context deadlines, not sleep-based scheduling; absent pointer cannot bypass a concurrent publication.
- [ ] Commit: feat: store immutable pricing versions and current policies.

### Task 3: Atomic Ride Request price snapshots

**Files:** Modify ride/ride.go, ride/postgres_repository.go; create ride/pricing_snapshot_integration_test.go; extend ride tests.

**Interfaces:** Add CreateInput.PricingPolicyVersion string. Define ride.PricingSnapshot with the copied policy fields above, PolicyID uuid.UUID, Version int64, SnapshotAt time.Time; add Request.PricingSnapshot *PricingSnapshot. Keep NewPostgresRepository(db) for disabled compatibility; add NewPostgresRepositoryWithPricing(db,*pricing.PostgresRepository,required bool). Repository calls pricing.LockCurrent inside its request transaction. The existing domain service retains generic money/location normalization; configured persistence enforces PKR/version requirements at the atomic boundary without calling a route provider.

- [ ] Write TestRequestPricingSnapshot, TestRequestPricingRollback, TestRequestPricingPolicyChanged and TestRequestPricingPublicationRace. Assert a manually edited proposal=110000 stays 110000, copied rates exactly match policy A, zero writes for mismatched/cross-service versions, and snapshot-insert failure rolls back request creation. Legacy requests remain readable.
- [ ] Run go test ./internal/ride -run TestRequestPricing -count=1; confirm red.
- [ ] Implement transaction wrapping existing request insertion/service activation/deadlines. Parse the public version as policy UUID; enabled creation requires its exact current match and PKR. Read server rates only. Compare after locks, insert request then snapshot, commit once. Current policy changes return pricing.ErrPolicyChanged; missing policy ErrUnavailable. Disabled compatibility keeps current manual behavior.
- [ ] Run the focused tests including two race outcomes: publisher first => conflict/no rows; creator first => immutable A snapshot and later B publication. Verify catalog disable/create concurrency and that no request/offer/trip query derives price from the current pointer.
- [ ] Commit: feat: snapshot request pricing atomically without repricing.

### Task 4: Priced previews, rollout and catalog contracts

**Files:** Modify cmd/api/{config.go,config_test.go,application.go}, httpapi/{api.go,ride.go,ride_previews.go,ride_services.go}, rideservice/{catalog.go,postgres_repository.go}; extend corresponding HTTP/catalog integration tests; modify .env.example.

**Interfaces:** SUGGESTED_FARES_ENABLED is true only for an explicit trimmed "true", default false; add config.SuggestedFaresEnabled bool. Dependencies carries Pricing *pricing.PostgresRepository and SuggestedFaresEnabled bool, shared by preview/create/catalog composition. Add rideservice.Option.PricingRequired bool and NewPostgresRepositoryWithPricing(db,required bool), preserving the old constructor. No mobile build flag.

- [ ] Write TestPricedPreview, TestPricingRollout and TestPricedRideRequestHTTP. Assert one routing call and matching amount/version, zero calls for invalid/unpriced services, old flag-off response/manual requests, default-off config, and strict flag-on request rejection when version is missing. Assert Driver route endpoint is still route-only. Catalog synthetic service appears only when priced and becomes hidden after disable in enabled mode.
- [ ] Run go test ./cmd/api ./internal/httpapi ./internal/rideservice -count=1; confirm relevant red cases.
- [ ] Wire pricing with the single config flag. Enabled catalog filters current PKR policies and emits pricing_required=true per entry; disabled emits false and preserves current catalog. Preview returns suggested_fare plus pricing_policy_version (policy UUID string), from one route. Create input passes version to Task 3.
- [ ] Implement error transport: {"error":"pricing_policy_changed","message":"Rates changed. Refresh the suggestion and submit again."} at 409; pricing_policy_unavailable at 503; invalid service/version/currency at 400. Preserve existing auth, route validation and provider status mapping.
- [ ] Run focused commands; expect PASS including service filtering, version provenance and restart/flag compatibility. Confirm no client-supplied rates or suggestion amount enters authoritative storage.
- [ ] Commit: feat: expose gated fare previews and version-checked booking.

### Task 5: Admin tariff publication

**Files:** Create httpapi/admin_pricing_policy.go and admin_pricing_policy_test.go; modify routes.go and link from admin_marketplace_policy.go.

**Interfaces:** GET /admin/operations/pricing; POST /admin/operations/pricing/publish; POST /admin/operations/pricing/disable. JSON admin endpoints GET/POST /v1/admin/pricing-policies and POST /v1/admin/pricing-policies/{policy_id}/disable reuse the existing admin middleware. Publication uses Task 2 Publish and Task 1 ParsePKR; disable resolves identity to service/currency and checks the target remains current before disabling, returning 409 for a stale page.

- [ ] Write TestAdminPricingAuthorization, TestAdminPricingPublish and TestAdminPricingDisable. Assert unauthorized/invalid origin writes nothing, exact rupee parsing, publication displays/records complete tariff and reviewer, historical versions survive, hidden/inactive services fail, stale disable cannot clear a newly published version, and publication does not enable the rollout flag.
- [ ] Run go test ./internal/httpapi -run TestAdminPricing -count=1; confirm red.
- [ ] Implement focused templates/handlers with existing Basic Auth/origin protections. Form labels show PKR rupees and all five fields, publication explicitly approves displayed values and affects new previews only. JSON rejects unknown fields; safe validation messages and no provider/secrets in errors.
- [ ] Run focused tests; expect PASS. Use Task 2 Disable with expectedPolicyID consistently in both HTML and JSON transports.
- [ ] Commit: feat: publish approved pricing versions through operations.

### Task 6: Flutter proposal ownership and conflict recovery

**Files:** Modify rider_request/{domain/route_preview.dart,domain/ride_service.dart,data/ride_request_repository.dart,application/rider_request_controller.dart,presentation/rider_request_screen.dart}, core/providers.dart; create application/rider_fare_controller.dart and mobile/test/rider_fare_controller_test.dart; extend rider_route_preview_test.dart, ride_request_repository_test.dart, rider_request_submission_test.dart, rider_service_catalog_test.dart, test_doubles.dart.

**Interfaces:** RoutePreview adds nullable Money suggestedFare and String pricingPolicyVersion; validate both together, positive PKR amount and nonempty version. RideService adds bool pricingRequired default false. Repository.create and RequestController.submit add String? pricingPolicyVersion; submit retains ApiException code in state for UI recovery. RiderFareController exposes selectionKey, proposal text, userEdited, version and applyPreview(selectionKey,preview), edit(text), invalidate(selectionKey); same-input refresh preserves edits, new-input invalidation clears ownership. Build selection keys from coordinates, Place IDs and service, and scope providers to account.

- [ ] Write test cases asserting prefill, edited proposal preservation on Retry, changed-input clearing, late result suppression, version serialization and pricing-conflict resubmission. Widget tests assert original dropdown, submitted amount=edited value, no automatic second POST after 409, account/catalog/Place-ID invalidation, preserved booked line and route-only Driver parsing.
- [ ] Run flutter test test/rider_fare_controller_test.dart test/rider_request_submission_test.dart test/rider_route_preview_test.dart; confirm red for missing behavior.
- [ ] Implement model/repository/controller changes and account-scoped fare provider. Enabled submission requires a current priced preview matching inputs and catalog selection; flag-off manual submission retains current behavior. Programmatic prefill must not mark the field userEdited.
- [ ] Connect the existing fare field and preview card to the fare controller. Display suggestion separately from edited proposal. On pricing_policy_changed, refresh preview and explain the change, keep edits and require another tap. On active-request recovery, render stored proposal and retain existing booked polyline; price previews cannot overwrite active state.
- [ ] Run the focused tests plus repository/service catalog tests; expect PASS. Update all fake interface implementations and generated models only where their source schema changes.
- [ ] Commit: feat: show editable suggestions and recover tariff conflicts.

### Task 7: Lifecycle regression, rollout evidence and final review

**Files:** Create backend/internal/marketplace/pricing_lifecycle_integration_test.go; extend relevant Flutter booking/recovery tests; update docs/{ADR-0013-google-maps-booking-preview-and-fare.md,pilot-google-booking-acceptance.md,pilot-deployment.md} and .env.example.

- [ ] Write TestPricingChangesDoNotRepriceRide: create request under A with proposal=110000, publish B, assert offer bounds 99000..143000, select offer=120000, publish C, complete/cash-collect/recover and assert request=110000, agreed/settled=120000, snapshot=A. Repeat with legacy request, cancellation, disabled service and completed-unsettled recovery.
- [ ] Run go test ./internal/marketplace -run TestPricingChangesDoNotRepriceRide -count=1. The behavior may already pass because prior lifecycle uses immutable amounts; assertions must now also prove the new snapshot survives.
- [ ] Add Android acceptance steps for publication during open offers, conflict before submission, edited proposals, dropdown/gestures, both Driver routes and cash receipt. Record physical validation as pending until actually performed. Document staged deployment with flag off, approved tariffs and compatible client, then explicit enable.
- [ ] Run full Readiness on the exact final head: gofmt/vet/go test -p 1 ./... with PostgreSQL; dart format/build_runner/clean generated diff/flutter analyze/flutter test. Remove any temporary diagnostic workflow before that run. Inspect output and resolve every failure.
- [ ] Perform a whole-branch independent review using requesting-code-review; verify race tests, lock ordering, no current-policy lookups in lifecycle, admin stale writes and fare ownership. Fix findings and rerun affected/full required checks when changes warrant it.
- [ ] Commit docs and open a draft PR stacked on the Driver route branch (or its confirmed integrated successor); include exact-head CI, immutable snapshot/race evidence and physical-device status. Do not merge or activate tariffs. Keep the branch for review.

## Self-review and execution handoff

Coverage: calculation/policy publication in Tasks 1–2; atomic request snapshots/races in Task 3; contracts/rollout/catalog in Task 4; admin approval in Task 5; editable UI/stale reply/conflicts in Task 6; lifecycle, docs and release evidence in Task 7. Review Focus cases are pinned to their listed tasks. No persisted Google metrics, guessed tariffs, mutable history or automatic repricing.

Disable uses expectedPolicyID throughout. Version transport is UUID identity; the numeric Version field is monotonic admin history, not the public request matching key.

Recommended execution: Native, task-by-task in this session, followed by one whole-branch independent review. These tasks share schema/transport interfaces closely; native execution avoids repeated context while retaining PostgreSQL race and lifecycle verification. Implementation begins after owner plan review and execution-method selection.
