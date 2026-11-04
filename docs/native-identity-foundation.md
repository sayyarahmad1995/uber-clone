# Go-native identity foundation — slice 1

## Scope and runtime status

This slice prepares future native credentials and sessions while **leaving the
running Kratos-backed authentication routes unchanged**. No native login,
registration, recovery, refresh, or session-authentication endpoint is wired.
Do not infer that users can sign in with the new subsystem yet.

It follows [ADR-0014](ADR-0014-phone-password-authentication.md), and can
land in a stacked PR after that decision is reviewed.

## Implemented

- `backend/internal/identity/native/password.go`: salted Argon2id PHC hashes,
  15-character minimum and 1024-byte maximum; bounded parsing to prevent
  stored-parameter resource exhaustion; wrong-password constant-time compare;
  rehash hint for future parameter changes.
- `native/sessions.go`: opaque 256-bit random access/refresh bearer tokens
  with nonsecret UUID session identifiers, 15-minute access lifetime,
  30-day **absolute** refresh lifetime, hashed-token persistence contract,
  single-use refresh rotation, logout and per-account revocation.
- `native/postgres.go`: credential storage, session lookup requiring both
  session UUID and digest, and atomic SQL compare-and-swap refresh rotation.
- Migration 036: new, empty native credential/session tables. **No password
  or session migration** from Kratos is attempted or needed: HiGO has no
  production users.
- Unit tests: hash salt uniqueness, credential policy, malformed hash bounds,
  wrong password, session expiry, revocation, refresh replay/concurrency.
- PostgreSQL integration tests: persisted credentials, digest-only session
  storage, refresh rotation, reuse rejection and account-wide revocation,
  scoped to disposable `*_test` databases.

## Security boundaries / remaining work

**Not safe to enable as public auth yet.** This package is infrastructure,
not a verified-user authentication coordinator.

Next slices must implement:
1. A native, verified phone-to-user mapping independent of Kratos subjects.
   **Skip migration of Kratos accounts, passwords and sessions:** no real
   users exist. Any disposable development identities must be treated as
   test data, with no automatic destructive database reset.
2. Phone+password login with uniform errors, fake-hash timing work for
   unknown users, account status checks, Redis-backed abuse throttles,
   and server-issued native sessions only after full authentication.
3. Registration linked to single-use, purpose-bound phone OTP proof with
   atomic account creation. No public phone-existence API.
4. Recovery, SIM-swap risk policy, session revocation upon credential
   reset, recovery OTP budgets, and app-owned session routes.
5. Refresh-token replay detection and incident handling beyond basic
   single-use rotation; session limits/cleanup, TLS, secure Flutter storage,
   native protected-route integration and Kratos retirement testing.
6. Updated Flutter entry UX and full physical-device testing.

The current Twilio issuance test route remains separately development-only.

**Retirement rule:** Keep Kratos behind the existing live boundary only until
native registration, login, recovery and protected-route authorization are
demonstrably working. Then remove Kratos adapters, routes, Compose services and
configuration rather than maintaining permanent compatibility. Do not remove
shared `users` or other business-domain data as a side effect. Review the
current `external_identities` and `user_verified_phones` foreign-key
relationships before altering schema; reuse the OTP verification core, not
Kratos-specific registration coordination.

Run `go test ./internal/identity/native` and the ordinary backend CI;
PostgreSQL integration requires `TEST_DATABASE_URL` ending in `_test`.
