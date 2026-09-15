CREATE TABLE marketplace_timing_policy (
    id SMALLINT PRIMARY KEY CHECK (id = 1),
    ride_request_ttl_seconds INTEGER NOT NULL CHECK (ride_request_ttl_seconds BETWEEN 60 AND 600),
    driver_opportunity_ttl_seconds INTEGER NOT NULL CHECK (driver_opportunity_ttl_seconds BETWEEN 10 AND 60),
    offer_decision_ttl_seconds INTEGER NOT NULL CHECK (offer_decision_ttl_seconds BETWEEN 5 AND 60),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by TEXT NOT NULL CHECK (btrim(updated_by) <> '')
);

INSERT INTO marketplace_timing_policy (
    id,
    ride_request_ttl_seconds,
    driver_opportunity_ttl_seconds,
    offer_decision_ttl_seconds,
    updated_by
) VALUES (1, 180, 30, 10, 'migration');
