# ADR-0007: Unified Ride Request Marketplace Model

## Status

Accepted for the MVP.

This decision is authoritative for future ride-request, matching, offer, and dispatch work unless it is explicitly superseded by a later architecture decision.

## Context

The product experience is a single Rider ride-request flow. A Rider chooses pickup and destination and proposes the fare they are willing to pay. The Rider does not choose between separate "automatic" and "offers" products.

Earlier implementation slices introduced an internal `booking_mode` distinction (`automatic` / `offers`) so offer-marketplace behavior could be added without breaking the existing automatic-candidate flow. That distinction was useful as an incremental implementation mechanism, but it is not the intended Rider-facing product model.

An earlier revision of this ADR also allowed a Driver who accepted the Rider proposed fare to assign the Trip immediately. Product review showed that this over-optimizes Driver response speed and can remove meaningful Rider choice: a farther Driver or a Driver with a less-preferred vehicle could win simply by responding first.

## Decision

A Ride Request is one marketplace request with:

- pickup;
- destination;
- Rider proposed fare and currency;
- lifecycle status and timestamps;
- application-owned eligibility/distribution state as needed.

Eligible Drivers receive or discover the Ride Request according to application-owned marketplace policy.

For an actionable Ride Request, a Driver submits one actionable commercial response:

1. **Accept the Rider proposed fare.** This creates or updates a Driver offer at exactly the Rider proposed fare. It does not assign the Trip.
2. **Submit a counteroffer.** This creates or updates a Driver offer at a different allowed fare. It does not assign the Trip.

The Rider sees actionable Driver offers and explicitly selects one. Rider selection is the assignment boundary: the selected Driver and Ride Request are revalidated atomically, and a Trip is created only if both remain eligible and available.

An offer equal to the Rider proposed fare should be presented as accepting the Rider's price (for example, a `Your fare` label). This is presentation derived from the offer amount matching the Ride Request proposed fare; it is not a separate booking mode or assignment path.

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
      |                 |
      +--------+--------+
               v
       Rider sees offers
       + Driver details
       + vehicle details
       + pickup distance
               |
               v
        Rider selects one
               |
               v
      atomic assignment
               |
               v
              Trip
```

## Rider choice and MVP presentation

Rider choice is authoritative for marketplace assignment. Driver responses express willingness to serve the Ride Request; they do not reserve or own it.

The Rider-facing offer view should be able to include:

- Driver name and photo when profile support exists;
- vehicle make/model and vehicle photo when available;
- current pickup distance where geographic data is available;
- offered fare;
- a presentation marker when the offered fare equals the Rider proposed fare.

Pickup ETA is explicitly deferred for the MVP. Straight-line geographic distance must not be presented as a trustworthy arrival-time estimate. Routing/traffic-based ETA remains a later capability.

## Marketplace competition invariant

A Driver offer does not reserve the Ride Request and does not reserve the Driver for that Ride Request.

Multiple Drivers may have pending offers on the same Ride Request. A Driver may also have pending offers on multiple Ride Requests because an offer is not an assignment.

Only Rider selection of an actionable Driver offer may create a Trip in the unified marketplace flow.

At Rider selection time the assignment transaction must revalidate that:

- the Ride Request is still open and unassigned;
- the selected offer is still actionable;
- the selected Driver remains operationally eligible;
- the selected Driver has no conflicting active Trip or assignment commitment.

Once assignment commits:

- the Ride Request is no longer actionable for other Drivers;
- competing pending offers become closed/non-actionable;
- the winning Driver's other pending marketplace offers must not permit a second active Trip;
- subsequent selections or Driver responses cannot create another Trip for the Ride Request.

## Geographic matching role

Geographic matching remains valuable, but its business role is marketplace eligibility, distribution, ranking, and Rider decision support rather than defining a separate Rider booking mode.

The current location rules remain reusable:

- Driver must be operationally eligible;
- Driver location must be fresh enough for the marketplace policy;
- straight-line pickup distance may be used for MVP ordering and display;
- deterministic tie-breaking is required;
- active Driver commitments and Trips remain authoritative exclusions at assignment time.

A fixed pickup radius, service-area boundary, routing ETA, and PostGIS remain separate decisions and must not be invented without a concrete product requirement.

## Assignment invariant

Automatic-candidate implementation may coexist temporarily while the code is reconciled, but the unified marketplace path must converge on Rider-selected assignment:

- at most one winning Driver per Ride Request;
- at most one active Trip per Driver;
- assignment-time Driver eligibility is revalidated atomically;
- losing offers/candidates cannot create a second assignment;
- Trip execution is agnostic to how the Driver priced their offer.

## Implementation migration rule

Existing `booking_mode` persistence and automatic-candidate endpoints are considered implementation debt, not a product contract to expand.

Until a focused migration removes or repurposes that distinction:

- do not expose booking-mode choice in the Rider UX or new Rider API contract;
- new Rider Ride Requests require a proposed fare;
- exact-fare Driver responses must not assign the Trip in the unified marketplace path;
- prefer new APIs and read models that fit Rider-selected offers;
- preserve backward compatibility only as necessary to migrate safely;
- reuse existing geographic, offer, candidate, concurrency, and Trip logic where it still represents valid business invariants.

## Consequences

This model keeps the Rider experience simple while preserving Rider choice over fare, proximity, and vehicle. It also makes exact-fare acceptance and counteroffers one coherent Driver-offer concept instead of separate assignment strategies.

The next implementation work should reconcile Rider creation, Driver ride discovery/responses, and Rider offer selection with this decision before adding more features to the legacy automatic-candidate abstraction.
