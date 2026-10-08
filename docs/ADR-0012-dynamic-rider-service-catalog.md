# ADR-0012: Dynamic Rider Ride Service Catalog

## Status

Accepted for the next MVP implementation milestone. Not implemented yet.

## Date

2026-09-29

## Context

ADR-0009 already models ride services as approved Driver/vehicle products, not account capabilities. Migration 018 introduced the PostgreSQL `driver_service_catalog` with stable `code`, `display_name`, `description`, `is_active` and `sort_order`. The Driver onboarding API already reads that catalog.

Rider booking is not yet dynamic: `RiderServicePicker` lists only Economy and Comfort, and the Go Ride Request service rejects any other service code even when the database catalog contains it. The upcoming Google Maps/route/fare booking experience must not reproduce those restrictions. After one compatible client update, the owner wants to activate a new ordinary car category without requiring another Android app release.

## Decision

1. **One catalog.** Reuse `driver_service_catalog` as the source of truth. Its service `code` is a stable business identifier used across Ride Requests, Driver enrollments, offers, Trip snapshots, pricing policies and history. Display names/order can change without changing the identifier. Extend the existing catalog minimally for Rider visibility or safe presentation metadata if needed; do not create an unrelated Rider catalog.
2. **Rider-facing API.** Add authenticated `GET /v1/ride-services`, returning currently bookable, compatible entries in deterministic order. The minimum response per entry is `code`, `display_name`, `description` and `display_order`; optional `presentation_token` is advisory only. Do not expose Driver-only vehicle eligibility rules, approval state, private details or internal cost data in this response.
3. **Server-owned availability.** `is_active` and any explicit Rider-visibility state are authoritative for *new* requests. A disabled/hidden/unknown service cannot be newly selected by a Rider, regardless of stale client state. The same catalogue check belongs in the backend Ride Request creation path and in the database write boundary. Disabling a service does not rewrite or invalidate already assigned Trip snapshots or historical receipt data.
4. **Dynamic mobile presentation.** The Flutter Rider app fetches the catalog when entering the Rider dashboard and again when returning to the foreground or retrying an unavailable catalog. It builds the picker/cards from the response, using a generic fallback icon for an unfamiliar presentation token. No service enum, switch or bundled graphic is required for a compatible new car category. After a change, an already-installed compatible app displays the new catalog on its **next successful refresh**, not via an instantaneous pushed update.
5. **Selection changes.** On refresh, preserve a selection only while its code remains bookable. If it disappears, clear the selection and any route/fare preview derived from it; require the Rider to choose an available service. Do not silently substitute a different service in an active request.
6. **Driver enrollment remains independent.** Publishing a service does not enroll Drivers. Driver offers and assignment still require an explicit, active approved enrollment for the request's service on the online Driver's selected verified operating vehicle, plus all existing eligibility/freshness checks. Derived technical eligibility does not create an enrollment.
7. **Service-specific pricing is a related, separate versioned policy.** Each bookable service in the Google-preview-enabled flow must have an approved active fare policy for the relevant currency (initially PKR). Pricing parameters and revisions live in a separate table keyed by service code, currency and policy version, not duplicated in Flutter and not overwritten in-place on every rate change. PR 1 may retain the *existing manual-fare booking path* until PR 5 activates suggestions; it must not incorrectly require nonexistent pricing policies during that transition. Once the new price-backed flow launches, a service with no approved active pricing policy is unavailable for new price-backed booking. Existing rides remain readable.
8. **Small rollout mechanism.** Initially add/activate catalog entries and approved pricing through reviewed migrations or narrowly controlled operator configuration. A full administrator product or rules engine is unnecessary. Verify the selected service's real Driver supply and approved policy before making it visible.

### Planned contract

`GET /v1/ride-services` (authenticated; no Rider/Driver account split):

```json
{
  "services": [
    {
      "code": "economy",
      "display_name": "Economy",
      "description": "Standard ride service",
      "display_order": 10,
      "presentation_token": "car"
    }
  ]
}
```

Values are illustrative. Do not treat this display description or token as new approved production data. The API is app-owned; provider- or admin-internal fields stay out of the response.

## Compatibility and acceptance

- An unchanged *post-PR-1* Flutter build shows a synthetic third compatible ride service added through backend configuration following the next catalog refresh. It can submit a Ride Request once the service is enabled and the backend's other requirements are met.
- Reordering, renaming and disabling the synthetic service require no additional app build. Disabled services cannot accept new Ride Requests, but older Trips and receipts still show their original service.
- Unknown icon tokens fall back to a generic card. No catalog or network response yields a recoverable UI state, never a fabricated list presented as current.
- After pricing launches, the third service must have an approved applicable pricing policy to appear as price-backed bookable, and the backend must reject stale or disabled selections.
- No change to one Rider request, Driver exact-fare/counteroffer, Rider-selected assignment, current offer bounds, immutable agreed fare, or cash settlement.

## Non-goals

Do not use the catalog to launch Courier/Freight, multi-stop rides, reservations, per-service custom screens, a broad admin panel, configurable rule engines, or new marketplace assignment modes. Services requiring different workflow or capabilities may still require a new application release.

## Relationship to other decisions

- Extends ADR-0009's vehicle/service model; does not change its enrollment and verification requirements.
- Preserves ADR-0007's single Rider request and Rider-selected assignment.
- Does not modify ADR-0008's shared map/panel interaction.
- Service-specific advisory pricing and Google routing are defined separately in ADR-0013; ADR-0011's cash settlement remains based on the selected offer's agreed fare.
