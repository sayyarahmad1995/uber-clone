# Authentication

## Client contract

The client communicates only with application APIs.

The client owns all authentication UI.

The client does not communicate with Ory services directly and does not depend on
provider-specific token formats or APIs.

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

`access_token` is the application API field name. In the current MVP implementation,
its value is an Ory Kratos session token surfaced through the application-owned
authentication boundary; it is not a Hydra-issued OIDC access token.

## Internal implementation

The Authentication domain and HTTP middleware depend on provider-neutral interfaces.

Ory Kratos is the current internal identity/session implementation. The application
owns registration, login, logout, and authenticated-user HTTP contracts and uses
Kratos behind those contracts.

Business domains do not depend on Kratos types or APIs. A future provider or session
strategy may replace the internal adapter without changing Rider, Driver,
marketplace, or Trip domain contracts.

See ADR-0005 for the active MVP session strategy. ADR-0003's Hydra + Authorization
Code + PKCE client-session decision is superseded for the MVP.
