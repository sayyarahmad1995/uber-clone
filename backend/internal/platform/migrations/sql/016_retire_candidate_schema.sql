-- Run with old API processes stopped: they still reference the retired schema.
-- Preserve Trips as the durable assignment history. Never silently discard an
-- accepted, unreleased commitment that has not been represented by a Trip.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM ride_driver_candidates c
        JOIN ride_requests r ON r.id = c.ride_request_id
        WHERE c.status = 'accepted' AND c.released_at IS NULL
          AND NOT EXISTS (
              SELECT 1 FROM trips t
              WHERE t.ride_request_id = c.ride_request_id
                AND t.driver_user_id = c.driver_user_id
                AND t.rider_user_id = r.rider_user_id
          )
    ) THEN
        RAISE EXCEPTION 'cannot retire candidates: reconcile accepted unreleased candidates with matching Trips first';
    END IF;
END $$;

ALTER TABLE ride_requests
    DROP CONSTRAINT ride_requests_market_fare_check,
    DROP CONSTRAINT ride_requests_booking_mode_check,
    DROP COLUMN booking_mode;

-- Legacy rides may have no recorded fare. Keep them readable without inventing
-- prices; marketplace queries exclude them and new requests require a fare.
-- Explicit null checks prevent PostgreSQL CHECK's UNKNOWN result from allowing
-- a half-populated fare.
ALTER TABLE ride_requests
    ADD CONSTRAINT ride_requests_market_fare_check CHECK (
        (proposed_fare_minor IS NULL AND currency IS NULL)
        OR
        (proposed_fare_minor IS NOT NULL AND currency IS NOT NULL
         AND proposed_fare_minor > 0 AND proposed_fare_minor <= 1000000000000
         AND currency ~ '^[A-Z]{3}$')
    );

-- Pending candidates do not become offers or Trips. Retirement releases that
-- obsolete reservation; only an existing Trip can keep a Driver busy.
DROP TABLE ride_driver_candidates;
