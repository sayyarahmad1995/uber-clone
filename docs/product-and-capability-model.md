# Product and Capability Model

## Product model

The product is a shared ride/logistics application built around a single user identity and multiple optional capabilities.

A user has one account. The same account may eventually participate in different business capabilities without creating another account.

## User and capability relationship

```text
User Account
│
├── Rider capability       (default; MVP)
├── Driver capability      (future/next implementation)
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

## MVP capability boundary

The MVP begins with Rider.

The Driver boundary should be preserved because the core ride-hailing flow will eventually require it, but future capability business logic should not be implemented before its corresponding vertical slice.

Courier, Freight, and similar capabilities are future domains. Their boundaries are acknowledged in the product model but they are not part of the MVP implementation.

## Capability selection

Capability selection is an application concern built on top of the user's account and available capabilities. It must not create a second account.

Future examples:

- A Rider can become a Driver by gaining the Driver capability.
- A user with Rider and Driver capabilities can switch between those experiences.
- The same identity, profile, and authentication relationship remain associated with the one account.

## Design rule

Keep the boundary between identity and business capability explicit. Add a new capability as a new business domain and vertical slice rather than cloning the user/account system.
