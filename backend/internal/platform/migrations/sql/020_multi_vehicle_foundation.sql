ALTER TABLE driver_vehicles
    DROP CONSTRAINT IF EXISTS driver_vehicles_driver_user_id_key;

CREATE UNIQUE INDEX driver_vehicles_driver_license_plate_key
    ON driver_vehicles (driver_user_id, license_plate);

CREATE INDEX driver_vehicles_driver_created_idx
    ON driver_vehicles (driver_user_id, created_at, id);
