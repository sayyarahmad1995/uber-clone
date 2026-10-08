CREATE TABLE ride_pricing_policy_versions (
 id UUID PRIMARY KEY,
 service_code TEXT NOT NULL REFERENCES driver_service_catalog(code),
 currency TEXT NOT NULL CHECK(currency='PKR'),
 version BIGINT NOT NULL CHECK(version>0),
 base_fare_minor BIGINT NOT NULL CHECK(base_fare_minor BETWEEN 0 AND 1000000000000),
 rate_minor_per_km BIGINT NOT NULL CHECK(rate_minor_per_km BETWEEN 0 AND 1000000000000),
 rate_minor_per_minute BIGINT NOT NULL CHECK(rate_minor_per_minute BETWEEN 0 AND 1000000000000),
 minimum_fare_minor BIGINT NOT NULL CHECK(minimum_fare_minor BETWEEN 1 AND 1000000000000),
 rounding_increment_minor BIGINT NOT NULL CHECK(rounding_increment_minor BETWEEN 1 AND 1000000000000),
 calculation_rule TEXT NOT NULL CHECK(calculation_rule='half_up_once_v1'),
 effective_from TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
 published_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
 published_by TEXT NOT NULL CHECK(btrim(published_by)<>''),
 UNIQUE(service_code,currency,version),
 UNIQUE(id,service_code,currency),
 UNIQUE(id,service_code,currency,version)
);
CREATE TABLE ride_pricing_policy_current (
 service_code TEXT NOT NULL REFERENCES driver_service_catalog(code),
 currency TEXT NOT NULL CHECK(currency='PKR'),
 current_policy_id UUID,
 updated_by TEXT NOT NULL,
 updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
 PRIMARY KEY(service_code,currency),
 FOREIGN KEY(current_policy_id,service_code,currency) REFERENCES ride_pricing_policy_versions(id,service_code,currency)
);
CREATE TABLE ride_request_pricing_snapshots (
 ride_request_id UUID PRIMARY KEY REFERENCES ride_requests(id),
 policy_id UUID NOT NULL,
 service_code TEXT NOT NULL,
 currency TEXT NOT NULL CHECK(currency='PKR'),
 version BIGINT NOT NULL,
 base_fare_minor BIGINT NOT NULL CHECK(base_fare_minor BETWEEN 0 AND 1000000000000),
 rate_minor_per_km BIGINT NOT NULL CHECK(rate_minor_per_km BETWEEN 0 AND 1000000000000),
 rate_minor_per_minute BIGINT NOT NULL CHECK(rate_minor_per_minute BETWEEN 0 AND 1000000000000),
 minimum_fare_minor BIGINT NOT NULL CHECK(minimum_fare_minor BETWEEN 1 AND 1000000000000),
 rounding_increment_minor BIGINT NOT NULL CHECK(rounding_increment_minor BETWEEN 1 AND 1000000000000),
 calculation_rule TEXT NOT NULL CHECK(calculation_rule='half_up_once_v1'),
 snapshot_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
 FOREIGN KEY(policy_id,service_code,currency,version) REFERENCES ride_pricing_policy_versions(id,service_code,currency,version)
);
CREATE FUNCTION reject_pricing_history_mutation() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 RAISE EXCEPTION 'pricing history is immutable';
END;
$$;
CREATE TRIGGER immutable_pricing_version BEFORE UPDATE OR DELETE ON ride_pricing_policy_versions FOR EACH ROW EXECUTE FUNCTION reject_pricing_history_mutation();
CREATE TRIGGER immutable_request_pricing BEFORE UPDATE OR DELETE ON ride_request_pricing_snapshots FOR EACH ROW EXECUTE FUNCTION reject_pricing_history_mutation();
