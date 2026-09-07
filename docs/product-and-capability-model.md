# Product and Capability Model

## Product model

The product is a shared ride/logistics application built around a single user identity and multiple optional capabilities.

A user has one account. The same account may eventually participate in different business capabilities without creating another account.

## User and capability relationship

```text
User Account
│
├── Rider capability       (default; MVP)
├── Driver capability      (MVP ride-hailing flow)
├── Courier capability     (future)
└── Freight capability     (future)
```

A capability represents what the user can do in the product. It is not a separate identity.

Ride services such as Economy or Comfort are not user capabilities. They are operational products a Driver may offer with a specific vehicle after that vehicle/service combination is approved.

## Shared client

There is one Flutter application shared by all capabilities.

The application has a current capability context. The user initially enters through Rider and can later choose another capability when that capability exists on the account.

Conceptually:

```text
One User Account
       │
       ▼
Available Capabilities
       │
       ▼
Current Capability
       │
       ▼
Capability-specific experience
```

## Ride-hailing product interaction

The Rider has one Ride Request flow:

```text
Rider
  ↓
pickup + destination + proposed fare
  ↓
Ride Request
```

The Rider does not choose between separate automatic and offers products.

Eligible Drivers receive or discover Ride Requests according to application-owned marketplace policy. For each actionable request, a Driver may:

- accept the Rider proposed fare by creating or updating a pending offer at that fare; or
- submit a counteroffer, which the Rider may accept or reject.

Both paths produce offers for Rider selection. Neither Driver response assigns
a Trip. The Rider selects an offer, and the application revalidates the request
and Driver atomically before assigning the Trip.

This product behavior is defined in [ADR-0007: Unified Ride Request Marketplace Model](ADR-0007-ride-request-marketplace-model.md). The legacy booking-mode column and candidate table are retired by migration 016.

## Driver onboarding and service model

A Rider becomes a Driver by gaining the Driver capability on the same account. Driver onboarding then creates the business information required to apply for one ride service with one vehicle.

Initial onboarding is:

```text
Become a Driver
  ↓
Enter Driver details
  ↓
Choose one desired service
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

The Driver chooses the service; administration does not silently assign one from vehicle condition. Deterministic rules may reject obviously incompatible applications before final submission, while final verification/review remains authoritative where evidence or judgment is required.

A Driver may own multiple vehicles. Each vehicle has its own verification state and its own service eligibility/enrollment relationships.

```text
Driver
  └── Vehicle
       ├── Vehicle verification
       └── Services
            ├── Economy  — approved
            ├── Comfort  — pending
            └── Premium  — ineligible
```

Initial onboarding applies for one service only. After one service is approved, the Driver may apply for additional services for the same vehicle without registering the vehicle again.

A service hierarchy may explicitly establish derived technical eligibility. For example, if Comfort requirements are defined as a strict superset of Economy requirements, a Comfort-approved vehicle may be shown as Economy-eligible. The Driver is not enrolled in Economy automatically; the Driver must explicitly add that service. If current approved evidence already satisfies Economy, that later enrollment may use a lightweight or automatic approval path.

## Driver online operating context

For the MVP, an online Driver operates with one selected verified vehicle and one selected approved service at a time.

Conceptually:

```text
Select verified vehicle
  ↓
Select approved service for that vehicle
  ↓
Publish current location
  ↓
Go online
```

If a vehicle has exactly one approved service, the client may omit the redundant service selector.

Changing vehicle or service requires going offline first. Marketplace and Trip records preserve the vehicle/service context used for the relevant offer, assignment, and Trip rather than following later Driver selections.

## Approved Driver and vehicle information

Approved Driver and vehicle information remains authoritative until a submitted revision is approved. Editing is not an immediate overwrite operation.

A local draft may be discarded. After submission, the server provides a cancellation window during which the Driver may cancel the revision directly. After that deadline, direct cancellation is unavailable and withdrawal requires an appeal. Approval, rejection, cancellation, withdrawal, and historical revisions are retained for control and history.

The concrete revision persistence, review UI, appeal workflow, and administrator tooling are deferred until that vertical slice. Current implementation must nevertheless avoid introducing semantics that assume approved information can be overwritten immediately.

See [ADR-0009: Driver Service and Vehicle Eligibility Model](ADR-0009-driver-service-vehicle-eligibility.md).

## MVP capability boundary

Rider and Driver are both required to validate the core ride-hailing journey.

Courier, Freight, and similar capabilities are future domains. Their boundaries are acknowledged in the product model but they are not part of the current MVP implementation.

## Capability selection

Capability selection is an application concern built on top of the user's account and available capabilities. It must not create a second account.

Examples:

- A Rider can become a Driver by gaining the Driver capability.
- A user with Rider and Driver capabilities can switch between those experiences.
- The same identity, profile, and authentication relationship remain associated with the one account.

Ride-service selection is not capability selection. Selecting Economy, Comfort, or another service changes the Driver's service application or operating context, not the user's account capabilities.

## Design rule

Keep the boundary between identity and business capability explicit. Add a new capability as a new business domain and vertical slice rather than cloning the user/account system.

Capability selection must not be used to encode ride-request commercial strategy or Driver service products. Accepting the Rider fare versus counteroffering is a Driver response inside the ride marketplace, not a separate account capability or Rider booking mode.
