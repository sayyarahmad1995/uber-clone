-- Verified-phone ownership is durable account state, never OTP cache state.
-- This table is empty on migration: existing users are NOT assumed verified.
CREATE TABLE user_verified_phones (
    phone_number TEXT PRIMARY KEY
        CHECK (phone_number ~ '^[+][1-9][0-9]{7,14}$'),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    issuer TEXT NOT NULL,
    subject TEXT NOT NULL,
    verified_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    FOREIGN KEY (issuer, subject) REFERENCES external_identities(issuer, subject)
);
CREATE INDEX user_verified_phones_identity_lookup
    ON user_verified_phones (issuer, subject);
-- The resolver joins external_identities on user_id, issuer and subject to
-- verify that the referenced identity actually belongs to this account.
