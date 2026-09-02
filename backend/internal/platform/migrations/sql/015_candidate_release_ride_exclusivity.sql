DROP INDEX ride_driver_candidates_active_ride_idx;

CREATE UNIQUE INDEX ride_driver_candidates_active_ride_idx
    ON ride_driver_candidates(ride_request_id)
    WHERE status IN ('pending', 'accepted')
      AND released_at IS NULL;
