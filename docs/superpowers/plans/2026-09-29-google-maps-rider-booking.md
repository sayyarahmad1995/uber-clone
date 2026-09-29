# HiGO Google Maps Rider Booking Experience — implementation plan

**Status:** Proposed; planning only. No feature code is authorized by this document alone.
**Prepared:** 2026-09-29
**Baseline reviewed:** main at 2285d8febf455e31db4049f0d2e1464bdcd410af (2026-09-28)
**Owner:** HiGO ride-hailing MVP
**Execution model:** Small, independently reviewable pull requests; test-driven changes; physical Android acceptance before pilot release.

## 1. Goal and release boundary

Deliver a complete Rider booking preview using the owner's existing working Google Maps Platform setup. A signed-in Rider can find pickup and destination, adjust them on a Google map, see a valid **driving** route and polyline, view road distance and estimated duration, see a backend-calculated **suggested** fare for Economy or Comfort, optionally change the proposed fare, and submit the existing Ride Request.

This is a core booking-experience milestone, not a dispatch rewrite, new booking mode, full navigation product, or general payments platform. A single shared Flutter application and a Go modular monolith remain authoritative.

### Existing baseline

- Shared Flutter Android app, Riverpod, Dio, current map-first Rider/Driver dashboard, flutter_map/OpenStreetMap, geolocator.
- Go/PostgreSQL backend; Rider requests, geographic Haversine discovery, Driver exact-fare offers/counteroffers, Rider-only selection, Trip lifecycle, cash settlement, history.
- Existing background Android Driver location publishing and server freshness policy; existing authenticated app-owned API surface.
- Live Google Cloud Console credentials/APIs are already configured and working, per the project owner. **Do not repeat Cloud setup.** Only verify project/API separation, key restrictions, app signing configuration, and quotas as part of release hygiene.
- Current WORKLOG still describes PR #83 cash settlement as future work even though settlement API and UI are implemented; reconcile it in the documentation PR.

### Success criteria

1. Google map renders for both existing Rider and Driver dashboards without changing their interaction contract.
2. Pickup/destination can be selected by address search or map adjustment, with meaningful labels.
3. Backend returns a valid road route, distance in meters, duration in seconds, and encoded polyline; Flutter draws it and fits the camera.
4. The Go backend returns a suggested PKR fare from an approved, versioned Economy/Comfort pricing policy, never from client-side math.
5. Rider can retain or edit the suggested amount as their proposed fare and create a Ride Request; current exact-fare/counteroffer/selection behavior is unchanged.
6. No route/no results/location denial/provider error lead to explicit, recoverable states; no Haversine distance is mislabeled as driving distance or ETA.
7. The full route-preview-to-cash-receipt path passes backend, Flutter and physical-device acceptance, including recovery.

## 2. Prerequisite — Gate 0: existing cash-ride acceptance

Before releasing this milestone to the pilot, complete docs/pilot-cash-ride-acceptance.md on current main and record build/commit, device, accounts, Economy/Comfort, selected fares, linked ride/trip IDs, pass/fail and recovery evidence. Verify exact-fare and counteroffer paths, Rider selection, both cancellation paths, completion, completed/unsettled recovery, cash collection, receipts and database invariants. Any blocking failure is repaired before expanding the pilot.

Gate 0 may be executed while the new documentation PR is prepared, but is a release gate. The repository checklist is evidence of planned tests, **not** evidence that every test has passed. Do not claim pilot readiness based on CI alone.

## 3. Decisions and invariants

### Preserve accepted architecture

- Go modular monolith, PostgreSQL transactional ownership, application-owned transport and provider interfaces; compose concrete adapters in backend/cmd/api.
- One account and one shared Flutter app. Rider is default. Driver eligibility derives from the operating vehicle's active approved service enrollments.
- One Rider request with pickup, destination, service, proposed fare and currency. Drivers may accept that fare or counteroffer. **Only Rider offer selection creates a Trip.** Selection and offer revision checks stay atomic.
- Existing offer bounds (currently 90–130% of the **Rider-proposed** fare) are not silently rebased to the calculated suggestion.
- Selected offer fare and Driver/vehicle/service context remain immutable Trip snapshots; completion and cash settlement do not recalculate the fare.
- Haversine distance continues to rank/filter inexpensive marketplace discovery; routing is for the Rider's chosen trip preview, not a Google request for every Driver/request pair.
- ADR-0008 RideDashboardScaffold behavior is unchanged: Rider 18%, Driver 16% collapsed, maximum 60%, shared committed extent and the existing gesture/scroll contract.
- Keep HTTP polling for the current ride-flow pilot, Android first, and the existing server-authoritative Driver location lease.

