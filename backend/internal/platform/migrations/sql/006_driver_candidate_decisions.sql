ALTER TABLE ride_driver_candidates
    ADD COLUMN status TEXT NOT NULL DEFAULT 'pending',
    ADD COLUMN decided_at TIMESTAMPTZ;

ALTER TABLE ride_driver_candidates
    ADD CONSTRAINT ride_driver_candidates_status_check
        CHECK (status IN ('pending', 'accepted', 'rejected')),
    ADD CONSTRAINT ride_driver_candidates_decision_time_check
        CHECK (
            (status = 'pending' AND decided_at IS NULL)
            OR (status IN ('accepted', 'rejected') AND decided_at IS NOT NULL)
        );
