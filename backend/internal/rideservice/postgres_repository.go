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
        SELECT code, display_name, description, sort_order, presentation_token
        FROM driver_service_catalog
        WHERE is_active = TRUE AND rider_visible = TRUE
        ORDER BY sort_order, display_name, code
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
