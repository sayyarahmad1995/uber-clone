ALTER TABLE driver_profiles
    ADD COLUMN display_name TEXT;

ALTER TABLE driver_profiles
    ADD CONSTRAINT driver_profiles_display_name_not_blank CHECK (
        display_name IS NULL OR BTRIM(display_name) <> ''
    );

ALTER TABLE driver_vehicles
    ADD COLUMN model_year INTEGER;

ALTER TABLE driver_vehicles
    ADD CONSTRAINT driver_vehicles_model_year_check CHECK (
        model_year IS NULL OR model_year BETWEEN 1886 AND 9999
    );
