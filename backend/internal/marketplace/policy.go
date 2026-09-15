package marketplace

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

const (
	MinRideRequestTTLSeconds       int64 = 60
	MaxRideRequestTTLSeconds       int64 = 600
	MinDriverOpportunityTTLSeconds int64 = 10
	MaxDriverOpportunityTTLSeconds int64 = 60
	MinOfferDecisionTTLSeconds     int64 = 5
	MaxOfferDecisionTTLSeconds     int64 = 60
)

var ErrInvalidTimingPolicy = errors.New("invalid marketplace timing policy")

type TimingPolicy struct {
	RideRequestTTLSeconds       int64     `json:"ride_request_ttl_seconds"`
	DriverOpportunityTTLSeconds int64     `json:"driver_opportunity_ttl_seconds"`
	OfferDecisionTTLSeconds     int64     `json:"offer_decision_ttl_seconds"`
	UpdatedAt                   time.Time `json:"updated_at"`
	UpdatedBy                   string    `json:"updated_by"`
}

type QueryRower interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func LoadTimingPolicy(ctx context.Context, db QueryRower) (TimingPolicy, error) {
	var policy TimingPolicy
	if err := db.QueryRowContext(ctx, `
		SELECT ride_request_ttl_seconds, driver_opportunity_ttl_seconds,
		       offer_decision_ttl_seconds, updated_at, updated_by
		FROM marketplace_timing_policy
		WHERE id = 1
	`).Scan(
		&policy.RideRequestTTLSeconds,
		&policy.DriverOpportunityTTLSeconds,
		&policy.OfferDecisionTTLSeconds,
		&policy.UpdatedAt,
		&policy.UpdatedBy,
	); err != nil {
		return TimingPolicy{}, err
	}
	return policy, nil
}

func UpdateTimingPolicy(ctx context.Context, db QueryRower, policy TimingPolicy, actor string) (TimingPolicy, error) {
	actor = strings.TrimSpace(actor)
	if actor == "" || !ValidTimingPolicy(policy) {
		return TimingPolicy{}, ErrInvalidTimingPolicy
	}
	var updated TimingPolicy
	if err := db.QueryRowContext(ctx, `
		UPDATE marketplace_timing_policy
		SET ride_request_ttl_seconds = $1,
		    driver_opportunity_ttl_seconds = $2,
		    offer_decision_ttl_seconds = $3,
		    updated_at = statement_timestamp(),
		    updated_by = $4
		WHERE id = 1
		RETURNING ride_request_ttl_seconds, driver_opportunity_ttl_seconds,
		          offer_decision_ttl_seconds, updated_at, updated_by
	`,
		policy.RideRequestTTLSeconds,
		policy.DriverOpportunityTTLSeconds,
		policy.OfferDecisionTTLSeconds,
		actor,
	).Scan(
		&updated.RideRequestTTLSeconds,
		&updated.DriverOpportunityTTLSeconds,
		&updated.OfferDecisionTTLSeconds,
		&updated.UpdatedAt,
		&updated.UpdatedBy,
	); err != nil {
		return TimingPolicy{}, err
	}
	return updated, nil
}

func ValidTimingPolicy(policy TimingPolicy) bool {
	return policy.RideRequestTTLSeconds >= MinRideRequestTTLSeconds &&
		policy.RideRequestTTLSeconds <= MaxRideRequestTTLSeconds &&
		policy.DriverOpportunityTTLSeconds >= MinDriverOpportunityTTLSeconds &&
		policy.DriverOpportunityTTLSeconds <= MaxDriverOpportunityTTLSeconds &&
		policy.OfferDecisionTTLSeconds >= MinOfferDecisionTTLSeconds &&
		policy.OfferDecisionTTLSeconds <= MaxOfferDecisionTTLSeconds
}
