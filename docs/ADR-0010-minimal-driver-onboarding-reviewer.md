# ADR-0010: Minimal Driver Onboarding Reviewer

## Status

Accepted.

## Context

ADR-0009 requires Driver onboarding submissions to be reviewed before submitted Driver, vehicle, or service information becomes approved operational data.

The MVP does not yet have a general administrator product. A complete administration platform would add unnecessary scope, but Driver onboarding cannot be validated end to end without an authorized reviewer path.

The reviewer path must not be implemented as a Driver capability, a public Driver endpoint, or direct database editing. Review is an operator responsibility with a separate authorization boundary.

## Decision

### 1. Build only the reviewer capability required by Driver onboarding

The current administration slice is limited to:

- listing pending Driver onboarding applications;
- viewing one submitted application snapshot;
- approving a pending application;
- rejecting a pending application with a required reason.

User management, ride operations, analytics, pricing, promotions, support tooling, generalized role management, and other administrator features remain out of scope.

### 2. Review uses a dedicated operator boundary

Administrator review is not an account capability such as Rider or Driver.

For the current internal MVP, reviewer routes are enabled only when both `ADMIN_REVIEW_USERNAME` and `ADMIN_REVIEW_PASSWORD` are configured. The backend protects the reviewer UI and reviewer API with HTTP Basic authentication.

This credential mechanism is intentionally temporary and internal. It is suitable for local development or an HTTPS-protected internal deployment. It must be replaced by a proper administrator identity/authorization mechanism before a remotely exposed production administration surface is introduced.

The review business service is independent of HTTP Basic authentication so the authentication mechanism can be replaced without rewriting review rules.

### 3. Review decisions are backend-authoritative and transactional

Approval and rejection are business operations, not raw database edits.

The backend locks the pending application and records exactly one terminal decision. An application that is no longer pending cannot be decided again through the reviewer surface.

Approval atomically:

1. verifies the application is still pending;
2. creates the approved Driver profile from the submitted snapshot;
3. creates the approved vehicle record from the submitted snapshot;
4. creates an approved service enrollment for that vehicle and selected service;
5. records application status `approved`, decision time, and reviewer identity.

Rejection atomically:

1. verifies the application is still pending;
2. requires a non-blank rejection reason;
3. creates no Driver, vehicle, or service enrollment;
4. records application status `rejected`, rejection reason, decision time, and reviewer identity.

### 4. Approval does not yet enable online operation

The current operational Driver path predates ADR-0009 and does not yet select a verified vehicle and approved service before going online.

Therefore a newly approved onboarding application creates a Driver profile with status `approved`, not `active`. It is visible as approved data but cannot go online or enter marketplace discovery through the existing `active`-only operational checks.

A later multi-vehicle/service operational slice will introduce backend-authoritative vehicle/service selection and the transition to operational `active` state.

Existing legacy `active` Drivers remain compatible until that migration is implemented.

### 5. Approved service enrollment is explicit

Approval creates an explicit vehicle/service enrollment record. Presence of that record means the vehicle has been approved for that selected service.

Approval of a stricter service does not silently enroll the vehicle in an implied lower service. Derived technical eligibility remains separate from enrollment, as defined by ADR-0009.

### 6. Provide both a small reviewer API and a small internal UI

The backend exposes a protected reviewer API for the business operations and a minimal server-rendered reviewer page that uses the same review service.

The UI is intentionally plain. It exists to complete the current product workflow and physical-device validation, not to establish a permanent administration frontend architecture.

### 7. Decision history is retained

Approved and rejected onboarding application rows remain in history. The reviewer surface does not delete or rewrite decided applications.

Later versioned Driver/vehicle edit review, cancellation windows, withdrawal appeals, and broader audit requirements remain governed by ADR-0009 and are not implemented by this slice.

## Consequences

- Driver onboarding can now be exercised end to end without direct SQL or a temporary Driver-accessible approval endpoint.
- The review domain remains reusable when a future administrator identity system or dedicated administration frontend is added.
- Newly approved Drivers intentionally remain non-operational until vehicle/service online selection is implemented.
- The reviewer surface must remain narrow; adding unrelated administration features requires a concrete product dependency and an explicit scope decision.
