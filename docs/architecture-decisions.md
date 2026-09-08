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

Ride services such as Economy or Comfort are not account capabilities. They are operational service products that a Driver may offer using a specific approved vehicle/service enrollment.

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

A complete administrator product remains deferred.

Only administrative functionality that becomes a concrete dependency of an MVP vertical slice will be designed and implemented. Driver onboarding approval/rejection is now such a dependency, so ADR-0010 authorizes a narrow internal reviewer surface for that workflow only.

Administrator review is not modeled as a Rider/Driver account capability. The current reviewer authentication mechanism is temporary and internal; broader administrator identity, role management, analytics, support tooling, and unrelated operations remain out of scope.

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

The user's committed panel extent is shared presentation state for the signed-in app
session. Expanded stays expanded across panel-content/business-state changes and
Rider/Driver capability switches; collapsed stays collapsed using the target
capability's own collapsed size. Content changes still reset panel content scroll and
invalidate active gestures. Explicit logout resets the shared extent so a new session
starts collapsed.

See [ADR-0008: Dashboard Panel Interaction Contract](ADR-0008-dashboard-panel-interaction-contract.md)
for the authoritative state machine and change-control rule.

## 14. Driver services, vehicles, and approved information

Driver onboarding, vehicle ownership, service eligibility, online operating context,
and future approved-information changes follow ADR-0009.

A Driver may register multiple vehicles. Initial onboarding applies one vehicle for
one Driver-selected service. After approval, the same vehicle may apply for additional
services without being registered again. Vehicle verification and service enrollment
are separate concerns.

For the MVP, an online Driver operates with one selected verified vehicle and one
selected approved service at a time. Marketplace, offer, assignment, and Trip history
must preserve the vehicle/service context that applied when the business action
occurred.

Driver and vehicle edits must not be designed as immediate replacement of approved
information. Approved values remain authoritative until a submitted revision is
approved. Draft discard, a server-owned post-submission cancellation window, and
post-window withdrawal by appeal are accepted product rules; their concrete review
and versioning implementation remains deferred until its vertical slice.

See [ADR-0009: Driver Service and Vehicle Eligibility Model](ADR-0009-driver-service-vehicle-eligibility.md).

## 15. Minimal Driver onboarding reviewer

Driver onboarding approval/rejection uses a dedicated backend review service and a narrow internal reviewer surface defined by ADR-0010.

Approval and rejection are transactional, backend-authoritative decisions. Approval promotes the submitted Driver/vehicle/service snapshot into approved records; rejection requires a reason and creates no operational records. Decided applications remain in history.

Newly approved Drivers remain non-operational until the vehicle/service operating-context slice is implemented. The reviewer slice must not reactivate the legacy immediate-write or legacy online bypass.

See [ADR-0010: Minimal Driver Onboarding Reviewer](ADR-0010-minimal-driver-onboarding-reviewer.md).

## 16. Capability shell navigation

Authenticated secondary navigation belongs to the capability shell, not to the Rider or Driver dashboard panel.

`CapabilityHomeScreen` owns a standard Flutter `Drawer`. Rider/Driver capability switching, enabling Driver access, read-only account/Driver destinations as they become real, settings/history destinations when implemented, and logout belong to this shell-level navigation.

Operational controls stay on their capability dashboard. The Drawer must not contain ride-request controls, Driver availability/location controls, marketplace actions, offers, assignment actions, or Trip controls.

The previous bottom Rider/Driver segmented switch is retired once the Drawer owns capability switching. Drawer entries must correspond to implemented destinations or real actions; placeholder or dead navigation items are not added merely to preview future information architecture.

The Drawer is outside ADR-0008. Adding or changing shell navigation must not modify `RideDashboardScaffold` panel sizing, gesture ownership, scrolling, or snap behavior except through an explicit ADR-0008 contract change.

## 17. Read-only Driver shell destinations

Driver-capable accounts may open `Driver details` and `Vehicles` from the capability shell. These destinations are informational surfaces, not operational dashboard controls and not approved-information editing flows.

The client reads the best currently authoritative Driver information available to it. Existing operational Drivers use the current `DriverProfile`. When the operational Driver read is intentionally unavailable for a pending, approved, or rejected onboarding Driver, the client uses the latest `DriverOnboardingApplication` snapshot instead of fabricating an operational profile.

`Driver details` may show real Driver identity/status, availability when an operational profile exists, selected onboarding service, review reason, and submission/decision dates when those fields exist. `Vehicles` is plural to match the accepted multi-vehicle target, but this slice displays only the real currently available vehicle snapshot. It must not imply add/remove/switch/edit support before the backend multi-vehicle and approved-information workflows exist.

Both surfaces remain read-only. Approved Driver/vehicle changes continue to follow ADR-0009 governance and must not be introduced as direct client-side overwrites.
