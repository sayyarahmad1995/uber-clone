CREATE TABLE ride_driver_candidates (
    ride_request_id UUID PRIMARY KEY REFERENCES ride_requests(id) ON DELETE CASCADE,
    driver_user_id UUID NOT NULL REFERENCES driver_profiles(user_id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX ride_driver_candidates_driver_user_id_idx
    ON ride_driver_candidates(driver_user_id);
