CREATE TABLE driver_locations (
    driver_user_id UUID PRIMARY KEY REFERENCES driver_profiles(user_id) ON DELETE CASCADE,
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT driver_locations_latitude_check CHECK (latitude >= -90 AND latitude <= 90),
    CONSTRAINT driver_locations_longitude_check CHECK (longitude >= -180 AND longitude <= 180)
);
