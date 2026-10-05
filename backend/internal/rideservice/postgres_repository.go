package rideservice

import (
	"context"
	"database/sql"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) PostgresRepository {
	return PostgresRepository{db: db}
}

func (r PostgresRepository) List(ctx context.Context) ([]Option, error) {
	rows, err := r.db.QueryContext(ctx, `
        SELECT c.code, c.display_name, c.description, c.sort_order, c.presentation_token
        FROM driver_service_catalog c
        JOIN ride_pricing_policies p
          ON p.service_code = c.code
         AND p.currency = 'PKR'
         AND p.is_active = TRUE
         AND p.is_approved = TRUE
         AND p.effective_from <= statement_timestamp()
         AND (p.effective_until IS NULL OR p.effective_until > statement_timestamp())
        WHERE c.is_active = TRUE AND c.rider_visible = TRUE
        ORDER BY c.sort_order, c.display_name, c.code
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]Option, 0)
	for rows.Next() {
		var option Option
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
