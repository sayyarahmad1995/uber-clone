# Driver route estimates and Google Maps handoff

## Scope

Owner-approved bounded extension on 2026-10-08. Reuse the existing Driver
stage-route response; there is no new server route endpoint, route refresh
timer, fare calculation or Trip transition.

- Assigned: show the estimated driving distance and duration from the fetched
  Driver-location-to-pickup route. Navigate targets the saved Trip pickup.
- In progress: show the **full pickup-to-destination route estimate**, not
  remaining distance/time from the moving Driver. Navigate targets the saved
  Trip destination.
- Estimate values come from the same response as the displayed polyline.
  Duration rounds up to whole minutes; distance displays kilometres to one
  decimal. These are preview estimates, not a live countdown or arrival promise.
- Navigation is an explicit Driver action inside the trip panel. It remains
  available if the HiGO route lookup fails, provided the Trip has valid saved
  coordinates. Loading/error states do not invent an estimate.
- Completion/cancellation removes estimates and navigation controls. A late
  launcher failure cannot show an error for a different Trip/account/stage.
  Repeated taps while that handoff is pending do not launch duplicates.

## Android boundary

The application-owned `DrivingNavigation` port receives only the selected
destination coordinates. A narrow Android method channel builds a fixed
Google Maps directions URL with `api=1`, `travelmode=driving` and
`dir_action=navigate`. Origin is omitted so Google Maps uses its current
device location. Prefer the Google Maps app; if unavailable, open the same
Google Maps website through an available handler. If neither can open it,
show a recoverable error without mutating the Trip.

Google Maps owns external navigation, permissions and any subsequent route
choice. Its current navigation route may differ from the HiGO preview.
Opening it does not start or complete a Trip. No API key or extra HiGO route
request is needed for the handoff. Current platform support is Android.

References:
- [Google Maps URLs](https://developers.google.com/maps/documentation/urls/get-started)
- [Android external intents](https://developer.android.com/training/basics/intents/sending)

## Device acceptance — pending

- [ ] Assigned Trip: read pickup estimate, navigate, and verify the saved pickup
  is selected in Google Maps.
- [ ] Start trip: read full trip estimate and verify navigation targets the
  saved destination from the Driver's current device position.
- [ ] Return from Maps; verify current Trip/recovery and existing location
  presence remain correct. Do not interpret this handoff as a change to
  background active-trip tracking.
- [ ] Complete and cancel: navigation and route estimates disappear.
- [ ] Test Maps installed, Maps disabled/unavailable with website fallback,
  and no usable handler. Failure must be recoverable.
- [ ] On a small device with expanded panel/text scaling and route error,
  navigation remains usable without obscuring status or changing panel gestures.
