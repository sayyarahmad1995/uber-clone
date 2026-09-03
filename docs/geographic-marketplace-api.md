# Geographic marketplace API

This slice implements the geographic contract in ADR-0007 within the existing
Go modular monolith. It reuses Driver location and vehicle storage, adds no
migration, and introduces no external provider, background worker, Redis, or
PostGIS dependency. Business packages remain independent of HTTP.

## Driver discovery

`GET /v1/driver/marketplace/ride-requests` keeps its existing response and adds
`pickup_distance_meters` to each request. This is a numeric straight-line
distance; zero is a valid distance, and it is not a route length or an ETA.

Discovery requires Driver capability, an active online profile, a vehicle,
location no older than two minutes and not in the future, and no assigned or
in-progress Trip. An ineligible Driver receives an empty request list. Existing
authentication/capability errors at the HTTP boundary are unchanged.

Eligible requests are ordered by distance, creation time (oldest first), then
request UUID. Ranking precedes the existing 50-request limit. Own requests,
cancelled requests, assigned requests, and legacy requests without fares remain
excluded. There is no fixed pickup radius or same-city restriction.

## Rider offer comparison

`GET /v1/ride-requests/{ride_request_id}/offers` keeps existing offer fields and
adds these fields to each item:

```json
{
  "vehicle": {"make": "Toyota", "model": "Corolla", "color": "White"},
  "pickup_distance_meters": 1234.5,
  "matches_proposed_fare": true,
  "selectable": true
}
```

`vehicle` is null when unavailable. Distance is null for missing, stale, or
future-dated Driver location. The fare-match flag compares both amount and
currency with the Rider's proposal; a client may use it for a “Your fare” label.
It is derived presentation data, not a stored booking mode.

Only pending offers from currently eligible Drivers are selectable. Rejected,
closed, and temporarily unavailable offers stay in the list with
`selectable: false`. Temporary unavailability does not mutate offer status;
refreshing a location or becoming available can restore a pending offer.

Order is selectable first, fare ascending, distance ascending (null last),
creation time ascending, then Driver UUID. This preserves fare comparison while
adding distance as a tie-breaker. The view requires the owning Rider; another
Rider gets the existing not-found response.

The response exposes no raw Driver coordinates, license plate, contact details,
or provider identity. Existing Driver identifiers remain for offer selection.
Names/photos await application profile support. Precise active-Trip location
reads retain their existing ownership boundary.

## Responses and assignment

Driver exact-fare acceptance and counteroffers continue to create/update pending
offers. Both now require the same eligibility and freshness policy. Rejection
by the Rider does not require the Driver to be online or freshly located.

Submission and Rider selection lock the Driver profile, vehicle, capability,
and location records within the existing ride transaction, then recheck using
the database's current statement time after any wait. Active-Trip checks and
the unique active-Trip index remain authoritative for assignment exclusivity.
An ineligible response or selection returns the existing HTTP 409 eligibility
error. Neither the comparison view nor a pending offer reserves a Driver.

Completion and cancellation do not require a fresh location: an already assigned
Trip can finish even if the Driver loses location updates.

Clients must keep publishing Driver location while participating in the
marketplace. Requests that previously worked without location now return an
empty discovery feed or an eligibility error; this is the intended enforcement
of ADR-0007.

## Validation

Use an isolated PostgreSQL database ending in `_test`, set `TEST_DATABASE_URL`,
and run `go test -count=1 -p 1 ./...` and `go vet ./...` from `backend`.

Coverage includes distance ranking before limiting, deterministic ties,
antimeridian/antipodal distances, stale/missing/future locations, offline/busy
Drivers, missing vehicles/capabilities, fare matching, ownership, nullable
comparison fields, location refresh, rejected-offer history, selection waiting
on a location update, and existing concurrent assignment/lifecycle tests.
