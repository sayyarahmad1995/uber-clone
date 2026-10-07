# ADR-0013: Google Maps Booking Preview and Advisory Service-Specific Fare

## Status

Accepted for the next MVP implementation milestone. Not implemented yet.

## Date

2026-09-29

## Context

The current Android Rider dashboard uses `flutter_map` with OpenStreetMap tiles and markers. Riders select pickup/destination by map tap, see coordinates, and type a proposed fare manually. The marketplace reports **Haversine straight-line** Driver-to-pickup distance; there is no road route polyline, driving distance, journey duration or routing-based fare estimate.

The owner has already configured working Google Maps Platform APIs. The next MVP slice is an understandable Rider booking journey, built *after* ADR-0012's dynamic Rider service catalog. The earlier routing and ETA deferrals in ADR-0007 and the cash-settlement non-goals in ADR-0011 described their **earlier milestones**, not a permanent ban on a separate, approved booking-preview slice.

## Decision

### 1. Map and Google integration

- Replace the shared Flutter `flutter_map`/OpenStreetMap rendering adapter with Google Maps SDK for Android through a compatible `google_maps_flutter` version verified during implementation. Both Rider and Driver dashboards migrate together. Keep provider types within `mobile/lib/core/maps`; Rider/Driver features use application-owned marker/coordinate/route/camera interfaces.
- Preserve existing `DeviceLocation`/geolocator and all ADR-0008 `RideDashboardScaffold` rules, including panel sizes, shared committed extent, gesture ownership and scroll behavior. Fit route geometry with padding that preserves the visible map above the panel. Respect explicit user map panning instead of recentering on every Driver update.
- Google Places Autocomplete (New), Place Details (New), and Geocoding v4 use narrow authenticated application-owned Go HTTP endpoints and an injected Google adapter. The client owns separate autocomplete session tokens per search selection and sends the token through the concluding Place Details operation. A Rider-confirmed free map pin is authoritative: reverse geocoding may supply a readable formatted address, but it must never move or replace that coordinate, and any Place ID returned by reverse geocoding is descriptive only and must not become a routing endpoint. Separately, typed Place selections and explicit taps on Google-rendered POI labels retain the selected Google Place ID together with the canonical Place Details coordinate. This explicit named-place selection is not automatic snapping. Use field masks, debounce, an admin-configurable user-location `locationRestriction` (5–50 km, default 50 km), and request limits; no unbounded Google passthrough endpoint. This is a runtime metro search window, not a permanent service-area boundary: when device location is available, autocomplete must exclude distant matches. Within Google's matched local prediction set, HiGO re-ranks by Google `userRatingCount` as a popularity proxy, with a 24-hour backend cache; equal/unknown popularity preserves Google's original order. Popularity must never introduce a place that Google did not already return for the typed query.
- Google Routes API is called **from Go** through a provider-neutral routing port with configured timeout, cancellation and bounded transient-error behavior. Rider booking previews use `DRIVE`, `TRAFFIC_AWARE_OPTIMAL`, and `BEST_GUESS`, with `computeAlternativeRoutes=false`. HiGO accepts Google's single recommended/default traffic-aware route rather than requesting shorter-distance or alternative candidates. For an explicitly selected Google Place, the routing waypoint uses that retained Place ID so Google can apply place/access-point routing semantics; for a free pin/current location, the waypoint uses the exact selected latitude/longitude. Request only driving distance, traffic-aware duration and encoded polyline needed for the preview. Flutter renders that one route. Route unavailability is explicit: never draw the great-circle line or display Haversine as a road route/ETA.
- The Android Maps SDK key and server-side Google web-service key are separate, API-restricted and appropriately app/server-restricted; no secrets enter source control or mobile public API responses. Review current Google Maps Platform display, attribution, caching and retention terms before storing provider-derived data. Do not persist detailed polylines or location trails for this slice.

### 2. Independent distance semantics

