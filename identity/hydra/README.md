# Hydra Configuration

Hydra is the OAuth 2.0 / OpenID Connect authorization server.

The application does not call Hydra-specific APIs from business domains. Hydra is an infrastructure implementation behind the provider-neutral OIDC boundary.

For the mobile MVP, the client must use Authorization Code + PKCE and must not embed a client secret in the mobile application.

The login and consent URLs point to an integration service. That service is responsible for connecting Hydra's authorization requests to the selected identity experience.
