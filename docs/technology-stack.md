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
| Real-time | WebSocket | Use only where live updates are required |
| Maps | TBD | Provider to be selected before map-dependent slice |
| Device location | TBD | Provider/package to be selected before location-dependent slice |

## Backend

| Area | Decision |
|---|---|
| Language | Go |
| Architecture | Modular monolith |
| API | HTTP API; concrete API style finalized with first backend slice |
| Database | PostgreSQL |
| Cache / fast ephemeral data | Redis |
| Authentication | External OIDC provider |
| Deployment | Containers |
| Infrastructure | Self-managed server |
| Scaling | Vertical scaling initially |

## Organization

Backend code is organized by business domain. Domain directories own the code needed to implement that domain; there are no global controller/service/repository layers.

The client is organized so that shared application infrastructure and business capabilities remain distinct. Platform-specific code is isolated where needed.

## Technology selection rule

Package-level choices remain revisable until the relevant vertical slice requires them. A tool should be introduced because the current product slice needs it, not because a future enterprise architecture might eventually need it.
