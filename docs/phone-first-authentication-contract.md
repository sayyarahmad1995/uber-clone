# HiGO phone-first authentication contract (proposed v1 extension)

Status: **design contract — not implemented**. Existing routes and semantics remain unchanged until handlers, persistence, provider integration, abuse controls and tests land. This document records the agreed Flutter/backend requirements and establishes implementation acceptance criteria.

## Boundary and identity

- Flutter talks only to the HiGO Go API, never directly to Kratos or an SMS provider.
- Go remains responsible for account lifecycle, verification challenge authorization, profile state and Rider/Driver authorization. Kratos remains the identity/credential/session provider behind the Go boundary.
- The canonical login identifier is a normalized E.164 phone number. The backend verifies and binds it to one account; client-side normalization is not trusted.
- Optional email is a contact address, not a substitute for the required verified phone. Do not fabricate placeholder emails to satisfy the existing Kratos schema.
- Do not expose Kratos flow IDs, internal schema fields or provider errors to clients.

## Proposed endpoints (not yet available)

All request and response field names below are targets for implementation, not claims about the running API.

### POST /v1/auth/phone/challenges

Request: `{"phone_number":"+923001234567","purpose":"sign_in_or_register"}`.
Response: HTTP 202 `{"challenge_id":"opaque","resend_after_seconds":30}`.

The API sends a short-lived OTP to the phone through a configured SMS provider (or explicitly marked local-only test transport). Responses must not disclose whether the phone belongs to an account. IP, phone, device and challenge-level abuse controls apply. A local test transport must never be enabled in production.

### POST /v1/auth/phone/challenges/complete

Request: `{"challenge_id":"opaque","code":"123456"}`.
Response: HTTP 200 `{"verification_token":"opaque_short_lived_token","next_step":"login"}` or `"register"`.

Only a successfully completed challenge permits account-existence-based branching. Verification tokens are short lived, single use, server bound to the verified phone and the intended flow. OTPs cannot be accepted solely because they have six digits; enforce expiry and attempt limits, with atomic consumption to prevent replay. Avoid logging OTPs or tokens.

### POST /v1/auth/register (extension)

Target request: `{"verification_token":"opaque","full_name":"Example Person","email":null,"password":"..."}`. The phone is derived from the server-side verification context, not supplied as a trusted flag. Optional email is syntax-validated and stored as **unverified**. Successful registration must be compatible with the selected Kratos identity strategy and must never imply email ownership. Return an application-owned response and guide the client into login or a session only after all applicable verification requirements are fulfilled.

### POST /v1/auth/login (extension)

Target request: `{"identifier":"+923001234567","password":"..."}`, together with a server-verifiable phone-challenge context where required by the approved sign-in policy. On success return the existing application contract `{"access_token":"...","expires_in":3600}`. Never issue a usable token for an unverified account. Do not include verification markers supplied by the client as authorization.

### GET /v1/me (extension)

Target response: `{"id":"uuid","capabilities":["rider"],"phone_number":"+923001234567","phone_verified":true,"email":null,"email_verified":false}`. The server alone sets verification status. If an email is present but unverified, return `email_verified: false`.

### Logout/session restore

Preserve existing `POST /v1/auth/logout`, `GET /v1/me`, `POST /v1/auth/session/extend` behavior until migrated. Flutter uses secure token storage; bootstrapping requires an authenticated `/v1/me` request. No unimplemented refresh endpoint should be assumed.

## Request efficiency and server-load policy

The goal is to minimize **unnecessary** requests, not to skip security enforcement.

