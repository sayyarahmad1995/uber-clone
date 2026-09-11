CREATE TABLE driver_operating_selections (
    driver_user_id UUID PRIMARY KEY REFERENCES driver_profiles(user_id) ON DELETE CASCADE,
    vehicle_id UUID NOT NULL,
    service_code TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    FOREIGN KEY (vehicle_id, service_code)
        REFERENCES driver_vehicle_service_enrollments(vehicle_id, service_code) ON DELETE CASCADE
);
-- Existing online flags predate enforced operating selection. Require a fresh opt-in.
UPDATE driver_profiles SET is_online = FALSE WHERE is_online;
