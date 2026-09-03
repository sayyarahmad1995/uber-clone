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

- Application-owned Driver display name when available; photos await media support;
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
- active Trips remain authoritative exclusions at assignment time; pending offers do not reserve a Driver.

A fixed pickup radius, service-area boundary, routing ETA, and PostGIS remain separate decisions and must not be invented without a concrete product requirement.

### Geographic marketplace slice contract

- Reuse the established two-minute Driver location freshness window. A location
  is fresh when its server timestamp is between the evaluation time minus two
  minutes and the evaluation time, inclusive. Missing, older, or future-dated
  locations do not qualify.
- Discovery, fare responses, and Rider selection require Driver capability,
  an active online profile, a vehicle, fresh location, and no active Trip.
- Rank Driver discovery by straight-line pickup distance, then request creation
  time (oldest first), then request UUID. Apply the feed limit after ranking all
  eligible requests. There is no distance cutoff.
- Rider offer views include current Driver display name and vehicle make/model/model year/color, nullable pickup
  distance in meters, a derived `matches_proposed_fare` flag, and `selectable`.
  Stale/missing locations produce a null distance and a non-selectable offer.
  Offer status/history is retained; temporary unavailability does not reject or
  close a pending offer. A refreshed location may make it selectable again.
- Rider lists place selectable offers first, then fare, distance (unknown last),
  creation time, and Driver UUID. Only the owning Rider may read the view.
- Before assignment, lock the Driver eligibility records and recheck freshness
  and availability after lock acquisition. Read views are snapshots and do not
  reserve a Driver or guarantee a later selection will succeed.
- Before assignment, expose neither raw Driver coordinates nor license plates,
  contact details, or provider identity. Photos remain deferred until
  supported by an application-owned profile capability.
- These fields describe distance, not arrival time; no ETA is derived from them.

## Assignment invariant

The unified marketplace uses Rider-selected assignment:

- at most one winning Driver per Ride Request;
- at most one active Trip per Driver;
- assignment-time Driver eligibility is revalidated atomically;
- losing offers cannot create a second assignment;
- Trip execution is agnostic to how the Driver priced their offer.

## Implementation migration rule

Migration 016 removes `booking_mode` and `ride_driver_candidates`. Historical
migrations remain unchanged. Runtime eligibility and lifecycle code use Trips
and offers without candidate reservations.

- New Rider Ride Requests require a proposed fare.
- Historical rides with no fare remain readable and excluded from marketplace
  discovery/selection; no synthetic fare is backfilled.
- Existing Trips and offers remain unchanged.
- Pending candidates are retired without creating a Trip or offer.
- An accepted, unreleased candidate must have a Trip for the same ride, Rider,
  and Driver; otherwise migration stops for explicit data reconciliation.
- Old API processes must be stopped before applying this schema removal.

See [the rollout guide](candidate-retirement-rollout.md).

## Consequences

This model keeps the Rider experience simple while preserving Rider choice over fare, proximity, and vehicle. It also makes exact-fare acceptance and counteroffers one coherent Driver-offer concept instead of separate assignment strategies.

Rider creation, Driver responses, and Rider selection follow this decision. The
remaining geographic discovery and Rider offer presentation work should build
on this marketplace without restoring an automatic-candidate abstraction.

Migration 017 adds nullable Driver display name and vehicle model year for legacy compatibility. New/re-onboarding requires both; missing presentation data does not disable legacy Drivers. Unknown model years are returned as null. See [Driver public presentation](driver-public-presentation.md).
