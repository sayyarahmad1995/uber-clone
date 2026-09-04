# Uber Clone mobile client

The shared Flutter client starts with the account entry journey: registration,
email verification, login, secure session restoration, Rider entry, optional
Driver capability switching, and logout.

The Rider experience also supports map-based pickup/destination selection, current
device location for pickup, PKR fare proposals, request creation, status refresh,
state restoration, and cancellation.

The Android emulator reaches a backend running on the development computer at
`http://10.0.2.2:8080` by default. Override it for another environment:

```bash
flutter run --dart-define=API_BASE_URL=https://api.example.com
```

Run generated-model updates and verification with:

```bash
dart run build_runner build
flutter analyze
flutter test
```

The client calls only application-owned HTTP endpoints. Provider-specific
identity types and interfaces must remain behind the backend authentication
boundary.
