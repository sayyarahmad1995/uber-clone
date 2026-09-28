# Historical Hydra Configuration

This directory preserves the superseded Hydra/PKCE design for historical
reference. Hydra is deferred and is not part of the active MVP runtime.

The superseded design used Hydra as the OAuth 2.0 / OpenID Connect authorization
server.

That design kept Hydra-specific APIs outside business domains behind a
provider-neutral OIDC boundary.

The superseded mobile design used Authorization Code + PKCE and did not embed a
client secret in the mobile application.

The superseded login and consent flow used an integration service to connect
Hydra authorization requests to the selected identity experience.
