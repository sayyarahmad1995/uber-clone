# ADR-0005: MVP Session Strategy

## Status

Accepted.

## Decision

For the MVP, the application uses Ory Kratos session tokens as the authenticated client session.

The application API returns the session token after successful login and validates it internally through the identity provider boundary.

The client sends the token to application APIs using the Authorization header.

## Deferred

Application-managed refresh tokens are deferred. Ory Kratos session tokens are not treated as refresh tokens.

The refresh endpoint is therefore intentionally unavailable in the MVP and returns HTTP 501.

## Rationale

This keeps the MVP working without introducing a second token/session system prematurely.

A future session strategy can add application-managed refresh tokens behind the existing Authentication domain boundary without changing business domains.
