package locationsearch

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

const (
	MinNamedPlaceSnapRadiusMeters     int64 = 1
	MaxNamedPlaceSnapRadiusMeters     int64 = 50
	DefaultNamedPlaceSnapRadiusMeters int64 = 5
)

var ErrInvalidPolicy = errors.New("invalid location search policy")

type Policy struct {
	NamedPlaceSnapRadiusMeters int64     `json:"named_place_snap_radius_meters"`
	UpdatedAt                  time.Time `json:"updated_at"`
	UpdatedBy                  string    `json:"updated_by"`
}

type PolicyQueryRower interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

type PolicyService interface {
	Load(context.Context) (Policy, error)
	Update(context.Context, Policy, string) (Policy, error)
}

type policyService struct {
	db PolicyQueryRower
}

func NewPolicyService(db PolicyQueryRower) PolicyService {
	return policyService{db: db}
}

func (s policyService) Load(ctx context.Context) (Policy, error) {
	var policy Policy
	if err := s.db.QueryRowContext(ctx, `
		SELECT named_place_snap_radius_meters, updated_at, updated_by
		FROM location_search_policy
		WHERE id = 1
	`).Scan(
		&policy.NamedPlaceSnapRadiusMeters,
		&policy.UpdatedAt,
		&policy.UpdatedBy,
	); err != nil {
		return Policy{}, err
	}
	return policy, nil
}

func (s policyService) Update(ctx context.Context, policy Policy, actor string) (Policy, error) {
	actor = strings.TrimSpace(actor)
	if actor == "" || !ValidPolicy(policy) {
		return Policy{}, ErrInvalidPolicy
	}
	var updated Policy
	if err := s.db.QueryRowContext(ctx, `
		UPDATE location_search_policy
		SET named_place_snap_radius_meters = $1,
		    updated_at = statement_timestamp(),
		    updated_by = $2
		WHERE id = 1
		RETURNING named_place_snap_radius_meters, updated_at, updated_by
	`,
		policy.NamedPlaceSnapRadiusMeters,
		actor,
	).Scan(
		&updated.NamedPlaceSnapRadiusMeters,
		&updated.UpdatedAt,
		&updated.UpdatedBy,
	); err != nil {
		return Policy{}, err
	}
	return updated, nil
}

func ValidPolicy(policy Policy) bool {
	return policy.NamedPlaceSnapRadiusMeters >= MinNamedPlaceSnapRadiusMeters &&
		policy.NamedPlaceSnapRadiusMeters <= MaxNamedPlaceSnapRadiusMeters
}
