# Pilot Cash Ride End-to-End Acceptance Checklist

## Purpose

This checklist validates that the MVP cash ride loop is ready for a small, controlled pilot test.

It is not a new product roadmap. It is the acceptance boundary for the current Rider ↔ Driver ride-hailing flow after marketplace assignment, Trip execution, cash settlement, and receipt/history work.

A pass means the current Android client, Go API, PostgreSQL schema, and documented MVP flow agree on the same business behavior for a cash ride.

## Scope

Validate one complete ride-hailing loop:

```text
Rider creates request
  ↓
Driver discovers request
  ↓
Driver responds with exact fare or counteroffer
  ↓
Rider selects an offer
  ↓
Trip is assigned
  ↓
Driver starts Trip
  ↓
Driver completes Trip
  ↓
Rider still sees completed/unsettled Trip
  ↓
Driver confirms cash collected
  ↓
Rider and Driver see settled receipt/history state
```

## Non-goals

This checklist does not validate or authorize:

- Stripe, card payments, wallets, payouts, refunds, tips, commissions, promotions, or cancellation fees.
- Courier, Freight, or other future capabilities.
- Full administrator operations beyond the existing narrow Driver onboarding reviewer.
- Advanced dispatch, routing ETA, surge pricing, service-area policy, PostGIS, Redis, MongoDB, or multi-node infrastructure.
- Driver/vehicle approved-information editing, appeals, or revision review flows.

## Preconditions

Use a current checkout that includes the ride-request lifecycle and cash-settlement work through PR #88.

The local stack should be rebuilt and running:

```bash
docker compose up -d --build
```

The API should be healthy:

```bash
curl http://localhost:8080/health
```

Expected response:

```json
{"status":"ok"}
```

Use one Rider account and one Driver-capable account. The Driver must have one approved vehicle/service operating context and must be able to go online for the service being tested.

For local-only test recovery, it is acceptable to clear stale development data. Do not use destructive truncation against pilot or production data.

## Evidence to capture

For every acceptance run, record:

- date/time of the run;
- branch or commit under test;
- Android build under test;
- Rider account identifier;
- Driver account identifier;
- requested service: Economy or Comfort;
- proposed fare and selected offer fare;
- final `ride_requests.id` / `trips.ride_request_id`;
- pass/fail notes for every section below.

## Acceptance matrix

### 1. Rider request creation

Steps:

1. Sign in as the Rider.
2. Enter Rider capability.
3. Select pickup and destination.
4. Select Economy.
5. Enter a proposed fare.
6. Submit the ride request.

Expected result:

- The Rider remains in the active request flow.
- The Rider sees a waiting-for-offers state.
- The database has one new `ride_requests` row with `status = 'requested'`.
- The row includes pickup, destination, service code, proposed fare, and currency.

Repeat the same creation path with Comfort in a separate run.

### 2. Driver online and request discovery

Steps:

1. Sign in as the Driver.
2. Enter Driver capability.
3. Ensure the Driver has a selected approved vehicle/service context.
4. Publish a fresh location by going online.
5. Open the Driver marketplace/request list.

Expected result:

- The Driver can discover only eligible requests for the selected approved service.
- The request card shows Rider fare, pickup, destination, and pickup distance when available.
- The request is not assigned merely because the Driver can see it.

### 3. Exact-fare offer path

Steps:

1. From the Driver request card, tap `Accept Rider fare`.
2. Return to or refresh the Rider view.

Expected result:

- The Driver response creates a pending offer at the Rider proposed fare.
- The Rider sees the offer in the active comparison panel.
- No Trip is assigned yet.
- The Rider can still choose among pending offers.

### 4. Counteroffer path

Steps:

1. Create a fresh request or use another pending request.
2. From the Driver request card, tap `Propose another fare`.
3. Enter a valid counteroffer within the allowed range.
4. Return to or refresh the Rider view.

Expected result:

- The Rider sees the counteroffer.
- The displayed fare is the counteroffer amount, not the original proposed fare.
- The Driver is not assigned until the Rider explicitly selects the offer.

### 5. Offer rejection and resubmission

Steps:

1. Have the Driver submit an offer.
2. Reject the offer from the Rider side.
3. Have the Driver submit a fresh offer while the request remains open and eligible.

Expected result:

- The rejected offer leaves the active Rider comparison.
- The new Driver response reopens as a pending offer.
- The Rider can select the fresh offer.

### 6. Rider offer selection and assignment

Steps:

1. From the Rider comparison panel, choose one Driver offer.
2. Confirm the selection.
3. Refresh both Rider and Driver views.

Expected result:

