# Architecture Decisions

## Status

Living document. Decisions are recorded for the MVP and may be revised only when new requirements justify a change. Accepted ADRs and this document are authoritative for implementation until explicitly superseded.

## 1. Vertical development

We will build the product vertically rather than designing and implementing the entire platform upfront.

Each vertical slice should take a usable business flow through the relevant client, API, domain, persistence, and runtime concerns.

We will avoid creating large speculative architecture documents or infrastructure that the current slice does not need.

## 2. MVP complexity

The MVP will not implement enterprise-level business logic merely because a future production system might need it.

We prefer the simplest design that correctly supports the current product requirements while keeping meaningful domain boundaries intact.

## 3. Backend architectural style

The initial backend will be a modular monolith.

Business domains are kept internally separate, but they are deployed as one application for the MVP. We will not introduce microservices until actual scale, ownership, deployment, or domain requirements justify them.

## 4. Project organization

The project is organized by business domain, not by global technical layers.

We do not want top-level structures such as `controllers/`, `services/`, and `repositories/` that scatter one business capability across the application.

Technical implementation concerns can exist inside a domain where needed.

## 5. Identity and capabilities

A person has one user account.

The account can have multiple business capabilities. Capabilities are not separate user accounts.

The default capability is Rider.

The initial product will implement only the capabilities required by the MVP. The architecture must leave a clear boundary for future capabilities such as Driver, Courier, and Freight without implementing those future business domains prematurely.

## 6. Shared client application

There is one shared mobile application for the user account.

The application changes its experience according to the capabilities available to and selected by the user.

The MVP is Android-first. iOS is a future implementation using the same shared application architecture.

Platform-specific behavior must be isolated behind clear boundaries where it is genuinely platform-specific. Business logic should remain platform-independent.

## 7. Authentication

Authentication and authorization will use an external OIDC provider.

We will not introduce a custom OIDC identity provider for the MVP.

## 8. Primary data technologies

PostgreSQL is the primary transactional database.

Redis is used for fast, ephemeral, or cache-oriented data only when a concrete vertical slice needs it.

We will not use Redis as a replacement for transactional persistence.

## 9. Deployment and scaling

The MVP will be deployed in containers on a self-managed server.

The initial scaling strategy is vertical scaling rather than a distributed multi-node architecture.

## 10. Administration

Administrator operations are deferred.

Only administrative functionality that becomes a concrete dependency of an MVP vertical slice will be designed and implemented. A complete enterprise administration system is out of scope for the initial build.

## 11. Ride request marketplace model

The Rider has one ride-request product flow. The Rider provides pickup, destination, and a proposed fare; the Rider does not choose between separate automatic and offers booking modes.

Eligible Drivers receive or discover actionable Ride Requests according to application-owned marketplace eligibility, distribution, and ranking policy.

A Driver may either accept the Rider proposed fare or submit a counteroffer. Both responses create pending offers. Only explicit Rider selection assigns a Trip, after atomic eligibility and availability checks, as defined by ADR-0007.

Migration 016 retired `booking_mode` and candidates. Geographic matching serves marketplace eligibility/distribution/ranking and Rider choice within the one ride-request flow.

See [ADR-0007: Unified Ride Request Marketplace Model](ADR-0007-ride-request-marketplace-model.md) for the authoritative decision and migration rules.

## 12. Documentation authority

Architecture and product documents are part of the implementation contract.

Future implementation must be checked against the accepted ADRs, this document, the product/capability model, MVP scope, and current worklog before a slice is designed. A code change must not silently redefine a documented business flow.

When implementation and documentation conflict, determine whether the code drifted or the requirement truly changed. If the product requirement changed, update or supersede the relevant architecture decision explicitly before building deeper dependencies on the new model.

## 13. Shared dashboard interaction

Rider and Driver map-first surfaces use one shared dashboard panel interaction
contract. Capability features supply content and business state without redefining
panel sizes, direct finger tracking, drag ownership, scroll locking, or
release-gated snap behavior.

See [ADR-0008: Dashboard Panel Interaction Contract](ADR-0008-dashboard-panel-interaction-contract.md)
for the authoritative state machine and change-control rule.
