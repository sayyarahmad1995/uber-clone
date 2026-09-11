# Rider ↔ Driver ride milestone

This milestone follows Driver operational readiness (PR #79) and ADR-0007/0009.
It connects the existing marketplace and trip lifecycle to both mobile dashboards.
Device acceptance is required before merging.

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

The mobile client refreshes every five seconds while foregrounded, serializes
commands with refreshes, and reloads after a failed command response. User
commands wait for an in-flight refresh rather than being silently dropped.
Network errors remain visible; the UI does not manufacture success. Driver
location is published every twenty seconds while online or on an active
foreground trip. Closing the app still ends marketplace presence; it does not
cancel an assignment. Rider location display drops stale/future samples.
Distances are not ETAs.

## Rollout

Deploy backend and mobile from `feat/rider-driver-ride-milestone` together.
Migration 022 adds request service and nullable offer/trip snapshots. Historical
requests keep unknown service and cannot enter service-specific discovery; make
a new request for acceptance testing. Historical trips remain readable.
New requests omitting service use Economy for compatibility. The current mobile
picker offers Economy and Comfort; the backend rejects inactive catalog services.
The accept-offer HTTP endpoint now requires `{"updated_at":"<displayed offer timestamp>"}`.

On the existing dev server, after switching/pulling the branch:

```sh
docker compose --env-file .env.dev -f docker-compose.yml -f docker-compose.dev.yml up -d --build --no-deps api
```

Keep the existing dev environment files and database volume. Rebuild the mobile
app from the same branch with its existing API configuration.

## Device acceptance

Use separate Rider and approved Driver accounts on two devices (or one device
and a second independent client).

- Create an Economy request; the online Economy Driver discovers it. A Comfort selection must not see that Economy request.
- Send the Rider fare, then a counteroffer. Rider sees the corresponding price and vehicle, and can reject or select. Neither Driver response assigns a trip by itself.
- Select an offer. Both clients show assignment and the same agreed fare/vehicle. Another Driver cannot win that ride.
- Reopen both apps during assignment and during an in-progress trip. Assignment survives; Driver marketplace presence does not silently resume.
- Start and complete. Both clients show terminal history with the original agreed details.
- Repeat with Rider cancellation and Driver cancellation. The other client observes the cancellation and the Driver becomes available for an explicit return online.
- Interrupt networking during a response/selection, restore it, and refresh. There must be no duplicate trip or false success.

Physical-device follow-up on 2026-09-11 found that a Driver counteroffer could be
silently skipped when the five-second foreground poll already owned the shared
ride-flow controller. Accepting the Rider fare was less exposed because it did
not have the counteroffer dialog transition window. The controller now waits for
an active refresh before executing a user command, and regression coverage sends
a counteroffer while a refresh is blocked and verifies that the PUT executes
exactly once with the proposed amount. Repeat the counteroffer device case on the
current branch head before considering acceptance complete.

Automated PostgreSQL coverage checks changed selections, wrong services, revoked
approval, immutable history, revised fares, concurrent assignment, and existing
cancellation/location rules. Mobile tests check state recovery, lost command
responses, polling lifetime, exact fare parsing, counteroffer serialization, and
stale-location presentation.

Payments, routing ETA, push notifications/background trip tracking, media,
additional vehicle registration and approval workflows remain outside this milestone.
