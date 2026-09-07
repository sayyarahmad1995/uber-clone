# Driver onboarding reviewer

This is the minimal internal reviewer surface defined by ADR-0010. It exists only to review initial Driver onboarding applications while a broader administration product is out of scope.

## Enable the reviewer

The reviewer routes are not registered unless both credentials are configured.

For Docker Compose, add strong local values to `.env`:

```text
ADMIN_REVIEW_USERNAME=reviewer
ADMIN_REVIEW_PASSWORD=replace-with-a-strong-local-password
```

Then rebuild/restart the API so migration 019 and the environment values are applied.

For a directly executed backend in Windows PowerShell:

```powershell
$env:ADMIN_REVIEW_USERNAME = "reviewer"
$env:ADMIN_REVIEW_PASSWORD = "replace-with-a-strong-local-password"
go run ./cmd/api
```

Do not expose this temporary HTTP Basic reviewer over untrusted plain HTTP. Use it on the local development host, or behind HTTPS on a trusted internal deployment. A future administrator identity system will replace this temporary authentication mechanism.

## Open the reviewer

With the backend on the default port, open:

```text
http://localhost:8080/admin/driver-onboarding
```

The browser will request the configured Basic-auth credentials.

The page lists pending applications. Open one application to inspect:

- Driver user ID and display name;
- selected service;
- vehicle make/model/model year/color/license plate;
- submission time;
- current decision state.

## Approve

Approval is transactional. It:

1. locks the still-pending application;
2. creates the approved Driver profile from the submitted snapshot;
3. creates the approved vehicle from the submitted snapshot;
4. creates the explicit enrollment for the selected service;
5. records `approved`, decision time, and reviewer identity on the application.

Approval does **not** silently enroll an implied lower service. For example, approving Comfort does not automatically enroll Economy.

The new Driver profile is created with status `approved`, not `active`. The Driver therefore cannot use the legacy online/marketplace path yet. The next vehicle/service operating-context slice will define the transition to active operation.

## Reject

Rejection requires a reason. It records the rejection and reviewer identity but creates no Driver, vehicle, or service-enrollment records.

The Driver can refresh/re-enter the mobile Driver workspace and see the rejected application and reason from the normal onboarding API.

## Reviewer API

The same Basic-auth boundary protects the small reviewer API:

```text
GET  /v1/admin/driver-onboarding-applications
GET  /v1/admin/driver-onboarding-applications/{application_id}
POST /v1/admin/driver-onboarding-applications/{application_id}/approve
POST /v1/admin/driver-onboarding-applications/{application_id}/reject
```

Reject JSON body:

```json
{
  "reason": "Vehicle condition does not meet the selected service requirements"
}
```

The server rejects a second decision once an application is no longer pending.

## Physical-device review gate

1. Submit a Driver onboarding application from the phone.
2. Confirm it appears in the reviewer pending list.
3. Approve it and refresh/re-enter the Driver workspace; confirm the application shows approved and does not expose the legacy Go online control.
4. Repeat with a separate account/application and reject with a reason; confirm the reason restores on the phone.
5. Confirm a decided application cannot be decided again.
6. Confirm the reviewer routes disappear when either reviewer credential is absent.
