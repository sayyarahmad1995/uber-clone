DROP INDEX ride_driver_candidates_active_driver_idx;

CREATE UNIQUE INDEX ride_driver_candidates_active_driver_idx
    ON ride_driver_candidates(driver_user_id)
    WHERE status IN ('pending', 'accepted')
      AND released_at IS NULL;
