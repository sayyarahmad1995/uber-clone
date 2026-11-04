# Authentication

## Client contract

The client communicates only with application APIs.

The client owns all authentication UI.

The client does not communicate with Ory services directly.

## MVP endpoints

- `POST /v1/auth/register`
- `POST /v1/auth/login`
- `POST /v1/auth/logout`
- `GET /v1/me`

`POST /v1/auth/refresh` is reserved but intentionally unavailable for the MVP.

## Login response

A successful login returns an application authentication session containing:

- `access_token`
- `expires_in`

The token is sent to authenticated endpoints as:

`Authorization: Bearer <access_token>`

## Internal implementation

The Authentication domain depends on provider-neutral interfaces.

Ory Kratos is the initial internal implementation.

Supabase Auth (GoTrue) is the selected **target** under
[ADR-0014](ADR-0014-supabase-auth-provider.md). The current public
endpoints and Kratos session contract are unchanged until a reviewed
implementation cutover. Password login after phone verification must not
send a routine OTP; secure phone-only recovery is a launch gate.
GoTrue issues and refreshes tokens; HiGO authorizes business operations.

Business domains do not depend on Kratos types or APIs.
