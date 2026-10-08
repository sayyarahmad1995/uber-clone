package ride

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/marketplace"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/pricing"
)

func (r PostgresRepository) createPriced(ctx context.Context, user uuid.UUID, input CreateInput) (Request, error) {
	if !r.pricingConfigured {
		return Request{}, pricing.ErrUnavailable
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Request{}, err
	}
	defer tx.Rollback()
	result, err := r.createPricedInTx(ctx, tx, user, input)
	if err != nil {
		return Request{}, err
	}
	if err := tx.Commit(); err != nil {
		return Request{}, err
	}
	return result, nil
}
func (r PostgresRepository) createPricedInTx(ctx context.Context, tx *sql.Tx, user uuid.UUID, input CreateInput) (Request, error) {
	if input.ProposedFare == nil || input.ProposedFare.Currency != "PKR" {
		return Request{}, pricing.ErrInvalidPolicy
	}
	expected, err := uuid.Parse(input.PricingPolicyVersion)
	if err != nil {
		return Request{}, pricing.ErrInvalidPolicy
	}
	policy, err := pricing.LockCurrent(ctx, tx, input.ServiceCode, "PKR")
	if err != nil {
		return Request{}, err
	}
	if expected != policy.ID {
		return Request{}, pricing.ErrPolicyChanged
	}
	timing, err := marketplace.LoadTimingPolicy(ctx, tx)
	if err != nil {
		return Request{}, err
	}
	var result Request
	result.ID = uuid.New()
	result.ServiceCode = input.ServiceCode
	result.RiderUserID = user
	result.Pickup = input.Pickup
	result.Destination = input.Destination
	// Copy the proposal rather than retaining a mutable input pointer.
	result.ProposedFare = &Money{AmountMinor: input.ProposedFare.AmountMinor, Currency: input.ProposedFare.Currency}
	result.Status = StatusRequested
	err = tx.QueryRowContext(ctx, "INSERT INTO ride_requests(id,rider_user_id,pickup_latitude,pickup_longitude,destination_latitude,destination_longitude,proposed_fare_minor,currency,status,service_code,expires_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,statement_timestamp()+($11*INTERVAL '1 second')) RETURNING created_at,expires_at", result.ID, user, input.Pickup.Latitude, input.Pickup.Longitude, input.Destination.Latitude, input.Destination.Longitude, input.ProposedFare.AmountMinor, input.ProposedFare.Currency, StatusRequested, input.ServiceCode, timing.RideRequestTTLSeconds).Scan(&result.CreatedAt, &result.ExpiresAt)
	if err != nil {
		return Request{}, err
	}
	snapshot := PricingSnapshot{PolicyID: policy.ID, Version: policy.Version, ServiceCode: policy.ServiceCode, Currency: policy.Currency, BaseFareMinor: policy.BaseFareMinor, RateMinorPerKm: policy.RateMinorPerKm, RateMinorPerMinute: policy.RateMinorPerMinute, MinimumFareMinor: policy.MinimumFareMinor, RoundingIncrementMinor: policy.RoundingIncrementMinor, CalculationRule: policy.CalculationRule}
	err = tx.QueryRowContext(ctx, "INSERT INTO ride_request_pricing_snapshots(ride_request_id,policy_id,service_code,currency,version,base_fare_minor,rate_minor_per_km,rate_minor_per_minute,minimum_fare_minor,rounding_increment_minor,calculation_rule) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING snapshot_at", result.ID, snapshot.PolicyID, snapshot.ServiceCode, snapshot.Currency, snapshot.Version, snapshot.BaseFareMinor, snapshot.RateMinorPerKm, snapshot.RateMinorPerMinute, snapshot.MinimumFareMinor, snapshot.RoundingIncrementMinor, snapshot.CalculationRule).Scan(&snapshot.SnapshotAt)
	if err != nil {
		return Request{}, err
	}
	result.PricingSnapshot = &snapshot
	return result, nil
}
