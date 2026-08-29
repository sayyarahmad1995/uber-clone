CREATE TABLE users (
    id UUID PRIMARY KEY,
    external_subject TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE user_capabilities (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    capability TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, capability),
    CONSTRAINT user_capabilities_capability_check CHECK (capability IN ('rider', 'driver', 'courier', 'freight'))
);
