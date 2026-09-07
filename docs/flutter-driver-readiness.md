# Flutter Driver setup and readiness

The Rider app bar offers **Become a Driver** to accounts without Driver capability.
It calls `PUT /v1/me/capabilities/driver`, uses the returned account capabilities,
and enters Driver on the same account. Login still defaults to Rider.

ADR-0009 supersedes the original one-vehicle onboarding semantics introduced by the
first Driver-readiness slice. The current implementation is transitional until the
service/vehicle onboarding foundation replaces it.

## Driver onboarding target

The client must distinguish the Driver account capability from ride services.
Economy, Comfort, and similar products are not capabilities.

Initial onboarding is:

```text
Become a Driver
  ↓
Enter Driver details
  ↓
Choose one desired service
  ↓
Review service requirements
  ↓
Enter vehicle details
  ↓
Pre-eligibility check
  ↓
Review application
  ↓
Submit for review
  ↓
Approved / Rejected with reason
```

Initial onboarding applies one vehicle to one service. After a service is approved,
the same vehicle may apply for additional services without re-entering or duplicating
its approved vehicle identity.

The client may perform deterministic pre-eligibility checks supplied by application
policy, but it must not invent verification or approval state. Final service/vehicle
state is backend-authoritative.

## Vehicle and service readiness

A Driver may own multiple vehicles. Vehicle verification and service enrollment are
separate states. A verified vehicle may have one approved service, another pending,
and another technically eligible but not enrolled.

The client must not silently enroll a Driver in a lower service merely because a
stricter service is approved. When the backend explicitly derives eligibility from a
service hierarchy, the UI may offer a lightweight **Add service** action that still
requires Driver intent.

## Availability and operating context

For the MVP, going online uses one backend-confirmed verified vehicle and one approved
service for that vehicle.

```text
Select verified vehicle
  ↓
Select approved service
  ↓
Publish current location
  ↓
Go online
```

If the selected vehicle has exactly one approved service, the service selector may be
omitted. Changing vehicle or service requires going offline first.

Going online obtains device location and publishes it before the online transition.
A permission or publish failure prevents the online write. Going offline never requires
location access. The screen displays only server-confirmed operational state.

Location publishes on explicit actions only. There is no background tracking, timer,
or automatic offline write on navigation/logout in this slice. Backend freshness
remains authoritative.

## Approved information and future edits

Driver and vehicle detail screens show approved information. A future edit flow must
submit a revision instead of immediately replacing approved values.

A local draft can be discarded. After submission, a server-owned cancellation window
allows direct cancellation. Once that deadline expires, withdrawal requires an appeal.
The concrete versioning/review/appeal implementation remains deferred, but current
client work must not add immediate-overwrite semantics that conflict with ADR-0009.

## Dashboard contract

Operational controls stay on the Driver dashboard. Profile, vehicle, service-application,
history, and settings navigation belong outside `RideDashboardScaffold`. ADR-0008 panel
geometry, thresholds, animation, pointer ownership, and scrolling remain unchanged.

## Next physical-device gate

1. Sign in as Rider and choose Become a Driver.
2. Enter Driver details, choose one service, and enter the first vehicle.
3. Verify pre-eligibility feedback and application review before submission.
4. Restore pending/approved/rejected application state from the backend.
5. After approval, verify the vehicle appears with the approved service.
6. Add a second service for the same vehicle without re-registering it.
7. Add a second vehicle independently.
8. Select only a verified vehicle and approved service for going online.
9. Go online with location permission granted; verify location publication precedes availability.
10. Go offline without requiring location permission.
11. Verify Rider/Driver dashboard gesture behavior still satisfies ADR-0008.

See [ADR-0009: Driver Service and Vehicle Eligibility Model](ADR-0009-driver-service-vehicle-eligibility.md).
