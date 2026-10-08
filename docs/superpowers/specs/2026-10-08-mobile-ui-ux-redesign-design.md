# HiGO mobile UI/UX redesign specification

Date: 2026-10-08
Status: approved by owner on 2026-10-08 after review of the five visual specification boards.
Baseline: main at 3a9fdfa2ca9305d20102ed1b5a61e84129452bfc.
Delivery: horizontal foundations across Rider and Driver, then shared screen groups.

## 1. Intent and decisions already agreed

Replace the mixed-purpose MVP interface with a coherent ride app while keeping
the proven backend and commercial behavior. The owner accepted the concept
boards' white/slate surfaces, blue primary actions, readable typography and
focused task cards, and selected horizontal implementation.

The 30 generated concept screens are visual references, not literal interaction
or data contracts. This document governs navigation, states and controls.
Generated sample names, dates, images, prices and requirements are not production
data. No unverified asset, vehicle photograph, minimum password length, six-digit
verification-code rule or Google map artwork is to be copied into production.

The selected approach is a journey redesign using reusable components.
A cosmetic-only refresh would retain the content hierarchy problems; an
app/backend rewrite adds unnecessary regression risk.

The owner approved this written spec and its visual specification on 2026-10-08.
The implementation plan is docs/superpowers/plans/2026-10-08-mobile-ui-ux-redesign.md.
The current audit is docs/ux/2026-10-08-mobile-ux-audit.md.

### Included

Shared visual system, signed-in navigation, map/task composition, Rider booking
and offers, Driver requests and responses, both active-trip experiences, cash
collection, history/receipts, authentication, onboarding and read-only vehicle/
Driver management. Empty, pending, disabled and recoverable failure states are
part of each screen, not an optional final polish task.

### Excluded

Operations/admin redesign, new backend endpoints/schema, persisted street names,
new reverse-geocoding calls, live arrival ETA, browsing-request road ETA, chat,
phone calls, ratings, push notifications, wallets/cards/refunds/promotions,
background tracking changes and new Arrived or no-show business transitions.
The accepted external Google Maps navigation stays.

## 2. Shared visual and component contract

Use the existing AppTheme and application-owned widgets. Do not build a parallel
theme system or let each feature define its own buttons and surface rules.

| Token | Proposed value/rule |
|---|---|
| Primary text | Graphite #111827 |
| Main action | Blue #2563EB, white label |
| Background / surface | Slate #F8FAFC / white #FFFFFF |
| Secondary text | Slate #475569 |
| Borders | Slate #CBD5E1 |
| Success / warning / error | #16A34A / #D97706 / #DC2626; text/icon also communicates meaning |
| Spacing | Existing 4, 8, 12, 16, 24, 32 logical-pixel scale |
| Card radius | 16; input/button radius 12 |
| Typography | System sans serif; no new bundled font dependency |
| Type roles | Screen title 24/32 semibold; section 20/28; main fare 28/36 bold; body 16/24; metadata 14/20 |
| Main controls | Minimum 48 logical-pixel height/touch target; grow when labels wrap |
| Content padding | 16 normal; safe-area-aware edges |
| Layout width | Focused forms centered with max width 600 on larger displays; maps can fill available width |

Color roles are semantic, not unconditional foreground combinations. Verify
normal-text contrast at least 4.5:1 and large-text/non-text control contrast at
least 3:1. Adjust foreground/surface pairing where required. No meaning is carried
only by green/red or by low-contrast faded text.

Text scales with the platform. Do not use fixed-height cards, clip labels or
shrink fonts to force a fit. Prices retain amount/currency together and render
from actual money objects; dates use existing localization/presentation choices.

Reusable pieces:

