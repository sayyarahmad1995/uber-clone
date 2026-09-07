# Uber Clone mobile client

The shared Flutter client starts with the account entry journey: registration,
email verification, login, secure session restoration, Rider entry, optional
Driver capability switching, and logout.

The Rider experience supports map-based pickup/destination selection, current
device location for pickup, PKR fare proposals, request creation, status refresh,
state restoration, and cancellation.

Driver onboarding now follows the ADR-0009 service-application model. Enabling the
Driver capability does not immediately create an operational Driver profile. A new
Driver chooses one service, enters Driver and vehicle details, passes the available
deterministic pre-eligibility checks, reviews the application, and submits it for
review. Pending, approved, and rejected application states are restored from the
backend. Submitted details do not become approved information immediately.

The existing one-vehicle operational Driver path remains temporarily available for
previously onboarded accounts. It preserves online/offline availability and manual
location publication while the multi-vehicle/service operating foundation is built.
Going online still publishes location first; going offline needs no location
permission. New marketplace client work must not deepen the legacy one-vehicle
assumption. See `docs/flutter-driver-readiness.md` and ADR-0009 for the target model.

## Client architecture direction

The client is a cross-platform Flutter app. Product behavior belongs in `lib/`
and must remain independent of Android or iOS runner code. Platform folders are
runners only.

New Rider and Driver workflows should use the shared map-first dashboard model:

- full-screen map as the base layer,
- small floating status widgets for context,
- bottom panels for the active business task,
- theme tokens from `core/theme` instead of ad-hoc styling,
- reusable map/dashboard primitives from `core/maps` and `core/dashboard`.

Dashboard panels follow the accepted interaction contract in
`docs/ADR-0008-dashboard-panel-interaction-contract.md`. Future feature panels use
the shared scaffold and must not redefine its sizing, direct finger tracking,
scrolling, snap animation, or gesture state machine without an explicit product
decision that updates or supersedes the ADR.

Driver/vehicle administration and service applications are distinct from dashboard
operational state. Approved Driver/vehicle information must not be overwritten by a
client edit. The future edit flow submits versioned changes under the governance
rules in ADR-0009; the concrete review, cancellation-window, and appeal implementation
is intentionally deferred.

This keeps the MVP light while avoiding a future rewrite when the product grows
from Rider request creation into Driver marketplace, offer selection, trip
execution, and later administrative surfaces.

## Environments

The Android emulator reaches a backend running on the development computer at
`http://10.0.2.2:8080` by default. Override it for another environment:

```bash
flutter run --dart-define=API_BASE_URL=https://api.example.com
```

For iOS Simulator with the backend on the same Mac:

```bash
flutter run --dart-define=API_BASE_URL=http://localhost:8080
```

For a physical Android or iOS device, use the development computer's LAN IP
address, with both devices on the same network. For example (replace the IP):

```bash
flutter run --dart-define=API_BASE_URL=http://192.168.1.100:8080
```

The backend port must be reachable from the phone; localhost refers to the phone
itself on a physical device.

If the iOS runner is not present locally, generate it from inside `mobile/` with:

```bash
flutter create --platforms=ios .
```

## Verification

Run generated-model updates and verification with:

```bash
dart run build_runner build
flutter analyze
flutter test
```

The client calls only application-owned HTTP endpoints. Provider-specific
identity types and interfaces must remain behind the backend authentication
boundary.
