ALTER TABLE ride_requests
    ADD COLUMN booking_mode TEXT NOT NULL DEFAULT 'automatic',
    ADD COLUMN proposed_fare_minor BIGINT,
    ADD COLUMN currency TEXT;

ALTER TABLE ride_requests
    ADD CONSTRAINT ride_requests_booking_mode_check CHECK (booking_mode IN ('automatic', 'offers')),
    ADD CONSTRAINT ride_requests_market_fare_check CHECK (
        (booking_mode = 'automatic' AND proposed_fare_minor IS NULL AND currency IS NULL)
        OR
        (booking_mode = 'offers' AND proposed_fare_minor > 0 AND proposed_fare_minor <= 1000000000000 AND currency ~ '^[A-Z]{3}$')
    );

CREATE TABLE ride_offers (
    ride_request_id UUID NOT NULL REFERENCES ride_requests(id) ON DELETE RESTRICT,
    driver_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    amount_minor BIGINT NOT NULL,
    currency TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (ride_request_id, driver_user_id),
    CONSTRAINT ride_offers_amount_check CHECK (amount_minor > 0 AND amount_minor <= 1300000000000),
    CONSTRAINT ride_offers_currency_check CHECK (currency ~ '^[A-Z]{3}$')
);

CREATE INDEX ride_offers_ride_request_amount_idx ON ride_offers (ride_request_id, amount_minor, created_at);