### Provider architecture

- Map display: Google Maps SDK for Android via an evaluated/pinned compatible google_maps_flutter release. Confine plugin types to mobile/lib/core/maps, expose app-owned map markers, coordinates, polylines and camera operations to feature screens. Replace flutter_map/MapTiles usage on both Rider and Driver screens as one migration; do not leave Google Places/Routes content displayed over OSM.
- Location: preserve the existing DeviceLocation/geolocator abstraction. SDK location features need not replace it.
- Place search/details and optional reverse geocoding: Go-backed, authenticated application APIs with a Google adapter; Flutter never holds or calls the unrestricted server API key. The Android map key is a separate Android-app-restricted key.
- Route preview: Go-owned routing port and Google Routes API adapter; no Google-specific SDK models in ride, offer, marketplace, Trip or public domain contracts.
- Fare suggestion: a small application-owned Go pricing component consuming validated route distance/duration and selected ride service. Its parameters and version are not hard-coded in Flutter.

### New documentation decision

Before writing implementation code, create ADR-0012: Google Maps booking preview and advisory fare policy. It must explicitly move **selected-trip route preview** out of ADR-0007's earlier deferral while preserving Haversine discovery and Rider-selected assignment. Update docs/technology-stack.md, docs/flutter-client-architecture.md, docs/mvp-scope.md, docs/architecture-decisions.md and docs/WORKLOG.md accordingly. ADR-0011 remains the historical scope of **cash settlement**; do not rewrite it as if it authorized pricing.

## 4. Delivery sequence

### PR 0 — Document the next slice

- [ ] Record ADR-0012, the architectural ownership, MVP limit, route/fare data semantics, and clear deferrals.
- [ ] Update docs listed above; mark PR #83/#86 cash settlement complete, distinguish physical pilot validation from completed code.
- [ ] Note the current vehicle-only Driver operating context, not the outdated single-global-service wording in the WORKLOG.
- [ ] Add a focused acceptance matrix for the new booking preview; do not replace the existing cash checklist.
- **Done when:** Documentation no longer names cash settlement as the next feature, accepted marketplace/dashboard/settlement rules remain consistent, and the PR is documentation-only.

### PR 1 — Shared Google Maps rendering foundation

- [ ] Verify compatible, pinned google_maps_flutter/Android Maps SDK versions and Android build requirements before updating dependencies.
- [ ] Implement the Google map adapter in mobile/lib/core/maps; adapt RideMap's public inputs, markers, map-tap callback and camera commands without passing provider classes into Rider/Driver features.
- [ ] Migrate BOTH current map surfaces, current-location focus, marker presentation and cached-center handling; remove obsolete MapTiles/OpenStreetMap references once all callers are migrated.
- [ ] Preserve dashboard panel gesture/extent regression behavior and marker access restrictions; no new ride business behavior.
- [ ] Provide environment-specific Android API key injection without committing any key. The SDK key must be Android application-restricted (package plus signing fingerprint) and Maps-SDK-restricted; keep the backend key server-side.
- [ ] Explicit user-initiated focus/fit should control the camera. Do not automatically recenter on each Driver location update after the user pans the map.
- **Tests:** shared-map widget/adapter tests; Rider and Driver dashboard gesture/persistence regression; Flutter analyze/test; Android debug/release build smoke; physical-device map/permission/rotation/background checks.
- **Done when:** Existing pickup/destination and Driver markers render correctly on Google maps and the prior booking behavior still works.

### PR 2 — Pickup/destination search and adjustment

- [ ] Add authenticated, app-owned endpoints for Places Autocomplete (New), selected Place Details (New), and reverse geocoding when an arbitrary dragged pin needs a readable label. Choose narrow field masks; no generic Google API proxy.
- [ ] Generate a fresh Places session token for each autocomplete session (pickup and destination independently), pass it through to the concluding Place Details request, and use credentials from the same Google Cloud project.
- [ ] Use sensible local location bias rather than a fabricated permanent service boundary. Add a short debounce, minimum input length, stale-response cancellation and per-user rate limiting to control latency and billable calls.
- [ ] Build Flutter pickup/destination inputs, suggestion selection, map pin adjustment and reverse-geocoded display. Show meaningful address text while retaining precise coordinates for booking.
- [ ] Handle empty predictions, inaccurate GPS, location permission denied, invalid place, lost network, request cancellation, and unsupported areas without inventing an address.
- **Tests:** fake Places adapter plus Go HTTP contract tests; Flutter search-session lifecycle and stale-result tests; physical Android search/select/drag flow.
- **Done when:** Both endpoints can be searched or map-adjusted and the user can correct an imprecise pin; the current Ride Request contract is not yet changed.

