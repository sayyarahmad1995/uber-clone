ALTER TABLE ride_requests
    ADD COLUMN expires_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() + INTERVAL '3 minutes');

UPDATE ride_requests
SET expires_at = created_at + INTERVAL '3 minutes';

ALTER TABLE ride_requests
    DROP CONSTRAINT ride_requests_status_check,
    DROP CONSTRAINT ride_requests_cancellation_check;

ALTER TABLE ride_requests
    ADD CONSTRAINT ride_requests_status_check
        CHECK (status IN ('requested', 'accepted', 'cancelled', 'expired')),
    ADD CONSTRAINT ride_requests_cancellation_check CHECK (
        (status IN ('requested', 'accepted', 'expired') AND cancelled_at IS NULL AND cancelled_by IS NULL)
        OR
        (status = 'cancelled' AND cancelled_at IS NOT NULL AND cancelled_by IN ('rider', 'driver'))
    ),
    ADD CONSTRAINT ride_requests_expiry_time_check CHECK (expires_at > created_at);

ALTER TABLE ride_offers
    ADD COLUMN expires_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() + INTERVAL '10 seconds');

UPDATE ride_offers
SET expires_at = created_at + INTERVAL '10 seconds';

ALTER TABLE ride_offers
    DROP CONSTRAINT ride_offers_status_check,
    DROP CONSTRAINT ride_offers_decision_time_check;

ALTER TABLE ride_offers
    ADD CONSTRAINT ride_offers_status_check
        CHECK (status IN ('pending', 'accepted', 'rejected', 'expired', 'closed')),
    ADD CONSTRAINT ride_offers_decision_time_check CHECK (
        (status = 'pending' AND decided_at IS NULL)
        OR
        (status IN ('accepted', 'rejected', 'expired', 'closed') AND decided_at IS NOT NULL)
    ),
    ADD CONSTRAINT ride_offers_expiry_time_check CHECK (expires_at > created_at);

CREATE TABLE driver_ride_request_opportunities (
    ride_request_id UUID NOT NULL REFERENCES ride_requests(id) ON DELETE CASCADE,
    driver_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'open',
    opened_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    visible_until TIMESTAMPTZ NOT NULL,
    responded_at TIMESTAMPTZ,
    PRIMARY KEY (ride_request_id, driver_user_id),
    CONSTRAINT driver_ride_request_opportunities_status_check CHECK (
        status IN (
            'open',
            'driver_skipped',
            'window_expired',
            'offered',
            'rider_rejected',
            'offer_expired',
            'accepted',
            'closed'
        )
    ),
    CONSTRAINT driver_ride_request_opportunities_time_check CHECK (visible_until > opened_at),
    CONSTRAINT driver_ride_request_opportunities_response_check CHECK (
        (status = 'open' AND responded_at IS NULL)
        OR
        (status <> 'open' AND responded_at IS NOT NULL)
    )
);

CREATE INDEX ride_requests_requested_expiry_idx
    ON ride_requests (expires_at)
    WHERE status = 'requested';

CREATE INDEX ride_offers_pending_expiry_idx
    ON ride_offers (expires_at)
    WHERE status = 'pending';

CREATE INDEX driver_ride_request_opportunities_open_expiry_idx
    ON driver_ride_request_opportunities (visible_until)
    WHERE status = 'open';

CREATE INDEX driver_ride_request_opportunities_driver_status_idx
    ON driver_ride_request_opportunities (driver_user_id, status, visible_until);
