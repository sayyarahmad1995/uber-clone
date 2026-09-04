# Flutter Rider ride request

## Journey

The Rider client loads existing ride requests when the Rider experience opens. If
there is an active request, it presents the latest backend status and supports
manual refresh and cancellation. Otherwise, the Rider selects pickup and
destination coordinates, proposes a fare in PKR, and creates one request through
the unified marketplace API.

Driver response and Rider offer selection remain later client slices. Creating a
request does not select a booking mode and does not assign a Trip.

## Location and map boundaries

The ride feature owns only application `GeoPoint` values. Device position is read
through the `DeviceLocation` port; Geolocator is the Android adapter. Map tiles are
read through the `MapTiles` configuration port. The initial composition uses an
OpenStreetMap tile surface with attribution for local MVP testing.

Map tile URLs, device-permission types, and package APIs do not enter ride-domain
models or backend contracts. Routing, geocoding, address search, ETA, and service
areas remain deferred. Replacing the map or location provider should change the
composition adapter and presentation, not the marketplace API.

## API and state

- `GET /v1/ride-requests` restores Rider request state after app restart.
- `POST /v1/ride-requests` sends pickup, destination, and `proposed_fare`.
- `GET /v1/ride-requests/{id}` refreshes authoritative request/Trip status.
- `POST /v1/ride-requests/{id}/cancel` cancels through existing shared lifecycle
  behavior.

All calls use the secure bearer session. Fare input is converted from displayed
PKR units to integer minor units before submission. The client sends no legacy
`booking_mode` field.

## Platform behavior

Android declares coarse and fine location permissions. The client asks only when
the Rider chooses current location and reports disabled services or denied access
without preventing manual map selection.
