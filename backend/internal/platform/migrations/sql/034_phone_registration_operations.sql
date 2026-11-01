-- Durable, non-secret recovery metadata for phone-first registration.
-- This does not store OTPs, passwords, SMS payloads, or verification tokens.
CREATE TABLE phone_registration_operations (
    id UUID PRIMARY KEY,
    phone_number TEXT NOT NULL CHECK (phone_number ~ '^[+][1-9][0-9]{7,14}$'),
    issuer TEXT NOT NULL,
    state TEXT NOT NULL DEFAULT 'reserved'
        CHECK (state IN ('reserved','identity_created','completed','needs_reconciliation')),
    identity_subject TEXT,
    user_id UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    CONSTRAINT phone_registration_state_fields CHECK (
        (state='reserved' AND identity_subject IS NULL AND user_id IS NULL) OR
        (state IN ('identity_created','needs_reconciliation') AND identity_subject IS NOT NULL AND user_id IS NULL) OR
        (state='completed' AND identity_subject IS NOT NULL AND user_id IS NOT NULL)
    ),
    UNIQUE (phone_number, issuer)
);
CREATE UNIQUE INDEX phone_registration_subject_unique
    ON phone_registration_operations(issuer, identity_subject)
    WHERE identity_subject IS NOT NULL;
