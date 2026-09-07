CREATE TABLE driver_service_catalog (
    code TEXT PRIMARY KEY,
    display_name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    minimum_model_year INTEGER,
    implied_service_code TEXT REFERENCES driver_service_catalog(code),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT driver_service_catalog_code_not_blank CHECK (BTRIM(code) <> ''),
    CONSTRAINT driver_service_catalog_display_name_not_blank CHECK (BTRIM(display_name) <> ''),
    CONSTRAINT driver_service_catalog_model_year_check CHECK (
        minimum_model_year IS NULL OR minimum_model_year >= 1886
    )
);

INSERT INTO driver_service_catalog (
    code,
    display_name,
    description,
    minimum_model_year,
    implied_service_code,
    sort_order
) VALUES
    ('economy', 'Economy', 'Vehicle eligibility is confirmed during review.', NULL, NULL, 10),
    ('comfort', 'Comfort', 'Vehicle condition and service eligibility are confirmed during review.', NULL, 'economy', 20)
ON CONFLICT (code) DO NOTHING;

CREATE TABLE driver_onboarding_applications (
    id UUID PRIMARY KEY,
    driver_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    display_name TEXT NOT NULL,
    service_code TEXT NOT NULL REFERENCES driver_service_catalog(code),
    vehicle_make TEXT NOT NULL,
    vehicle_model TEXT NOT NULL,
    vehicle_model_year INTEGER NOT NULL,
    vehicle_color TEXT NOT NULL,
    vehicle_license_plate TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    rejection_reason TEXT,
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    decided_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT driver_onboarding_display_name_not_blank CHECK (BTRIM(display_name) <> ''),
    CONSTRAINT driver_onboarding_vehicle_make_not_blank CHECK (BTRIM(vehicle_make) <> ''),
    CONSTRAINT driver_onboarding_vehicle_model_not_blank CHECK (BTRIM(vehicle_model) <> ''),
    CONSTRAINT driver_onboarding_vehicle_color_not_blank CHECK (BTRIM(vehicle_color) <> ''),
    CONSTRAINT driver_onboarding_vehicle_license_plate_not_blank CHECK (BTRIM(vehicle_license_plate) <> ''),
    CONSTRAINT driver_onboarding_vehicle_model_year_check CHECK (vehicle_model_year >= 1886),
    CONSTRAINT driver_onboarding_status_check CHECK (status IN ('pending', 'approved', 'rejected')),
    CONSTRAINT driver_onboarding_decision_check CHECK (
        (status = 'pending' AND decided_at IS NULL AND rejection_reason IS NULL)
        OR (status = 'approved' AND decided_at IS NOT NULL AND rejection_reason IS NULL)
        OR (status = 'rejected' AND decided_at IS NOT NULL AND BTRIM(COALESCE(rejection_reason, '')) <> '')
    )
);

CREATE UNIQUE INDEX driver_onboarding_one_pending_per_driver
    ON driver_onboarding_applications (driver_user_id)
    WHERE status = 'pending';

CREATE INDEX driver_onboarding_driver_submitted_idx
    ON driver_onboarding_applications (driver_user_id, submitted_at DESC);
