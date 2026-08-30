CREATE TABLE driver_profiles (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    status TEXT NOT NULL,
    is_online BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT driver_profiles_status_check CHECK (status IN ('active'))
);

CREATE TABLE driver_vehicles (
    id UUID PRIMARY KEY,
    driver_user_id UUID NOT NULL UNIQUE REFERENCES driver_profiles(user_id) ON DELETE CASCADE,
    make TEXT NOT NULL,
    model TEXT NOT NULL,
    color TEXT NOT NULL,
    license_plate TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT driver_vehicles_make_not_blank CHECK (BTRIM(make) <> ''),
    CONSTRAINT driver_vehicles_model_not_blank CHECK (BTRIM(model) <> ''),
    CONSTRAINT driver_vehicles_color_not_blank CHECK (BTRIM(color) <> ''),
    CONSTRAINT driver_vehicles_license_plate_not_blank CHECK (BTRIM(license_plate) <> '')
);
