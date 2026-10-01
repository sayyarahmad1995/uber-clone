# Dynamic Ride Services and Google Maps Booking — Acceptance Matrix

**Status:** Planned validation for ADR-0012 and ADR-0013; not a record of completed tests.

Do not replace [the current cash-ride pilot acceptance checklist](pilot-cash-ride-acceptance.md). Confirm that existing flow separately before allowing expanded booking to enter the pilot.

## Record evidence for every physical run

Record test date, current branch/SHA, Android app build and device/version, environment, request/Trip IDs, service code and pricing policy version, route distance/duration, suggested fare, Rider-proposed fare, selected offer amount, screenshots/screen recording where helpful, and explicit pass/fail for each case. Do not record API keys or unnecessary raw location traces.

## 1. Dynamic service catalog

**Physical-device evidence — 2026-09-30:** On the same installed PR #120 Android build, the owner added a compatible third service and it appeared after refresh without rebuild/reinstall. The Rider created a request for it. An unenrolled Driver could not receive/respond; after approved vehicle/service enrollment, the Driver could discover the request, accept/counteroffer, and the Rider could select it. After service disablement, new requests were blocked while existing Trip/history context remained readable. Rename/reorder, unknown-token fallback, empty/error catalog UI and future price-policy gating remain separate checks.

- [ ] Existing Economy/Comfort appear from authenticated `GET /v1/ride-services` in documented order, with readable descriptions; the app does not hardcode the service list.
- [x] On the **same installed post-PR-1 Android build**, activate an ordinary third test car service in the existing catalog; refreshing the Rider dashboard displays it without reinstalling or rebuilding the app.
- [ ] Rename, reorder and disable the test service from the backend. Its presentation changes on the next successful refresh; a disabled service cannot create a new Ride Request. A historical Trip with that code remains readable.
- [ ] Unknown presentation tokens render with a generic icon. A failed/empty catalog shows an explicit recoverable state, not silently fabricated current availability.
- [ ] A selected service removed while the app is backgrounded is deselected on foreground refresh, and any prior route/fare preview is invalidated.
- [x] Driver eligibility for the test service requires real approved vehicle/service enrollment. Merely activating a catalog entry does not authorize an unenrolled Driver.
- [ ] Once priced booking launches, a test service without a current approved fare policy is not presented as ready for price-backed booking.

## 2. Shared map and places

**Physical-device evidence — 2026-10-01:** Rider and Driver Google Maps rendering, map taps, pickup/destination/Driver markers, explicit current-location focus, manual pan without unwanted re-centering, ADR-0008 panel interactions, rotation/background recovery and the existing Rider→Driver marketplace flow passed on Android. An online-Driver close/reopen auto-center regression was found during the first pass, fixed on `97bfd4b75b20072e19d339a4fc8b716d6ea3ac84`, and the retest passed. Places/search items below remain for PR 3.

- [x] Rider and Driver dashboards render Google Maps, markers, current-location focus and cached location using the shared app-owned map adapter.
- [x] Shared panel gesture sizes/ownership/scrolling and expanded/collapsed persistence still pass ADR-0008 regression on both capabilities.
- [ ] Rider searches for pickup and destination, selects Places results and sees readable labels and accurate pins. Dragging/adjusting a pin triggers coordinate and label reconciliation.
- [ ] Validate no prediction, denied location permission, poor GPS, lost internet and a stale autocomplete response. No wrong-place selection or hidden provider error.
- [ ] Google content displays over the Google map with appropriate provider attribution. No API key is shipped in committed source.

## 3. Routing and suggested fares

- [ ] Two plausible local locations return one road-driving route; Flutter displays the encoded polyline and fits both endpoints with the bottom panel visible.
- [ ] Display distance in road kilometres and estimated journey duration using the **same** backend route result. Verify no-route response shows an error rather than a Haversine fallback labeled as a road route.
- [ ] Changing either endpoint or the selected service invalidates prior route/fare displays and ignores out-of-order responses.
- [ ] Owner-approved Economy and Comfort parameters produce reproducible suggested PKR fares for controlled fixture distances/durations; backend owns math, version and final integer rounding.
- [ ] A third test service has its **own** active pricing policy and its own reproducible suggestion; deleting/disabling its policy prevents new price-backed booking without changing previously agreed prices.
- [ ] The Rider may edit the suggestion. Existing Driver exact-fare and counteroffer limits apply to the **Rider's submitted fare**, not the suggested amount.

## 4. Complete ride and recovery

- [ ] With one approved Driver, perform an Economy ride and a separate Comfort ride: search → route preview → estimate → Rider proposal → request → Driver exact/counteroffer → Rider selection → assigned → in progress → completed/unsettled → cash collected → receipts.
- [ ] A third compatible service follows the same journey without an app update after its activation, policy approval and Driver enrollment.
- [ ] The agreed Trip fare equals the **selected Driver offer**, not the earlier suggested fare. Reopening or settling the Trip does not reprice it.
- [ ] Restart, background/foreground, provider timeout and transient network loss never display a stale fare as current or strand an active Trip.
- [ ] Disable a service after an assignment: no new requests may use it, but the existing Trip can finish and its immutable service/amount/history remain readable.

## Validation gate

Before each PR merges, use the existing Readiness workflow on its exact head: Go formatting/vet/PostgreSQL tests and Flutter format/build-runner/drift/analyze/full tests. Android map/plugin PRs also require physical build/device checks. Do not call the expanded booking flow pilot-ready until this matrix **and** the existing cash-loop gate are recorded as passed. Any failed critical case becomes the next fix.