| Component | Responsibility |
|---|---|
| Primary/secondary/destructive action | Consistent prominence, loading, disabled and confirmation semantics |
| Location summary | Role label, available place label or saved-map-pin fallback, optional detail disclosure |
| Fare summary | Read-only suggested/requested/agreed fare; ownership explanation where relevant |
| Vehicle summary | Real captured vehicle details; no pre-assignment plate |
| Offer/request card | Comparable facts, actionable status and clear response/selection |
| Trip summary + action area | Current stage, relevant endpoint, fare, estimates where supported and available commands |
| Feedback banner | Contextual loading/error/unavailable message with recovery action |
| Empty state | Honest state with useful next step; no invented requests or offers |
| Receipt/status row | Immutable operation/settlement facts and terminal outcome |
| Form section | Accessible field label, server-provided requirements and validation |
| Confirmation surface | Action/fare-specific consequences; pending and failure state |

These widgets receive data/callbacks and render it. They do not calculate fares,
make network requests or decide server eligibility.

## 3. Navigation and presentation structure

### Signed-in shell

Three persistent destinations:
1. Ride in Rider mode / Drive in Driver mode.
2. History for the current capability.
3. Account for the one signed-in identity.

There is no separate Map destination. History opens a receipt/details route.
Account owns the explicit Rider/Driver switch, existing Become a Driver action,
Driver details, Vehicles & services, and logout. Show management entries only
when existing access rules permit them. Approved profile information remains
read-only; do not add unsupported edit or support actions.

Switch through SessionController's existing capability operation and error
handling. Preserve server/account restrictions and do not construct a second
identity. Changing a destination is not switching capability.

An active ride remains reachable through Ride/Drive, with a compact "Active ride"
or "Cash due" reminder in other destinations when authoritative state supports
it. Reminder taps return to the current task; they do not create a new controller
or trip. Do not display names/fares from another account or capability.

Focused child screens (search, request/offer review, setup and receipt) have a
back affordance; they need not display the bottom navigation. System back returns
to the parent presentation without cancelling requests, going offline, starting
a trip or discarding a submitted application. Confirmation occurs only when
the user explicitly invokes a destructive/domain action.

Draft location/vehicle form state is preserved during in-session back/navigation
as today; do not introduce persistent drafts across logout or process death.
Read-only confirmation screens may return to edit the local draft.

### Map and task surface

Replace the large floating dashboard status card plus repeated panel headings
with one authoritative task summary. Errors/status belong to that summary or a
contextual banner, not duplicated in multiple cards.

The shared scaffold supports:
- Map with compact/expanded task panel for home, waiting, pickup and active trip.
- Focused scrollable task page for search, complex request/offer comparison,
  onboarding and receipts.
- Explicit set-on-map screen with one current field, pin and Confirm location;
  no ambiguous alternating pickup/destination mode.

Compact task content contains the stage, relevant location, fare where applicable
and essential action area. Expanded content adds secondary details and permitted
secondary actions. Do not place history, applications or diagnostics in a trip
panel. Never fabricate a primary button for a pure waiting state.

**Proposed ADR-0008 amendment:** the compact panel height is measured from its
stage summary/action content instead of a fixed 16%/18%. Expanded map panels
remain capped at 60% of available task height; map viewports must retain at least
160 logical pixels. Available height excludes app navigation, safe areas and
keyboard. If compact content cannot fit within the cap, or a usable map cannot
remain, use the focused scrollable page with an optional map preview. This is a
responsive fallback, not text shrinking or hidden lifecycle controls.

Compact/expanded is a session-owned choice, preserved across capability and
business-state changes where that layout exists. Content changes reset body
scroll and invalidate old gestures, not user choice. During focused-page fallback,
preserve the choice for returning to a map layout.

Retain ADR-0008's one-pointer ownership, direct drag feedback, release-based
snapping, cancel-to-start behavior, control-versus-drag isolation and expanded
scroll handoff. The compact panel's secondary body is hidden rather than an
unscrollable clipped long form. Expanded body scrolls after snap completes.
Expose accessible Expand/Collapse actions in addition to gestures.

The action area is inside the panel surface and stays visible while expanded
body details scroll. Navigation and Start/Complete have separate touch targets.
Controls cannot accidentally resize the panel. Full-screen forms use ordinary
scrolling; keyboard insets scroll the focused field and action into view.

Use RideMap's existing padding contract to keep pins, Google attribution and
provider controls above occupied panel/action regions and below header regions.
Avoid triggering a new route/Places call during panel resizing. Camera fitting
responds to route/endpoint changes, not each build or frame of a drag.
Actual maps use the existing integration; static generated artwork is not an asset.

