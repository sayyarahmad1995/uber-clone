CREATE TABLE ride_pricing_policies (
    id BIGSERIAL PRIMARY KEY,
    service_code TEXT NOT NULL REFERENCES driver_service_catalog(code),
    currency CHAR(3) NOT NULL,
    version BIGINT NOT NULL CHECK (version > 0),
    base_fare_minor BIGINT NOT NULL CHECK (base_fare_minor >= 0),
    rate_minor_per_km BIGINT NOT NULL CHECK (rate_minor_per_km >= 0),
    rate_minor_per_minute BIGINT NOT NULL CHECK (rate_minor_per_minute >= 0),
    minimum_fare_minor BIGINT NOT NULL CHECK (minimum_fare_minor > 0),
    rounding_increment_minor BIGINT NOT NULL CHECK (rounding_increment_minor > 0),
    effective_from TIMESTAMPTZ NOT NULL,
    effective_until TIMESTAMPTZ,
    is_approved BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT statement_timestamp(),
    created_by TEXT NOT NULL,
    CONSTRAINT ride_pricing_policy_currency_upper CHECK (currency = UPPER(currency)),
    CONSTRAINT ride_pricing_policy_window CHECK (
        effective_until IS NULL OR effective_until > effective_from
    ),
    CONSTRAINT ride_pricing_policy_version_unique UNIQUE (
        service_code,
        currency,
        version
    )
);

CREATE UNIQUE INDEX ride_pricing_policy_one_active
    ON ride_pricing_policies(service_code, currency)
    WHERE is_active = TRUE;

-- Intentionally no Economy/Comfort tariff rows are seeded here. Publishing a
-- pricing policy is an explicit operations approval action.
