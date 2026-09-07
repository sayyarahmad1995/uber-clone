ALTER TABLE driver_profiles
    DROP CONSTRAINT driver_profiles_status_check;

ALTER TABLE driver_profiles
    ADD CONSTRAINT driver_profiles_status_check
    CHECK (status IN ('approved', 'active'));

ALTER TABLE driver_onboarding_applications
    ADD COLUMN decided_by TEXT;

CREATE TABLE driver_vehicle_service_enrollments (
    vehicle_id UUID NOT NULL REFERENCES driver_vehicles(id) ON DELETE CASCADE,
    service_code TEXT NOT NULL REFERENCES driver_service_catalog(code),
    approved_at TIMESTAMPTZ NOT NULL,
    approved_by TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (vehicle_id, service_code),
    CONSTRAINT driver_vehicle_service_enrollments_approved_by_not_blank
        CHECK (BTRIM(approved_by) <> '')
);

CREATE INDEX driver_vehicle_service_enrollments_service_idx
    ON driver_vehicle_service_enrollments (service_code, vehicle_id);