Update/supersede ADR-0008 and its tests in the same shared-scaffold PR that changes
the contract. Until then, existing screens keep the old behavior; local feature
exceptions to the ADR are not permitted.

## 4. Journey and action contract

These are presentation states derived from current controllers, not new backend
statuses. Server facts win over optimistic visual progression.

### Rider

| State | Visible priority | Actions |
|---|---|---|
| Home | Destination entry, pickup summary, location alternative | Choose destination; contextual current-location action |
| Location selection | Current field, search results, selected endpoint | Choose result; set on map; confirm selected field |
| Request review | Endpoints, current service, route options/estimate, suggested fare | Request ride; edit locations/service/route before submission |
| Request waiting | Stored request price and waiting message | Cancel with confirmation; retry only when needed |
| Offer comparison | Captured Driver/vehicle, fare, straight-line distance, selectable state | Choose Driver then confirm exact displayed fare/revision; decline individually |
| Assigned | Driver/vehicle/plate, pickup, agreed fare, fresh location or unavailable state | Existing permitted cancellation; no fabricated ETA/contact action |
| In progress | Destination, agreed fare and authoritative stage | Existing permitted cancellation; no Rider complete/start action |
| Completed | Fare, outcome, cash due or collected record | Booking/history; no Rider cash-confirm command |
| Cancelled/expired | Truthful outcome and recorded commercial facts | Return to booking; no automatic fee/refund claim |

Only show server-available services, not fixed Economy/Comfort lists.
The currently priced suggestion is a display summary, not an editable field.
Any genuinely supported legacy/manual-pricing behavior remains explicit and
uses its existing controller rules; it is not silently changed in a UI PR.
Request submission still validates route/price binding and policy version.
Changed inputs invalidate stale preview/suggestion; price/service conflicts
require visible recovery and deliberate resubmission.

Offer comparison may have a local selected-card highlight, but selection is not
assignment until the existing command confirms/reloads authoritative state.
Unavailable/changed offers are not silently substituted. Multiple selectable
offers have equal comparison structure; do not assume cheapest is best or add
unsupported sorting/rating criteria.

### Driver

| State | Visible priority | Actions |
|---|---|---|
| Setup required | Requirements and next application step | Existing onboarding/precheck/review/submit |
| Under review/rejected | Server application status/reason | Refresh; existing apply-again flow when allowed |
| Offline/ready | Selected verified vehicle, active services/readiness | Go online; change vehicle while offline |
| Online/waiting | Readiness and request discovery | Go offline; request list/review |
| Request review | Map pins/endpoints, Rider price, straight-line distance | Offer requested fare; counteroffer; decline where existing rules permit |
| Offer pending | Offered amount, waiting-for-selection state | Continue reviewing other requests; update/decline through existing commands |
| Assigned | Pickup, agreed fare, route-to-pickup estimate | Navigate to pickup primary; Start trip distinct secondary with confirmation; cancel secondary |
| In progress | Destination, agreed fare, full-route estimate | Navigate to destination primary; Complete trip distinct secondary with confirmation; cancel secondary |
| Completed/cash due | Exact agreed amount and outstanding cash | Confirm cash collected with existing confirmation |
| Cash collected/cancelled | Outcome, receipt and authoritative availability | Return to Drive; no automatic go-online/presence resumption |

Offers do not reserve or assign the Driver. A pending-offer screen must not
block access to the rest of the marketplace or pretend there can be only one
pending offer. Request deadline is displayed only when actual data supports it;
server expiry/eligibility stays authoritative.

Counteroffers retain backend limits and existing validation. Review confirmations
show the amount sent. Driver Start requires Rider on board; Complete confirms
destination reached. Navigation is an external launch and cannot trigger either.

Assigned route duration is the fetched Driver-to-pickup estimate. In-progress
wording remains "Full trip route"; it is not remaining time from current GPS.
Route failure suppresses estimate/polyline but preserves valid saved-endpoint
navigation and legal lifecycle actions. Completed/cancelled trips hide navigation.

