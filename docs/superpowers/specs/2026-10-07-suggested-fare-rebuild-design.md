# Suggested fare rebuild and immutable Ride Request pricing snapshots

Date: 2026-10-07
Status: Approved by the owner on 2026-10-07; implementation plan awaiting review. Application implementation has not started.
Baseline: fix/driver-trip-route-polylines at b88e412c74c7132cb4490750afc9a79bdddf9f71.
This is a fresh implementation. Do not restore, cherry-pick or copy the implementation from closed PR #124.

## Intent and approved constraints

Help a Rider choose a proposed PKR fare using the selected service and road route, while keeping negotiation and settlement authoritative. The owner approved rebuilding Suggested fare and explicitly requires price/rate snapshots when the Ride Request is created: tariff changes must not affect an open request or ongoing Trip.

Keep the original service dropdown, shared dashboard interactions, Rider and Driver route fixes, Rider-only assignment, Driver eligibility, polling, and existing cash settlement. The suggestion is advisory; the Rider may edit it. Real Economy/Comfort tariffs require explicit owner approval; fixture numbers are never production tariffs.

Current code already persists proposed_fare_minor and currency in ride_requests. Offer bounds use that stored proposal. The new pricing snapshot adds policy provenance and preserves applied rates; it must never introduce dynamic repricing.

## Price ownership

| Value | Authority | Lifetime |
| --- | --- | --- |
| Suggested fare | Go computation from one route result and one approved policy | Current preview only |
| Rider proposed fare | Rider's submitted amount, validated and persisted by Go | Immutable for that Ride Request |
| Pricing snapshot | Server policy version and copied rate fields | Created atomically with the Ride Request; immutable |
| Agreed Trip fare | Driver offer selected by the Rider | Immutable through completion, cash settlement and history |

Snapshot the submitted amount as the request's price. Do not replace it with the suggestion if the Rider edited it. The request stores copied rates and the policy version as pricing context, not a formula that is rerun during offers or settlement.

The calculated preview suggestion remains transient in this slice. Do not store a client-echoed suggestion as a trusted backend result. Persisting a verified historical suggestion would require additional preview provenance; it is unnecessary to protect the proposed or agreed fare and is outside this design. No Google geometry, distance or duration is persisted as a side effect.

## Architecture

Add a small backend/internal/pricing component containing policy validation, exact calculation and repository interfaces. PostgreSQL owns immutable policy versions, the current-policy pointer, publication and request snapshots. HTTP handlers translate app-owned contracts; Flutter only displays the returned suggestion and collects the Rider proposal.

Reuse the existing Admin Operations authentication and reviewer identity for pricing publication. Add a focused pricing page/form linked from Operations rather than expanding its current handler into a large pricing implementation. Composition stays in backend/cmd/api. Request snapshot persistence belongs to the existing request creation transaction; core ride/offer/trip packages must not depend on Google adapters.

Considered alternatives:
- A current-rate lookup every time a request or Trip is displayed would violate the owner's immutability rule.
- A policy-version reference alone preserves provenance if historical policies are immutable, but copying the rates additionally satisfies the explicit snapshot requirement.
- Selected approach: immutable policy history plus a server-written per-request snapshot and the existing stored proposal.

## Policy model and publication

Each policy version has a server identifier, stable service code, PKR currency, monotonic service/currency version, base_fare_minor, rate_minor_per_km, rate_minor_per_minute, minimum_fare_minor, rounding_increment_minor, calculation rule identifier, effective_from, and publisher identity/time.

Rates/base may be zero; minimum and rounding increment must be positive. All money fields must be representable within the existing ride.MaxFareMinor limit. Unsupported currencies, unknown/inactive/Rider-hidden services, malformed decimals, negative values and out-of-range fields are rejected. Policy calculation must separately reject output above the fare limit.

Policy publication is an explicit approval action in Admin Operations. Publish immediately using database time; future scheduling is excluded. Each publication inserts an immutable new version and changes the current service/currency pointer atomically. Previously published versions remain readable and are never overwritten or deleted through application APIs. Disabling a policy clears the current pointer without modifying history or existing requests.

The admin form shows PKR amounts in rupees, parses at most two decimal places exactly, and converts to integer minor units in Go. The publish action shows the complete proposed tariff and states that it affects new previews only. It records the authenticated reviewer. Reuse existing admin authorization/form protections; JSON transport rejects unknown fields. Concurrent publication is serialized per service/currency.

## Exact calculation

For positive validated route distance d in meters and duration t in seconds:

raw = base + perKm*d/1000 + perMinute*t/60
suggested = max(minimum, round_half_up_once(raw, increment))

Use exact integer/rational arithmetic, not binary floating point. Retain the fractions through addition and perform one half-up rounding to the configured increment. Apply the minimum after rounding, consistent with ADR-0013. A minimum need not be a multiple of the increment. Use checked arithmetic or arbitrary-precision intermediates and reject a result that exceeds the existing supported fare range before converting to int64.

No surge, tolls, taxes, waiting charges, discounts, cancellation fees or measured-trip repricing.

## Preview contract

Extend authenticated POST /v1/ride-previews. Keep the existing pickup, destination and service_code inputs. In the enabled pricing flow, validate the selected catalog service and current approved PKR policy before calling routing. Calculate route and price from exactly one routing response.

