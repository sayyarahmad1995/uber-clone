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

Business domains do not depend on Kratos types or APIs.
