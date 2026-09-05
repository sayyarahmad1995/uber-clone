# Flutter Driver setup and readiness

The Rider app bar offers **Become a Driver** to accounts without Driver capability.
It calls `PUT /v1/me/capabilities/driver`, uses the returned account capabilities,
and enters Driver on the same account. Login still defaults to Rider.

The Driver feature owns its API repository, models, operation controller, and
presentation. It reuses the existing application DeviceLocation port and coordinate
model, Dio/session composition, map layer, and ADR-0008 dashboard unchanged.

## Profile lifecycle

`GET /v1/driver` restores profile and online status. Only 404 means onboarding is
needed; other failures show retry rather than an empty form. Setup and editing use
`PUT /v1/driver` with display name, vehicle make/model/year/color/license plate.
Required fields and model year are validated locally and by the existing backend.
Legacy null display names/model years remain readable and can be edited.
Capability enablement is separate from profile creation; a failed setup can be retried.

## Availability and location

Going online obtains device location and calls `PUT /v1/driver/location` before
`PUT /v1/driver/availability`. A permission or publish failure prevents the online
write. Going offline never requires location access. The screen displays only
server-confirmed availability; failed requests retain the last confirmed state
and offer refresh/retry. Going offline does not cancel an existing Trip.

Location publishes on explicit actions only. There is no background tracking,
timer, or automatic offline write on navigation/logout. Leaving during device
lookup prevents subsequent writes; an already-sent request can still complete on
the server. Re-entry restores server state. Updates are serialized to prevent
duplicate taps and conflicting operations.

The server-returned timestamp is shown as the last publication during this visit,
not as proof of current marketplace eligibility. Backend freshness remains two
minutes. This slice does not claim that being online alone makes a Driver eligible
or that a pending offer assigns a Trip. Discovery and offers are the next slice.

## Dashboard contract

Loading, setup/edit, and readiness have distinct panel identities. Transitions
open collapsed; refresh and availability changes preserve readiness identity.
Every panel control is wrapped in `DashboardPanelControl`. Panel geometry,
thresholds, animation, pointer ownership, and scrolling remain in the shared shell.

## Device check

1. Sign in as a Rider, choose Become a Driver, and expand the setup panel.
2. Save complete Driver and vehicle details; verify readiness opens collapsed.
3. Go online with location permission granted; verify the confirmed online state.
4. Update current location, then go offline with location disabled.
5. Reopen Driver and confirm persisted details/status; edit and save details.
6. Check denied permission and unavailable network errors, then retry.

Automated coverage includes API contracts, legacy data, permission and server
failures, duplicate operations, disposal during lookup, capability navigation,
setup submission, and the existing shared dashboard regression suite.

Validation: all 32 Flutter tests pass and `flutter analyze --no-pub` reports no
issues. Physical-device verification of this slice remains pending.
