# ADR-0007: Unified Ride Request Marketplace Model

## Status

Accepted for the MVP.

This decision is authoritative for future ride-request, matching, offer, and dispatch work unless it is explicitly superseded by a later architecture decision.

## Context

The product experience is a single Rider ride-request flow. A Rider chooses pickup and destination and proposes the fare they are willing to pay. The Rider does not choose between separate "automatic" and "offers" products.

Earlier implementation slices introduced an internal `booking_mode` distinction (`automatic` / `offers`) so offer-marketplace behavior could be added without breaking the existing automatic-candidate flow. That distinction was useful as an incremental implementation mechanism, but it is not the intended Rider-facing product model.

## Decision

A Ride Request is one marketplace request with:

- pickup;
- destination;
- Rider proposed fare and currency;
- lifecycle status and timestamps;
- application-owned eligibility/distribution state as needed.

Eligible Drivers receive or discover the Ride Request according to application-owned marketplace policy.

For an actionable Ride Request, a Driver has two commercial responses:

1. **Accept the Rider proposed fare.** This may assign the Trip immediately if the Driver and Ride Request are still eligible and unassigned.
2. **Submit a counteroffer.** The Rider may accept or reject that offer. Accepting a counteroffer assigns the Trip atomically if the Driver and Ride Request are still eligible and unassigned.

The Rider is not asked to choose an automatic or offers booking mode.

Conceptually:

```text
Rider creates Ride Request
(pickup + destination + proposed fare)
              |
              v
Application selects eligible Drivers
and distributes/ranks the request
              |
      +-------+-------+
      |               |
      v               v
Driver accepts    Driver counteroffers
Rider fare             |
      |                 v
      |           Rider accepts/rejects
      |                 |
      +--------+--------+
               v
              Trip
```

## Geographic matching role

Geographic matching remains valuable, but its business role is marketplace eligibility, distribution, and ranking rather than defining a separate Rider booking mode.

The current location rules remain reusable:

- Driver must be operationally eligible;
- Driver location must be fresh enough for the marketplace policy;
- straight-line pickup distance may be used for MVP ordering;
- deterministic tie-breaking is required;
- active Driver commitments and Trips remain authoritative exclusions.

A fixed pickup radius, service-area boundary, routing ETA, and PostGIS remain separate decisions and must not be invented without a concrete product requirement.

## Assignment invariant

Automatic-candidate and offer-selection implementation paths may coexist temporarily while the code is reconciled, but they must converge on one authoritative assignment boundary:

- at most one winning Driver per Ride Request;
- at most one active Trip per Driver;
- assignment-time Driver eligibility is revalidated atomically;
- losing offers/candidates cannot create a second assignment;
- Trip execution is agnostic to how the commercial agreement was reached.

## Implementation migration rule

Existing `booking_mode` persistence and endpoints are considered implementation debt, not a product contract to expand.

Until a focused migration removes or repurposes that distinction:

- do not expose booking-mode choice in the Rider UX;
- do not add new Rider-facing behavior that requires choosing `automatic` versus `offers`;
- prefer new APIs and read models that fit the unified marketplace model;
- preserve backward compatibility only as necessary to migrate safely;
- reuse existing geographic, offer, candidate, concurrency, and Trip logic where it still represents valid business invariants.

## Consequences

This model keeps the Rider experience simple while allowing Driver price competition. It also avoids duplicating Ride Request and Trip models for different matching strategies.

The next implementation work should reconcile Driver ride discovery/actionability with this decision before adding more features to the legacy automatic-candidate abstraction.
