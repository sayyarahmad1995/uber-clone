# Technology Stack

## Client

| Area | MVP decision | Future direction |
|---|---|---|
| Framework | Flutter | Continue shared Flutter application |
| Language | Dart | Continue shared Dart codebase |
| Initial platform | Android | Add iOS later |
| State management | Riverpod | Re-evaluate only if requirements justify change |
| Navigation | GoRouter | Shared across platforms |
| HTTP client | Dio | Shared across platforms |
| Data models | Freezed + json_serializable | Shared across platforms |
| Secure storage | Flutter Secure Storage | Platform-specific implementations hidden behind clear boundaries |
| Simple preferences | SharedPreferences | Keep lightweight unless offline requirements require a database |
| Real-time/update transport | HTTP polling for ride-flow pilot | Introduce another transport only when a concrete live-update requirement justifies it |
| Maps | Google Maps SDK for Android via pinned google_maps_flutter 2.18.2 behind the app-owned `core/maps` boundary; provider-neutral draggable markers support Rider pin adjustment | Route polylines and route-aware camera fit are subsequent slices |
| Device location | geolocator behind DeviceLocation | Preserve through the map migration; existing background Driver presence remains independent | 
| Pickup/destination search | Google Places Autocomplete (New), Place Details (New), Nearby Search (New) for confirmed named-place pin snapping, and Geocoding v4 formatted-address fallback through authenticated app-owned Go APIs; separate pickup/destination sessions, debounced search and stale-result protection | Add richer place semantics only when a booking workflow requires them |

## Backend

| Area | Decision |
|---|---|
| Language | Go |
| Architecture | Modular monolith |
| API | HTTP API; concrete API style finalized with first backend slice |
| Database | PostgreSQL |
| Ride services | Existing driver_service_catalog + authenticated Rider catalog API; Flutter renders server-driven compatible services. Add operator tooling only when justified. |
| Routing | No road route API yet; marketplace uses Haversine. Next slice: Google Routes behind a provider-neutral Go port returning road route, duration and encoded polyline. |
| Fare estimate | Rider-proposed manual fare; no server calculator. Next slice: service/currency-keyed versioned fare policy and advisory estimator in Go using the selected route. |
| Cache / fast ephemeral data | none currently required; Redis deferred until justified |
| Authentication | application-owned HTTP/session contracts with Ory Kratos adapter |
| Deployment | Containers |
| Infrastructure | Self-managed server |
| Scaling | Vertical scaling initially |

## Organization

Backend code is organized by business domain. Domain directories own the code needed to implement that domain; there are no global controller/service/repository layers.

The client is organized so that shared application infrastructure and business capabilities remain distinct. Platform-specific code is isolated where needed.

## Technology selection rule

Package-level choices remain revisable until the relevant vertical slice requires them. A tool should be introduced because the current product slice needs it, not because a future enterprise architecture might eventually need it.
