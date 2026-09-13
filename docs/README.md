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
- Treat ride services as Driver/vehicle products, not account capabilities.
- Let a Driver register multiple vehicles and apply each vehicle to one or more services.
- Initial Driver onboarding applies one vehicle to one Driver-selected service; later service applications reuse the same vehicle.
- Operate online with one verified vehicle and one approved service at a time for the MVP.
- Keep approved Driver/vehicle information authoritative until a submitted revision is approved.
- Use a separate, narrow operator boundary for Driver onboarding approval/rejection; administrator review is not a Rider/Driver capability.
- Use one shared client application.
- Build Android first with clear boundaries for a later iOS implementation.
- Use an external OIDC provider for authentication and authorization.
- Use PostgreSQL for primary persistence and Redis only where a concrete slice needs fast/ephemeral state.
- Deploy containerized on a self-managed server and scale vertically.
- Use one Rider Ride Request flow: pickup, destination, and proposed fare.
- Let eligible Drivers either accept the Rider proposed fare or submit a counteroffer.
- Treat geographic logic as marketplace eligibility/distribution/ranking policy, not a Rider-selected booking mode.
- Close the MVP ride loop with minimal cash settlement and receipt display before introducing any sophisticated payment infrastructure.

## Documents

- [Architecture Decisions](architecture-decisions.md)
- [ADR-0007: Unified Ride Request Marketplace Model](ADR-0007-ride-request-marketplace-model.md)
- [ADR-0008: Dashboard Panel Interaction Contract](ADR-0008-dashboard-panel-interaction-contract.md)
- [ADR-0009: Driver Service and Vehicle Eligibility Model](ADR-0009-driver-service-vehicle-eligibility.md)
- [ADR-0010: Minimal Driver Onboarding Reviewer](ADR-0010-minimal-driver-onboarding-reviewer.md)
- [ADR-0011: Minimal Cash Settlement and Trip Receipt](ADR-0011-minimal-cash-settlement-and-receipt.md)
- [Driver onboarding reviewer runbook](driver-onboarding-reviewer.md)
- [Product and Capability Model](product-and-capability-model.md)
- [Technology Stack](technology-stack.md)
- [MVP Scope](mvp-scope.md)
- [Work Log](WORKLOG.md)

These are living documents, but architectural changes must be explicit rather than emerging accidentally from implementation details.
