# Rider ↔ Driver ride milestone

**Status: completed on 2026-09-12.** PR #80 is merged into `main` at merge commit
`a6fb64975629b595346dfea47448441b8604ef47`. Automated validation and the required
physical-device acceptance passed.

This milestone follows Driver operational readiness (PR #79) and ADR-0007/0009.
It connects the marketplace and trip lifecycle to both mobile dashboards.

## Flow

1. Rider chooses pickup, destination, Economy or Comfort, and a proposed PKR fare.
2. An online Driver with the matching selected, explicitly approved vehicle/service and fresh location sees requests ranked by straight-line pickup distance.
3. Driver accepts the Rider fare or counteroffers within 90–130%. Responses do not reserve the Driver or assign a trip.
4. Rider compares the captured Driver/vehicle details, fare, and current pickup distance and chooses a selectable offer. Rejection is also available.
5. Selection checks the displayed offer's `updated_at` revision and current operating eligibility under transaction locks. A revised fare or changed vehicle cannot silently replace the Rider's choice.
6. Assigned Driver starts, completes, or cancels the trip. Rider can also cancel. Both sides recover server-owned state after reopening and see terminal rides in history.

Offers capture the selected vehicle, service, Driver presentation, and fare. Trips
copy that snapshot at assignment; later profile/selection/enrollment changes do
not rewrite history. License plates stay out of pre-assignment offer responses.
Pending offers without context remain in history but require a new Driver response.
Rejected offers remain persisted for lifecycle/history purposes but are excluded
from the Rider's active comparison; a later Driver resubmission reopens that offer
as pending with the new/current fare.

The mobile client refreshes every five seconds while foregrounded, serializes
commands with refreshes, and reloads after a failed command response. User
commands wait for an in-flight refresh rather than being silently dropped.
Network errors remain visible; the UI does not manufacture success. Driver
location is published every twenty seconds while online or on an active
foreground trip. Closing the app still ends marketplace presence; it does not
cancel an assignment. Rider location display drops stale/future samples.
Distances are not ETAs.

## Rollout

Deploy backend and mobile from `main` at or after PR #80's merge commit
`a6fb64975629b595346dfea47448441b8604ef47`.

Migration 022 adds request service and nullable offer/trip snapshots. Historical
requests keep unknown service and cannot enter service-specific discovery; make
a new request for marketplace testing. Historical trips remain readable.
New requests omitting service use Economy for compatibility. The current mobile
picker offers Economy and Comfort; the backend rejects inactive catalog services.
The accept-offer HTTP endpoint requires
`{"updated_at":"<displayed offer timestamp>"}`.

On the existing dev server after pulling `main`:

```sh
docker compose --env-file .env.dev -f docker-compose.yml -f docker-compose.dev.yml up -d --build --no-deps api
```

Keep the existing dev environment files and database volume. Rebuild the mobile
app from the same `main` revision with its existing API configuration when mobile
code has changed.

## Device acceptance

Physical-device acceptance was completed successfully on 2026-09-12 using
independent Rider and Driver clients.

- [x] Create an Economy request; the online Economy Driver discovers it. A Comfort selection does not see that Economy request.
- [x] Send the Rider fare and a Driver counteroffer. Rider sees the corresponding price and vehicle and can reject or select. Neither Driver response assigns a trip by itself.
- [x] Reject an offer. It disappears from the Rider's active comparison. A fresh response from the same Driver can reappear as a new pending offer with the current fare.
- [x] Select an offer. Both clients show assignment and the same agreed fare/vehicle. Another Driver cannot win that ride.
- [x] Reopen both apps during assignment and during an in-progress trip. Server-owned trip state is restored correctly; Driver marketplace presence does not silently resume.
- [x] Start and complete. Both clients show terminal history with the original agreed details.
- [x] Repeat with Rider cancellation and Driver cancellation. The other client observes the cancellation and the Driver can explicitly return online where applicable.
- [x] Interrupt networking during marketplace/trip interaction, restore it, and refresh. The clients recover the authoritative server state without a duplicate trip or false success.

## Device follow-ups closed during acceptance

A physical-device follow-up found that a Driver counteroffer could be silently
skipped when the five-second foreground poll already owned the shared ride-flow
controller. Accepting the Rider fare was less exposed because it did not have the
counteroffer dialog transition window. The controller now waits for an active
refresh before executing a user command, regression coverage verifies the
counteroffer PUT executes exactly once, and the counteroffer path was retested
successfully on-device.

A later device follow-up found that rejected offers remained visible in the Rider
comparison. The backend comparison query now returns pending offers only while
retaining rejected rows for history. PostgreSQL integration coverage verifies
pending → rejected/hidden → Driver resubmission/pending, and the corrected behavior
was included in the completed device acceptance.

## Automated validation

Readiness validation passed on the final PR head before merge:

- backend: `go test -p 1 ./...` with PostgreSQL integration coverage;
- mobile: `flutter analyze`;
- mobile: `flutter test`.

Coverage includes changed operating selections, wrong services, revoked approval,
immutable historical snapshots, revised fares, concurrent assignment, rejected
offer visibility/resubmission, cancellation/location rules, state recovery, lost
command responses, polling lifetime, counteroffer serialization, and stale-location
presentation.

## Closure

The marketplace-to-trip slice is accepted and complete. PR #80 is merged into
`main`; no remaining acceptance item is open for this milestone.

Payments, routing ETA, push notifications/background trip tracking, media, and
additional vehicle registration/approval workflows remain outside this milestone.
