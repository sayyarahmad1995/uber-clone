CREATE TABLE external_identities (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    issuer TEXT NOT NULL,
    subject TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (issuer, subject)
);

INSERT INTO external_identities (user_id, issuer, subject, created_at)
SELECT id, 'legacy', external_subject, created_at
FROM users;

ALTER TABLE users DROP CONSTRAINT users_external_subject_key;
ALTER TABLE users DROP COLUMN external_subject;
