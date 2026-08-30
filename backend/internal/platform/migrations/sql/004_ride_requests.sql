CREATE TABLE ride_requests (
    id UUID PRIMARY KEY,
    rider_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    pickup_latitude DOUBLE PRECISION NOT NULL,
    pickup_longitude DOUBLE PRECISION NOT NULL,
    destination_latitude DOUBLE PRECISION NOT NULL,
    destination_longitude DOUBLE PRECISION NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ride_requests_pickup_latitude_check CHECK (pickup_latitude BETWEEN -90 AND 90),
    CONSTRAINT ride_requests_pickup_longitude_check CHECK (pickup_longitude BETWEEN -180 AND 180),
    CONSTRAINT ride_requests_destination_latitude_check CHECK (destination_latitude BETWEEN -90 AND 90),
    CONSTRAINT ride_requests_destination_longitude_check CHECK (destination_longitude BETWEEN -180 AND 180),
    CONSTRAINT ride_requests_status_check CHECK (status IN ('requested'))
);

CREATE INDEX ride_requests_rider_user_id_created_at_idx
    ON ride_requests (rider_user_id, created_at DESC);
