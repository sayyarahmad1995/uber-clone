# ADR-0014: Phone-identified, password-authenticated accounts

## Status

Accepted product/architecture decision — 2026-11-04. **Design decision only:** this ADR
does not assert that the future Go-native Identity Module, new routes, or session
refresh have been implemented.

Supersedes the *OTP-before-every-sign-in* flow proposed in
[phone-first-authentication-contract.md](phone-first-authentication-contract.md).
Preserves the application-owned API boundary of ADR-0004 and provider-neutral
external-system boundary of ADR-0006. Existing Kratos-backed login and sessions
remain in place until an explicit, tested migration replaces them.

## Motivation

At an observed Twilio Pakistan rate of USD 0.4734 per SMS segment, sending a
code for every routine login would make authentication unnecessarily expensive
and introduce an external delivery dependency into normal sign-in.

Phone remains the primary account identifier; it must **not** be conflated
with a requirement to send SMS during every authentication.

## Decision: user experience and account lifecycle

1. **Phone entry.** The shared Flutter app asks for a phone number, normalized
   by Go to E.164. Entering or submitting the number does **not** send an OTP.
2. **Privacy-safe next step.** Present a password sign-in path and a
   **Create account** action regardless of whether the phone is registered.
   Do **not** expose a public existence-check endpoint, or automatically
   redirect to registration versus login based on an unauthenticated lookup.
   Such branching would enable account enumeration. After the user supplies
   valid credentials or proves possession through an authorized verification
   flow, the API may disclose the next step. This deliberately refines the
   originally proposed automatic redirect for account privacy.
3. **Registration.** Collect the minimum account information, set a password,
   and verify possession of the intended phone using a short-lived, single-use
   OTP before activating/claiming the phone-bound account. Do not create a
   usable unverified account or allow an existing phone to be taken over.
   Treat attempts involving existing accounts with privacy-safe responses
   and offer sign-in/recovery options without revealing existence to callers.
   Keep optional email unverified until independently verified.
4. **Normal sign-in.** A verified, existing account signs in using
   phone + password. On success, create an authenticated application session.
   **No SMS/OTP is required for ordinary password sign-in.**
5. **Returning to the app.** Restore a still-valid, securely stored session;
   rotate/refresh credentials only through an intentionally designed and
   implemented session lifecycle. Do not ask for a password or SMS merely
   because the app restarts. The existing MVP has no general refresh endpoint;
   do not make the client depend on one before implementing it.
6. **Forgot password.** Provide a separate, rate-limited recovery flow that
   verifies control of an eligible recovery factor (initially the verified
   phone via an OTP) before resetting credentials. Recovery requests must
   use generic responses, bounded attempts, expiry and one-time consumption.
   On successful recovery, revoke other appropriate sessions. SIM swaps,
   lost-number recovery and high-value Driver/operator actions need explicit
   risk treatment; phone possession by itself is not universally sufficient
   proof of identity.
7. **Risk-based step-up.** Use additional verification when warranted by
   security-sensitive actions, unusual authentication risk or changes to
   recovery factors; do not require it for every routine app launch.

One HiGO account continues to own Rider and optional Driver capabilities.
Neither capability creates a second identity.

## Security requirements

- **No account enumeration:** generic invalid-login, registration-collision
  and password-recovery responses; control timing and response-size
  differences as practical; prevent per-account-specific rate-limit messages.
  Rate limits and bot/abuse controls are necessary but **do not** make a
  public `phone_exists` response private.
- **Password storage:** once the Go-native Identity Module is introduced,
  store a suitably parameterized, salted Argon2id password hash; never
  plaintext or a plain fast hash. Support rehash-on-login as parameters evolve.
  While Kratos remains the credential provider, keep password processing
  within its approved credential boundary.
- **Credential abuse:** defend against password spraying, credential
  stuffing, brute force and login enumeration with Redis-backed or otherwise
  shared per-IP, per-phone/account, device-aware and risk-aware throttles;
  avoid permanent, attacker-triggerable account lockouts. Do not leak
  whether the password or account identifier was wrong.
- **SMS abuse/cost control:** provider-neutral OTP delivery; per-destination,
  per-source/device and per-purpose limits, idempotency, resend cooldown,
  sending budgets, geographic restrictions, monitoring, and a bounded
  recovery pathway. A 202 issuance response is not proof of delivery.
- **OTP and proof safety:** cryptographically random codes, short expiry,
  attempt limits, atomic one-time verification, purpose/phone binding and
  replay resistance. Never log raw OTPs, passwords, tokens or proofs.
- **Session safety:** TLS in deployment; bounded token lifetimes; secure
  platform storage in Flutter; rotation and server-side revocation when
  those mechanisms are introduced; protect refresh credentials against
  replay; logout and compromised-account recovery must invalidate the
  appropriate credentials. Restore state from authenticated `/v1/me`,
  not from a client-only verified flag.
- **Authorization:** verify account status and Rider/Driver/operator access
  server-side on every protected operation. Phone verification is not
  authorization to drive, approve applications or change sensitive data.
- **Privacy/operations:** minimize retained phone data and sensitive logs,
  protect recovery against SIM swaps, collect monitoring without exposing
  credentials, and keep registration/OTP abuse defenses testable.

## Implementation boundaries and sequence

- Keep the existing **Go modular monolith**, PostgreSQL identity/account
  persistence, and Redis ephemeral OTP/rate-limit infrastructure.
- Reuse the already tested provider-neutral OTP engine, SMS outbox and
  Twilio adapter for **registration and recovery**. Do not delete the
  in-progress development-only issuance endpoint; do not expose it broadly.
- Implement Go-native identity/credential/session ownership through a
  separate, reviewed migration plan, reconciling existing Kratos subjects,
  user mappings and sessions before changing runtime auth. Existing
  Kratos-based endpoints stay operational during cutover.
- Maintain application-owned auth APIs so the client never calls Kratos
  or messaging providers directly. Final route names and response schemas
  must be specified and tested before Flutter cutover; avoid claiming
  draft endpoints already exist.
- Prefer one vertical slice at a time: identity model and migration;
  password login and secure session lifecycle; registration with OTP;
  password recovery; risk-based step-up and operational safeguards.
- Move the existing Flutter sign-in flow to the approved model only
  after backend integration and device testing.

## Acceptance gates

1. An unknown vs. known phone cannot be distinguished through an anonymous
   phone-lookup response; login and recovery errors are privacy-safe.
2. Registration binds exactly one account to a verified E.164 phone, with
   atomic proof consumption and protection against duplicate identities.
3. Correct phone + password creates an authorized session **without SMS**;
   wrong credentials do not; unverified or disabled accounts are denied.
4. A valid stored session can restore the app; logout and applicable
   revocation/expiry terminate access.
5. Recovery requires a verified factor and rejects wrong, expired, replayed
   or over-attempt OTPs; credential reset invalidates intended sessions.
6. Password storage and login throttling withstand offline and online
   credential attacks; no passwords, OTPs or tokens reach logs.
7. SMS frequency/cost limits and retries are enforced server-side; Twilio
   can be replaced behind the delivery port without changing identity APIs.
8. Current Rider/Driver authorization, account continuity and migration
   rollback are validated before retiring Kratos.

## Deliberate non-goals

No mandatory SMS for every login, no public `is_registered` API, no
microservices, no speculative multi-provider identity framework, and no
assumption that passwordless or passkey sign-in is ready in this slice.