- The new **Rider trip preview** is one Google-recommended traffic-aware driving route between the selected pickup and destination: `distance_meters`, `duration_seconds`, and `encoded_polyline`. HiGO does not separately optimize for shortest distance or expose alternatives during booking. The preview remains an estimate rather than a guarantee of local road quality or the Driver's eventual navigation path. Label all values as estimates, not measured trip distance.
- The current Driver-to-pickup **discovery distance** remains cheap Haversine straight-line distance for marketplace ordering and eligibility. Do not make a paid Google Routes request for every Driver/ride pair or relabel this distance as Driver arrival time.
- Driver-to-pickup routing, traffic-aware ETA, live rerouting and turn-by-turn navigation remain subsequent product decisions, not part of this first preview milestone.

### 3. One recommended route preview, one advisory fare

- Add a small Go pricing component consuming validated route distance/duration and the selected **active catalog service**. Each service/currency has its own independently versioned, approved policy. Initial currency is PKR; initial services are Economy and Comfort. A synthetic third service tests the architecture without adding a production tariff.
- Required policy fields: `service_code`, `currency`, `version`, `base_fare_minor`, `rate_minor_per_km`, `rate_minor_per_minute`, `minimum_fare_minor`, `rounding_increment_minor`, `effective_from`, database publication time and reviewer. Publication takes effect immediately through one current pointer; disabling clears that pointer with an expected policy identity. Future scheduling is outside this milestone. Changes create new versions; historical policies are not silently overwritten.
- Proposed computation, using integer minor units and a defined **single final rounding step**:

  ```text
  raw = base_fare_minor
      + rate_minor_per_km * distance_meters / 1000
      + rate_minor_per_minute * duration_seconds / 60
  suggested_fare_minor = max(minimum_fare_minor,
                             round(raw, rounding_increment_minor))
  ```

  Use exact rational/integer intermediates, round half up once to the increment, then enforce the minimum. Reject invalid input or results above 1,000,000,000,000 minor units; use no binary floating point for money. Rate values and rounding increments are **business configuration**, not invented constants. The owner must approve actual Economy and Comfort tariffs before production fare suggestions are enabled.
- Extend the single authenticated `POST /v1/ride-previews` call to return the one recommended route. During PR 4 it contains distance/duration/polyline. PR 5 adds `suggested_fare {amount_minor,currency}` and `pricing_policy_version`, calculated from that same route without another Google call. After pricing activation, no price-backed booking proceeds without a valid active policy.
- This is a **suggestion, not a fare quote or automatic charge**. The Rider may accept or edit the suggested amount as their **proposed fare**. Existing 90–130% Driver response bounds remain relative to the Rider's submitted proposal, not the system suggestion. Neither previewing nor Driver accepting the Rider's amount assigns a Trip.
- Ride Request creation remains `POST /v1/ride-requests` with coordinates, selected service and Rider-proposed fare. Revalidate service activation and the preview policy UUID at create time. The editable proposal is Rider input. Copy the authoritative policy identity, numeric version, five money parameters, calculation rule and snapshot time in the same transaction as the request. Store no client-supplied suggestion, Google route metrics or geometry. Shared catalog/pointer locks serialize creation against publication: publication first returns 409 with no request; creation first permanently keeps its snapshot. Later tariff changes cannot change open request proposals, offer bounds or Trip fares. Rider offer selection is the only assignment boundary; the selected offer's **agreed fare snapshot** is final for the Trip and subsequent cash settlement. Do not retrospectively reprice existing Trips or use actual travelled distance to change the agreed fare in this milestone.

### 4. Client states and failure behavior

- Address search and map adjustment must invalidate stale route/fare previews. Service changes, service deactivation and late network responses likewise invalidate stale previews. Disable request submission until the user has a current successful preview in the launched price-backed flow.
- Show understandable loading, unavailable route, permission denied, quota/provider error, offline and retry states. Do not invent a route, distance, ETA or price in a failure state.
- Preserve existing Ride Request, offer, Trip and cash-settlement recovery after app restart and interrupted network requests. Do not introduce WebSockets, Redis, queues, PostGIS, new booking modes or additional lifecycle states for this slice.

### Proposed application-owned API