### PR 3 — Authoritative road route preview and polyline

- [ ] Add backend/internal/routing (or equivalent domain-owned package) with a provider-neutral Preview route operation, a Google Routes adapter and injected HTTP client/config.
- [ ] Request driving route with the minimal Google field mask for distanceMeters, duration and encodedPolyline; return only data needed for presentation. Define timeout, cancellation, bounded retry for transient errors, and deterministic error mapping. Do not automatically request alternative routes or detailed navigation instructions.
- [ ] Expose a narrow authenticated POST /v1/ride-previews endpoint accepting pickup, destination and service_code. Initially return route geometry/distance/duration; PR 4 adds an optional suggested_fare object and policy version.
- [ ] Normalize the public response to route.distance_meters, route.duration_seconds and route.encoded_polyline. Validate coordinates and exclude zero/invalid/unroutable or malformed responses. Keep Google-specific wire types inside the adapter.
- [ ] Add a Flutter route preview controller; decode/draw the route polyline on Google Maps and fit it with panel-aware padding. Invalidate outdated previews whenever pickup/destination changes; discard late replies and allow explicit retry.
- [ ] Clearly label all numbers as *route estimates*. Keep Haversine pickup distance in marketplace discovery unchanged. Do not confuse trip distance with Driver-to-pickup distance.
- **Tests:** fake Routes HTTP fixtures, field mask/timeout/error/coordinate validation tests; encoded-polyline/viewport widget tests; contract checks; physical route on Android.
- **Done when:** A Rider selecting two routable locations sees the road polyline, road distance and estimated duration; an unroutable pair shows a real error, not a straight line presented as a road route.

### PR 4 — Versioned suggested fare, without changing negotiation

- [ ] Add an app-owned pricing policy keyed by service and currency, initially Economy/Comfort and PKR. For each service approve: minimum fare, base fare, minor-units-per-km, minor-units-per-minute, rounding rule, version and effective-date policy. **Production values require explicit business approval**; test-fixture numbers are not tariffs.
- [ ] Compute suggested fare in Go from the valid route distance and duration using integer minor units and explicitly defined rounding/overflow guards. Recommended starting formula: max(minimum, round(base + perKm * distanceMeters/1000 + perMinute * durationSeconds/60)). No surge, tolls, taxes, discounts or cancellation charges in this slice.
- [ ] Extend POST /v1/ride-previews to return suggested_fare (amount_minor, currency) and pricing_policy_version alongside the corresponding route. Calculate each combined preview from one route response rather than duplicating Google calls.
- [ ] The result is **advisory**. The Rider may accept or edit it as proposed_fare; the existing offer bounds still apply to that proposal. The backend must never accept a client-computed suggestion as a trusted charge or silently mutate the settled amount.
- [ ] Do not add a persistent Google polyline/duration history as a side effect. Retain user-owned pickup/destination and the already immutable selected-offer fare. Any proposal to persist additional Google-derived content must pass a fresh terms/caching review.
- [ ] If a route or pricing policy is unavailable, do not fabricate a price. For the initial Google-preview-enabled path, block price-backed request submission with a clear retry state; manual-only fallback is a separate explicit product decision.
- **Tests:** zero/near-zero, long, invalid, overflow, rounding, policy version, Economy vs Comfort, route failure and price display/edited-proposal tests.
- **Done when:** Backend reproduces an approved fare for the exact preview inputs and the Rider can propose a different amount without altering assignment or settlement.

### PR 5 — Complete Rider booking UX and end-to-end verification

