# ADR-0014: Supabase Auth (GoTrue) as HiGO's target authentication provider

## Status

Accepted target provider and phone/password product policy (2026-11-04).
**Implementation and provider cutover remain gated by tests.** The current
backend continues using Kratos until an end-to-end replacement is approved.

## Decision and rationale

- Self-host **Supabase Auth (GoTrue)**, the standalone open-source Go service,
  for identities, salted password credentials, login, SMS verification,
  recovery and authentication sessions. HiGO's Go backend is **not** a second
  password/session server.
- Keep app-owned HTTP endpoints and Flutter-owned login UI. Flutter does not
  need Supabase Database, Storage, Studio, PostgREST, or direct GoTrue access.
- The HiGO backend validates presented bearer credentials through a trusted
  GoTrue identity adapter, maps provider subject to one application account,
  and authorizes Rider, Driver, vehicle and Trip operations itself.
- No real users exist, so no Kratos password/account/session migration is
  required. Retain disposable test environments separately; never reset a
  shared PostgreSQL database as an automatic part of provider cutover.
- The rationale is **usable phone authentication and future provider
  flexibility**, not simply the total number of offered sign-in methods.
  Self-hosting still incurs infrastructure, SMS and security maintenance costs.

## Product contract

- E.164 phone number is the primary account identifier.
- Both **Sign in** and **Create account** remain accessible after phone entry.
  There is no public phone-existence lookup that could enumerate accounts.
- Registration: phone + password, verified through a short-lived SMS OTP
  before HiGO enables Rider/Driver access.
- Normal login: phone + password creates/restores a session **without SMS**.
  Valid sessions persist across app restarts and use provider-native refresh.
- Recovery: a verified, rate-limited account-recovery flow must be proven
  against the pinned GoTrue build **before launch**. Do not assume that
  ordinary phone passwordless OTP equals safe password recovery; recycled
  numbers and SIM swaps require risk controls and recovery-session policy.
- Email is optional and unverified until separately confirmed. Do not add
  fake email addresses to satisfy an auth schema.

## Boundaries

1. **GoTrue** stores identity credentials, verification state and sessions
   in its own PostgreSQL schema/database.
2. **HiGO** stores user/business data and provider-neutral identity mappings;
   no business domain imports GoTrue types.
3. **SMS** sends provider-owned OTP messages only for verified registration,
   recovery and justified step-up. Prefer GoTrue's Send SMS Hook (after the
   exact self-hosted version and networking contract are tested) to route
   to a local Pakistani SMS carrier. Do **not** run a second competing HiGO
   OTP verification state machine for the same GoTrue flow.
4. **Secrets** such as GoTrue signing keys and service-role tokens are never
   exposed to Flutter or logged. The Go API must not treat a decoded but
   unverified JWT as authentication.
5. SMS test codes, autoconfirm, HTTP-only endpoints and developer test
   credentials are strictly forbidden in production.
6. Existing Kratos and its OTP experimental code remain present until the
   replacement passes all gates; they are retired only in a separately
   reviewable cleanup PR with database dependency analysis.

## Acceptance gates before replacing Kratos

1. Isolated pinned GoTrue container starts with its **own disposable database**.
2. Phone + password signup triggers a challenge (without autoconfirm);
   valid OTP verifies ownership; wrong/expired/replayed OTP fails.
3. Verified account signs in with phone + password with **zero SMS sends**;
   wrong password / unknown phone yield privacy-safe errors and throttling.
4. The Go API accepts only validated GoTrue identities and **requires
   phone_confirmed_at** as well as valid E.164 phone ownership; unverified,
   anonymous, banned and revoked sessions fail closed.
5. GoTrue subject `id` maps to one HiGO account with the normal default
   Rider capability. Identity changes cannot take over another account.
6. Refresh, logout, account revocation and app restart are tested, including
   stale access tokens and race conditions. Note that self-contained JWTs
   may remain valid until expiration unless checked server-side.
7. **Phone-only password recovery** works end to end with adequate
   account-recovery assurance; password change revokes appropriate
   sessions, and SIM-swap/recycled-number risks are considered.
8. SMS rate limits, idempotency, cost budgets and local-provider hook
   authentication are tested with no leakage of OTPs, passwords or tokens.
9. Existing Rider/Driver authorization and Flutter session restoration pass
   integration/physical-device tests.
10. After gate completion: remove Kratos config/containers/code and
    unused native-auth PRs without removing unrelated marketplace data.

## Initial implementation

The `feat/supabase-auth-evaluation` branch provides:
- an isolated loopback-only `docker-compose.gotrue-dev.yml` with a fake OTP
  mapping and independently named disposable database volume;
- a **dormant**, unit-tested `backend/internal/identity/gotrue` adapter
  for authenticated `GET /user` and verified phone checks;
- an evaluation/runbook. No public endpoint is switched.

The previous ADR-0014 proposed in unmerged PR #135 and native identity
code in PR #136 are superseded; the Twilio issuance-only PR #134 is not
the target OTP protocol under GoTrue.
