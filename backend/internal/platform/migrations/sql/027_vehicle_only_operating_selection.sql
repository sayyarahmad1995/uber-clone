ALTER TABLE driver_operating_selections
    DROP CONSTRAINT driver_operating_selections_vehicle_id_service_code_fkey;

ALTER TABLE driver_operating_selections
    DROP COLUMN service_code;

ALTER TABLE driver_operating_selections
    ADD CONSTRAINT driver_operating_selections_vehicle_id_fkey
    FOREIGN KEY (vehicle_id) REFERENCES driver_vehicles(id) ON DELETE CASCADE;
