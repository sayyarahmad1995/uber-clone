# HiGO phone OTP / Twilio smoke test

Status: **development-only smoke-test route**, disabled by default. This is
not general phone-first login, nor proof of successful SMS delivery until the
test phone actually receives the message.

The implementation builds on the existing Redis challenge store, atomic
encrypted outbox, Twilio Programmable Messaging sender, and background worker.
It does **not** require the Kratos phone registration feature to be enabled.
The eventual Go-native Identity Module will reuse this provider-neutral OTP
boundary.

## Prerequisites

- Access to an operational Twilio account, API key and sender configured for
  the destination's country. Confirm Pakistan (+92) SMS permissions and any
  sender or account requirements before testing.
- PostgreSQL, Redis, and the existing HiGO API infrastructure running.
- A personally controlled, explicitly consented test phone number in E.164.
- The same Redis instance, OTP HMAC secret, and dispatch encryption key used
  by both issuance and the worker.
- Do not store Twilio credentials, OTP codes or personal test numbers in
  Git, public issue comments, or shared logs.

## Local configuration

Set these values in the secure local environment of the HiGO API process:

```dotenv
PHONE_SMS_ISSUANCE_ENABLED=true
PHONE_SMS_WORKER_ENABLED=true
PHONE_REGISTRATION_ENABLED=false
PHONE_SMS_TEST_ALLOWED_NUMBERS=+92300REPLACE_WITH_REAL_TEST_NUMBER
PHONE_OTP_HMAC_SECRET=<unique random secret at least 32 bytes>
PHONE_SMS_DISPATCH_ENCRYPTION_KEY=<base64 of exactly 32 random bytes>
TWILIO_ACCOUNT_SID=<account SID>
TWILIO_API_KEY_SID=<API key SID>
TWILIO_API_KEY_SECRET=<API key secret>
TWILIO_MESSAGING_SERVICE_SID=<approved messaging service SID>
# Or TWILIO_FROM_NUMBER=<approved sender>, never both.
REDIS_URL=<existing Redis connection URL>
```

The example phone value is a placeholder, **not** a valid number. All
credentials must come from the existing secure configuration workflow.
Keep `AUTH_PROVIDER=kratos` for the currently running backend until the
native identity migration is completed.

When `PHONE_SMS_ISSUANCE_ENABLED=true`, startup refuses to proceed without
`PHONE_SMS_WORKER_ENABLED=true`. The API also fails closed for an empty,
malformed, wildcard, or overbroad test-number allowlist. At most ten
explicit destination numbers are accepted.

Do not enable this development route on the public pilot deployment until
the proxy/source identity, abuse monitoring, rate limits, and network policy
are reviewed. The current rate limiter uses the TCP peer from
`http.Request.RemoteAddr` and **never** trusts unvalidated
`X-Forwarded-For`. Behind a reverse proxy, many users can share a source
bucket; this requires a deliberate policy before public rollout.

## Exercise the full local API path

After starting the API and worker, select one allowlisted test number and
create an unpredictable idempotency key. Keep the **same key** when retrying
the same request after a network timeout.

```bash
# Replace these examples in a secure local terminal.
export HIGO_TEST_PHONE='+92...'
export HIGO_TEST_IDEMPOTENCY_KEY="$(uuidgen)"

curl --fail-with-body --show-error -i \
  -X POST http://127.0.0.1:8080/v1/auth/phone/challenges \
  -H 'Content-Type: application/json' \
  -H "Idempotency-Key: $HIGO_TEST_IDEMPOTENCY_KEY" \
  --data "{\"phone_number\":\"$HIGO_TEST_PHONE\",\"purpose\":\"sign_in_or_register\"}"
```

Expected: HTTP **202 Accepted** with `challenge_id` and
`resend_after_seconds`. A 202 means the SMS is **queued**, not confirmed
as delivered.

Check the Twilio Messaging Logs and the recipient device, then enter the
actual received code into the verification endpoint:

```bash
curl --fail-with-body --show-error -i \
  -X POST http://127.0.0.1:8080/v1/auth/phone/challenges/verify \
  -H 'Content-Type: application/json' \
  --data '{"challenge_id":"<ID_FROM_202_RESPONSE>","code":"<RECEIVED_6_DIGIT_CODE>"}'
```

Expected: HTTP 200 with a **short-lived proof token**, not a login session.
Never print that token in shared logs. A second use of the same OTP must
fail, as must an incorrect/expired code.

## Test matrix

1. Enabled SMS issuance is restricted to allowlisted numbers; other numbers
   receive HTTP 403 and cannot cause a Twilio submission.
2. Invalid payloads, missing idempotency keys and unsupported purposes never
   enqueue an SMS.
3. A retry using the **same** idempotency key returns the previous challenge
   without sending a second SMS.
4. A fresh key during the 30-second resend cooldown returns HTTP 429 with a
   retry hint.
5. Source-level and phone-level rate limits reject excessive attempts.
6. Twilio logs show an accepted outbound message and the test phone actually
   receives it. Verify sender delivery constraints for +92.
7. Verification is single-use; incorrect codes consume an attempt and cannot
   grant a proof.
8. Disconnect Redis or disable the SMS worker in a safe test environment to
   validate fail-closed startup / error handling.
9. Test ambiguous Twilio outcomes: the queue must not blindly retry and send
   duplicates after a potentially accepted submission.

## Known limitations / next steps

- The new issuer is for **explicitly allowlisted development testing** only.
  It is not the final public OTP API.
- SMS delivery status is external and asynchronous. A queued challenge does
  not mean that the code was delivered or that the phone is verified.
- Twilio is used through Programmable Messaging; HiGO generates and verifies
  the OTP. **Twilio Verify is not used** by this implementation.
- The current backend still uses Kratos for account credentials and sessions;
  moving to a native Go Identity Module remains a separate task.
- Before public rollout: approve trusted-proxy source extraction, configure
  abuse alarms and sender restrictions, settle live delivery monitoring, and
  complete session and registration lifecycle tests.

When the test is finished, set `PHONE_SMS_ISSUANCE_ENABLED=false` and
`PHONE_SMS_WORKER_ENABLED=false` for environments that do not need them.
