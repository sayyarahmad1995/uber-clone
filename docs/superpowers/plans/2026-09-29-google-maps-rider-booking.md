# HiGO Dynamic Ride Services and Google Maps Rider Booking Experience — implementation plan

**Status:** Implementation merged into `main`; complete milestone acceptance and rollout evidence remain open. This is the current development plan, not a pilot-release approval.
**Prepared:** 2026-09-29
**Revised:** 2026-10-08 — reconciled with merged PRs #119–#123, #125, #127–#130 and the owner's read-only suggested-fare revision.
**Baseline reviewed:** main at 2285d8febf455e31db4049f0d2e1464bdcd410af (2026-09-28)
**Owner:** HiGO ride-hailing MVP
**Execution model:** Small, independently reviewable pull requests; test-driven changes; physical Android acceptance before pilot release.

## Current progress and evidence — 2026-10-08

The ten implementation/documentation PRs were merged into `main` in dependency order. The final integration commit is `afa95bd98f0e42d46c7644aa206e18d6ffc37d52`. [Final main Readiness](https://github.com/sayyarahmad1995/uber-clone/actions/runs/37744424250) passed backend formatting/vet/full PostgreSQL tests and Flutter formatting/generated drift/analyze/all 233 tests. These are automated results, not a substitute for case-by-case device evidence.

| Plan slice | Merged implementation | Current status |
| --- | --- | --- |
| PR 0 — architecture and plan | #119 | Documentation merged; this update reconciles its original future-tense status. |
| PR 1 — dynamic Rider services | #120; submission lifetime fix #125 | Implemented. Owner reported third-service discovery on the same installed app, Driver enrollment enforcement, selection and disablement on 2026-09-30. Remaining catalog presentation/error cases stay in the acceptance matrix. |
| PR 2 — shared Google Maps | #121 | Implemented; Android debug/release smoke and owner-reported map/interaction checks, including online Driver restart retest, passed. Later Places/route interaction changes still require their own regression checks. |
| PR 3 — Places and map selection | #122, refined in #123 | Implemented. Search and the active-request scroll regression have owner-reported device evidence; the full search/pin/permission/error matrix is not recorded as passed. Free pins remain exact; explicit Google POIs use canonical Place Details coordinates. |
| PR 4 — recommended road route | #123; booked Rider line fix #127 | Implemented and automated checks passed. #127 retains the booking preview within the current session; it does not reconstruct a Rider booked route after a fresh restart. Complete physical route/viewport/failure evidence remains open. |
| PR 5 — suggested fare | #129, read-only revision and race follow-up #130 | Implemented. Owner reported the feature device test passed on 2026-10-08 and confirmed disabling pricing removes the service from selection. Full pricing-change-to-cash acceptance and production tariff/rollout evidence are not recorded here. |
| Driver stage routes — approved extension | #128 | Implemented: fresh Driver location to pickup after assignment, pickup to destination after Start trip, line cleared at completion/cancellation. Automated stage/retry/stale-response tests passed; dedicated physical stage/recovery evidence remains open. |
| PR 6 — integrated booking and verification | Integrated code across the above PRs | Booking integration is implemented. Remaining work is a focused gap audit and complete end-to-end/device/release evidence; do not rebuild the completed slices merely because the original checkboxes were unchecked. |

**Approved fare behavior:** #130 supersedes the original editable suggestion. With suggested fares enabled, the Rider cannot edit the calculated amount. The server re-routes and re-calculates before creating a request; a fare/policy change refreshes the suggestion and requires another tap. Drivers can accept or counteroffer; the Rider-selected Driver offer is the immutable final payable fare. Rollout-off manual booking remains available for compatibility.

**Coverage follow-up resolved:** `TestRequestPricingFirstPublicationRace` covers creator-first and publisher-first orderings with no pre-existing policy/current pointer. Both cases passed ten repeated PostgreSQL runs and failed when the catalog lock was removed in the CI workspace. The production lock was unchanged, the mutation was not committed, and temporary diagnostics were removed.

**Evidence limits:** The 2026-10-08 feature pass did not include device/build identifiers or per-case outcomes. Do not mark the entire physical acceptance matrix, cash-loop gate, provider latency/load evaluation or live rollout as passed from that report. Preserve the merged plan branches as development references, per the owner.

## 1. Goal and release boundary

First make ride services server-configurable and discoverable by the existing Flutter Rider app without subsequent app-store releases. Then deliver a complete Rider booking preview using the owner's existing working Google Maps Platform setup. A signed-in Rider can discover currently bookable services, find pickup and destination, adjust them on a Google map, see a valid **driving** route and polyline, view road distance and estimated duration, see a backend-calculated **suggested** fare for the selected service, view that suggestion read-only, and submit the existing Ride Request at the server-validated amount.

This is a core booking-experience milestone, not a dispatch rewrite, new booking mode, full navigation product, or general payments platform. A single shared Flutter application and a Go modular monolith remain authoritative.

### Existing baseline

- Shared Flutter Android app, Riverpod, Dio, current map-first Rider/Driver dashboard, flutter_map/OpenStreetMap, geolocator.
- Go/PostgreSQL backend; Rider requests, geographic Haversine discovery, Driver exact-fare offers/counteroffers, Rider-only selection, Trip lifecycle, cash settlement, history. The existing driver_service_catalog and authenticated GET /v1/driver/services serve Driver onboarding, but no Rider-facing catalog endpoint exists. Go ride creation currently hardcodes only economy/comfort, and Flutter's RiderServicePicker hardcodes the same two options.
- Existing background Android Driver location publishing and server freshness policy; existing authenticated app-owned API surface.
- Live Google Cloud Console credentials/APIs are already configured and working, per the project owner. **Do not repeat Cloud setup.** Only verify project/API separation, key restrictions, app signing configuration, and quotas as part of release hygiene.
- Current WORKLOG still describes PR #83 cash settlement as future work even though settlement API and UI are implemented; reconcile it in the documentation PR.

### Success criteria

1. After one client update to introduce this feature, the Rider app fetches a backend-managed catalog and renders any currently bookable **compatible** ride service with a generic card/icon fallback; adding, disabling or reordering such services later requires no further app-store update. Refresh on Rider dashboard entry and app foreground/recovery, not a new real-time transport.
2. The backend validates service selection from current catalog state rather than hardcoded economy/comfort names, and Driver eligibility still requires explicit approved vehicle/service enrollment. A disabled service cannot receive new Ride Requests, but historical snapshots remain readable.
3. Google map renders for both existing Rider and Driver dashboards without changing their interaction contract.
4. Pickup/destination can be selected by address search or map adjustment, with meaningful labels.
5. Backend returns a valid road route, distance in meters, duration in seconds, and encoded polyline; Flutter draws it and fits the camera.
6. The Go backend returns a suggested PKR fare from an approved, versioned pricing policy keyed by catalog service, never from client-side math. Economy and Comfort launch parameters require approval; a synthetic third service proves extensibility in tests.
7. In priced booking, the Rider sees a read-only suggestion and creates a Ride Request at the server-validated amount. Drivers may accept or counteroffer within existing bounds; Rider selection alone assigns the Trip and fixes the final payable fare.
8. No route/no results/location denial/provider error lead to explicit, recoverable states; no Haversine distance is mislabeled as driving distance or ETA.
9. The full route-preview-to-cash-receipt path passes backend, Flutter and physical-device acceptance, including recovery.

## 2. Prerequisite — Gate 0: existing cash-ride acceptance

Before releasing this milestone to the pilot, complete docs/pilot-cash-ride-acceptance.md on current main and record build/commit, device, accounts, Economy/Comfort, selected fares, linked ride/trip IDs, pass/fail and recovery evidence. Verify exact-fare and counteroffer paths, Rider selection, both cancellation paths, completion, completed/unsettled recovery, cash collection, receipts and database invariants. Any blocking failure is repaired before expanding the pilot.

Gate 0 may be executed while the new documentation PR is prepared, but is a release gate. The repository checklist is evidence of planned tests, **not** evidence that every test has passed. Do not claim pilot readiness based on CI alone.

## 3. Decisions and invariants

### Preserve accepted architecture

- Go modular monolith, PostgreSQL transactional ownership, application-owned transport and provider interfaces; compose concrete adapters in backend/cmd/api.
- One account and one shared Flutter app. Rider is default. Driver eligibility derives from the operating vehicle's active approved service enrollments. Catalog codes are stable business identifiers; presentation metadata and service ordering come from the server, not Flutter enums.
- One Rider request with pickup, destination, service, proposed fare and currency. Drivers may accept that fare or counteroffer. **Only Rider offer selection creates a Trip.** Selection and offer revision checks stay atomic.
- Existing offer bounds (currently 90–130% of the **saved request amount**) remain fixed for that request. In priced booking this amount is the validated suggestion; later tariff changes do not rebase the bounds.
- Selected offer fare and Driver/vehicle/service context remain immutable Trip snapshots; completion and cash settlement do not recalculate the fare.
- Haversine distance continues to rank/filter inexpensive marketplace discovery; routing is for the Rider's chosen trip preview, not a Google request for every Driver/request pair.
- ADR-0008 RideDashboardScaffold behavior is unchanged: Rider 18%, Driver 16% collapsed, maximum 60%, shared committed extent and the existing gesture/scroll contract.
- Keep HTTP polling for the current ride-flow pilot, Android first, and the existing server-authoritative Driver location lease.

### Dynamic service-catalog architecture

- Reuse the existing PostgreSQL driver_service_catalog as the source of truth; evolve it minimally with Rider booking visibility/display metadata only if needed. Do not create a competing Rider catalog or a second service identity. Driver-oriented eligibility fields must not leak into Rider cards without a product need.
- Add a narrow authenticated GET /v1/ride-services API returning *currently bookable* services with stable code, display_name, description, display_order and a safe presentation token with an app-provided generic fallback. Design activation as an explicit server-side state; do not assume a Driver's enrollment alone makes a service bookable in every product state.
- Replace the hardcoded RiderServicePicker with API-backed rendering and a deliberate default/selection policy. Reload on initial Rider dashboard entry and app foreground; invalidate the selected service and any stale route/fare preview if it becomes unavailable. No push/WebSocket is needed to propagate catalog changes on the next refresh.
- Replace the hardcoded economy/comfort check in backend/internal/ride/ride.go with catalog-backed validation (the existing persistence active-service check remains a defensive invariant). Keep a service_code string in existing Ride Requests/offers/Trips rather than introducing service-specific code paths.
- Pricing policy must be keyed by catalog service and currency with explicit activation/versioning. A service that requires a suggested fare must not be displayed as ready for the new booking flow until a valid pricing policy is available; no silent zero-price or default-to-Economy behavior. Changes to approved tariff values affect new previews only, not already agreed fares.
- Launching an additional **compatible car category** after this infrastructure is deployed should require catalog/pricing configuration plus approved Driver vehicle/service enrollment, not a new Rider app build. A category with distinct scheduling, delivery, multi-stop, capacity or other new interaction needs may still require a versioned client feature and app release.

### Provider architecture

- Map display: Google Maps SDK for Android via an evaluated/pinned compatible google_maps_flutter release. Confine plugin types to mobile/lib/core/maps, expose app-owned map markers, coordinates, polylines and camera operations to feature screens. Replace flutter_map/MapTiles usage on both Rider and Driver screens as one migration; do not leave Google Places/Routes content displayed over OSM.
- Location: preserve the existing DeviceLocation/geolocator abstraction. SDK location features need not replace it.
- Place search/details and optional reverse geocoding: Go-backed, authenticated application APIs with a Google adapter; Flutter never holds or calls the unrestricted server API key. The Android map key is a separate Android-app-restricted key.
- Route preview: Go-owned routing port and Google Routes API adapter; no Google-specific SDK models in ride, offer, marketplace, Trip or public domain contracts.
- Fare suggestion: a small application-owned Go pricing component consuming validated route distance/duration and selected ride service. Its parameters and version are not hard-coded in Flutter.

### New documentation decisions

PR #119 recorded **ADR-0012: Dynamic Rider Ride Service Catalog** and **ADR-0013: Google Maps booking preview and advisory fare policy** before implementation. ADR-0012 extends ADR-0009's existing vehicle/service enrollments into a Rider-facing catalog without a second booking mode. ADR-0013 explicitly moves **selected-trip route preview** out of ADR-0007's earlier deferral while preserving Haversine discovery and Rider-selected assignment. Update docs/technology-stack.md, docs/flutter-client-architecture.md, docs/mvp-scope.md, docs/architecture-decisions.md and docs/WORKLOG.md accordingly. ADR-0011 remains the historical scope of **cash settlement**; do not rewrite it as if it authorized pricing.

## 4. Delivery sequence

### PR 0 — Document the next slice

- [x] Record ADR-0012 (dynamic catalog) and ADR-0013 (Google Maps/route preview/advisory fares), including their architectural ownership, MVP limit, route/fare data semantics, server-driven UI rules and clear deferrals.
- [x] Update docs listed above; mark PR #83/#86 cash settlement complete, distinguish physical pilot validation from completed code.
- [x] Note the current vehicle-only Driver operating context, not the outdated single-global-service wording in the WORKLOG.
- [x] Add a focused acceptance matrix for dynamic catalog refresh, future compatible services and the new booking preview; do not replace the existing cash checklist.
- **Done when:** Documentation no longer names cash settlement as the next feature, accepted marketplace/dashboard/settlement rules remain consistent, and the PR is documentation-only.

### PR 1 — Dynamic Rider service catalog (backend + Flutter)

- [x] Reuse driver_service_catalog and existing GET /v1/driver/services. Add only the Rider-facing visibility/order/presentation metadata required to serve a stable catalog; provide a data migration with Economy and Comfort intact.
- [x] Add authenticated GET /v1/ride-services returning only currently bookable compatible services (code, display_name, description, display_order, optional presentation token). Avoid a general-purpose configuration/admin service and do not expose Driver eligibility internals.
- [x] Replace the Go Ride Request service's economy/comfort-only validation with repository/catalog-backed checks. Keep the PostgreSQL active-service guard; reject unknown or disabled services, and preserve Rider-only offer selection, vehicle/service eligibility and immutable service snapshots.
- [x] Implement an API-backed Flutter RiderServicePicker. Re-fetch when entering the Rider dashboard and on app foreground/recovery; handle loading, empty/error, disabled selected service and reorder changes. Render known icons when available and a generic default for unknown service codes. Never require a compile-time enum or asset for a newly configured compatible service.
- [x] Bind service changes to route/fare preview invalidation; once pricing is introduced, list as bookable in the new booking flow only those services with an active approved compatible fare policy. Do not silently substitute Economy parameters.
- [x] Define an owner-controlled rollout procedure: create service metadata and pricing configuration, set visible/bookable only after backend and Driver enrollment readiness, and verify a catalog refresh on an already-installed compatible app. An operator UI is **not** required for this MVP.
- **Tests:** Go PostgreSQL and HTTP cases for active/disabled/unknown services, deterministic display order and concurrent deactivation; Flutter contract/controller/widget tests with a synthetic third service, unknown icon fallback, catalog refresh and removal of an obsolete selection; regression for Driver service eligibility and historical Trip context.
- **Done when:** A third synthetic service can appear, be selected and create a valid Ride Request (with eligible approved Driver enrollment), then disappear from new booking after disablement, **without changing or rebuilding the already-updated Flutter client**. Economy/Comfort and existing lifecycle remain unaffected.

### PR 2 — Shared Google Maps rendering foundation

- [x] Verify compatible, pinned google_maps_flutter/Android Maps SDK versions and Android build requirements before updating dependencies.
- [x] Implement the Google map adapter in mobile/lib/core/maps; adapt RideMap's public inputs, markers, map-tap callback and camera commands without passing provider classes into Rider/Driver features.
- [x] Migrate BOTH current map surfaces, current-location focus, marker presentation and cached-center handling; remove obsolete MapTiles/OpenStreetMap references once all callers are migrated.
- [x] Preserve dashboard panel gesture/extent regression behavior and marker access restrictions in code/automated regressions; physical Android interaction acceptance remains outstanding.
- [x] Provide environment-specific Android API key injection without committing any key. The SDK key must be Android application-restricted (package plus signing fingerprint) and Maps-SDK-restricted; keep the backend key server-side.
- [x] Explicit user-initiated focus/fit should control the camera. Do not automatically recenter on each Driver location update after the user pans the map.
- **Tests:** shared-map widget/adapter tests; Rider and Driver dashboard gesture/persistence regression; Flutter analyze/test; Android debug/release build smoke; physical-device map/permission/rotation/background checks.

**PR 2 evidence:** Readiness passed with the pinned dependency and provider-boundary regression test. A temporary CI smoke run built both Android debug and release APKs successfully with the non-secret fallback key, then the workflow was restored. The owner reported the shared map/interaction physical checks and online-Driver restart retest passed; later Places/route changes remain subject to their own acceptance cases.
- **Done when:** Existing pickup/destination and Driver markers render correctly on Google maps and the prior booking behavior still works. **Done on physical Android 2026-10-01 after retesting the online-Driver restart auto-center fix.**

### PR 3 — Pickup/destination search and adjustment

- [x] Add authenticated, app-owned endpoints for Places Autocomplete (New), selected Place Details (New), and reverse geocoding for readable map-pin labels. The confirmed Rider coordinate remains authoritative and is never moved by reverse geocoding. Choose narrow field masks; no generic Google API proxy.
- [x] Generate a fresh Places session token for each autocomplete session (pickup and destination independently), pass it through to the concluding Place Details request, and use credentials from the same Google Cloud project.
- [x] Use the Rider device location as a runtime local search center. Apply the admin-configured Google Autocomplete `locationRestriction` (5–50 km, default 50 km) when the center is available so distant strong text matches do not outrank nearby POIs/businesses; this remains a search window rather than a persisted service-area boundary. Re-rank Google's returned local predictions by Google `userRatingCount` as the popularity proxy, cache that signal for 24 hours, and preserve Google order when popularity is equal or unavailable. Do not use popularity to inject places that were not text-matched by Autocomplete. Add a short debounce, minimum input length, stale-response cancellation and per-user rate limiting to control latency and billable calls.
- [x] Build Flutter pickup/destination inputs, suggestion selection, explicit center-pin map selection plus draggable-pin adjustment, direct selection of Google-rendered POIs by tap, and reverse-geocoded display. Free-pin selection retains the exact Rider coordinate; an explicitly tapped POI uses that POI's canonical Place Details coordinate.
- [x] Handle empty predictions, inaccurate GPS, location permission denied, invalid place, lost network, request cancellation, and unsupported areas without inventing an address.
- **Tests:** fake Places adapter plus Go HTTP contract tests; Flutter search-session lifecycle and stale-result tests; physical Android search/select/drag flow.
**Automated implementation evidence:** Go provider/service/HTTP tests cover narrow provider contracts, exact-coordinate reverse geocoding, formatted-address resolution and application error mapping. Flutter repository/controller tests cover authenticated app-owned endpoints, independent pickup/destination sessions, fresh post-selection sessions, stale-result rejection and exact-coordinate retention. Physical Android search/select/drag/error-state validation remains required.

- **Done when:** Both endpoints can be searched or map-adjusted and the user can correct an imprecise pin; the current Ride Request contract is not yet changed.

### PR 4 — Authoritative road route preview and polyline

- [x] Add backend/internal/routing (or equivalent domain-owned package) with a provider-neutral Preview route operation, a Google Routes adapter and injected HTTP client/config.
- [x] Request one `DRIVE` route with `TRAFFIC_AWARE_OPTIMAL`, `BEST_GUESS`, and `computeAlternativeRoutes=false`. Do not request `SHORTER_DISTANCE` or other reference routes. For typed Place/explicit POI selections, preserve the selected Place ID through Flutter state and `POST /v1/ride-previews`, then use a Google Routes `placeId` waypoint; for free pins/current location, omit `place_id` and use the exact lat/lng waypoint. Use the narrow field mask for distanceMeters, duration and encodedPolyline only. Treat Google's returned default route as the booking recommendation. Define timeout, cancellation, bounded retry for transient errors, and deterministic error mapping. Do not request detailed navigation instructions.
- [x] Expose a narrow authenticated POST /v1/ride-previews endpoint accepting pickup/destination required coordinates, optional per-endpoint `place_id` for explicit Google Place selections, and service_code. Coordinates remain present for app state/validation; the Google adapter prefers `placeId` only when explicitly supplied. Initially return route geometry/distance/duration; PR 5 adds an optional suggested_fare object and policy version.
- [x] Normalize the public response to one `route` with `distance_meters`, `duration_seconds` and `encoded_polyline`. Reject zero/invalid/unroutable or malformed responses. Keep Google-specific wire types inside the adapter.
- [x] Add a Flutter route preview controller; decode/draw the single recommended route polyline on Google Maps and fit it once when the route changes. Invalidate outdated previews whenever pickup/destination changes; discard late replies and allow explicit retry. Panel expand/collapse must not move or re-fit the map.
- [x] Clearly label all numbers as *route estimates*. Keep Haversine pickup distance in marketplace discovery unchanged. Do not confuse trip distance with Driver-to-pickup distance.
- **Tests:** fake Routes HTTP fixtures, field mask/timeout/error/coordinate validation tests; encoded-polyline/viewport widget tests; contract checks; physical route on Android.
- **Done when:** A Rider selecting two routable locations sees the road polyline, road distance and estimated duration; an unroutable pair shows a real error, not a straight line presented as a road route.

### PR 5 — Versioned suggested fare, without changing negotiation

- [x] Add an app-owned pricing policy keyed by catalog service and currency, initially Economy/Comfort and PKR; a synthetic third service is required for extensibility tests. The implemented immutable policy contains minimum fare, base fare, minor-units-per-km, minor-units-per-minute, rounding increment, version and immediate publication metadata. Exact calculation uses one final half-up increment rounding step, then enforces the minimum. **Production values require explicit business approval**; test-fixture numbers are not tariffs.
- [ ] Record the owner's approval of actual launch tariffs and the enabled deployment's service/policy identities. Publishing test values or a passing calculation test does not establish production approval.
- [x] Compute suggested fare in Go from the valid route distance and duration using integer minor units and explicitly defined rounding/overflow guards. Recommended starting formula: max(minimum, round(base + perKm * distanceMeters/1000 + perMinute * durationSeconds/60)). No surge, tolls, taxes, discounts or cancellation charges in this slice.
- [x] Extend POST /v1/ride-previews with suggested_fare (amount_minor, currency) and pricing_policy_version for the same recommended route. Calculate the fare from that route's distance/duration without duplicating Google calls.
- [x] The result is the read-only Rider request amount in priced booking. The backend resolves selected Place IDs, checks canonical coordinates, obtains a fresh route and calculates the fare again before creation. A mismatch returns `suggested_fare_changed` (409), with no request/snapshot; the app refreshes and requires another tap. The atomic policy UUID check remains. Driver counteroffers and Rider selection fix the final Trip/settlement amount.
- [x] Do not add a persistent Google polyline/duration history as a side effect. Retain user-owned pickup/destination and the already immutable selected-offer fare. Any proposal to persist additional Google-derived content must pass a fresh terms/caching review.
- [x] If a route or pricing policy is unavailable, do not fabricate a price. For the initial Google-preview-enabled path, block price-backed request submission with a clear retry state; manual-only fallback is a separate explicit product decision.
- **Tests:** zero/near-zero, long, invalid, overflow, rounding, policy version, Economy vs Comfort vs synthetic third service, disabled/unpriced service rejection, route failure and read-only price display, tampered amount/endpoint rejection, explicit conflict resubmission and first-publication race tests.
- **Done when:** Backend reproduces the configured fare for the preview inputs, the Rider cannot edit it in priced booking, and the selected Driver offer remains the final payable fare. Production tariff approval is a separate rollout gate.

### PR 6 — Complete Rider booking UX and end-to-end verification

- [x] Connect dynamic catalog loading/refresh, location search, map adjustment, chosen service, combined route/fare preview, read-only suggested fare and the existing POST /v1/ride-requests action. #125 keeps catalog state alive during submission; #127 retains the current-session booked route.
- [ ] Keep ADR-0008 panel behavior intact. Ensure the polyline/markers are visible with an expanded bottom sheet, keyboard, small device and text scaling; show appropriate loading, empty, permission and retry states.
- [x] Do not show a stale route/price after any pickup, destination or service change or catalog deactivation. Prevent accidental duplicate submission and recover the backend-authoritative active ride after restart or request timeout.
- [x] Preserve Rider offer comparison, rejection, chosen offer/assignment, Driver operations, cash confirmation, and both histories. The agreed fare shown in history must remain the selected Driver offer rather than the earlier suggestion.
- [x] Extend docs/pilot-cash-ride-acceptance.md (or a companion feature checklist) with Google routing/search/fare/UX physical-device cases, while retaining all existing lifecycle and settlement checks.
- **Tests:** Flutter booking-controller, widget/interaction and full regression; Go provider/HTTP/pricing and PostgreSQL regression; synthetic third-service configuration/refresh/deactivation on unchanged client; physical Android complete Economy and Comfort ride paths including no-route and network recovery; exact-HEAD readiness CI.
- **Done when:** Record complete Economy and Comfort real-device route-preview-to-cash-receipt cases, including read-only suggestion, exact-fare/counteroffer selection and recoverable failures, with build/device and case outcomes. Automated integration is complete; this final evidence gate is not yet recorded as passed.

### Remaining PR 6 work — acceptance audit, not a new feature build

- [ ] Reconcile each open item in [the booking matrix](../../pilot-google-booking-acceptance.md) and [cash-loop checklist](../../pilot-cash-ride-acceptance.md) with concrete existing evidence. Request only missing cases; do not repeat the already reported suggested-fare and pricing-disable checks without a reason.
- [ ] Verify the integrated small-screen/keyboard/text-scaling/panel/map flow, catalog invalidation, duplicate-submit protection and timeout/restart recovery on the merged build.
- [ ] Record Economy and Comfort exact-fare/counteroffer journeys through unsettled recovery, cash collection and receipts; verify final fare equals the selected offer across tariff changes and disablement.
- [ ] Record Driver stage-route/restoration/retry/late-response and camera-gesture checks. Assess the known current-session-only Rider booked-route limitation against required restart UX before proposing a fix.
- [ ] Record actual launch tariff approvals, restricted key/permission checks and staged deployment/flag state before pilot release. Merging implementation is not activation evidence.

Any demonstrated failure becomes a bounded fix with a regression test. Choose the next implementation only after this audit identifies an actual gap.

## 5. Proposed application contracts (finalize in PR 0)

- GET /v1/ride-services: authenticated Rider-visible catalog with services [{code, display_name, description, display_order, presentation_token?}]; backend returns only currently bookable compatible entries in stable display order. Do not mix Driver onboarding requirements into this response. Re-fetch on Rider screen entry and app foreground/recovery; services appear after the next successful refresh, not an instantaneous pushed UI update.
- POST /v1/places/autocomplete: input, session_token, optional user-location center; Google Autocomplete is restricted to the current admin-configured 5–50 km circle around that center when present, returning place IDs and display text rather than a raw provider object.
- GET /v1/places/{place_id}: session_token; returns normalized place ID, label and coordinates. The matching autocomplete token concludes here.
- GET /v1/places/{place_id}: returns normalized Place Details. Autocomplete selections include their session_token; direct taps on a Google-rendered POI resolve the tapped Place ID without fabricating an autocomplete session.\n- POST /v1/places/reverse-geocode: confirmed free-pin coordinates; returns a readable label while preserving the exact submitted coordinates. Reverse geocoding is descriptive only and must never snap or reposition the Rider's pin.
- POST /v1/ride-previews: pickup and destination require latitude/longitude and may include `place_id` only for an explicitly selected Google Place; reverse-geocoded free pins omit it. Include dynamic catalog service_code; return route {distance_meters, duration_seconds, encoded_polyline}, suggested_fare {amount_minor, currency}, pricing_policy_version, and request correlation metadata. During PR 4, suggested_fare may be null; it is required only after an active pricing policy is configured for the released PR 5+ experience.
- Existing POST /v1/ride-requests remains the booking command (dynamic catalog service_code, pickup, destination, proposed_fare). Validate service activation, selected Place-ID/coordinate consistency, current policy UUID and fresh server-calculated fare on each priced creation. Do not introduce a second booking mode or require a durable quote token for this read-only suggestion flow. Fresh server-side validation and atomic policy snapshots are implemented instead.

All new endpoints are authenticated; validate payload sizes and coordinates, bound external provider calls, and avoid placing raw secrets or precise location trails in operational logs. Public response structs must not leak Google billing/error internals.

## 6. Verification and release gates

Before merging each implementation PR:
- Go: gofmt on changed files, go vet ./..., go test -p 1 ./... with the dedicated _test PostgreSQL database, including focused fake-provider tests.
- Flutter: dart format, build_runner generation and clean diff, flutter analyze, flutter test, plus Android compilation where SDK/plugin integration is touched.
- CI: existing .github/workflows/readiness.yml green on the exact head; never merge with only local tests when CI is required.
- Physical Android: document device, Android version, build SHA, screenshot/video of relevant journey, and actual pass/fail. Include a backend-enabled third compatible test service that becomes visible following a catalog refresh without a client rebuild. No live Google API requests in ordinary CI unit tests.
- Privacy/permissions: verify location permission UX, Google branding/attributions, pre-assignment Driver location protections, safe error messages and restricted keys.
- Cost: field masks, autocomplete session handling, debounce, authentication/rate limits, route requests only on settled inputs, API quotas/billing alerts.
- Google terms: use Google-derived route/Places content with the Google map. Do not archive route geometry or build a competing routing dataset; review current retention terms before persisting derived provider data.

## 7. Deferred follow-ups (not part of this milestone)

1. Driver-to-pickup and pickup-to-destination stage routes were brought forward and implemented in #128. Live arrival ETA presentation, continuous rerouting and optional open-in-Google-Maps navigation remain separate follow-ups; do not reimplement the delivered stage polylines.
2. Smooth Driver marker animation and measured live trip tracking; keep the existing server presence policy until deliberately revised.
3. Broader visual design-system and accessibility refresh across authentication, Driver onboarding and history after the Rider journey is validated.
4. Stage-specific arrival/pickup verification, no-show and interrupted-trip policies based on pilot findings.
5. Full turn-by-turn navigation, automatic pricing/charging, surge, sophisticated dispatch, Route Matrix at large scale, service geofences, multi-region backend, wallets/commissions and Courier/Freight.

Do not introduce Redis, WebSockets, PostGIS, queues, microservices or Kubernetes just for this route preview.

## 8. Explicit stop conditions and risks

- Gate 0 fails: fix the existing cash ride loop before exposing the expanded booking flow to the pilot.
- Map SDK changes violate ADR-0008 gestures/camera behavior or break the Driver dashboard: repair within PR 2 before proceeding.
- Places tokens, Google key restrictions, billing quotas or provider terms are misunderstood: resolve before sending real user traffic.
- No approved real pricing configuration exists: PR 5 tests may pass with fixtures but suggested fares stay disabled in the pilot until owner approves the actual policy. A newly introduced service must not go bookable in the new pricing-enabled flow without an active approved service-specific policy.
- Google cannot find a road route: show an unavailable state; never substitute straight-line distance as if it were the route.
- A proposed PR changes Rider-only assignment or immutable agreed-fare behavior: stop and revise the PR or formally supersede the relevant ADR before implementation.

## 9. Source of truth and vendor references

Existing project contracts:
- docs/README.md
- docs/architecture-decisions.md
- docs/mvp-scope.md
- docs/technology-stack.md
- docs/ADR-0007-ride-request-marketplace-model.md
- docs/ADR-0008-dashboard-panel-interaction-contract.md
- docs/ADR-0009-driver-service-vehicle-eligibility.md
- docs/ADR-0011-minimal-cash-settlement-and-receipt.md
- docs/flutter-client-architecture.md
- docs/pilot-cash-ride-acceptance.md
- docs/WORKLOG.md

Current Google documentation to re-check when each adapter is implemented:
- Routes API: https://developers.google.com/maps/documentation/routes/compute_route_directions
- Places Autocomplete (New): https://developers.google.com/maps/documentation/places/web-service/place-autocomplete
- Places session tokens: https://developers.google.com/maps/documentation/places/web-service/place-session-tokens
- Maps SDK Android: https://developers.google.com/maps/documentation/android-sdk/start
- Google Maps Platform terms: https://cloud.google.com/maps-platform/terms
- Service-specific terms: https://cloud.google.com/maps-platform/terms/maps-service-terms

**Current integration state:** PRs #119–#123, #125 and #127–#130 are merged. The original documentation/development branches are retained as references. Implementation checkboxes above describe delivered code; open physical/release checkboxes require evidence. Do not enable production fare suggestions without the owner's explicit approval of real per-service tariffs.
