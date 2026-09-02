# Uber Clone — Architecture Documentation

This directory contains the living architecture and product documentation for the HiGO ride-hailing MVP.

The documentation is intentionally incremental. We record decisions that are useful for the current vertical slice and defer enterprise-level design until it is justified by the product.

## Authority

Accepted ADRs and the architecture/product documents in this directory are part of the implementation contract.

Future implementation must be checked against these documents before a slice is designed. Code must not silently redefine a documented product flow. When a requirement genuinely changes, update or supersede the relevant architecture decision explicitly before building deeper dependencies on the new model.

## Current approach

- Build vertically, feature by feature.
- Structure the system by business domain, not by global technical layers.
- Use a modular monolith for the MVP backend.
- Keep the MVP intentionally small; do not implement enterprise business logic prematurely.
- Use one user account with multiple capabilities.
- Rider is the default capability; Driver is part of the MVP ride-hailing journey.
- Use one shared client application.
- Build Android first with clear boundaries for a later iOS implementation.
- Use an external OIDC provider for authentication and authorization.
- Use PostgreSQL for primary persistence and Redis only where a concrete slice needs fast/ephemeral state.
- Deploy containerized on a self-managed server and scale vertically.
- Use one Rider Ride Request flow: pickup, destination, and proposed fare.
- Let eligible Drivers either accept the Rider proposed fare or submit a counteroffer.
- Treat geographic logic as marketplace eligibility/distribution/ranking policy, not a Rider-selected booking mode.

## Documents

- [Architecture Decisions](architecture-decisions.md)
- [ADR-0007: Unified Ride Request Marketplace Model](ADR-0007-ride-request-marketplace-model.md)
- [Product and Capability Model](product-and-capability-model.md)
- [Technology Stack](technology-stack.md)
- [MVP Scope](mvp-scope.md)
- [Work Log](WORKLOG.md)

These are living documents, but architectural changes must be explicit rather than emerging accidentally from implementation details.
