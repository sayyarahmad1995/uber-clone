# ADR-0003: Ory Component Scope for MVP Authentication

## Status

Accepted.

## Decision

The initial self-hosted identity stack uses:

- Ory Kratos for identity lifecycle and credentials.
- Ory Hydra for OAuth 2.0 and OpenID Connect token issuance.

The mobile client uses Authorization Code + PKCE.

The application backend validates OIDC tokens through the provider-neutral identity boundary and maps the verified issuer + subject to an application user.

## Explicit boundary

Hydra and Kratos are infrastructure components. Application business domains do not import provider-specific SDKs or use provider-specific identity types.

## Consequence

A login/consent integration component is required to bridge Hydra authorization requests with the selected identity experience. This component is infrastructure-facing and is not part of the User business domain.
