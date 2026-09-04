# ADR-0008: Dashboard Panel Interaction Contract

## Status

Accepted.

## Context

Rider and Driver workflows share a map-first dashboard with a floating task panel.
Small differences in panel sizing, drag ownership, scrolling, and snap timing make
the client feel inconsistent and can obscure the map or trigger accidental state
changes. Future vertical slices need one interaction contract rather than defining
their own bottom-sheet behavior.

## Decision

`RideDashboardScaffold` owns the dashboard panel state machine and is the shared
entry point for Rider and Driver task panels.

- Rider panels start collapsed at 18 percent of available dashboard height.
- Driver panels start collapsed at 16 percent of available dashboard height.
- Both expand to a maximum of 60 percent, leaving at least 40 percent of the map
  visible.
- The top handle supports continuous panel resizing throughout its full range.
- Releasing an upward handle drag snaps to the maximum extent. Releasing a
  downward handle drag snaps to the capability's collapsed extent.
- Panel-body scrolling is locked at every extent below the maximum and unlocked
  only at the maximum extent.
- From the collapsed state, an upward panel-body pull moves the panel continuously
  with the pointer, providing direct visual feedback. Crossing the shared intent
  threshold selects expansion; on pointer release, the panel snaps to the maximum
  extent. A shorter pull snaps back to the collapsed extent.
- At the maximum extent, content scrolls normally. A downward gesture that merely
  returns content to its top does not minimize the panel. A new downward pull that
  begins at the top moves the panel continuously with the pointer. Crossing the
  shared intent threshold selects minimization; on pointer release, the panel
  snaps to the collapsed extent. A shorter pull snaps back to the maximum extent.
- Body pulls provide continuous visual movement, but the final state transition
  executes only on pointer release. A cancelled gesture returns to its starting
  extent.
- Panel transitions use the shared dashboard animation rather than feature-owned
  animations.

Feature panels own their content and business behavior. They receive the shared
content controller and scroll-enabled state; they do not implement independent
panel sizing or gesture policies.

## Change control

Future Rider and Driver slices must reuse this contract. They must not override the
collapsed sizes, maximum extent, direct finger tracking, gesture ownership, scroll
lock, snap targets, or release behavior as a local implementation choice. UX tuning
values such as the intent threshold and animation duration remain shared design
constants protected by regression tests; feature slices cannot override them. A
different interaction requirement must be stated explicitly and this ADR must be
updated or superseded before the shared behavior changes.

## Consequences

- Map visibility and panel behavior remain consistent across capabilities.
- Content controls cannot accidentally resize the panel.
- Users see the panel follow their body gesture before a deliberate,
  release-gated snap completes the transition.
- Widget tests protect the state transitions as part of the client architecture.
