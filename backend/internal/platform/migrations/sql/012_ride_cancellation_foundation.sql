ALTER TABLE ride_requests
    DROP CONSTRAINT ride_requests_status_check,
    ADD COLUMN cancelled_at TIMESTAMPTZ,
    ADD COLUMN cancelled_by TEXT;

ALTER TABLE ride_requests
    ADD CONSTRAINT ride_requests_status_check CHECK (status IN ('requested', 'cancelled')),
    ADD CONSTRAINT ride_requests_cancellation_check CHECK (
        (status = 'requested' AND cancelled_at IS NULL AND cancelled_by IS NULL)
        OR
        (status = 'cancelled' AND cancelled_at IS NOT NULL AND cancelled_by IN ('rider', 'driver'))
    );

ALTER TABLE trips
    DROP CONSTRAINT trips_status_check,
    DROP CONSTRAINT trips_started_at_check,
    ADD COLUMN cancelled_at TIMESTAMPTZ;

ALTER TABLE trips
    ADD CONSTRAINT trips_status_check CHECK (status IN ('assigned', 'in_progress', 'completed', 'cancelled')),
    ADD CONSTRAINT trips_started_at_check CHECK (status IN ('assigned', 'cancelled') OR started_at IS NOT NULL),
    ADD CONSTRAINT trips_cancellation_check CHECK (
        (status = 'cancelled' AND cancelled_at IS NOT NULL)
        OR
        (status <> 'cancelled' AND cancelled_at IS NULL)
    );
