CREATE TABLE rider_place_search_policy (
    id SMALLINT PRIMARY KEY CHECK (id = 1),
    autocomplete_radius_meters BIGINT NOT NULL
        CHECK (autocomplete_radius_meters BETWEEN 5000 AND 50000),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT statement_timestamp(),
    updated_by TEXT NOT NULL
);

INSERT INTO rider_place_search_policy (
    id,
    autocomplete_radius_meters,
    updated_by
) VALUES (
    1,
    50000,
    'migration'
);
