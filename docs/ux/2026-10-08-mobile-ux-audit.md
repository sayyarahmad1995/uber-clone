# Mobile UI/UX audit and proposed journey structure

Date: 2026-10-08
Status: audit complete; proposed structure awaits owner review.
Baseline: main at 3a9fdfa2ca9305d20102ed1b5a61e84129452bfc (PR #132 merged).
Scope: shared Flutter Rider/Driver app. Operations/admin redesign is excluded.

## Brief

The owner dislikes the overall MVP interface and wants a coherent Rider and
Driver experience. Preserve the working marketplace, pricing, trip, cash
settlement and Google Maps handoff behavior. This phase audits the current
implementation and proposes information structure; it does not approve a
visual style, a final design specification, or product-code changes.

Success: users can identify their ride stage, relevant price, location and next
action without opening unrelated management sections. Reduce competing controls
and repeated information, preserve recoverable failures, and make core actions
usable with small screens, larger text and the keyboard.

## Evidence and limits

This is a source-based structural audit, not a screenshot or live-device visual
inspection. Hierarchy and content findings are grounded in widget order, labels,
state models and shared contracts. Contrast, actual truncation, perceived polish,
gesture comfort and task speed still require rendered/device validation.
No numerical usability improvement is claimed.

Primary source files:
- mobile/lib/features/rider_request/presentation/rider_request_screen.dart
- mobile/lib/features/driver_workspace/presentation/driver_workspace_screen.dart
- mobile/lib/features/ride_flow/ride_flow_panels.dart
- mobile/lib/features/capabilities/presentation/capability_home_screen.dart
- mobile/lib/features/driver_workspace/presentation/operating_selection.dart
- mobile/lib/features/driver_workspace/presentation/vehicle_service_application_section.dart
- mobile/lib/features/driver_workspace/presentation/driver_readonly_surfaces.dart
- mobile/lib/core/dashboard/ride_dashboard_scaffold.dart
- mobile/lib/core/theme/app_theme.dart
- mobile/lib/features/ride_flow/domain/{trip,marketplace_request,ride_execution}.dart
- docs/ADR-0008-dashboard-panel-interaction-contract.md

Some historical docs describe editable suggested fares, fixed service choices
or superseded location behavior. They must not drive the redesign. Current
implementation and the later owner decisions are authoritative: the priced
suggestion is read-only, and the selected Driver offer is the agreed fare.
Read the current contracts before defining labels or movement between screens.

## Current surface inventory

| Surface/state | Current content and actions | Structural finding |
|---|---|---|
| Sign-in/registration | Email/password and registration switch | Supporting flow; visual QA later, no authentication replacement |
| Email verification | Code, verify, resend | Keep pending/error/retry states |
| Capability shell | Drawer for Rider, Driver, Become a Driver, Driver details, Vehicles, logout | One account already has separate capabilities; history is embedded elsewhere |
| Rider booking | History, instruction, service picker, pickup/destination selector, search fields, map selection, current-location action, duplicate point summaries, route choices, fare field, request action | A long mixed-purpose panel, with service choice before place fields and history near the heading |
| Rider open request/offers | Request summary, proposed fare, refresh/cancel, offer cards, Driver/vehicle/fare/distance, choose/decline | Repeated summaries and actions compete with offer comparison |
| Rider assigned/in-progress | Status, Driver marker, operation details, settlement and location-update text | Information is text-heavy; arrival ETA is absent and must not be invented |
| Rider history | Expandable recent rides with operation and settlement details | History interrupts booking instead of having its own destination |
| Driver onboarding/review | Service requirements, Driver/vehicle form, review, pending/rejected state | Distinct management journey currently shares the workspace |
| Driver idle/offline/online | Profile, operating vehicle, additional application, availability, requests, history, published-location text, manual update/refresh | Setup, administration and daily driving compete in one scrollable panel |
| Driver request review | Fare, coordinate endpoints, straight-line pickup distance, accept fare/counteroffer/decline | Coordinates are hard to interpret; accepting a fare is an offer, not assignment |
| Driver assigned | Navigation, active-trip heading, operation/fare/settlement, endpoints, start/cancel, history, availability/location tools | Navigation and lifecycle actions are separated by other information |
| Driver in progress | Full-trip route estimate, navigation, complete/cancel, operation details | Full-route duration is not remaining time from the moving Driver |
| Driver completed/cash due | Completion state, agreed fare and confirm-cash action | Completion and cash collection are distinct states and both must remain recoverable |
| Driver details/vehicles | Read-only profile, review and approved vehicle/service information | Existing dedicated management destinations can be extended rather than duplicating dashboard content |

## Prioritized findings

| Priority | Evidence | Proposed response |
|---|---|---|
| P1 | Booking and driving share their panels with history/setup/diagnostics | Separate current activity from account, vehicle and history destinations |
| P1 | Collapsed panels are 18% Rider / 16% Driver; content scrolls only at 60% | Design a useful compact state with stage/location/price/action; explicitly review ADR-0008 if fixed sizes prevent it |
| P1 | Status card and panel repeat status or send users elsewhere to find it | Establish one authoritative activity summary and one deliberate action hierarchy |
| P1 | Trip/request location display uses latitude/longitude | Specify a truthful location-label strategy before drawing mockups with street names |
| P1 | Fare display remains a read-only TextField for priced requests | Present it as a price summary with a short explanation of offer-selected final fare |
| P1 | Driver response can be mistaken for getting the ride | Label submission as an offer; waiting for Rider selection is an explicit presentation state |
| P1 | Trip completion is followed by cash collection | Give cash due its own screen/state; do not send the Driver directly to idle |
| P2 | Multiple refresh controls and published-location timestamps | Put refresh in contextual recovery; keep technical diagnostics out of the default activity view |
| P2 | Repeated point summaries, service details and operational text | Show only stage-relevant facts; place secondary details behind a deliberate disclosure |
| P2 | Shared tokens exist but typography/component state rules are incomplete | Build on AppTheme; define type scale, semantic components and loading/error/disabled variants |
| P2 | Lifecycle/polling behavior resides partly in RideFlowPanel | Preserve foreground refresh and recovery when panels are relocated or replaced |

These are design priorities, not severity ratings for production defects.

## Proposed information architecture

Keep one app/account and an explicit Rider/Driver switch. Use the current
capability as the main activity destination. Put history, account and Driver
vehicle/service management in secondary destinations; drawer versus tabs is
a later visual/navigation decision, not selected in this audit.

Keep a map where location or movement helps the task. Place search, offer
comparison, onboarding and receipts may use focused surfaces rather than
forcing every task into a small map sheet. Active driving keeps a map with
a compact stage-specific trip summary.

### Rider journey

| Presentation state | Essential content | Main action |
|---|---|---|
| Ready to book | Destination entry, pickup summary, current-location option | Choose destination |
| Choose locations | Place search/results, selected field, map selection alternative | Confirm pickup/destination |
| Review request | Pickup/destination, selected service, route choice/estimate, read-only suggested fare | Request ride |
| Waiting for responses | Request summary, waiting state, recoverable status | No forced primary action; cancel is secondary |
| Compare offers | Comparable Driver/vehicle, offered price, truthful pickup distance, availability | Select an offer and confirm its fare |
| Driver assigned | Driver/vehicle/plate, pickup, agreed fare, fresh location or unavailable state | No invented action; cancellation is secondary |
| Trip in progress | Destination, agreed fare, trip status, available map context | No invented action; retain permitted cancellation |
| Completed/cancelled | Outcome, agreed fare, truthful settlement, receipt/history | Return to booking |

These are UI states; waiting/offers/location selection do not introduce new
backend trip statuses. Actual request expiry, changed offers, no services,
pricing conflicts and errors must branch to recoverable views.

### Driver journey

| Presentation state | Essential content | Main action |
|---|---|---|
| Setup/review required | Eligibility/application progress and next required step | Finish setup or inspect review |
| Offline/ready | Selected verified vehicle, active services, readiness blockers | Go online |
| Online/waiting | Availability, location readiness, unobtrusive selected vehicle | No fabricated request; go offline remains available |
| Review a request | Pickup/destination, Rider request fare, straight-line distance, deadline where supported | Send fare offer; counteroffer is an explicit alternative |
| Offer pending | Submitted fare and waiting-for-selection message | No false assignment; preserve existing response/decline semantics |
| Assigned/heading to pickup | Saved pickup, road distance/duration estimate, agreed fare | Navigate to pickup; Start trip is separately available with existing confirmation |
| Trip in progress | Destination, agreed fare, clearly labeled full-route estimate | Navigate to destination; Complete trip is separately available with existing confirmation |
| Completed/cash due | Agreed amount and cash outstanding | Confirm cash collected |
| Cash confirmed/cancelled | Outcome and receipt/history | Return to appropriate availability state; never auto-resume presence |

One primary action does not mean hiding essential lifecycle controls. Navigation
and start/complete must both be reachable, with clear prominence and separation.
No new Arrived, chat, call, rating, wallet or payment behavior is assumed.

## Data and behavior constraints

- Marketplace requests and Trip endpoints carry coordinates; readable place
  names are not present in their reviewed models. Existing Rider place-search
  labels are local presentation and may not survive recovery. Options to assess:
  truthful coordinate fallback, bounded reverse-geocoding/cache, or persisted
  endpoint-label snapshots. Costs, provenance and privacy need a separate
  decision before adding backend/provider work.
- Never relabel straight-line pickup distance as driving distance or derive
  arrival time by guessing speed. Browsing-request ETA and live Rider arrival
  ETA remain deferred features.
- The assigned Driver route estimate uses Driver-location-to-pickup. In-progress
  estimate is the full pickup-to-destination route, not a live remaining ETA.
- Keep the external Google Maps handoff accepted by the owner.
- Preserve dynamic service eligibility, read-only priced suggestion, pricing
  refresh/resubmission, immutable selected-offer fare and stale-offer checks.
- Do not expose pre-assignment license plates or invent missing Driver ratings,
  rider names, contact actions or unsupported vehicle/profile edits.
- Keep authoritative transitions, confirmations, cash due/collected distinction,
  cancellation permissions, duplicate-command protection, loading, network
  recovery, location freshness and account/capability isolation.
- Preserve polling and presence behavior when moving widgets. A visually hidden
  management panel must not disable lifecycle observation or accidentally resume
  online status.
- ADR-0008 currently fixes extent/gesture/session behavior. New height, sticky
  action placement or full-screen task requirements must be proposed explicitly
  in the design spec and reconciled with that ADR and its regression tests.

## Approach comparison

| Approach | Benefit | Limitation |
|---|---|---|
| Cosmetic refresh | Small implementation scope | Retains mixed-purpose panels and poor hierarchy |
| Journey redesign with reusable components (recommended) | Addresses structure and visuals while preserving working domain behavior | Requires prototype review and careful lifecycle migration |
| App/backend rewrite | Maximum freedom | Adds regression risk and new product scope without evidence it is needed |

## Next design decisions and acceptance

First review the proposed journey structure. Then compare two visual directions
on the same screens: Rider request review/offer comparison and Driver pickup.
Expand the chosen direction into a complete clickable prototype before writing
the final design spec and small-PR implementation plan.

Open decisions for design work: panel contract, navigation destinations,
persistent endpoint labels, visual direction and implementation sequence.
These are explicit decisions to resolve, not permission to implement guesses.

Prototype/device checks:
- Stage, location, fare and next action are recognizable without management content.
- Cash due remains visible and recoverable after completion/restart.
- No streets, ETA, ratings, contact actions or payment states are fabricated.
- Essential actions fit at a 360x640 logical viewport and 1.5x text scale;
  keyboard and smaller supported viewports receive separate layout checks.
- Map attribution, safe areas, touch targets, readable contrast and screen-reader
  labels are verified on rendered prototypes/app screens.
- Changed fares/offers, services unavailable, route failure, stale location,
  cancellation, request expiry and network recovery have explicit designs.
- Existing business regression tests remain; update gesture tests only for
  deliberately approved shared interaction changes.

No visual direction or implementation is approved by this document.