- Flutter validates E.164 formatting locally for UX; Go independently validates all untrusted input. Do not call the API on every digit typed or widget rebuild.
- **One phone challenge issuance** per deliberate action. The server returns `resend_after_seconds`; the client counts down locally without polling. Server-side throttling is authoritative. A duplicate tap while a request is in flight is suppressed.
- A verification completion response includes `next_step` so Flutter does **not** issue a second account-existence lookup. Never reveal account existence before successful phone verification.
- Registration and login requests are initiated once per user submission. Disable submit while pending and do not blindly retry mutations on timeout. For challenge issuance, support a bounded idempotency key so retrying after network loss does not send a duplicate SMS; scope keys to the requester/phone/purpose and expire them.
- A successful login's `access_token` and `expires_in` are retained securely. Restore sessions with **one** authenticated `GET /v1/me` at app startup, not a check on each screen or navigation. Do not assume a refresh endpoint exists.
- Share an in-memory current-user snapshot across screens. Deduplicate concurrent `GET /v1/me` calls and refresh only after account mutations, explicit user action, authorization failure, or session restoration. Never use cached capabilities to authorize backend operations.
- Email verification status must not be polled. Prefer completion responses that already describe the updated state, or a single `GET /v1/me` after successful mutation.
- HTTP clients should reuse connections, use bounded timeouts, cancel obsolete calls and avoid automatic retry storms. Never persist unverified state as if it were authoritative.
- Backend OTP and login abuse controls are enforced on every relevant action independent of client behavior. Use persistent/distributed rate limiting and atomic challenge state, not per-process in-memory counters, when multiple API replicas are deployed.
- Apply conservative, explicit rate limits per phone, IP, account and purpose; return `429` with `Retry-After` for throttling. Do not disclose account existence through limits or response differences. Track SMS sends, failed OTP completions, login failures and duplicate requests with privacy-safe metrics.
- Aim for the normal successful **existing-user** sign-in flow to require one challenge-issuance call, one challenge-completion call, one password-login call, and one authenticated user fetch as needed; registration substitutes a registration call and subsequent authenticated bootstrap. Do not add a separate lookup call or repeated status polling.

## Verified email rules

- Email is optional at registration; malformed nonempty email is rejected.
- Adding an email or changing an unverified email never creates a verified email.
- An unverified email can be removed or replaced.
- **Once an email is verified, it cannot be removed.**
- Replacing a verified email requires a fresh OTP to the account's registered **verified** phone, followed by proof of ownership of the **new** email address.
- Keep the old verified email active throughout the pending replacement. Change the primary email and its verified flag in a single atomic transaction **only after both challenges succeed**. Expired or failed challenges leave the old email unchanged.
- Bind challenges to account, pending address, purpose and expiration; make them single use and rate limited. Notify the old verified email of requested/completed changes. A high-assurance recovery procedure for loss of phone access is separate from ordinary email change.
- Unverified email must not be trusted for password recovery or other security-critical purposes.

## Required technical prerequisites

The current implementation is email-first: `identity/kratos/identity.schema.json` requires `traits.email`, the Kratos registration adapter passes `traits.email = identifier`, and `backend/internal/auth/http.go` accepts only identifier/password registration. `identity/kratos/kratos.yml` disables recovery. These **must be changed** before the phone-first client flow can be enabled.

The engineering implementation must settle Kratos phone identity/verification support for the pinned version and choose a compatible design, define a persistent challenge store and a production SMS delivery adapter, and ensure rollback/migration safety. Do not introduce incompatible endpoints before those dependencies are ready.

## Acceptance tests

1. OTP issuance does not enumerate accounts; resend throttling and expiry work.
2. Wrong/expired/replayed OTPs are rejected; challenge completion is concurrency safe.
3. A verified phone can register and authenticate; an unverified phone cannot.
4. Optional email can be absent or unverified and cannot be spoofed as verified.
5. A verified email cannot be removed, cannot be changed without fresh phone OTP, and stays active until the new email is verified.
6. Attempt limits, rate limits and generic errors work; sensitive values are not logged.
7. Authenticated `/v1/me`, server revocation/logout, session restoration and Rider/Driver capability authorization are tested.
8. Existing Go endpoints stay stable until the versioned extension and Flutter cutover are validated.
9. OTP resend countdowns do not poll; parallel profile reads coalesce; repeat taps do not issue duplicate mutations.
10. Duplicate challenge issuance using the same idempotency key does not generate a second SMS, and server rate limits return a bounded retry window.
