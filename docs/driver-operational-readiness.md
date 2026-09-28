# Driver operational readiness milestone

This milestone delivers operating selection and Go Online enforcement. Marketplace expansion and the full ride vertical remain the next milestone.

## Approval evidence

For this MVP, explicit reviewed vehicle/service enrollments are the authorization evidence for the services a vehicle may offer. Vehicle existence and legacy active status are insufficient. This does not invent an independent vehicle-verification flag or claim physical inspection occurred. Reviewers remain responsible for their actual review policy.

## Contract

Authenticated Driver-only GET and PUT `/v1/driver/operating-selection` return `selection` (null or `vehicle_id` and `valid`) and `can_change`. PUT accepts only `vehicle_id`. The server requires ownership and at least one active approved enrollment. Selection is persisted across sessions. Changes require offline state and no assigned/in-progress trip.

GET `/v1/driver` now exposes approved profiles as well as active ones. Approved profiles may publish location. Going online requires a selected eligible vehicle, Driver capability, fresh location under the existing two-minute policy, and no active trip. Marketplace requests match any active approved enrollment on that vehicle. Successful Go Online promotes approved status to active. Going offline remains available without selection or fresh location.

Selection and availability acquire the Driver profile lock also used by marketplace assignment. Authorization rows are locked through the write. Migration 021 clears preexisting online flags so old sessions must explicitly select an approved vehicle. Migration 027 preserves selected vehicle IDs while removing the obsolete selected-service column. No vehicles or enrollments are automatically selected or granted.

## Acceptance

- Approved Driver opens readiness dashboard without manual database activation.
- Select owned approved vehicle/service; restart the app and confirm it is restored.
- Missing approval or selection prevents going online.
- Device permission/location failure prevents going online; offline remains available.
- Online or active-trip selection changes are rejected on the backend.
- Revoked enrollment or inactive service blocks subsequent Go Online.
- Two sessions racing selection and availability do not bypass the profile lock.

CI runs backend tests against PostgreSQL 17 and Flutter tests. Physical device acceptance still requires the developer's devices. Server deployment must explicitly use its `.env.dev` and `docker-compose.dev.yml` configuration; no server-specific files are committed here.

## Renewable presence lease

Online availability is an explicit Driver session backed by a renewable server-side location lease. It is not tied to the Driver screen or Flutter controller lifetime. Minimizing the Android app, locking the screen, navigating away from the Driver screen, or disposing its controller does not request offline. Reopening or resuming reloads the Driver profile and operating selection from the API and reconciles the local publisher with server truth.

On Android, successful Go Online starts one notification-backed foreground service of type `location`. This service is the sole publisher for online marketplace presence, including while the Flutter UI is foregrounded. Every renewal checks the existing secure session, authenticated Driver identity, and server `is_online` state before obtaining a fresh location and publishing it. It does not poll the marketplace or own trip state. Android requires foreground-service and notification declarations; a persistent “Driver is online” notification is expected while the service runs. This milestone has no iOS project, so iOS background presence is unsupported and unverified.

Go Online remains ordered as fresh location publication, server-confirmed online availability, then foreground-service startup. If service startup or its first authenticated location publication fails, the client reconciles toward server-confirmed offline state rather than claiming durable online presence. Explicit Go Offline updates the server first and then stops the service. If stopping the Android service fails after the server confirms offline, its next renewal observes `is_online = false`, requests no GPS fix, publishes nothing, and self-stops. Sign-out stops the service before clearing credentials; missing or changed credentials make any lingering task harmless.

Process termination, force-stop, lost connectivity, revoked location permission, and unexpected service loss cannot reliably send a final offline request. Marketplace rejects a stale location immediately at the existing two-minute deadline. Server cleanup runs every ten seconds and clears stale online flags; reads also reconcile expired presence. Late location heartbeats cannot turn an expired Driver back online. No immediate-offline guarantee is made for process kill.

Marketplace presence and trip lifecycle remain independent. Presence expiry or explicit offline does not cancel or corrupt an assigned/in-progress trip and does not remove the saved operating vehicle. Foreground offer/trip polling remains in the existing Rider and Driver controllers and is not moved into the background service.
