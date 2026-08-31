ALTER TABLE ride_driver_candidates
    ADD COLUMN released_at TIMESTAMPTZ;

DROP INDEX ride_driver_candidates_active_driver_idx;

CREATE UNIQUE INDEX ride_driver_candidates_active_driver_idx
    ON ride_driver_candidates(driver_user_id)
    WHERE status IN ('pending', 'accepted') AND released_at IS NULL;

CREATE TABLE trips (
    ride_request_id UUID PRIMARY KEY REFERENCES ride_requests(id) ON DELETE RESTRICT,
    rider_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    driver_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    status TEXT NOT NULL DEFAULT 'assigned',
    assigned_at TIMESTAMPTZ NOT NULL,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    CONSTRAINT trips_status_check CHECK (status IN ('assigned', 'in_progress', 'completed')),
    CONSTRAINT trips_driver_not_rider_check CHECK (driver_user_id <> rider_user_id),
    CONSTRAINT trips_started_at_check CHECK (status = 'assigned' OR started_at IS NOT NULL),
    CONSTRAINT trips_completed_at_check CHECK (status <> 'completed' OR completed_at IS NOT NULL)
);

CREATE INDEX trips_rider_user_id_assigned_at_idx
    ON trips (rider_user_id, assigned_at DESC);

CREATE INDEX trips_driver_user_id_assigned_at_idx
    ON trips (driver_user_id, assigned_at DESC);

INSERT INTO trips (
    ride_request_id,
    rider_user_id,
    driver_user_id,
    status,
    assigned_at
)
SELECT
    c.ride_request_id,
    r.rider_user_id,
    c.driver_user_id,
    'assigned',
    COALESCE(c.decided_at, c.created_at)
FROM ride_driver_candidates c
JOIN ride_requests r ON r.id = c.ride_request_id
WHERE c.status = 'accepted'
  AND c.released_at IS NULL
ON CONFLICT (ride_request_id) DO NOTHING;
