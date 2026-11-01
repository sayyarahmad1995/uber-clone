# Phone registration recovery state (inactive)

Migration `034_phone_registration_operations.sql` records **non-secret durable state**, with one operation per phone and configured identity issuer. It stores no OTP, verification token, password, or SMS body.

States:
- `reserved`: operation allocated; no Kratos identity ID recorded. **Do not blindly retry** an uncertain Kratos registration call from this state; a network timeout may mean Kratos already created an identity.
- `identity_created`: confirmed Kratos subject recorded; a coordinator may resume PostgreSQL account completion after independently verifying the Kratos identity matches the phone and the original verification authorization.
- `completed`: account, capability, verified-phone ownership and operation completion committed atomically to PostgreSQL.
- `needs_reconciliation`: reserved for explicit operator recovery (not automatically transitioned by this initial implementation).

New repository methods: `ReserveRegistration`, `GetRegistration`, `RecordRegistrationIdentity`, and `CompleteRegistration`. They are **internal primitives, not an active end-to-end registration coordinator**. A future coordinator must bind proof consumption to the operation, enforce authorization at each transition, and resolve uncertain Kratos outcomes without duplicating identities. No public phone routes are enabled.

Security note: the methods that record and complete a registration do **not** verify a Redis proof or query Kratos themselves. They must not be exposed as HTTP handlers or invoked by untrusted callers. Validate the real Kratos v1.3.1 registration payload and identity response before activating this workflow.

PostgreSQL stores durable recovery metadata; Redis remains the ephemeral OTP/proof store. Credentials remain in Kratos.

## Coordinator (internal only)

`PhoneRegistrationCoordinator.Register` consumes one Redis proof, reserves a durable operation, atomically claims the new `submitting` state, calls Kratos at most once, records a confirmed subject, and commits the account + verified-phone binding together. If creation times out or crashes after claim, the operation stays `submitting` and returns `ErrPhoneRegistrationPending` on retries. **Never automatically resubmit**; reconcile the identity against Kratos using a separately authorized maintenance path. Migration `035` introduces this state.

Known limitations: the coordinator is not wired to public HTTP routes, cannot automatically recover uncertain Kratos calls, and has not been tested against live Kratos v1.3.1. A successful proof consumption followed by database unavailability also needs an operator-guided recovery policy. Existing registration primitives require trusted internal callers and must not be directly exposed. No session is issued by this workflow.

## Live Kratos compatibility gate

Opt-in test `TestLiveKratosPhoneRegistration` makes actual registration and admin identity-read requests, verifies phone-only traits, and attempts to delete the test identity. It requires both `TEST_KRATOS_PUBLIC_URL` and `TEST_KRATOS_ADMIN_URL`. Run **only against disposable, isolated Kratos test data**, not a shared production database. The test is skipped in ordinary `go test ./...` runs. If it fails, do not activate phone authentication; Kratos 1.3.1 registration response shape/schema behavior must be corrected first. This smoke test does not cover OTP consumption, account recovery, or issuance of an authenticated session.

## Operator-directed recovery

Internal `ReconcilePhoneRegistration` can complete a stranded `submitting` or `identity_created` operation when an operator supplies its operation ID, phone and exact Kratos identity subject. It checks the operation belongs to that phone, uses a **fresh Kratos admin identity lookup** to match subject and phone, then records the subject and commits the PostgreSQL registration. Repeated completion is idempotent; mismatched phone, subject and operation ID are rejected.

**Not automated or exposed publicly.** An admin must first investigate the uncertain registration and check for duplicate Kratos identities. Matching the traits of one candidate does **not** prove no other identity exists. Do not invoke recovery using untrusted user-supplied identity IDs. Production activation still requires proof-to-reservation durability, live Kratos testing and guarded operator authentication/authorization.

## Proof-to-reservation hardening

The coordinator now uses `RedisPhoneProofs.PeekProof` to validate the short-lived proof, persists the PostgreSQL registration reservation, then consumes the same proof once before claiming `submitting`. When PostgreSQL reservation fails, the Redis proof remains usable until its original expiry. A regression test asserts this ordering. If the Redis proof expires or is consumed concurrently between peek and consume, no Kratos request is dispatched.

**Remaining limitation:** After the proof is consumed, a database failure before the `submitting` claim can leave a `reserved` operation without a usable proof. This requires a cross-store handoff/recovery protocol before public activation. Peeking does not extend proof expiry or authorize a session; do not expose `PeekProof` publicly.

## Durable claim before proof consumption

The coordinator now reserves the operation, atomically transitions it to `submitting`, **then** consumes the Redis proof, and only after successful consumption calls Kratos. This eliminates the prior gap where a proof could be irreversibly consumed while the operation was still `reserved`. A failed PostgreSQL claim leaves the proof unconsumed. A Redis failure after the claim leaves `submitting` as a durable, fail-closed marker, with **no Kratos request dispatched**.

A `submitting` state therefore means **submission may have happened**, not proof that it happened. Do not automatically reset or resubmit it: recovering an uncertain outcome requires privileged investigation and, where appropriate, renewed proof of possession. The operator reconciler requires an independently verified Kratos identity and must not infer one from `submitting` alone. This design favors preventing duplicate or unauthorized identity creation over fully automatic recovery. Public phone-auth endpoints remain disabled.
