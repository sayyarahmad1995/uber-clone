-- Durable, single-use phone OTP challenge records. No public endpoint activates this table.
-- OTP digests must be keyed HMACs created by the application, never raw codes or unkeyed hashes.
CREATE TABLE auth_phone_challenges (
    id UUID PRIMARY KEY,
    phone_number TEXT NOT NULL CHECK (phone_number ~ '^[+][1-9][0-9]{7,14}$'),
    purpose TEXT NOT NULL CHECK (purpose = 'sign_in_or_register'),
    code_digest BYTEA NOT NULL CHECK (octet_length(code_digest) = 32),
    idempotency_key TEXT NOT NULL CHECK (char_length(idempotency_key) BETWEEN 16 AND 128),
    attempt_count SMALLINT NOT NULL DEFAULT 0 CHECK (attempt_count >= 0),
    max_attempts SMALLINT NOT NULL DEFAULT 5 CHECK (max_attempts BETWEEN 1 AND 10),
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    expires_at TIMESTAMPTZ NOT NULL,
    resend_after TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ,
    CHECK (expires_at > created_at),
    CHECK (resend_after > created_at),
    CHECK (resend_after <= expires_at)
);
CREATE UNIQUE INDEX auth_phone_challenges_idempotency
    ON auth_phone_challenges (phone_number, purpose, idempotency_key);
CREATE INDEX auth_phone_challenges_pending
    ON auth_phone_challenges (phone_number, purpose, created_at DESC)
    WHERE consumed_at IS NULL;
CREATE INDEX auth_phone_challenges_expiry ON auth_phone_challenges (expires_at);
