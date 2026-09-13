ALTER TABLE trips
    ADD COLUMN settlement_status TEXT NOT NULL DEFAULT 'unsettled',
    ADD COLUMN settlement_method TEXT,
    ADD COLUMN cash_collected_at TIMESTAMPTZ,
    ADD COLUMN cash_collected_by UUID REFERENCES users(id) ON DELETE RESTRICT;

ALTER TABLE trips
    ADD CONSTRAINT trips_settlement_status_check CHECK (settlement_status IN ('unsettled', 'cash_collected')),
    ADD CONSTRAINT trips_cash_settlement_check CHECK (
        (
            settlement_status = 'unsettled'
            AND settlement_method IS NULL
            AND cash_collected_at IS NULL
            AND cash_collected_by IS NULL
        )
        OR
        (
            settlement_status = 'cash_collected'
            AND status = 'completed'
            AND settlement_method = 'cash'
            AND cash_collected_at IS NOT NULL
            AND cash_collected_by = driver_user_id
        )
    );

CREATE INDEX trips_driver_user_id_settlement_status_idx
    ON trips (driver_user_id, settlement_status, assigned_at DESC);
