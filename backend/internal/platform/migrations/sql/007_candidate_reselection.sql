ALTER TABLE ride_driver_candidates
    DROP CONSTRAINT ride_driver_candidates_pkey;

ALTER TABLE ride_driver_candidates
    ADD CONSTRAINT ride_driver_candidates_pkey
        PRIMARY KEY (ride_request_id, driver_user_id);

CREATE UNIQUE INDEX ride_driver_candidates_active_ride_idx
    ON ride_driver_candidates(ride_request_id)
    WHERE status IN ('pending', 'accepted');
