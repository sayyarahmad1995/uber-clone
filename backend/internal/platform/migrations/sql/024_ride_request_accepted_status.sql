ALTER TABLE ride_requests
    DROP CONSTRAINT ride_requests_status_check,
    DROP CONSTRAINT ride_requests_cancellation_check;

ALTER TABLE ride_requests
    ADD CONSTRAINT ride_requests_status_check CHECK (status IN ('requested', 'accepted', 'cancelled')),
    ADD CONSTRAINT ride_requests_cancellation_check CHECK (
        (status IN ('requested', 'accepted') AND cancelled_at IS NULL AND cancelled_by IS NULL)
        OR
        (status = 'cancelled' AND cancelled_at IS NOT NULL AND cancelled_by IN ('rider', 'driver'))
    );

UPDATE ride_requests rr
SET status = 'accepted'
WHERE status = 'requested'
  AND EXISTS (
      SELECT 1
      FROM trips t
      WHERE t.ride_request_id = rr.id
        AND t.status IN ('assigned', 'in_progress', 'completed')
  );
