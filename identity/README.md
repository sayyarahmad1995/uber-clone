# Identity Infrastructure

The application depends on the provider-neutral `internal/identity.Provider`
boundary. Provider-specific identity/session behavior stays in infrastructure and
authentication adapters rather than business domains.

## Current MVP runtime

The current self-hosted identity runtime uses Ory Kratos for identity lifecycle,
credentials, and authenticated sessions.

The mobile client does not communicate with Kratos directly. Registration, login,
logout, and authenticated-user access are application-owned HTTP contracts. The
application returns the active Kratos session token through its `access_token`
contract and validates that session internally through the provider-neutral identity
boundary.

No Ory SDK or provider-specific type belongs in application business domains.

## Hydra status

ADR-0003 originally selected Ory Hydra for OAuth 2.0 / OpenID Connect token issuance
and mobile Authorization Code + PKCE. ADR-0005 supersedes that client-session choice
for the MVP.

Hydra and its login/consent bridge are not part of the current MVP runtime. They may
be reconsidered only if a later session or OAuth/OIDC requirement justifies them.

## Deployment boundary

This directory contains the identity schema and configuration required by the current
Kratos-backed runtime. Application code must continue to depend on provider-neutral
identity/authentication boundaries so a future provider or token strategy does not
leak into Rider, Driver, marketplace, or Trip business domains.
