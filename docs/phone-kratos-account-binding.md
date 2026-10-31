# Phone proof to Kratos identity integration

Status: **account-binding contract and unit tests only; no public routes enabled**.

The current Ory Kratos v1.3.1 adapter is email-first: registration supplies an email trait and login uses an identifier/password flow. Its `EnsureVerifiedSession` check must remain enforced. The existing user model records a Kratos issuer/subject link but does not currently expose an authoritative verified-phone lookup. Consequently, a completed Redis phone OTP is **not sufficient** to register or authenticate a Kratos user today.

## Rules for the next implementation

1. The verified phone passed to `PhoneAccountChecker.Decide` must originate from a validated OTP/proof, never from client-controlled input. Account lookup must use PostgreSQL's authoritative verified-phone binding. Missing records select registration; existing records require the matching Kratos identity subject.
2. Only after phone possession is proven should a handler disclose `next_step=login|register`. Never return a Kratos identity ID in this public response.
3. Registration must establish the phone-to-Kratos link in a recoverable, idempotent transaction/saga, prevent one phone linking to multiple identities, and reconcile failures. Optional email must not be treated as verified merely because the phone was verified.
4. Existing-user login must continue to require valid Kratos credentials and a checked session. A phone proof cannot independently mint or elevate a Kratos session; identity returned from Kratos must match the account associated with the verified phone.
5. Do not fake an email trait to satisfy the current Kratos identity schema. Adapt the schema and provider method for optional-email, verified-phone-first identity before exposing the new endpoints.
6. Single-use proof consumption must be tied to the relevant login/registration transaction. If consuming a proof precedes a failing durable write, establish retry-safe application-level recovery without allowing proof replay or account linkage duplication.
7. Keep the current email auth routes operational until a migration and compatibility plan is tested. Redis stores temporary challenges/proofs; PostgreSQL stores durable phone ownership; Kratos owns credentials/sessions.
8. Production service-to-service connections must use verified TLS; no auth endpoint should be activated before SMS provider delivery and trusted-source request limits are complete.

## Implemented in this slice

`phone_account_binding.go` adds a minimal `PhoneAccountResolver` interface and a checker for verified-phone -> registered Kratos identity routing. Unit tests assert existing/new account classification, inconsistent binding rejection, and fail-closed resolver errors.

No implementation of the resolver, database migration, Kratos trait changes, login/session binding or HTTP handler is part of this slice.
