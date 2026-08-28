# Uber Clone — Architecture Documentation

This directory contains the living documentation for the Uber-like MVP.

The documentation is intentionally incremental. We record decisions that are useful for the current vertical slice and defer enterprise-level design until it is justified by the product.

## Current approach

- Build vertically, feature by feature.
- Structure the system by business domain, not by global technical layers.
- Use a modular monolith for the MVP backend.
- Keep the MVP intentionally small; do not implement enterprise business logic prematurely.
- Use one user account with multiple capabilities.
- Rider is the default capability.
- Driver is a future/next capability; Courier, Freight, and similar capabilities are future extensions.
- Use one shared client application.
- Build Android first with clear boundaries for a later iOS implementation.
- Use an external OIDC provider for authentication and authorization.
- Use PostgreSQL for primary persistence and Redis where fast/ephemeral data is needed.
- Deploy containerized on a self-managed server and scale vertically.

## Documents

- [Architecture Decisions](architecture-decisions.md)
- [Product and Capability Model](product-and-capability-model.md)
- [Technology Stack](technology-stack.md)
- [MVP Scope](mvp-scope.md)

These documents are living documents and should be updated when an architectural decision changes.
