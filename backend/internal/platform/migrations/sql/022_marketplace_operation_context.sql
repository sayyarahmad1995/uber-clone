-- Historical requests and trips keep unknown context; do not infer a vehicle.
ALTER TABLE ride_requests ADD COLUMN service_code TEXT REFERENCES driver_service_catalog(code);
ALTER TABLE ride_requests ALTER COLUMN service_code SET DEFAULT 'economy';
ALTER TABLE ride_offers ADD COLUMN operation_context JSONB;
ALTER TABLE trips ADD COLUMN operation_context JSONB;
-- Old offers retain history but cannot be selected without an explicit new response.
