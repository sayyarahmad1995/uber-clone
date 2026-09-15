# ADR-0005: MVP Session Strategy

## Status

Accepted.

## Supersedes

For the MVP client session, this decision supersedes ADR-0003's Hydra token issuance
and mobile Authorization Code + PKCE flow. It does not supersede ADR-0003's rule that
business domains remain provider-neutral.

## Decision

For the MVP, the application uses Ory Kratos session tokens as the authenticated
client session.

The application API returns the session token after successful login and validates
it internally through the identity provider boundary. The public `access_token`
field is an application API contract; clients do not depend on Kratos-specific token
semantics.

The client sends the token to application APIs using the Authorization header.

## Deferred

Application-managed refresh tokens are deferred. Ory Kratos session tokens are not
treated as refresh tokens.

The refresh endpoint is therefore intentionally unavailable in the MVP and returns
HTTP 501.

Hydra-issued OAuth/OIDC access tokens and mobile Authorization Code + PKCE are also
deferred from the active MVP session path.

## Rationale

This keeps the MVP working without introducing a second token/session system
prematurely.

A future session strategy can add application-managed refresh tokens or replace the
internal provider behind the existing Authentication and identity boundaries without
changing business domains.
