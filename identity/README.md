# Identity Infrastructure

The application depends on the provider-neutral `internal/identity.Provider` boundary.

The current deployment uses self-hosted Ory Kratos for identity lifecycle and
credentials. The application API owns registration, login, logout, and session
contracts; the mobile client does not call Ory services directly.

Hydra and the identity gateway remain historical/deferred infrastructure from the
superseded authorization-code design. They are not part of the active MVP runtime.

No Ory SDK or provider-specific type belongs in application business domains.

## MVP boundary

This repository contains the active Kratos deployment boundary and identity schema.
Provider-specific behavior remains isolated to backend adapters.
