ALTER TABLE ride_offers
    ADD COLUMN status TEXT NOT NULL DEFAULT 'pending',
    ADD COLUMN decided_at TIMESTAMPTZ;

ALTER TABLE ride_offers
    ADD CONSTRAINT ride_offers_status_check CHECK (status IN ('pending', 'accepted', 'rejected', 'closed')),
    ADD CONSTRAINT ride_offers_decision_time_check CHECK (
        (status = 'pending' AND decided_at IS NULL)
        OR
        (status IN ('accepted', 'rejected', 'closed') AND decided_at IS NOT NULL)
    );

CREATE UNIQUE INDEX ride_offers_one_accepted_per_ride_idx
    ON ride_offers (ride_request_id)
    WHERE status = 'accepted';

CREATE UNIQUE INDEX trips_active_driver_idx
    ON trips (driver_user_id)
    WHERE status IN ('assigned', 'in_progress');
