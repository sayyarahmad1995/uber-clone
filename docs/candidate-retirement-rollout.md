# Candidate retirement rollout

Migration `016_retire_candidate_schema.sql` removes `ride_driver_candidates` and
`ride_requests.booking_mode`. Migrations 001–015 stay unchanged so both existing
databases and new installations reach the same schema.

## Behavior and data preservation

- Trips remain the durable assignment and lifecycle record. Existing Trips and
  offers are not rewritten.
- Driver discovery and offer submission exclude Drivers with an assigned or
  in-progress Trip. Pending offers do not reserve a Driver.
- Completion and cancellation free the Driver through Trip status. The existing
  unique active-Trip index and assignment transaction prevent double assignment.
- Historical rides with neither a proposed fare nor currency remain readable.
  They are excluded from the marketplace, and prices are never invented.
  New requests require a valid fare through the application service.
- The database allows either a null fare/currency pair for legacy compatibility
  or a positive fare up to 1,000,000,000,000 minor units with a three-letter
  uppercase currency. Half-populated pairs are rejected.
- Pending candidates are discarded with the retired table, without assignment
  or conversion into offers. Candidate-only history is removed; retain a backup
  if that historical implementation data is needed.
- An accepted, unreleased candidate without a matching Trip stops migration.
  Matching requires the same ride, Rider, and Driver. This also catches a Trip
  assigned to a different Driver. Reconcile the source data using verified
  assignment records; do not invent Trips or mark commitments released merely
  to bypass the check.

## Deployment

1. Back up the database and retain the pre-upgrade application version.
2. Stop all old API instances and any database writers. This is a coordinated
   upgrade, not a rolling deployment: old code references the removed schema.
3. Deploy the new API. Its startup migration runner applies each pending
   migration in a transaction. If the safeguard rejects the upgrade, migration
   016 rolls back without a migration-ledger entry; correct the inconsistent
   records before retrying.
4. Check `/health` and `/ready`, then smoke-test Driver discovery, fare response,
   Rider selection, Trip execution, cancellation, and history.

There is no automatic down migration. Recovery requires the database backup and
the matching old application version, with an explicit plan for any writes made
after upgrade. Do not run the old binary against the upgraded schema.

## Verification

Use an isolated PostgreSQL database ending in `_test`, set `TEST_DATABASE_URL`,
and run from `backend`:

```bash
go test -p 1 ./...
go vet ./...
```

Integration coverage includes fresh installation, populated legacy upgrade,
repeat application, preserved Trips/offers, nullable historical fares, invalid
fare rejection, transactional refusal of missing/mismatched assignments,
concurrent Rider selection, Driver availability after completion/cancellation,
and the full discovery/offer/selection/cancellation journey.

The migration tests create and remove unique schemas inside that test database.
No application database should be used for these tests.
