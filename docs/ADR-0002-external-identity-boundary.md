# ADR-0002: Provider-Neutral External OIDC Boundary

## Status

Accepted.

## Decision

The application will authenticate through its own `identity.Provider` interface.

Business domains receive only:

- issuer
- subject

The initial implementation targets self-hosted Ory infrastructure, but Ory-specific code must remain outside business domains.

Application users are identified independently from external identities. The durable identity mapping is the tuple:

`issuer + subject`

## Consequences

- A provider can be replaced without rewriting the User domain.
- A future user may link more than one external identity.
- Provider-specific operational complexity is isolated.
