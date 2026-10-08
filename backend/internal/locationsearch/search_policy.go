package locationsearch

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

const (
	DefaultAutocompleteRadiusMeters int64 = 50000
	MinAutocompleteRadiusMeters     int64 = 5000
	MaxAutocompleteRadiusMeters     int64 = 50000
)

var ErrInvalidSearchPolicy = errors.New("invalid location search policy")

type SearchPolicy struct {
	AutocompleteRadiusMeters int64     `json:"autocomplete_radius_meters"`
	UpdatedAt                time.Time `json:"updated_at"`
	UpdatedBy                string    `json:"updated_by"`
}

type SearchPolicyService interface {
	Load(context.Context) (SearchPolicy, error)
	Update(context.Context, SearchPolicy, string) (SearchPolicy, error)
}

type searchPolicyService struct {
	db QueryRower
}

type QueryRower interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func NewSearchPolicyService(db QueryRower) SearchPolicyService {
	return searchPolicyService{db: db}
}

func (s searchPolicyService) Load(ctx context.Context) (SearchPolicy, error) {
	return LoadSearchPolicy(ctx, s.db)
}

func (s searchPolicyService) Update(
	ctx context.Context,
	policy SearchPolicy,
	actor string,
) (SearchPolicy, error) {
	return UpdateSearchPolicy(ctx, s.db, policy, actor)
}

func LoadSearchPolicy(ctx context.Context, db QueryRower) (SearchPolicy, error) {
	var policy SearchPolicy
	if err := db.QueryRowContext(ctx, `
		SELECT autocomplete_radius_meters, updated_at, updated_by
		FROM rider_place_search_policy
		WHERE id = 1
	`).Scan(
		&policy.AutocompleteRadiusMeters,
		&policy.UpdatedAt,
		&policy.UpdatedBy,
	); err != nil {
		return SearchPolicy{}, err
	}
	return policy, nil
}

func UpdateSearchPolicy(
	ctx context.Context,
	db QueryRower,
	policy SearchPolicy,
	actor string,
) (SearchPolicy, error) {
	actor = strings.TrimSpace(actor)
	if actor == "" || !ValidSearchPolicy(policy) {
		return SearchPolicy{}, ErrInvalidSearchPolicy
	}

	var updated SearchPolicy
	if err := db.QueryRowContext(ctx, `
		UPDATE rider_place_search_policy
		SET autocomplete_radius_meters = $1,
		    updated_at = statement_timestamp(),
		    updated_by = $2
		WHERE id = 1
		RETURNING autocomplete_radius_meters, updated_at, updated_by
	`,
		policy.AutocompleteRadiusMeters,
		actor,
	).Scan(
		&updated.AutocompleteRadiusMeters,
		&updated.UpdatedAt,
		&updated.UpdatedBy,
	); err != nil {
		return SearchPolicy{}, err
	}
	return updated, nil
}

func ValidSearchPolicy(policy SearchPolicy) bool {
	return policy.AutocompleteRadiusMeters >= MinAutocompleteRadiusMeters &&
		policy.AutocompleteRadiusMeters <= MaxAutocompleteRadiusMeters
}
