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

- accept the Rider proposed fare, allowing immediate assignment if the request and Driver are still eligible; or
- submit a counteroffer, which the Rider may accept or reject.

Both paths converge on the same Trip assignment and Trip lifecycle.

This product behavior is defined in [ADR-0007: Unified Ride Request Marketplace Model](ADR-0007-ride-request-marketplace-model.md). Existing implementation-specific `booking_mode` state must not be treated as a Rider-facing capability or product mode.

## MVP capability boundary

Rider and Driver are both required to validate the core ride-hailing journey.

Courier, Freight, and similar capabilities are future domains. Their boundaries are acknowledged in the product model but they are not part of the current MVP implementation.

## Capability selection

Capability selection is an application concern built on top of the user's account and available capabilities. It must not create a second account.

Examples:

- A Rider can become a Driver by gaining the Driver capability.
- A user with Rider and Driver capabilities can switch between those experiences.
- The same identity, profile, and authentication relationship remain associated with the one account.

## Design rule

Keep the boundary between identity and business capability explicit. Add a new capability as a new business domain and vertical slice rather than cloning the user/account system.

Capability selection must not be used to encode ride-request commercial strategy. Accepting the Rider fare versus counteroffering is a Driver response inside the ride marketplace, not a separate account capability or Rider booking mode.
