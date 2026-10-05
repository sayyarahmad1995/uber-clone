package pricing

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

type Draft struct {
	ServiceCode        string
	Currency           string
	BaseFareMinor      int64
	RateMinorPerKM     int64
	RateMinorPerMinute int64
	MinimumFareMinor   int64
	RoundingIncrement  int64
}

func (d Draft) Valid() bool {
	service := strings.ToLower(strings.TrimSpace(d.ServiceCode))
	currency := strings.ToUpper(strings.TrimSpace(d.Currency))
	return service != "" &&
		len(currency) == 3 &&
		d.BaseFareMinor >= 0 &&
		d.RateMinorPerKM >= 0 &&
		d.RateMinorPerMinute >= 0 &&
		d.MinimumFareMinor > 0 &&
		d.RoundingIncrement > 0
}

type PolicyManager interface {
	Publish(context.Context, Draft, string) (Policy, error)
	ListActive(context.Context) ([]Policy, error)
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) PostgresRepository {
	return PostgresRepository{db: db}
}

func (r PostgresRepository) ActivePolicy(
	ctx context.Context,
	serviceCode string,
	currency string,
	at time.Time,
) (Policy, error) {
	serviceCode = strings.ToLower(strings.TrimSpace(serviceCode))
	currency = strings.ToUpper(strings.TrimSpace(currency))

	var policy Policy
	var effectiveUntil sql.NullTime
	err := r.db.QueryRowContext(ctx, `
		SELECT
			service_code,
			currency,
			version,
			base_fare_minor,
			rate_minor_per_km,
			rate_minor_per_minute,
			minimum_fare_minor,
			rounding_increment_minor,
			effective_from,
			effective_until,
			is_approved,
			is_active
		FROM ride_pricing_policies
		WHERE service_code = $1
		  AND currency = $2
		  AND is_active = TRUE
		  AND is_approved = TRUE
		  AND effective_from <= $3
		  AND (effective_until IS NULL OR effective_until > $3)
	`, serviceCode, currency, at).Scan(
		&policy.ServiceCode,
		&policy.Currency,
		&policy.Version,
		&policy.BaseFareMinor,
		&policy.RateMinorPerKM,
		&policy.RateMinorPerMinute,
		&policy.MinimumFareMinor,
		&policy.RoundingIncrement,
		&policy.EffectiveFrom,
		&effectiveUntil,
		&policy.Approved,
		&policy.Active,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Policy{}, ErrPolicyUnavailable
	}
	if err != nil {
		return Policy{}, err
	}
	if effectiveUntil.Valid {
		policy.EffectiveUntil = &effectiveUntil.Time
	}
	return policy, nil
}

func (r PostgresRepository) Publish(
	ctx context.Context,
	draft Draft,
	actor string,
) (Policy, error) {
	draft.ServiceCode = strings.ToLower(strings.TrimSpace(draft.ServiceCode))
	draft.Currency = strings.ToUpper(strings.TrimSpace(draft.Currency))
	actor = strings.TrimSpace(actor)
	if !draft.Valid() || actor == "" {
		return Policy{}, ErrPolicyInvalid
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Policy{}, err
	}
	defer func() { _ = tx.Rollback() }()

	var serviceCode string
	if err := tx.QueryRowContext(ctx, `
		SELECT code
		FROM driver_service_catalog
		WHERE code = $1
		  AND is_active = TRUE
		  AND rider_visible = TRUE
		FOR UPDATE
	`, draft.ServiceCode).Scan(&serviceCode); errors.Is(err, sql.ErrNoRows) {
		return Policy{}, ErrPolicyInvalid
	} else if err != nil {
		return Policy{}, err
	}

	now := time.Now().UTC()
	if _, err := tx.ExecContext(ctx, `
		UPDATE ride_pricing_policies
		SET is_active = FALSE,
		    effective_until = $3
		WHERE service_code = $1
		  AND currency = $2
		  AND is_active = TRUE
	`, draft.ServiceCode, draft.Currency, now); err != nil {
		return Policy{}, err
	}

	var version int64
	if err := tx.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(version), 0) + 1
		FROM ride_pricing_policies
		WHERE service_code = $1 AND currency = $2
	`, draft.ServiceCode, draft.Currency).Scan(&version); err != nil {
		return Policy{}, err
	}

	var policy Policy
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO ride_pricing_policies (
			service_code,
			currency,
			version,
			base_fare_minor,
			rate_minor_per_km,
			rate_minor_per_minute,
			minimum_fare_minor,
			rounding_increment_minor,
			effective_from,
			is_approved,
			is_active,
			created_by
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,TRUE,TRUE,$10)
		RETURNING
			service_code,
			currency,
			version,
			base_fare_minor,
			rate_minor_per_km,
			rate_minor_per_minute,
			minimum_fare_minor,
			rounding_increment_minor,
			effective_from,
			is_approved,
			is_active
	`,
		draft.ServiceCode,
		draft.Currency,
		version,
		draft.BaseFareMinor,
		draft.RateMinorPerKM,
		draft.RateMinorPerMinute,
		draft.MinimumFareMinor,
		draft.RoundingIncrement,
		now,
		actor,
	).Scan(
		&policy.ServiceCode,
		&policy.Currency,
		&policy.Version,
		&policy.BaseFareMinor,
		&policy.RateMinorPerKM,
		&policy.RateMinorPerMinute,
		&policy.MinimumFareMinor,
		&policy.RoundingIncrement,
		&policy.EffectiveFrom,
		&policy.Approved,
		&policy.Active,
	); err != nil {
		return Policy{}, err
	}

	if err := tx.Commit(); err != nil {
		return Policy{}, err
	}
	return policy, nil
}

func (r PostgresRepository) ListActive(ctx context.Context) ([]Policy, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			service_code,
			currency,
			version,
			base_fare_minor,
			rate_minor_per_km,
			rate_minor_per_minute,
			minimum_fare_minor,
			rounding_increment_minor,
			effective_from,
			effective_until,
			is_approved,
			is_active
		FROM ride_pricing_policies
		WHERE is_active = TRUE
		ORDER BY service_code, currency
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	policies := make([]Policy, 0)
	for rows.Next() {
		var policy Policy
		var effectiveUntil sql.NullTime
		if err := rows.Scan(
			&policy.ServiceCode,
			&policy.Currency,
			&policy.Version,
			&policy.BaseFareMinor,
			&policy.RateMinorPerKM,
			&policy.RateMinorPerMinute,
			&policy.MinimumFareMinor,
			&policy.RoundingIncrement,
			&policy.EffectiveFrom,
			&effectiveUntil,
			&policy.Approved,
			&policy.Active,
		); err != nil {
			return nil, err
		}
		if effectiveUntil.Valid {
			policy.EffectiveUntil = &effectiveUntil.Time
		}
		policies = append(policies, policy)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return policies, nil
}
