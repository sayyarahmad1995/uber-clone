-- Historical requests and trips keep unknown context; do not infer a vehicle.
ALTER TABLE ride_requests ADD COLUMN service_code TEXT REFERENCES driver_service_catalog(code);
ALTER TABLE ride_requests ALTER COLUMN service_code SET DEFAULT 'economy';
ALTER TABLE ride_offers ADD COLUMN operation_context JSONB;
ALTER TABLE trips ADD COLUMN operation_context JSONB;
-- Old pending offers lack a reviewed operating snapshot and must be resubmitted.
UPDATE ride_offers SET status = 'closed', decided_at = NOW(), updated_at = NOW()
WHERE status = 'pending';