Return the existing route object plus:
- suggested_fare: amount_minor and currency;
- pricing_policy_version: an unambiguous version identifier for that service/currency.

Missing/unavailable policy is a distinct recoverable error with no fabricated price. Invalid services/coordinates return validation errors; provider failures preserve existing route error mapping. Driver POST /v1/driver/trip/route-preview remains route-only and does not require a fare policy.

No persistent quote record or quote token is introduced. The version identifier is pricing provenance, not authorization to charge or assign.

## Request creation and publication races

In the enabled pricing flow, POST /v1/ride-requests includes pricing_policy_version alongside the existing locations, service and editable proposed_fare. Go looks up its own policy data; the client cannot supply authoritative rates.

Within one database transaction:
1. Lock/revalidate the catalog service and active policy using a consistent lock order shared with publication/deactivation.
2. Confirm that the submitted preview version is the current approved version for the same service and PKR.
3. Insert the request with the Rider's validated proposal and existing deadline rules.
4. Insert its one-to-one snapshot containing the policy identifier/version, service, currency, all applied rate/minimum/rounding fields, calculation rule and snapshot time.
5. Commit both or neither.

If publication wins before request creation, return HTTP 409 with a specific pricing_policy_changed code. Do not silently replace the proposal, choose the new policy, or create a request. Flutter refreshes the preview, explains that rates changed and requires another explicit submission. Preserve a Rider-edited proposal during that refresh.

If request creation wins, its snapshot is fixed before the new version becomes current. Publication never scans or updates requests, offers, Trips or settlements. Unknown/cross-service/cross-currency versions are rejected. Disabled/unpriced services cannot create new requests in the enabled flow.

Use a unique request snapshot relationship and foreign keys to historical policy identity. Do not backfill invented snapshots for legacy requests. Existing lifecycle/history queries remain valid when legacy requests have no snapshot. Request reads never join the current-policy pointer to derive a price.

## Flutter behavior

Keep the dropdown and existing map route behavior. Parse optional pricing fields so route-only Driver responses remain compatible; Rider submission in the enabled flow requires a successful current priced preview.

Prefill an untouched proposed-fare field from the latest current suggestion. User edits establish ownership of that field. Retry or refresh for unchanged inputs must preserve edits. Pickup/destination/service changes invalidate both preview and the old proposal association; clear the old field and prefill after the new successful preview. Never submit a price or version belonging to previous inputs.

Show Suggested fare separately when the Rider has edited the proposal. Display the submitted proposal in request/offer views and the selected offer amount in Trip/receipt views.

Late route/fare replies cannot restore an obsolete version or overwrite edits. Pricing conflicts refresh the preview and require resubmission. Preserve duplicate-submit prevention and backend-authoritative request recovery after timeout/restart. Once a request exists, preview state must not drive its stored amount or clear its booked polyline.

## Rollout

Use one deployment-level suggested-fares enable setting, disabled by default. This is rollout compatibility for the existing booking flow, not a second booking product.

While disabled, retain current route-only/manual-proposal behavior and catalog semantics. Operators may prepare policies. Once enabled, Rider-visible bookable services must have an active approved PKR policy; preview and request creation enforce the same requirement. Do not seed real tariffs or auto-enable pricing after migration/publication.

Publish owner-approved Economy/Comfort tariffs, verify server/app compatibility and policies, then enable the new flow together. Existing requests/Trips continue unchanged even if their service or policy is disabled. Physical-device acceptance remains a release gate; CI is not evidence of a completed ride.

## Required verification

- Calculation: exact fractions, halfway rounding, single final rounding, minimum boundary, zero/negative inputs, long routes, output/intermediate overflow, and distinct Economy/Comfort/synthetic third-service fixtures.
- Policy PostgreSQL tests: immutable history, current-pointer uniqueness, approval identity, invalid service, publication/deactivation and concurrent publishers.
- Request PostgreSQL tests: atomic request/snapshot creation, rollback, exact copied rates and edited proposal, concurrent tariff publication with both lock outcomes, disabled-service races, and readable legacy requests.
- Lifecycle regression: create request under version A, change to B while offers are open, offer bounds still use the A request proposal, Rider selects an offer, change to C, complete/settle/recover and verify the unchanged selected amount.
- HTTP: auth/role enforcement, policy-unavailable vs route failure, one provider call per combined preview, policy-version mismatch/cross-service rejection and conflict semantics.
- Flutter: prefill/edit ownership, Retry, changed inputs, stale replies, policy-conflict resubmission, service removal, third service, preserved booked routes and route-only Driver responses.
- Full exact-head Readiness: Go formatting/vet/PostgreSQL suite; Flutter formatting/generated-code drift/analyze/full tests.
- Android: assigned route, Start trip route, priced/edited request through offer selection and cash receipt, publication during open requests, restart/network recovery and dropdown/panel gestures.

## Documentation and completion

During implementation update ADR-0013 and the booking acceptance checklist to document copied request rates, policy-version conflicts and rollout. The source plan remains the milestone roadmap.

Done means fresh implementation reviewed, exact-head CI green, request snapshots and pricing publication races verified, and a draft PR with explicit physical-test status. Do not merge or enable real tariffs without the relevant owner authorization. After written-spec approval, write the implementation plan for review and execution-method selection.
