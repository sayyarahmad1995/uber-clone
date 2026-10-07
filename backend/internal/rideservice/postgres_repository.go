package rideservice

import (
	"context"
	"database/sql"
)

type PostgresRepository struct {
	db              *sql.DB
	pricingRequired bool
}

func NewPostgresRepository(db *sql.DB) PostgresRepository {
	return PostgresRepository{db: db}
}

func NewPostgresRepositoryWithPricing(db *sql.DB, required bool) PostgresRepository {
	return PostgresRepository{db: db, pricingRequired: required}
}

func (r PostgresRepository) List(ctx context.Context) ([]Option, error) {
	rows, err := r.db.QueryContext(ctx, `
        SELECT code, display_name, description, sort_order, presentation_token
        FROM driver_service_catalog
        WHERE is_active = TRUE AND rider_visible = TRUE
 AND (NOT $1 OR EXISTS (SELECT 1 FROM ride_pricing_policy_current c JOIN ride_pricing_policy_versions p ON p.id=c.current_policy_id WHERE c.service_code=driver_service_catalog.code AND c.currency='PKR'))
        ORDER BY sort_order, display_name, code
    `, r.pricingRequired)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]Option, 0)
	for rows.Next() {
		var option Option
		option.PricingRequired = r.pricingRequired
		if err := rows.Scan(
			&option.Code,
			&option.DisplayName,
			&option.Description,
			&option.DisplayOrder,
			&option.PresentationToken,
		); err != nil {
			return nil, err
		}
		result = append(result, option)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
