-- Native identity preparation only. Existing Kratos authentication remains active.
-- No existing passwords or sessions are copied: Kratos credentials are not
-- recoverable plaintext and must not be silently converted to native hashes.

CREATE TABLE native_password_credentials (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    encoded_hash TEXT NOT NULL CHECK (octet_length(encoded_hash) BETWEEN 50 AND 256),
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);

-- Opaque bearer tokens are NEVER stored. Only SHA-256 digests of full
-- random 256-bit access and refresh tokens are persisted. Each session is
-- scoped to an existing HiGO user rather than a provider-specific subject.
CREATE TABLE native_auth_sessions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    access_digest BYTEA NOT NULL UNIQUE CHECK (octet_length(access_digest) = 32),
    refresh_digest BYTEA NOT NULL UNIQUE CHECK (octet_length(refresh_digest) = 32),
    access_expires_at TIMESTAMPTZ NOT NULL,
    refresh_expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    rotated_at TIMESTAMPTZ,
    CHECK (access_expires_at <= refresh_expires_at)
);

CREATE INDEX native_auth_sessions_live_user_idx
    ON native_auth_sessions (user_id) WHERE revoked_at IS NULL;
