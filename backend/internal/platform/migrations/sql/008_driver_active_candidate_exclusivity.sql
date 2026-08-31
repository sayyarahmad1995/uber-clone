DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM ride_driver_candidates
        WHERE status IN ('pending', 'accepted')
        GROUP BY driver_user_id
        HAVING COUNT(*) > 1
    ) THEN
        RAISE EXCEPTION 'cannot enforce active Driver candidate exclusivity while duplicate active candidates exist';
    END IF;
END $$;

CREATE UNIQUE INDEX ride_driver_candidates_active_driver_idx
    ON ride_driver_candidates(driver_user_id)
    WHERE status IN ('pending', 'accepted');
