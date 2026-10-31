# Kratos phone-first compatibility gate

The default Kratos identity schema now accepts either an email or an E.164 phone trait as a password identifier. Email remains optional **for schema validation only**; existing email/password registration and email verification annotations are preserved.

**Do not enable phone-only registration or public phone OTP endpoints yet.** The existing Go Kratos registration adapter still submits `traits.email`, and both `auth/kratos.Provider.EnsureVerifiedSession` and `identity/kratos.Provider.AuthenticateVerified` require at least one *Kratos-verified address*. The SMS challenge proof is owned by HiGO in Redis, and does not automatically mark any Kratos address as verified. Neither guard may simply be relaxed to accept an unverified phone trait.

Before activating phone-only identities:

1. Validate this schema against the actual Kratos v1.3.1 server and test that old email identities can still log in.
2. Choose and implement a supported verification bridge: authenticate a Kratos password session, resolve the Kratos identity subject, verify the session's phone matches an authoritative PostgreSQL verified-phone binding and the required OTP proof, then apply the same checks in both session guard paths. Account creation/binding must be recovery-safe.
3. Never accept a phone-number trait as proof of possession. Only the OTP proof and durable verified-phone binding can establish that fact.
4. Preserve the policy that optional email must be separately verified before it can grant email-based privileges.
5. The public phone routes are still disabled; production HTTP-to-Kratos, PostgreSQL and Redis connections require verified TLS.

The tests added in `backend/internal/auth/kratos/identity_schema_test.go` validate the intended schema contract, not a live Kratos migration or completed SMS registration flow.