- `POST /v1/places/autocomplete` — input, session_token and optional user-location center; the Google adapter applies the current admin-configured circular `locationRestriction` (5–50 km, default 50 km) when that center is present and returns normalized predictions only.
- `GET /v1/places/{place_id}` — matching session token; normalized label and coordinates.
- `POST /v1/places/reverse-geocode` — coordinates; readable label when available.
- `POST /v1/ride-previews` — pickup and destination each include required latitude/longitude plus optional `place_id` only when the Rider explicitly selected a Google Place; free pins omit `place_id`. The request also includes catalog `service_code`; response returns one normalized `route` with distance/duration/polyline and, once enabled, backend-generated suggested fare and policy version.

Exact request/response schemas, status codes and adapter dependency wiring are finalized in their own small implementation PRs; do not expose Google's raw wire format as the product API.

## Acceptance and rollout

Follow [the booking acceptance matrix](pilot-google-booking-acceptance.md). PRs remain vertical: shared Google map, place selection, route preview, pricing, then complete Rider booking UX. Validate fake-provider backend tests, PostgreSQL invariants, Flutter widget/controller tests, exact-head CI, Android builds and documented physical-device rides.

The preexisting [cash ride acceptance checklist](pilot-cash-ride-acceptance.md) remains a **separate release prerequisite**: verify or repair the current cash loop before enabling the new booking experience for pilot users. A passing CI run alone is not physical-device acceptance.

## Non-goals

No pricing surge, automatic final price, toll/tax estimation, waiting fees, cancellation/no-show fees, wallets, payments platform, large-scale Route Matrix, Driver in-app turn-by-turn navigation, global service areas, arbitrary pickup-radius policy, general support/administrator product, Courier or Freight.

## Relationship to other decisions

- Extends ADR-0007 **only for selected Rider trip route preview**. Its marketplace discovery Haversine policy, Rider choice, selection atomicity and preassignment Driver privacy stay unchanged.
- Uses ADR-0012's dynamic catalog and service-keyed versioned tariff policies.
- Preserves ADR-0008's shared Rider/Driver map/panel interaction contract.
- Does not retroactively expand ADR-0011's earlier *cash-settlement* scope; the agreed offer price and settlement rules remain authoritative.

## Suggested fare rebuild (2026-10-07)

`SUGGESTED_FARES_ENABLED` defaults to false and enables only for trimmed `true`. Disabled retains manual proposals and route-only previews. Enabled filters the catalog to active visible services with a current PKR policy and emits `pricing_required=true`. Rider previews add `suggested_fare` and `pricing_policy_version` (policy UUID); priced creation requires that UUID. Driver route responses remain route-only.

The original service dropdown remains. Suggestions prefill untouched proposals. Retry and a tariff conflict preserve edits; changes to coordinates, Place IDs, service or account reset ownership. A 409 refreshes the preview and requires a second explicit tap. Active-request views use the saved proposal and retain the booked route line.

Operations approves tariffs at `/admin/operations/pricing`; JSON admin endpoints share Basic Auth and origin protections. History and request snapshots reject update/delete. Publication alone does not enable the rollout. No production tariff values are seeded.


## Read-only Rider suggestion (owner revision, 2026-10-07)

This revision supersedes the editable Rider proposal behavior described above when suggested fares are enabled. The Rider sees a read-only suggested fare and requests at that amount. Drivers may still accept or counteroffer within existing limits; the offer selected by the Rider remains the immutable final payable Trip fare.

Priced request creation performs a fresh server-side route lookup and exact calculation, preserving selected Place IDs or exact pin coordinates. It rejects a submitted amount that differs with `suggested_fare_changed` (409), leaving no request or snapshot. The app refreshes the suggestion and requires another explicit tap. This adds one provider route call on submission and may require review if traffic changes the suggestion; no client route metrics or quote token are trusted. Concurrent tariff publication remains protected by the request transaction's policy UUID check. Saved requests, offers, Trip fares and settlements are never repriced.

The disabled-rollout manual flow remains available for rollback compatibility. Deploy the updated backend and Flutter app together before validating the enabled flow; older editable clients may receive a fare-change conflict.
