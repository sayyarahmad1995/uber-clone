# Identity Infrastructure

The application depends on the provider-neutral `internal/identity.Provider` boundary.

The current deployment uses self-hosted Ory components:

- Kratos: identity lifecycle and credentials.
- Hydra: OAuth 2.0 / OpenID Connect authorization server.

No Ory SDK or provider-specific type belongs in application business domains.

## MVP boundary

This repository contains the deployment boundary and identity schema. The remaining integration work is isolated to the Ory adapter and OAuth/OIDC client configuration.

The initial mobile flow is Authorization Code + PKCE.
