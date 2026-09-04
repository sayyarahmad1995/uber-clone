# Uber Clone mobile client

The shared Flutter client starts with the account entry journey: registration,
email verification, login, secure session restoration, Rider entry, optional
Driver capability switching, and logout.

The Rider experience supports map-based pickup/destination selection, current
device location for pickup, PKR fare proposals, request creation, status refresh,
state restoration, and cancellation.

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

This keeps the MVP light while avoiding a future rewrite when the product grows
from Rider request creation into Driver marketplace, offer selection, trip
execution, and enterprise-grade surfaces.

## Environments

The Android emulator reaches a backend running on the development computer at
`http://10.0.2.2:8080` by default. Override it for another environment:

```bash
flutter run --dart-define=API_BASE_URL=https://api.example.com
```

For iOS Simulator or a physical device, pass an environment-specific backend URL
instead of relying on the Android emulator default:

```bash
flutter run --dart-define=API_BASE_URL=http://localhost:8080
```

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