- Exactly one Trip is created for the ride request.
- The Trip status is `assigned`.
- The parent ride request status becomes `accepted`.
- Competing offers for the selected ride are closed.
- The selected offer fare is preserved as the Trip agreed fare snapshot.
- The Driver sees the assigned Trip as current.
- The Rider sees the assigned Driver/vehicle/service context.

### 7. Trip start

Steps:

1. On the Driver side, tap `Start trip`.
2. Refresh both Rider and Driver views.

Expected result:

- Trip status becomes `in_progress`.
- Rider and Driver both recover the in-progress state after app restart.
- The agreed fare and assigned vehicle/service context remain unchanged.

### 8. Trip completion before settlement

Steps:

1. On the Driver side, tap `Complete trip`.
2. Before tapping cash collection, refresh both Rider and Driver views.

Expected result:

- Trip status becomes `completed`.
- Settlement status remains `unsettled`.
- The Driver sees the completed Trip and the `Confirm cash collected` action.
- The Rider still sees the completed/unsettled Trip instead of losing the active ride context.
- The Driver cancel action is not available after completion.

### 9. Cash collection confirmation

Steps:

1. On the Driver side, tap `Confirm cash collected`.
2. Confirm the dialog.
3. Refresh both Rider and Driver views.

Expected result:

- Settlement status becomes `cash_collected`.
- Settlement method is `cash`.
- `cash_collected_at` is populated.
- `cash_collected_by` equals the assigned Driver user id.
- The Driver current Trip clears.
- The Rider active Trip clears or moves to terminal history/receipt state after refresh.
- Repeating the cash-collected action is idempotent and does not create a duplicate settlement.

### 10. Receipt and history display

Steps:

1. Open Rider recent rides/history.
2. Open Driver Trip history.

Expected result:

Both histories expose receipt-relevant facts from immutable Trip context:

- agreed fare amount and currency;
- service code/name where available;
- Driver/vehicle context captured at assignment;
- Trip status;
- completion time where available;
- settlement status;
- cash collection timestamp when available.

### 11. Database invariants

After the run, inspect the linked request and Trip:

```sql
SELECT rr.id,
       rr.status AS request_status,
       rr.service_code,
       rr.proposed_fare_minor,
       rr.currency,
       t.status AS trip_status,
       t.settlement_status,
       t.settlement_method,
       t.completed_at,
       t.cash_collected_at,
       t.cash_collected_by
FROM ride_requests rr
LEFT JOIN trips t ON t.ride_request_id = rr.id
WHERE rr.id = '<ride_request_id>';
```

Expected final state after cash collection:

```text
request_status = accepted
trip_status = completed
settlement_status = cash_collected
settlement_method = cash
completed_at is not null
cash_collected_at is not null
cash_collected_by = assigned driver user id
```

There must be no final completed Trip whose parent request remains `requested`.

There must be no completed cash-collected Trip without `settlement_method = 'cash'`, `cash_collected_at`, and `cash_collected_by`.

### 12. Cancellation coverage

Run at least one Rider cancellation and one Driver cancellation before Trip completion.

Expected result:

- Cancellation closes the ride flow for both sides.
- Pending offers for the ride are closed.
- The cancellation actor is recorded as Rider or Driver.
- Completed Trips cannot be cancelled.

### 13. Recovery coverage

Repeat selected steps with app restart or temporary network interruption:

- after request creation;
- after offer submission;
- after Rider selection/assignment;
- after Trip start;
- after Trip completion but before cash collection;
- after cash collection.

Expected result:

- Foreground recovery reloads authoritative server state.
- User commands do not silently disappear behind polling refreshes.
- Failed commands reload current state and allow the user to continue from the correct point.

## Pass/fail rule

The cash ride loop passes pilot acceptance only when:

- Economy and Comfort ride-request creation can both be exercised.
- At least one exact-fare path and one counteroffer path are verified.
- Assignment remains Rider-selected.
- Trip execution reaches completed state.
- Completed/unsettled state remains recoverable for Rider and Driver before cash collection.
- Driver cash collection moves settlement to `cash_collected`.
- Rider and Driver histories show receipt-relevant facts.
- Database state matches the expected request/Trip lifecycle.
- Restart or transient-network recovery does not strand the user in an impossible state.

Any failure should become the next implementation slice before starting unrelated product work.

## Next-slice rule

After this checklist passes, the next slice should be selected from actual pilot blockers or the smallest visible receipt/history polish still needed for the cash ride loop.

Do not move to Courier, Freight, sophisticated payments, broad admin tooling, advanced dispatch, multi-database architecture, or infrastructure scaling until the cash ride pilot loop is demonstrably stable.
