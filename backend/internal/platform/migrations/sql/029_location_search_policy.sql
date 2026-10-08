CREATE TABLE location_search_policy (
    id SMALLINT PRIMARY KEY CHECK (id = 1),
    named_place_snap_radius_meters INTEGER NOT NULL
        CHECK (named_place_snap_radius_meters BETWEEN 1 AND 50),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by TEXT NOT NULL CHECK (btrim(updated_by) <> '')
);

INSERT INTO location_search_policy (
    id,
    named_place_snap_radius_meters,
    updated_by
) VALUES (1, 5, 'migration');
