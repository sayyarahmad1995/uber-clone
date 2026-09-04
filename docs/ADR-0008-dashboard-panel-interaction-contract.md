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
- The top handle is the only surface that supports continuous panel resizing.
- Releasing an upward handle drag snaps to the maximum extent. Releasing a
  downward handle drag snaps to the capability's collapsed extent.
- Panel-body scrolling is locked at every extent below the maximum and unlocked
  only at the maximum extent.
- From the collapsed state, an upward panel-body pull of at least 56 logical pixels
  arms expansion. The panel remains collapsed until pointer release, then snaps to
  the maximum extent.
- At the maximum extent, content scrolls normally. A downward gesture that merely
  returns content to its top does not minimize the panel. A new downward pull that
  begins at the top and reaches 56 logical pixels arms minimization.
- Armed body transitions execute only on pointer release. Cancelled and
  sub-threshold gestures do not change panel extent.
- Panel transitions use the shared dashboard animation rather than feature-owned
  animations.

Feature panels own their content and business behavior. They receive the shared
content controller and scroll-enabled state; they do not implement independent
panel sizing or gesture policies.

## Change control

Future Rider and Driver slices must reuse this contract. They must not override the
collapsed sizes, maximum extent, gesture ownership, scroll lock, threshold, or
release behavior as a local implementation choice. A different product requirement
must be stated explicitly and this ADR must be updated or superseded before the
shared behavior changes.

## Consequences

- Map visibility and panel behavior remain consistent across capabilities.
- Content controls cannot accidentally resize the panel.
- Users can expand or minimize from the panel body through deliberate,
  release-gated gestures.
- Widget tests protect the state transitions as part of the client architecture.
