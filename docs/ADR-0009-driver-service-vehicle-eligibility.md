# ADR-0009: Driver Service and Vehicle Eligibility Model

## Status

Accepted.

## Context

The existing Driver onboarding implementation combines Driver profile creation and one vehicle record. That model does not support the required product behavior:

- one account may own the Driver capability;
- one Driver may register multiple vehicles;
- a vehicle may qualify for multiple ride services;
- the Driver should choose the service they want to join rather than having administration assign one implicitly;
- only approved service/vehicle combinations may be used operationally;
- Driver and vehicle information changes require controlled approval and history rather than immediate replacement of approved data.

Ride services such as Economy or Comfort are not account capabilities. `Driver` remains the account capability. Services are products offered by a Driver using a specific vehicle.

## Decision

### 1. Driver onboarding selects one service

Initial onboarding follows this product flow:

```text
Become a Driver
  ↓
Driver capability enabled
  ↓
Enter Driver details
  ↓
Choose one desired service
  ↓
Review service requirements
  ↓
Enter vehicle details
  ↓
Deterministic pre-eligibility check
  ↓
Review application
  ↓
Submit for review
  ↓
Approved / Rejected with reason
```

The pre-eligibility check may reject combinations that clearly violate published deterministic rules. It does not replace final verification or review where documents, condition, inspection, media, or judgment are required.

### 2. Driver capability and ride services are different concepts

`Driver` is a business capability attached to the user account.

A ride service is an operational product classification attached to a Driver vehicle through service eligibility/enrollment state. Capability selection must not be used to represent Economy, Comfort, Premium, or similar ride services.

### 3. A Driver may own multiple vehicles

A Driver may register multiple vehicles. Vehicle ownership is one-to-many from Driver to Vehicle.

Vehicle identity and verification are independent from service eligibility. A vehicle may be verified while being approved for one service, eligible but not enrolled in another, pending review for another, or ineligible for another.

Conceptually:

```text
Driver
  └── Vehicle
       ├── Vehicle verification
       └── Service eligibility
            ├── Economy  — approved
            ├── Comfort  — pending
            └── Premium  — ineligible
```

### 4. Initial onboarding applies for one service only

The Driver chooses one service during initial onboarding. Approval establishes enrollment for that specific vehicle/service combination.

After at least one service is approved for the vehicle, the Driver may apply for additional services for the same vehicle without registering the vehicle again.

### 5. Service hierarchy may infer eligibility, not enrollment

A service catalog may explicitly define that one service's requirements are a strict superset of another service's requirements.

When that relationship exists, approval for the stricter service may establish technical eligibility for the less restrictive service using the already-approved evidence. It must not silently enroll the Driver in that service.

Example:

```text
Comfort — approved
Economy — eligible, not enrolled
```

The Driver must explicitly choose to add Economy. If current approved evidence satisfies all Economy requirements, that later enrollment may use a lightweight or automatic approval path. Otherwise it enters the normal review flow.

### 6. Service decisions belong to vehicle/service combinations

Approval or rejection of one service does not automatically approve another service unless the service rules explicitly permit derived eligibility.

Rejection should use structured reasons where possible, including the selected service, failed requirement, actual value/evidence, required value/evidence, and eligible alternatives when known.

### 7. Online operation uses one vehicle and one approved service

For the MVP, an online Driver operates with exactly one selected vehicle and one selected approved service at a time.

The client presents only vehicles and services that the backend confirms are operationally eligible. Selection is server-authoritative.

Conceptually:

```text
Select verified vehicle
  ↓
Select one approved service for that vehicle
  ↓
Publish current location
  ↓
Go online
```

If the selected vehicle has only one approved service, the client may omit the redundant service selector.

Changing the operating vehicle or service requires the Driver to go offline first. Marketplace discovery, offers, assignment, and Trip history must preserve the vehicle/service context that applied at the time; later selection changes must not rewrite historical associations.

### 8. Operational eligibility is broader than online state

Marketplace eligibility requires all applicable invariants, including:

- Driver capability;
- active/approved Driver state;
- selected verified vehicle;
- selected approved service enrollment for that vehicle;
- online state;
- fresh location;
- no conflicting active Trip or other applicable commitment.

Being online alone never implies marketplace eligibility.

### 9. Approved details remain authoritative

Driver and vehicle detail editing must not immediately replace approved information.

The long-term governance model is:

```text
Approved version
  ↓
Local draft
  ├── Discard
  └── Submit
        ↓
Submitted revision
        ├── Cancel during server-owned cancellation window
        └── Cancellation window expires
              ↓
              Withdrawal requires appeal
```

Approval makes a submitted revision the new current approved version. Rejection, cancellation, or approved withdrawal leaves the previous approved version authoritative. Historical revisions and decisions are retained.

The server owns cancellation deadlines and approval state. The client must not infer approval or mutate approved state locally.

The concrete revision tables, review tooling, appeal workflow, and administration UI are deferred until that vertical slice is implemented. This ADR defines the product invariant now so current client/backend work does not introduce immediate-overwrite semantics.

## Consequences

- The current one-vehicle database constraint must be removed before deeper marketplace client work.
- Driver onboarding APIs must represent a selected service and a vehicle/service application rather than treating one vehicle as the permanent Driver profile.
- Service catalog and eligibility rules require an application-owned boundary.
- Vehicle verification and service enrollment remain distinct states.
- Driver dashboards should keep operational vehicle/service selection separate from profile and vehicle administration.
- Future Driver/vehicle edit flows must submit revisions rather than overwriting approved records.
- Full administrator operations, sophisticated rules engines, and generalized compliance infrastructure remain deferred until required by a concrete slice.
