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
| Ride-state refresh | HTTP polling | Add WebSockets or push only when a concrete latency/scale requirement justifies server-driven updates |
| Maps | `flutter_map` with attributed OpenStreetMap tiles behind application-owned map boundaries | Keep provider replaceable behind `MapTiles`/map adapters |
| Device location | `geolocator` behind the application-owned `DeviceLocation` port | Keep platform/provider APIs outside ride-domain models |

## Backend

| Area | Decision |
|---|---|
| Language | Go |
| Architecture | Modular monolith |
| API | Application-owned HTTP API |
| Database | PostgreSQL |
| Cache / fast ephemeral data | None currently; introduce Redis only when a concrete slice needs cache/ephemeral data |
| Authentication | Provider-neutral identity/authentication boundaries backed by Ory Kratos sessions for the MVP |
| Deployment | Containers |
| Infrastructure | Self-managed server |
| Scaling | Vertical scaling initially |

## Organization

Backend code is organized by business domain. Domain directories own the code needed to implement that domain; there are no global controller/service/repository layers.

The client is organized so that shared application infrastructure and business capabilities remain distinct. Platform-specific code is isolated where needed.

## Technology selection rule

Package-level choices remain revisable until the relevant vertical slice requires them. A tool should be introduced because the current product slice needs it, not because a future enterprise architecture might eventually need it.

The current absence of Redis, WebSockets, push infrastructure, Hydra, Kubernetes, or an event bus is intentional. Those tools require a concrete product or operational need before adoption.
