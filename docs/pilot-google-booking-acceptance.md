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

**PR 3 automated evidence:** Backend-owned Places/Geocoding adapters use narrow field masks and a five-second provider timeout; app endpoints require Rider capability and apply a configurable per-user request guard. Flutter uses independent UUID-style pickup/destination sessions, 300 ms debounce, cancellation plus stale-response guards, and provider-neutral draggable pins. Physical search/select/drag and adverse-network/location checks remain open.


**Physical-device evidence — 2026-10-01:** Rider and Driver Google Maps rendering, map taps, pickup/destination/Driver markers, explicit current-location focus, manual pan without unwanted re-centering, ADR-0008 panel interactions, rotation/background recovery and the existing Rider→Driver marketplace flow passed on Android. An online-Driver close/reopen auto-center regression was found during the first pass, fixed on `97bfd4b75b20072e19d339a4fc8b716d6ea3ac84`, and the retest passed. Places/search items below remain for PR 3.

**PR 3 physical regression — 2026-10-01:** Places search succeeded after the backend Google key was supplied to the development API container. During the active Rider request flow, a dashboard scroll-handoff defect was found: after content had scrolled away from the top, a continuing downward pull could reach the top without transferring ownership to panel collapse, making the panel feel locked. The shared ADR-0008 handoff is being corrected and requires device retest before PR 3 is accepted.

**PR 3 follow-up regression — 2026-10-01:** Device retest initially found intermittent scroll stalls. The shared raw-pointer handoff could claim panel dragging on a few pixels of downward touch jitter at the top, immediately disabling ListView scrolling even when the intended gesture then moved upward. Collapse ownership now requires Flutter touch slop before scroll physics are locked. The Rider active-request retest passed after this fix; broader Rider/Driver ADR-0008 coverage remains part of the PR gate.

**PR 3 pin-selection clarification — 2026-10-01:** Device review exposed that the implemented map path supported tap-to-drop and dragging existing markers but did not expose the expected visible center-pin selector. A first center-pin iteration also made confirmation hard to discover. PR 3 now gives Pickup and Destination their own field-level map buttons, keeps an explicit confirmation control visible over the map, and resolves the confirmed coordinate to a readable address without changing the coordinate. The Rider's pin is authoritative. Physical retest remains required.

- [x] Rider and Driver dashboards render Google Maps, markers, current-location focus and cached location using the shared app-owned map adapter.
- [ ] Shared panel gesture sizes/ownership/scrolling and expanded/collapsed persistence pass ADR-0008 regression on both capabilities after the PR 3 scroll-handoff fix.
- [ ] Rider searches for pickup and destination, selects Places results and sees readable labels and accurate pins. With device location available, short name searches are locally restricted and popularity-ranked: for example, from Islamabad, typing `faisal` should return relevant places within the configured local search window rather than a distant city such as Faisalabad, and among Google's matched local predictions more popular places should appear before less popular ones when Google supplies rating-count data. Each field has its own map-pin button. Set-on-map mode shows a visible center pin and explicit confirm action. Confirming, tapping blank map space, or dragging a free pin retains the exact Rider-selected coordinate; reverse geocoding may update only the readable address label. Tapping a Google-rendered POI label directly selects that named POI and its canonical coordinate without an automatic camera transition; the selected text must use Google's POI display name rather than substituting the formatted street address.
- [ ] Validate no prediction, denied location permission, poor GPS, lost internet and a stale autocomplete response. No wrong-place selection or hidden provider error.
- [ ] Google content displays over the Google map with appropriate provider attribution. No API key is shipped in committed source.

## 3. Routing and suggested fares

- [ ] Two plausible local locations request one `TRAFFIC_AWARE_OPTIMAL` Google-recommended driving route with `computeAlternativeRoutes=false`. For a typed Place or explicitly tapped Google POI, verify the preview sends/uses the selected Place ID and reaches the appropriate Google access point; for a free pin or current-location endpoint, verify the preview omits Place ID and routes from the exact coordinate. Flutter shows one polyline and one distance/ETA summary; there is no alternative-route selector and no shortest-distance reference request. During materially different traffic conditions, verify duration reflects current traffic.
- [ ] Display road kilometres and estimated journey duration from the same Google-recommended route. Physically sanity-check the route against known local drivability; if Google's own road graph is stale or wrong, record it as a provider-data limitation rather than substituting Haversine or a locally fabricated route. Verify no-route response shows an error rather than a Haversine fallback labeled as a road route.
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

## Driver trip route stages

Physical Android validation remains pending for the Driver route endpoint and map:

- [ ] Deploy the backend from the Driver route branch before testing the updated app.
- [ ] Select a Driver from the Rider app. The assigned Driver sees a Google road route from a fresh device location to the trip pickup.
- [ ] Start trip. The pickup route is replaced by pickup-to-destination, using the server-owned current trip endpoints.
- [ ] Complete or cancel the trip. The driving line disappears, including completed trips still awaiting cash settlement.
- [ ] Restore the app with an assigned or in-progress trip. The correct stage route is requested again.
- [ ] Test GPS permission failure, provider timeout/no route and Retry. No straight line is substituted for a failed driving route.
- [ ] Delay the pickup response, then Start trip or cancel. A late response cannot restore the previous route.
- [ ] Pan or zoom the map, allow unchanged trip polls and resume/refresh the same route. The camera is not repeatedly reset; a new trip or stage can fit its new route.
- [ ] Check the Driver panel extent and scroll/drag handoff with the route and Retry control visible.

## Validation gate

Before each PR merges, use the existing Readiness workflow on its exact head: Go formatting/vet/PostgreSQL tests and Flutter format/build-runner/drift/analyze/full tests. Android map/plugin PRs also require physical build/device checks. Do not call the expanded booking flow pilot-ready until this matrix **and** the existing cash-loop gate are recorded as passed. Any failed critical case becomes the next fix.
