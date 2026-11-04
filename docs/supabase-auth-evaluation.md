# HiGO Supabase Auth evaluation (development only)

## Intent

HiGO selected self-hosted Supabase Auth (GoTrue) as its target
authentication provider in [ADR-0014](ADR-0014-supabase-auth-provider.md).
This is a disposable **evaluation**, not a provider switch. Existing
Kratos login, registration and protected API routes still run unchanged.

The `backend/internal/identity/gotrue` package is a dormant identity
adapter. It calls GoTrue's authenticated `GET /user`, requires a
`phone_confirmed_at` timestamp and valid E.164 phone, and returns the
provider-neutral issuer/subject principal. It does **not** issue sessions
or provision HiGO accounts. Do not wire it to the API until all endpoints
and account mapping have been implemented and validated.

## Local isolated start

Use the standalone Compose file, not the existing production/pilot files.
This will **not** mount the main HiGO PostgreSQL volume or publish the
GoTrue development DB to the host.

Generate a unique, disposable 32+ character signing key:

```bash
export GOTRUE_DEV_JWT_SECRET="$(openssl rand -hex 32)"
sudo --preserve-env=GOTRUE_DEV_JWT_SECRET docker compose -f docker-compose.gotrue-dev.yml up -d
curl -fsS http://127.0.0.1:9999/health
```

The Compose stack runs a one-shot `gotrue-dev-schema-init` job that
idempotently creates the PostgreSQL `auth` namespace **before** GoTrue's
first migration. This is required even when the disposable database volume
already exists; do not reset the volume just to repair a missing schema.

If you started the stack before this init job was added, use the **same
terminal session** where you exported `GOTRUE_DEV_JWT_SECRET`:

```bash
git pull --ff-only
sudo --preserve-env=GOTRUE_DEV_JWT_SECRET docker compose -f docker-compose.gotrue-dev.yml up -d
curl -i http://127.0.0.1:9999/health
```

If GoTrue still fails, inspect its logs without a Compose interpolation
dependency: `sudo docker logs --tail=100 higo-gotrue-evaluation-gotrue-dev-1`.
Do not use `down -v` to recover from a schema initialization error.

The service binds the HTTP API to **localhost only**. It has **no live SMS
provider** and configures the fixed fake phone `+923001234567` with test
OTP `123456` until **2026-12-01**. The test phone/password are
disposable fixtures, never real credentials. This Compose file must
never be used for production. If evaluating after that date, use a
new, explicit short-lived expiry in this dev file; never simply remove
the expiry guard.

The pinned image is `supabase/gotrue:v2.197.0`, chosen from Supabase's
self-hosted Compose version at time of evaluation. Review updates
and security advisories before any real deployment.

## Manual API contract checks

Use an HTTP client against `http://127.0.0.1:9999` (never against the
HiGO API: the adapter is not connected). For the **fake** test number:

1. `POST /signup` with JSON
   `{"phone":"+923001234567","password":"dev-only-long-password-123"}`.
   Verify signup does **not** create an authorized HiGO account or
   return an authenticated session before phone verification.
2. `POST /verify` with JSON
   `{"type":"sms","phone":"+923001234567","token":"123456"}`.
   Check invalid code and replay behavior.
3. `POST /token?grant_type=password` with the same phone/password.
   Confirm tokens are returned **without requesting another SMS**.
4. `GET /user` with `Authorization: Bearer <access_token>`.
   Confirm `phone_confirmed_at`, `phone` and `id` are present.
5. Exercise `POST /token?grant_type=refresh_token` with a refresh
   token and `POST /logout` with the access token. Check exact
   revocation and access-token semantics for the pinned version.
6. Explicitly investigate **password recovery for phone-only users**:
   reset/reauthentication and stolen/recycled/SIM-swapped numbers.
   Standard `/recover` documentation describes email recovery;
   do not publish a phone-only recovery route before a secure
   implementation is reviewed.

Run backend unit tests:

```bash
cd backend
go test ./internal/identity/gotrue
```

Do not print real passwords, bearer tokens, OTPs or signing keys in
logs, screenshots or PRs. Real SMS testing requires a separate,
explicitly configured provider and send budget; this fixture must
not accidentally send billed messages.

## Cleanup

To stop the isolated test stack without deleting data:

```bash
sudo --preserve-env=GOTRUE_DEV_JWT_SECRET docker compose -f docker-compose.gotrue-dev.yml down
```

Deleting its **disposable test database volume** is a separate,
intentional choice (`down -v`). Do not apply destructive cleanup
commands to normal HiGO Compose files or the shared database.

## Production gates

- No test OTP, autoconfirm or loopback-only HTTP assumptions.
- Route **GoTrue-owned** SMS issuance through a supported authenticated
  Send SMS Hook or built-in provider integration; do not attach HiGO's
  separate OTP verifier to GoTrue's sessions.
- Set an explicit ingress, trusted proxy/rate-limiting policy, TLS,
  signing-key rotation, secret storage, monitoring and SMS budgets.
- Implement identity mapping, application-owned login/signup/refresh/
  logout HTTP API and Flutter session restore as coherent slices.
- Exercise Rider/Driver authorization and the password-recovery gate
  before Kratos retirement.
