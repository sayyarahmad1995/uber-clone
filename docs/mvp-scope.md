# MVP Scope

## Goal

Validate the core ride-hailing experience with a small, understandable system. The MVP is not an enterprise platform and should not carry speculative business complexity.

## Included foundation

- One user account per person.
- External OIDC authentication.
- Rider as the default capability.
- Shared Flutter mobile application.
- Android as the initial client platform.
- Go backend.
- PostgreSQL persistence.
- Redis only where a concrete vertical slice needs fast/ephemeral state.

## Core MVP journey

The MVP will be built as vertical slices toward this flow:

```text
Open app
  ↓
Authenticate
  ↓
Load user account
  ↓
Enter Rider capability
  ↓
Set pickup
  ↓
Set destination
  ↓
Propose fare
  ↓
Request ride
  ↓
Eligible Drivers receive/discover the request
  ↓
Driver accepts proposed fare OR sends counteroffer
  ↓
Rider selects an offer (original fare or counteroffer)
  ↓
Trip assigned
  ↓
Trip in progress
  ↓
Live location/status updates
  ↓
Trip completed
  ↓
Trip history
```

There is one Rider ride-request experience. The Rider does not select an `automatic` or `offers` booking mode. Migration 016 retires the legacy booking-mode persistence.

Driver commercial response is the differentiator:

- accepting the Rider proposed fare creates a pending offer at that fare;
- submitting a counteroffer creates a pending offer at a different allowed fare;
- either response requires Rider selection before assignment.

Geographic matching is marketplace policy used to determine which Drivers are eligible to receive/discover a request and how requests/Drivers are ranked. It is not a second Rider booking mode.

## Driver onboarding required by the MVP

Driver onboarding is part of the ride-hailing MVP because marketplace validation requires real Drivers with an approved operating context.

Initial onboarding applies one vehicle to one Driver-selected service:

```text
Become a Driver
  ↓
Enter Driver details
  ↓
Choose one service
  ↓
Review service requirements
  ↓
Enter vehicle details
  ↓
Pre-eligibility check
  ↓
Review application
  ↓
Submit for review
  ↓
Approved / Rejected with reason
```

A Driver may register multiple vehicles. After a service is approved for a vehicle, the Driver may apply the same vehicle to additional services without registering the vehicle again.

For the MVP, going online uses one selected verified vehicle and one selected approved service at a time. The backend remains authoritative for vehicle verification, service approval, location freshness, and marketplace eligibility.

The exact slice boundaries and data model are defined as implementation reaches each step, but they must follow ADR-0009 rather than the original single-vehicle onboarding assumption.

## Explicitly deferred

The MVP does not include complete implementations for:

- Courier capability
- Freight capability
- Other future capabilities
- Full administrator operations
- Enterprise dispatch algorithms
- Advanced surge/pricing systems
- Promotions/referrals/incentive systems
- Complex fraud/risk platforms
- Multi-region architecture
- Microservices
- Advanced analytics/data pipelines
- Elaborate audit/compliance platforms
- Sophisticated payment infrastructure unless a concrete MVP payment requirement is introduced
- Arbitrary fixed pickup-radius/service-area policy without a concrete launch requirement or operating data
- The complete Driver/vehicle revision review, cancellation-window, withdrawal-appeal, and administration UI; approved-information non-overwrite semantics remain an architecture requirement meanwhile

## Scope rule

A feature enters the MVP only when it is necessary to validate the current product flow or is a concrete dependency of a vertical slice.

Future extensibility should come primarily from clean business boundaries, not by implementing unused future functionality.

Accepted architecture decisions are authoritative for future slices. Implementation must not introduce a new Rider product mode, change Driver service/vehicle semantics, or materially change the marketplace flow without first updating or superseding the relevant architecture decision.
