# ADR-0006: Replaceable External Providers

## Status

Accepted.

## Decision

External systems implement application-defined ports. Provider-specific concepts must not define business-domain models or public application APIs.

For authentication and identity:

- The mobile client communicates only with application-owned authentication APIs.
- The application owns request/response terminology, error semantics, and business-facing identity concepts.
- Concrete providers such as Ory Kratos live behind internal adapters selected at the composition root.
- Provider-specific transaction IDs, error payloads, HTTP statuses, headers, and implementation terminology remain inside adapters.
- Business domains receive only application identity principals and application users; they do not import provider packages.
- Provider-specific operational settings, including session lifecycle controls, remain deployment/infrastructure configuration and do not become business-domain concepts.

This rule also applies to future external systems such as maps, routing, geocoding, payments, notifications, storage, and messaging.

## Identity source and migration

An external identity is mapped to an application user using an identity source plus the provider subject. The identity source is configuration, not a hard-coded vendor name. The initial application-owned namespace is `primary-identity-v1`.

Replacing an identity provider does not guarantee that provider subjects remain the same. If a future provider assigns different subjects, account continuity requires an explicit migration or account-linking procedure. This is a data migration concern, not a reason to couple business domains to a provider.

Active provider-issued sessions may become invalid during a provider replacement. Preserving active sessions across provider migrations is not an MVP requirement.

## Consequences

- Replacing a provider should primarily require a new adapter and composition/configuration change.
- Public API clients do not need to understand provider-specific concepts.
- Business domains remain stable when infrastructure providers change.
- Provider migration may still require explicit identity-data migration when external subject identifiers change.
- Provider-specific session configuration can evolve independently of public application contracts.
- The project does not build custom security-sensitive infrastructure solely to avoid an external dependency.
