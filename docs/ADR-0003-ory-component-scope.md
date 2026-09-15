# ADR-0003: Ory Component Scope for MVP Authentication

## Status

Superseded in part by ADR-0005.

The provider-neutral identity boundary remains accepted. The MVP decision to use
Hydra-issued OIDC tokens and mobile Authorization Code + PKCE is superseded by the
Kratos-session strategy in ADR-0005.

## Original decision

The initial self-hosted identity stack was planned to use:

- Ory Kratos for identity lifecycle and credentials.
- Ory Hydra for OAuth 2.0 and OpenID Connect token issuance.

The planned mobile client flow was Authorization Code + PKCE.

The application backend would validate OIDC tokens through the provider-neutral
identity boundary and map the verified issuer + subject to an application user.

## Retained boundary

Kratos, Hydra, or any future identity provider are infrastructure concerns.
Application business domains do not import provider-specific SDKs or use
provider-specific identity types.

The current MVP session mechanism is defined by ADR-0005 and must remain behind
that provider-neutral boundary.

## Consequence

The Hydra login/consent integration described by the original decision is not a
current MVP dependency. If a future session strategy reintroduces OAuth/OIDC token
issuance, that integration remains infrastructure-facing and outside the User
business domain.
