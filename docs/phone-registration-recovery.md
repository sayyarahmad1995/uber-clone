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

## Opt-in three-service integration test

`TestLivePhoneRegistrationCoordinator` is an opt-in test requiring `TEST_DATABASE_URL` targeting a `*_test` PostgreSQL database, `TEST_REDIS_URL`, and both Kratos URLs with hostname `kratos-test`. It creates an isolated PostgreSQL schema, a short-lived test proof directly in Redis, runs the registration coordinator against the isolated Kratos instance, validates the durable phone binding, and rejects proof replay. It attempts to delete the created Kratos identity and drop the test PostgreSQL schema. Use disposable test environments only.

This test **seeds** a valid-format proof in Redis; it does not validate the complete OTP challenge, SMS delivery, or real phone possession. It does not yet test phone/password login or protected-route access. If Kratos succeeds but the coordinator loses the identity result before recording the subject, manual orphan-identity cleanup may be needed; no production endpoints should be enabled on the strength of this test alone.

## Feature-flagged HTTP adapter

The API now supports `POST /v1/auth/phone/register` with JSON `{"proof_token":"<opaque proof>","password":"<password>"}` only when `PHONE_REGISTRATION_ENABLED=true`. Default is disabled (route not mounted). When enabled, the application initializes a Redis proof consumer using `REDIS_URL`, reuses the configured Kratos registration provider and PostgreSQL operation repository, and fails startup if Redis cannot be reached. Success returns HTTP 201 with a `user_id`. Invalid proofs return HTTP 400; unresolved/pending operations return HTTP 409. The handler does not expose credentials, proof, or internal errors.

**Do not enable in production yet.** There is no publicly wired OTP initiation/verification or SMS dispatcher worker; no complete external abuse-control policy or live API-level authorization test has been validated. The existing `POST /v1/auth/login` can authenticate phone credentials through Kratos, but phone registration should remain disabled until the OTP delivery path and HTTP integration are ready.

## Phone OTP verification HTTP adapter

When `PHONE_REGISTRATION_ENABLED=true` and `PHONE_OTP_HMAC_SECRET` is a stable secret of at least 32 bytes, the backend initializes the existing Redis OTP challenge verifier and mounts `POST /v1/auth/phone/challenges/verify` with JSON `{"challenge_id":"<uuid>","code":"<6 digits>"}`. A valid unused code returns `{"proof_token":"<opaque proof>"}` with `Cache-Control: no-store`. No authenticated session is issued. Invalid requests and Redis failures return non-sensitive errors.

**Challenge initiation remains deliberately unmounted.** The encrypted Redis SMS queue exists, but no configured production SMS sender/dispatcher and IP-level abuse controls have been validated. Enabling the feature flag does not make OTP issuance available. Before a public launch, configure actual SMS delivery, a worker with provider-side idempotency, issue-side IP/global limits, secrets management, and HTTP end-to-end tests. Do not publish this flow merely because the verification endpoint compiles.

## SMS dispatch worker foundation

`OTPDispatchWorker` now consumes the existing encrypted Redis dispatch queue through `Due`, `Claim`, and `Finish`. The provider-neutral `SMSSender` must support **provider-side idempotency** using the original challenge UUID: ambiguous provider results can be retried with the same ID. Each send is bounded by half the queue lease; failed sends are acknowledged as retryable by the queue, and a missing sender fails closed. The worker never logs OTP codes or recipients.

This commit intentionally **does not configure a provider, run a background worker, or expose OTP initiation**. The queue's retry ceiling is three attempts and its delivery window is bounded by the challenge TTL. Deployment additionally needs provider-specific idempotency verification, worker lifecycle management, metrics, rate limiting, and SMS delivery failure alerts. Until those exist, keep `PHONE_REGISTRATION_ENABLED=false` in publicly exposed deployments.