- [ ] Connect location search, map adjustment, chosen service, combined route/fare preview, price editing and the existing POST /v1/ride-requests action in a single comprehensible flow.
- [ ] Keep ADR-0008 panel behavior intact. Ensure the polyline/markers are visible with an expanded bottom sheet, keyboard, small device and text scaling; show appropriate loading, empty, permission and retry states.
- [ ] Do not show a stale route/price after any pickup, destination or service change. Prevent accidental duplicate submission and recover the backend-authoritative active ride after restart or request timeout.
- [ ] Preserve Rider offer comparison, rejection, chosen offer/assignment, Driver operations, cash confirmation, and both histories. The agreed fare shown in history must remain the selected Driver offer rather than the earlier suggestion.
- [ ] Extend docs/pilot-cash-ride-acceptance.md (or a companion feature checklist) with Google routing/search/fare/UX physical-device cases, while retaining all existing lifecycle and settlement checks.
- **Tests:** Flutter booking-controller, widget/interaction and full regression; Go provider/HTTP/pricing and PostgreSQL regression; physical Android complete Economy and Comfort ride paths including no-route and network recovery; exact-HEAD readiness CI.
- **Done when:** The owner can demonstrate one complete, real-device route-preview-to-cash-receipt ride, including edited proposed fare/counteroffer and recoverable failures.

## 5. Proposed application contracts (finalize in PR 0)

- POST /v1/places/autocomplete: input, session_token, optional user-location bias; returns place IDs and display text, not a raw provider object.
- GET /v1/places/{place_id}: session_token; returns normalized place ID, label and coordinates. The matching autocomplete token concludes here.
- POST /v1/places/reverse-geocode: coordinates; returns a readable label or an explicit unavailable result.
- POST /v1/ride-previews: pickup (latitude/longitude), destination (latitude/longitude), service_code; returns route {distance_meters, duration_seconds, encoded_polyline}, suggested_fare {amount_minor, currency}, pricing_policy_version, and request correlation metadata. During PR 3, suggested_fare may be null; it is required only after a pricing policy is configured for the released PR 4+ experience.
- Existing POST /v1/ride-requests remains the booking command (service_code, pickup, destination, proposed_fare). Do not introduce a second booking mode or require a durable quote token for an **editable suggestion** in this milestone.

All new endpoints are authenticated; validate payload sizes and coordinates, bound external provider calls, and avoid placing raw secrets or precise location trails in operational logs. Public response structs must not leak Google billing/error internals.

## 6. Verification and release gates

Before merging each implementation PR:
- Go: gofmt on changed files, go vet ./..., go test -p 1 ./... with the dedicated _test PostgreSQL database, including focused fake-provider tests.
- Flutter: dart format, build_runner generation and clean diff, flutter analyze, flutter test, plus Android compilation where SDK/plugin integration is touched.
- CI: existing .github/workflows/readiness.yml green on the exact head; never merge with only local tests when CI is required.
- Physical Android: document device, Android version, build SHA, screenshot/video of relevant journey, and actual pass/fail. No live Google API requests in ordinary CI unit tests.
- Privacy/permissions: verify location permission UX, Google branding/attributions, pre-assignment Driver location protections, safe error messages and restricted keys.
- Cost: field masks, autocomplete session handling, debounce, authentication/rate limits, route requests only on settled inputs, API quotas/billing alerts.
- Google terms: use Google-derived route/Places content with the Google map. Do not archive route geometry or build a competing routing dataset; review current retention terms before persisting derived provider data.

## 7. Deferred follow-ups (not part of this milestone)

1. Driver-to-pickup route, traffic-aware arrival estimates, and optional open-in-Google-Maps navigation.
2. Smooth Driver marker animation and measured live trip tracking; keep the existing server presence policy until deliberately revised.
3. Broader visual design-system and accessibility refresh across authentication, Driver onboarding and history after the Rider journey is validated.
4. Stage-specific arrival/pickup verification, no-show and interrupted-trip policies based on pilot findings.
5. Full turn-by-turn navigation, automatic pricing/charging, surge, sophisticated dispatch, Route Matrix at large scale, service geofences, multi-region backend, wallets/commissions and Courier/Freight.

Do not introduce Redis, WebSockets, PostGIS, queues, microservices or Kubernetes just for this route preview.

## 8. Explicit stop conditions and risks

- Gate 0 fails: fix the existing cash ride loop before exposing the expanded booking flow to the pilot.
- Map SDK changes violate ADR-0008 gestures/camera behavior or break the Driver dashboard: repair within PR 1 before proceeding.
- Places tokens, Google key restrictions, billing quotas or provider terms are misunderstood: resolve before sending real user traffic.
- No approved real pricing configuration exists: PR 4 tests may pass with fixtures but suggested fares stay disabled in the pilot until owner approves the actual policy.
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

**Planning branch only:** This document does not implement or merge any of PR 0–5. Start only after the plan and any required production tariff decisions are approved.