Cash collection is distinct from completion. Keep completed/unsettled trips
recoverable and visible until collection is confirmed. Payment state is read
from SettlementSnapshot, never inferred from navigation/return/receipt rendering.
Availability controls reflect current business rules; the redesign neither
blocks nor permits operations the existing controllers/backend disallow.

### Supporting experiences

History is a dedicated list by capability with receipt/detail navigation,
empty/error/loading and retry. Receipt facts come from immutable trip context.
No measured driven-distance row is added; the app does not collect that record.
Missing historical context receives an explicit unavailable fallback.

Account/Driver details/vehicles use existing access and read-only boundaries.
Vehicle/service additions retain requirements, precheck, review, approval/
rejection and offline-only operating changes. Pending vehicles cannot become an
operating choice. No immediate edits to approved information are introduced.

Authentication keeps current email/password/register/verify/resend contracts.
Display current field rules from implementation/server, not generated mockup
assumptions. Preserve autofill, password visibility where supported, keyboard
behavior, submission protection, validation and session restoration.

## 5. Data, lifecycle and failure handling

### Location names and missing data

Use existing resolved place labels for the current Rider draft only while they
match the selected endpoint. Do not reuse a label for a changed coordinate.
Recovered requests/trips without trustworthy labels show "Pickup · saved map pin"
and "Destination · saved map pin", with coordinates available through details
and relevant markers on the map. They do not masquerade as named streets.

There are no new geocoding calls, local persistent location cache or backend
address snapshots in this rollout. Better recovered location names can be a
separate bounded feature after this redesign. This limits presentation quality,
but is preferable to false location labels or hidden new billing/data work.

Only actual captured Driver/vehicle details are shown. Use neutral vehicle icons,
not fabricated photos or a displayed vehicle image implied to be a real upload.
Pre-assignment plates remain absent. Missing Driver location is stated plainly;
drop stale samples through the existing freshness logic.

### Responsibility and controller ownership

- Theme/components: pure presentation and accessible interaction.
- Shell/task layout: navigation, layout, panel session and app lifecycle bridging.
- Feature presentation: derive visual state from existing controller snapshots,
  retain relevant IDs/revisions, wire callbacks to existing commands.
- Controllers/repositories: authoritative refresh, polling, command serialization,
  recovery, pricing/route logic and native provider boundaries.

Replace mixed monolithic widget content with small screen/component units.
Avoid duplicated Rider/Driver business controllers or a new mega UI controller.
Any presentation-state resolver is pure and does not introduce persisted domain
state. Existing providers/repositories and account invalidation are retained.

Foreground observation currently partly resides in RideFlowPanel. Move that
observation to the signed-in capability shell when the old panel is replaced,
with exactly one lifecycle bridge per relevant controller. Keep selected-
capability controllers alive while navigating History/Account; otherwise tab
changes can dispose polling/presence owners. Returning to Ride/Drive reuses the
same state, rather than starting another polling loop.

Do not widen polling/presence to inactive capabilities or change background
service policy. Session/capability changes follow existing controller guards,
cancellation and disposal. Logout resets navigation, local draft/panel state and
account-bound providers. Native Maps return triggers the existing foreground
refresh/recovery behavior, not a duplicate launch or Trip transition.

### Required failure branches

| Condition | Presentation and recovery |
|---|---|
| Initial loading | Stable structure/loading indicator; unsafe commands unavailable |
| No services | Honest unavailable state and catalog Retry; no phantom service |
| Place lookup denied/failed | Contextual message and map/search alternative; preserve other input |
| Route/fare loading or failure | No fabricated estimate/fare; Retry; priced submit follows existing validity gate |
| Service/policy changed | Explain change, refresh current preview/catalog, explicit resubmission |
| No offers | Waiting state; no implicit timeout promise or assignment |
| Offer changed/unavailable | Refresh current comparison and require a fresh confirmation |
| Request expired | Terminal outcome and return-to-booking action |
| Network/command failure | Retain context, show authoritative recovery result; no duplicate command/false success |
| Driver location stale/unavailable | Suppress stale marker and state unavailable; keep trip details |
| Native Maps unavailable | Recoverable launch message; preserve trip, controls and preview |
| Cash-confirm failure | Stay cash due/unconfirmed until server reload confirms collection |
| Approval/eligibility changed | Explain actual blocker; current selection cannot bypass server policy |
| Missing historical fields | "Unavailable"/map-pin fallback, never invented commercial details |

