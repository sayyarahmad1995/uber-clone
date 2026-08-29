# ADR-0004: Client Authentication Boundary

## Status

Accepted.

## Decision

The mobile client communicates only with application-owned APIs.

The client does not call Ory Kratos, Ory Hydra, or any future external identity provider directly. The client also does not render provider-owned authentication interfaces.

Authentication screens are product UI owned by the client. Authentication API contracts are owned by the application.

## MVP request boundary

The public API will own endpoints such as:

- registration
- login
- logout
- refresh
- current user

The exact request and response contracts will be implemented inside the Authentication domain.

## Infrastructure boundary

Ory remains an internal implementation detail behind provider adapters.

The database and identity infrastructure are not exposed to the public internet. Docker Compose port publication is limited to the application API for the MVP.

## Abuse protection

The MVP uses pragmatic protection at the public API boundary:

- input validation
- bounded request sizes
- basic rate limiting/throttling
- generic authentication failure responses

Advanced fraud and enterprise security systems are deferred.
