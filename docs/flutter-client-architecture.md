# Flutter client architecture

## First client slice

The first Android slice owns application entry: registration and verification,
login, secure session persistence and restoration, account loading, Rider-first
capability selection, and logout. It consumes the application-owned authentication
and `/v1/me` contracts; the client never calls or models Ory APIs.

## Boundaries

- `core` contains configuration, application API models, secure session storage,
  simple capability preference storage, and dependency composition.
- `features/authentication` owns account-entry API access, session state, and UI.
- `features/capabilities` owns Rider/Driver navigation and capability shells.
- Feature code depends on `AuthRepository`, `SessionStore`, and `CapabilityStore`
  interfaces so transport and device storage remain replaceable in tests.

Riverpod composes state and dependencies. GoRouter derives access from the shared
session controller. Dio owns application HTTP transport. Freezed and JSON
serialization represent stable API contracts. The access token and its expiry are
stored in Flutter Secure Storage; the current capability is a non-secret preference.

## Session behavior

At startup, an unexpired stored token is validated by loading `/v1/me`. Missing,
expired, or unauthorized sessions return to application login. A successful login
stores the backend session and enters Rider even when Driver is also available.
The user may switch to Driver only when `/v1/me` lists that capability. Logout
attempts server invalidation, then always clears local session and preference data.

The backend session-extension endpoint is not used automatically in this slice.
ADR-0005 does not define an application refresh-token workflow, so extension can
be added later as an explicit provider-backed session policy without changing
business features.

## Configuration and platform scope

Flutter 3.47.2 and Dart 3.13.2 created the project; exact package resolution is
committed in `pubspec.lock`. `API_BASE_URL` is a compile-time definition and
defaults to the Android emulator host alias `http://10.0.2.2:8080`. Cleartext HTTP
is permitted only in Android debug builds; release builds require HTTPS.

Android is the only generated platform. WebSockets, push notifications, offline
databases, and media providers remain outside the implemented client slices.

## Rider request extension

The Rider request slice introduces application-owned `DeviceLocation` and
`MapTiles` ports. Geolocator and an attributed OpenStreetMap tile surface are the
initial client adapters. Ride-domain models contain only coordinates and money;
provider SDK types and tile configuration remain outside them. See
[Flutter Rider ride request](flutter-rider-request.md).

## Shared dashboard foundation

Rider and Driver surfaces use a map-first dashboard composed from reusable client
primitives. `core/theme` owns visual tokens, `core/maps` owns provider-independent
map rendering, and `core/dashboard` owns the draggable task-panel shell. Feature
screens supply business state, markers, controls, and panel contents without moving
business behavior into those shared presentation components.

Dashboard task panels open in their collapsed state and expand to at most 60
percent of the available dashboard height. Longer workflows scroll inside that
limit so the map remains visible.

Only the panel's top handle resizes it. Gestures within the panel body scroll its
contents without changing the panel extent. Releasing an upward handle drag snaps
the panel to its maximum extent; releasing a downward drag snaps it to its collapsed
extent.

Panel-body scrolling is locked below the maximum extent and unlocked once the panel
finishes expanding. While expanded, a downward gesture scrolls content to the top;
a fresh downward pull from the top past the interaction threshold minimizes the
panel only after the user releases the gesture. Crossing the threshold while the
finger remains down never changes the panel extent, and a cancelled gesture leaves
the panel expanded. While collapsed, an upward body gesture uses the same threshold
and release rule to maximize the panel without enabling content scrolling early.

The Rider dashboard requests location during startup so the map can focus on the
Rider automatically. On a first run, the platform may display its location-permission
prompt as part of this startup behavior. Explicit current-location actions remain
available for refocusing the map and selecting pickup. The Driver dashboard is
currently a visual foundation and does not imply that onboarding, availability,
location publishing, marketplace discovery, or offer behavior has been implemented.
