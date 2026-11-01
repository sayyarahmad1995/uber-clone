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
