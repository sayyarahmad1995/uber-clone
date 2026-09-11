# Driver operational readiness milestone

This milestone delivers operating selection and Go Online enforcement. Marketplace expansion and the full ride vertical remain the next milestone.

## Approval evidence

For this MVP, the explicit reviewed vehicle/service enrollment is the authorization evidence for that exact combination. Vehicle existence and legacy active status are insufficient. This does not invent an independent vehicle-verification flag or claim physical inspection occurred. Reviewers remain responsible for their actual review policy.

## Contract

Authenticated Driver-only GET and PUT `/v1/driver/operating-selection` return `selection` (null or vehicle_id, service_code, valid) and `can_change`. PUT accepts vehicle_id and service_code. The server requires ownership, approved enrollment, and an active catalog service. Selection is persisted across sessions. Changes require offline state and no assigned/in-progress trip.

GET `/v1/driver` now exposes approved profiles as well as active ones. Approved profiles may publish location. Going online requires a selected approved combination, Driver capability, fresh location under the existing two-minute policy, and no active trip. Successful Go Online promotes approved status to active. Going offline remains available without selection or fresh location.

Selection and availability acquire the Driver profile lock also used by marketplace assignment. Authorization rows are locked through the write. Migration 021 clears preexisting online flags so old sessions must explicitly select an approved combination. No vehicles or enrollments are automatically selected or granted.

## Acceptance

- Approved Driver opens readiness dashboard without manual database activation.
- Select owned approved vehicle/service; restart the app and confirm it is restored.
- Missing approval or selection prevents going online.
- Device permission/location failure prevents going online; offline remains available.
- Online or active-trip selection changes are rejected on the backend.
- Revoked enrollment or inactive service blocks subsequent Go Online.
- Two sessions racing selection and availability do not bypass the profile lock.

CI runs backend tests against PostgreSQL 17 and Flutter tests. Physical device acceptance still requires the developer's devices. Server deployment must explicitly use its `.env.dev` and `docker-compose.dev.yml` configuration; no server-specific files are committed here.

## Foreground presence

Online availability is temporary foreground presence, not a preference restored at startup. The foreground Driver controller republishes location every 20 seconds while online. Backgrounding requests offline; a new controller session also clears any old online flag before showing the restored selection. Permission-dialog inactive events do not themselves end presence.

Force termination and lost connectivity cannot reliably send offline requests. The API expires online flags after the existing two-minute location freshness window; cleanup runs every ten seconds and reads also reconcile expired presence. Marketplace already rejects stale location immediately at its deadline. Late location heartbeats cannot turn an expired Driver back online. Expiry does not cancel trips or remove the saved operating selection.