An error banner is not a promise that the network is definitely offline unless
the app has evidence. Safely show last-known content with stale/unavailable
labeling; do not override backend eligibility or existing command gates.

## 6. Horizontal rollout boundaries

This is delivery ordering, not the detailed implementation plan.

| Layer | Scope across both capabilities | Completion evidence |
|---|---|---|
| Shared design system | Existing tokens, type roles, actions, inputs, summaries, cards, feedback | Component rendering/state/accessibility checks; app still usable |
| Shell and task layout | Three destinations, capability/account guards, history routes, lifecycle owner, adaptive panel/focused pages | Navigation/back/account isolation and shared gesture/map-inset checks |
| Booking and offers | Rider home/search/review/wait/compare; Driver readiness/request/response | Fare, offer-revision, eligibility and marketplace regression/device checks |
| Active rides and cash | Both pickup/in-progress/terminal experiences, navigation, settlement | End-to-end two-client ride/cash/cancel/recovery checks |
| Supporting surfaces | Auth, onboarding, vehicles/details, histories/receipts | Existing contract preservation and form/accessibility checks |
| Final validation | Cross-flow visual/interaction consistency and cleanup | Exact-head readiness, Android build, device acceptance and documented results |

Each implementation PR must leave both capabilities functional. Define a shared
component contract before feature consumers diverge. Do not combine UI redesign
with unsolicited backend/domain refactors. Intermediate migration can adapt old
screens to new shared components; remove replaced paths when their consumers move.
No permanent competing navigation or design systems.

Detailed PR/task sizing, test commands, dependency ordering and execution method
are chosen in the implementation plan after this specification is reviewed.
Preserve development branches as requested. No automatic merge or deployment.

## 7. Acceptance and definition of done

- Normal layout checked at 360x640 and 390x844 logical pixels, plus 360x500
  short viewport, landscape and keyboard-visible forms. Respect safe areas.
- At 1.0x and 1.5x text scale all required actions are reachable; at 2.0x use
  responsive/focused fallback rather than clipping important content.
- Driver navigation and Start/Complete remain separate and reachable without
  expanding unrelated management content.
- Primary/secondary actions have distinguishable prominence, accessible labels,
  logical focus order and semantic Expand/Collapse affordances.
- Map attribution/provider controls and main actions stay unobscured; route
  fitting accounts for overlays so relevant endpoint markers remain in the usable
  viewport. Navigation, keyboard, banners and safe areas cannot cover them.
- All supported business stages and failure branches have rendered references
  and behavioral checks appropriate to their risk.
- Existing pricing, offer selection, route stage, trip/cancellation, cash,
  account isolation, polling and presence regressions remain green.
- Native Maps targets, return/recovery, terminal hiding and failure behavior
  are verified on Android. Two independent clients verify the complete cash
  ride and cancellation flows after the redesign.
- Visual review compares actual Flutter screens against the accepted style;
  generated mockups alone are not evidence of usable implementation.
- Exact-head repository readiness and Android builds pass before merge.
- Final documentation reconciles revised shared UX contracts and replaces
  obsolete UI instructions only within this redesign's scope.

## 8. Spec review focus

The owner should review four concrete decisions:
1. Three destinations: Ride/Drive, History and Account; no separate Map tab.
2. Adaptive compact map panels and full-screen tasks, with an explicit ADR-0008
   sizing amendment while preserving gesture/session safeguards.
3. Existing local place labels with saved-map-pin fallback; no new address-data
   infrastructure in this rollout.
4. Horizontal delivery order and preservation of current backend/ride semantics.

The visual direction, horizontal delivery and written/visual specification are
approved. Implementation-plan review and execution-method selection are pending;
product-code execution has not started.

