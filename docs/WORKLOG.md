# Work Log

## Purpose

This file tracks meaningful project progress and the current stopping point so work can resume without reconstructing previous decisions.

Update this file after every meaningful work session.

---

## Current Status

**Current engineering milestone:** User Entry and Rider Foundation — In Progress

**Current work item:** Implement the Ory adapter behind the completed provider-neutral external OIDC boundary.

**Current MVP planning approach:** Define the current milestone precisely, keep the next business milestone reasonably clear, and intentionally leave later slices flexible until we learn from completed work.

---

## Completed

- [x] Repository created: `sayyarahmad1995/uber-clone`
- [x] MVP-first development approach established
- [x] Enterprise-level business logic deferred
- [x] Project organization will be based on business domains, not global technical layers
- [x] Shared client application approach established
- [x] One user account can have multiple capabilities
- [x] Rider is the default capability
- [x] Driver may be added later
- [x] Courier capability deferred for future implementation
- [x] Freight capability deferred for future implementation
- [x] Android is the initial client platform
- [x] iOS support is deferred, with platform-specific boundaries kept clear where needed
- [x] Flutter selected as the client framework
- [x] Go selected for the backend
- [x] PostgreSQL selected as the primary database
- [x] Redis selected for required fast/transient data needs
- [x] External OIDC selected for authentication
- [x] Self-hosted Ory selected as the initial OIDC provider
- [x] Provider-neutral external OIDC boundary selected
- [x] Modular monolith selected as the initial backend architecture
- [x] Docker Compose selected as the current deployment mechanism
- [x] CI/CD explicitly deferred
- [x] Kubernetes and other orchestration platforms are not part of the current implementation
- [x] Deployment Foundation selected as the first engineering milestone
- [x] Deployment Foundation classified as a foundational engineering milestone, not a business vertical slice
- [x] First business vertical slice defined as User Entry and Rider Foundation
- [x] Later MVP slices intentionally kept flexible rather than fully designed upfront
- [x] Administrative operations deferred
- [x] Initial decision and scope documentation created
- [x] Work tracking added
- [x] Deployment Foundation implemented
- [x] Deployment Foundation runtime-verified locally with Docker Compose
- [x] PostgreSQL container verified healthy
- [x] API container verified running
- [x] `GET /health` verified successfully
- [x] Deployment Foundation merged into `main`

---

## MVP Delivery Direction

### Engineering Foundation

**Deployment Foundation**

Goal:

- Establish a minimal deployable system.
- Use Docker Compose.
- Make deployment repeatable with:
  `docker compose up -d --build`
- Verify the running system with a health endpoint.

This is a technical foundation, not a business vertical slice.

### Next Business Vertical Slice

**User Entry and Rider Foundation**

Expected end-to-end outcome:

`Open app → Authenticate → User account is created or retrieved → Rider capability exists → Enter the initial Rider experience`

This slice should be refined before implementation. It is not necessary to design all later MVP slices now.

### Beyond the Next Slice

The rough product direction includes:

- Creating a ride request
- Introducing the minimum driver participation required for the ride flow
- Basic matching
- Trip execution

The exact boundaries, order, and scope of these later slices remain intentionally flexible and will be refined based on completed work and emerging requirements.

---

## In Progress

Active branch: `feature/user-entry-rider-foundation`.

Implemented so far:

- PostgreSQL-backed user domain.
- User domain separated from external identity records.
- External identity keyed by OIDC issuer + subject.
- Provider-neutral identity provider interface and HTTP authentication boundary.
- Temporary identity header endpoint removed.
- Default Rider capability created idempotently.
- Capability storage designed to support future capabilities without separate accounts.
- Database migration runner.
- API readiness endpoint.
- Temporary integration endpoint for exercising user provisioning.

The application now has a provider-neutral identity boundary. The next implementation is the Ory adapter behind that boundary; Ory-specific code must not enter the User domain.

---

## Completed Deployment Verification

Verified successfully using:

`docker compose up -d --build`

Observed result:

- PostgreSQL became healthy.
- API started successfully.
- `curl http://localhost:8080/health` returned `{"status":"ok"}`.

---

## In Progress

- [x] Define the minimal repository structure for the Deployment Foundation
- [x] Create the minimal Go backend
- [x] Add the health endpoint
- [x] Containerize the backend
- [x] Add Docker Compose configuration
- [ ] Verify PostgreSQL starts through Docker Compose
- [ ] Verify the project starts with `docker compose up -d --build` in a Docker-capable environment

---

## Immediate Next Step

Implement the minimum self-hosted Ory components and adapter behind the provider-neutral identity boundary, then connect authenticated principals to user provisioning.

---

## Deployment Foundation Goal

The project must be runnable with:

```bash
docker compose up -d --build
```

The initial deployment must start:

- Go backend
- PostgreSQL

The backend must expose a minimal health endpoint.

This milestone is a deployable technical foundation, not a business vertical slice.

---

## Deferred

- CI/CD
- Kubernetes
- iOS implementation
- Driver implementation
- Courier capability
- Freight capability
- Administrator operations
- Enterprise-level business logic
- Advanced ride matching
- Payments
- Promotions
- Advanced analytics

---

## Important Working Principles

- Build the MVP incrementally.
- Business functionality should be developed as vertical slices.
- Do not fully design all future vertical slices upfront.
- Define the current milestone precisely.
- Keep the next milestone reasonably clear.
- Refine later slices when earlier work and product learning justify it.
- Establish a deployable foundation before implementing business features.
- Keep the MVP simple.
- Avoid premature enterprise complexity.
- Organize code primarily around business domains.
- Maintain boundaries for future capabilities without implementing unused functionality.
- Challenge architectural and technical decisions when they are weak, premature, or inconsistent with the project goals.
- Record meaningful progress and the current stopping point in this file.
