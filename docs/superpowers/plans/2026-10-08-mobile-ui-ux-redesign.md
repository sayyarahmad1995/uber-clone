# HiGO Mobile UI/UX Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the approved white/slate/blue mobile experience across Rider and Driver without changing ride, pricing, eligibility or settlement behavior.

**Architecture:** Extend AppTheme and introduce pure shared presentation components. Keep selected-capability controller ownership in the signed-in shell, separate navigation from domain commands, and replace mixed dashboard content with adaptive task surfaces and focused pages. Migrate both capabilities together in each horizontal layer.

**Tech Stack:** Existing Flutter 3.47.2, Dart, Material 3, Riverpod, go_router, Google Maps integration and current repository/controller interfaces; no new package is required.

**Spec:** `docs/superpowers/specs/2026-10-08-mobile-ui-ux-redesign-design.md` (owner approved on 2026-10-08 after reviewing five visual specification boards).

## Global Constraints

- Three persistent destinations: Ride in Rider mode / Drive in Driver mode, History for the current capability, and Account for the one signed-in identity. There is no separate Map destination.
- Main action: Blue #2563EB, white label. Primary text: Graphite #111827. Background / surface: Slate #F8FAFC / white #FFFFFF. Secondary text: Slate #475569. Borders: Slate #CBD5E1.
- Spacing: Existing 4, 8, 12, 16, 24, 32 logical-pixel scale. Card radius: 16; input/button radius 12. System sans serif; no new bundled font dependency.
- Screen title 24/32 semibold; section 20/28; main fare 28/36 bold; body 16/24; metadata 14/20. Minimum 48 logical-pixel height/touch target; grow when labels wrap. Focused forms centered with max width 600 on larger displays.
- Normal-text contrast at least 4.5:1 and large-text/non-text control contrast at least 3:1. Semantic success/warning/error color never communicates meaning alone.
- Expanded map panels remain capped at 60% of available task height; map viewports must retain at least 160 logical pixels. Available height excludes app navigation, safe areas and keyboard.
- Compact/expanded is a session-owned choice. Preserve one-pointer ownership, direct drag feedback, release-based snapping, cancel-to-start behavior, control-versus-drag isolation and expanded scroll handoff.
- The currently priced suggestion is a display summary, not an editable field. The accepted Driver offer is the agreed fare; current controller gates and revision binding remain authoritative.
- No new backend endpoints/schema, geocoding calls, address cache, live arrival ETA, browsing-request road ETA, chat, calls, ratings, payment methods or background tracking policy.
- Keep the accepted external Google Maps navigation. Navigation and Start/Complete have separate touch targets; terminal trips hide navigation.
- Keep selected-capability controllers alive while navigating History/Account. Do not widen polling/presence to inactive capabilities. Logout resets navigation, local draft/panel state and account-bound providers.
- Cash collection is distinct from completion. SettlementSnapshot is authoritative. No automatic go-online/presence resumption.
- Preserve development branches. No automatic merge or deployment.

## Review Focus

1. Delayed response after logout/capability change: old account data and callbacks cannot populate the new shell (Task 2 tests).
2. Offer changes while confirmation is open: the exact displayed fare/revision must be revalidated; no silent replacement (Task 4 tests).
3. Two-pointer drag interrupted by a new business state: old pointer cannot move new content; committed expansion survives (Task 3 tests).
4. Native Maps failure or return during a pending lifecycle command: navigation cannot duplicate Start/Complete or falsely advance the trip (Task 5 tests).
5. Restart after completion while cash confirmation failed: cash due stays recoverable; no invented receipt or availability transition (Tasks 5 and 7 tests).

## Delivery and file ownership

Baseline inspected: `56c42bb4502e573b9d15778b556d5fe9a4214a66`, documentation branch based on main `3a9fdfa2ca9305d20102ed1b5a61e84129452bfc`. Before execution, fetch current main and reconcile any newer changes. The scratch snapshot is partial, not a buildable checkout; execution requires a full isolated checkout and Flutter SDK.

| PR | Tasks | Deliverable |
|---|---|---|
| UI-1 | 1 | Shared theme and component contracts, used in both capabilities |
| UI-2 | 2 | Signed-in shell, dedicated history/account, lifecycle ownership |
| UI-3 | 3 | Adaptive map/task layout and ADR-0008 amendment |
| UI-4 | 4 | Rider booking/offers and Driver readiness/request responses |
| UI-5 | 5 | Both active-trip journeys, navigation and cash settlement |
| UI-6 | 6 | Authentication, onboarding, vehicles, history and receipts |
| UI-7 | 7 | Final cross-flow validation, documentation and removal of obsolete UI |

PRs are sequential; each leaves Rider and Driver usable. UI-2 moves existing history rendering before UI-6 polishes it. UI-3 supplies compact/body/action slots to legacy adapters before UI-4/UI-5 replace their content. No permanent second design system or navigation tree.

All paths below are repository-relative. Test commands run in `mobile/` unless stated otherwise. Extend the fakes in `test/test_doubles.dart` and existing test-specific fake repositories; use controllable Futures and call counters for races. Tests assert visible behavior, controller identity and domain calls, not widget implementation details. Never approve goldens automatically; inspect the images first.

### Task 1: Shared visual system and reusable presentation

**Files:** Modify `mobile/lib/core/theme/app_theme.dart`, `mobile/lib/features/rider_request/presentation/rider_request_screen.dart`, `mobile/lib/features/driver_workspace/presentation/driver_workspace_screen.dart`; create `mobile/lib/core/widgets/task_action.dart`, `task_summary.dart`, `feedback_banner.dart`, `focused_task_page.dart`, `action_confirmation.dart`; extend `mobile/test/app_theme_test.dart`; create `mobile/test/shared_task_components_test.dart`.

**Interfaces:**
- Preserve AppColors/AppSpacing/AppRadii/AppTheme.light. Keep graphite text token; set ColorScheme.primary to blue and onPrimary to white rather than turning all graphite text blue.
- `TaskAction({required String label, required VoidCallback? onPressed, TaskActionKind kind = TaskActionKind.primary, bool busy = false})`; enum values primary, secondary, destructive. Pending action disables invocation and retains an accessible label.
- `LocationSummary({required String role, required String label, Widget? details})`, `FareSummary({required String label, required String amountText, String? explanation})`, `VehicleSummaryCard({required Widget details})`, `TaskSummary({required String title, required Widget content})`. Formatting adapters use actual existing money objects; components contain no network/price logic.
- `FeedbackBanner({required String message, String? actionLabel, VoidCallback? onAction, bool loading = false})`; no retry affordance without a real callback.
- `FocusedTaskPage({required String title, required Widget body, Widget? actions})`: ordinary scroll, max width 600, keyboard/safe-area handling.
- `Future<bool> confirmTaskAction(BuildContext context, {required String title, required String message, required String actionLabel, bool destructive = false})`; command remains caller-owned and is invoked only after confirmation.

- [ ] **Step 1: Write component tests** named `blue_primary_readable_tokens`, `pending_action_cannot_submit_twice`, `wrapped_actions_remain_reachable`, `fare_is_summary_not_input`, `focused_form_keyboard_and_2x_text`. For primary color/target/fare assertions use:
  ```dart
  expect(AppTheme.light.colorScheme.primary, const Color(0xFF2563EB));
  expect(AppTheme.light.colorScheme.onPrimary, Colors.white);
  expect(tester.getSize(find.byType(FilledButton)).height, greaterThanOrEqualTo(48));
  expect(find.byType(TextField), findsNothing); // isolated FareSummary
  expect(tester.takeException(), isNull);
  ```
  Add semantic labels, disabled/loading state, contrast calculations for actual foreground/background pairs and 600px width cap assertions. Test 360x500 at 2.0x text and keyboard inset 240.
- [ ] **Step 2: Run** `flutter test test/app_theme_test.dart test/shared_task_components_test.dart`; expect failure for absent components or changed theme values.
- [ ] **Step 3: Implement** the interfaces and exact tokens above. Migrate one real task action/summary in each capability, retaining commands and existing keys where their semantics remain valid. No fixed-height text cards or generated photos/map artwork.
- [ ] **Step 4: Run** the same tests plus `flutter test test/rider_request_submission_test.dart test/driver_workspace_test.dart`; require all pass. Review actual component rendering at normal and large text sizes.
- [ ] **Step 5: Commit** the explicit files with message `feat(mobile): establish shared task presentation system`.

### Task 2: Signed-in shell and selected-capability lifecycle

**Files:** Modify `mobile/lib/core/providers.dart`, `mobile/lib/features/capabilities/presentation/capability_home_screen.dart`, `mobile/lib/features/driver_workspace/presentation/driver_workspace_screen.dart`, `mobile/lib/features/ride_flow/ride_flow_panels.dart`; create `mobile/lib/features/capabilities/presentation/capability_lifecycle_scope.dart`, `account_screen.dart`, `mobile/lib/features/ride_flow/presentation/ride_history_screen.dart`; extend `mobile/test/app_routing_test.dart`, `session_controller_test.dart`, `ride_flow_controller_split_test.dart`; create `mobile/test/capability_shell_lifecycle_test.dart`.

**Interfaces:**
- `enum CapabilityDestination { task, history, account }`; CapabilityHomeScreen retains its existing `capability` parameter and owns selected destination. A destination change never calls SessionController.selectCapability.
- `CapabilityLifecycleScope({required Capability capability, required Widget child})`: one WidgetsBindingObserver; watch only current-capability providers and retain them across destination/child navigation. Rider owns request/place/route/fare providers and active ride family only when an actual ID exists. Driver owns driver/readiness, marketplace/trip/route and onboarding only under their existing eligibility gates.
- Move existing driver.setActiveTrip synchronization into that scope too, so tab changes cannot stop readiness synchronization. Preserve initial controller behavior; do not add new command/polling loops.
- `AccountScreen({required Capability capability})`: existing capability switch, Become a Driver, management access guards, logout. `RideHistoryScreen({required Capability capability})`: use existing listRiderRides/listDriverTrips data and loading/error/retry rendering; initially reuse existing operation details.
- Child routes remain under the signed-in shell lifetime; return/back reuse the same controller objects. Retain current authorization redirects for Driver details/vehicles. Active-ride/cash-due reminder derives from current snapshots and changes only the selected destination.

- [ ] **Step 1: Write shell tests** `tab_navigation_retains_selected_controller`, `one_foreground_refresh_per_controller`, `inactive_permission_dialog_does_not_end_presence`, `inactive_capability_is_not_polled`, `logout_drops_late_account_response`, `back_does_not_issue_domain_command`, `active_cash_reminder_returns_to_existing_task`. Capture controller references before tab changes; assertions include:
  ```dart
  expect(identical(beforeController, afterController), isTrue);
  expect(refreshCallsAfterResume - refreshCallsBeforeResume, 1);
  expect(inactiveCapabilityPollCalls, 0);
  expect(domainCallsAfterBack, domainCallsBeforeBack);
  expect(find.text(oldAccountName), findsNothing);
  ```
  Delay repository completion, logout/login as another account, then complete it. Also fail capability switching and assert destination/data remain tied to the actual selected capability.
- [ ] **Step 2: Run** `flutter test test/capability_shell_lifecycle_test.dart test/app_routing_test.dart`; expect missing destination/lifetime assertions to fail on the old drawer shell.
- [ ] **Step 3: Implement** shell, routes and selected-provider retention; remove lifecycle observers from RideFlowPanel and DriverWorkspaceScreen in the same change. Preserve resumed refresh/route retry, ignore inactive state, and preserve paused/hidden/detached handling. Keep account ID invalidation; reset destination/panel/drafts on logout. Extract history out of trip panels so it is not duplicated.
- [ ] **Step 4: Run** the tests above plus `flutter test test/session_controller_test.dart test/driver_presence_service_test.dart test/ride_flow_controller_split_test.dart test/ride_flow_polling_loop_test.dart`; require pass. Inspect both capabilities with active and no-active snapshots.
- [ ] **Step 5: Commit** with message `feat(mobile): add capability shell with stable lifecycle ownership`.

### Task 3: Adaptive task panel, focused fallback and map insets

**Files:** Modify `mobile/lib/core/dashboard/ride_dashboard_scaffold.dart`, `dashboard_panel_session.dart`, `mobile/lib/core/maps/ride_map.dart` only if existing padding support needs correction, both task screen consumers, and `docs/ADR-0008-dashboard-panel-interaction-contract.md`; extend `mobile/test/dashboard_panel_scroll_handoff_test.dart`, `dashboard_panel_build_efficiency_test.dart`, `ride_map_contract_test.dart`; create `mobile/test/adaptive_task_layout_test.dart`.

**Interfaces:** Keep RideDashboardScaffold and DashboardPanelControl names. Introduce `typedef TaskMapBuilder = Widget Function(BuildContext context, EdgeInsets occupiedPadding)` and constructor slots `required TaskMapBuilder mapBuilder`, `required Widget summary`, `required DashboardPanelBuilder bodyBuilder`, `Widget? actions`, `required Object panelIdentity`, `Widget? mapControls`. Migrate all consumers and remove fixed minPanelSize/initialPanelSize overrides in this PR. DashboardPanelSession remains the single committed-expanded owner. Internal measured-layout state belongs to the scaffold.

- [ ] **Step 1: Write tests** `measured_compact_has_all_essential_actions`, `expanded_cap_and_minimum_map`, `short_keyboard_2x_falls_back`, `body_change_preserves_choice_resets_scroll`, `second_pointer_and_old_content_cannot_drag`, `actions_do_not_drag_and_stay_visible`, `drag_does_not_refetch_route_or_fit_camera`. Assert at 360x640 and 390x844, then 360x500/landscape/keyboard/2x:
  ```dart
  expect(expandedHeight, lessThanOrEqualTo(availableTaskHeight * 0.60));
  expect(visibleMapHeight, greaterThanOrEqualTo(160)); // map layout only
  expect(tester.takeException(), isNull);
  expect(session.expanded, previousCommittedChoice);
  expect(routeCallsAfterDrag, routeCallsBeforeDrag);
  expect(cameraFitsAfterDrag, cameraFitsBeforeDrag);
  ```
  Add fallback presence/action reachability, occupied RideMap.padding and handle Expand/Collapse semantics assertions. Preserve existing release snap/cancel/scroll-handoff/build-efficiency coverage with measured extents.
- [ ] **Step 2: Run** `flutter test test/adaptive_task_layout_test.dart test/dashboard_panel_scroll_handoff_test.dart test/dashboard_panel_build_efficiency_test.dart test/ride_map_contract_test.dart`; expect old fixed sizing and absent slots to fail.
- [ ] **Step 3: Implement** measured summary+action height, separate expanded scroll body and pinned action area. Expanded maximum is the smaller of 60% available height and height minus 160; use focused scrollable fallback when compact exceeds it. Exclude actual navigation/safe-area/keyboard occupancy once, without double padding. Keep one shared gesture policy and accessible Expand/Collapse. Use mapBuilder padding on RideMap; fit only when route/endpoints change. Adapt both old task screens; remove duplicated floating status. Amend ADR-0008's sizing clauses while preserving gesture/session clauses.
- [ ] **Step 4: Run** Step 2 commands plus `flutter test test/driver_route_map_test.dart test/driver_workspace_test.dart test/rider_request_submission_test.dart`; require pass. Physically check Google attribution and endpoint visibility on Android, since placeholder-map widget tests cannot prove native-map layout.
- [ ] **Step 5: Commit** with message `feat(mobile): make shared task layout adaptive`.

### Task 4: Booking and marketplace presentation across Rider and Driver

**Files:** Modify `mobile/lib/features/rider_request/presentation/rider_request_screen.dart`, `mobile/lib/features/driver_workspace/presentation/driver_workspace_screen.dart`, `mobile/lib/features/ride_flow/ride_flow_panels.dart`; create `mobile/lib/features/rider_request/presentation/location_selection_screen.dart`, `request_review_screen.dart`, `mobile/lib/features/ride_flow/presentation/rider_offer_screen.dart`, `driver_request_screen.dart`, `mobile/lib/features/driver_workspace/presentation/driver_readiness_card.dart`; extend rider place/catalog/fare/submission and ride-flow tests; create `mobile/test/booking_marketplace_presentation_test.dart`.

**Interfaces:** LocationSelectionScreen receives the existing RiderPlaceField and operates on the retained rider controllers; RequestReviewScreen has no new domain state. `RiderOfferScreen({required String rideId})` consumes riderActiveRideControllerProvider(rideId). `DriverRequestScreen({required MarketplaceRequest request})` consumes driverMarketplaceControllerProvider. Keep current methods `selectOffer(String driverUserId, DateTime updatedAt)`, `acceptProposedFare`, `submitOffer` and decline operations. DriverReadinessCard receives rendered readiness content/callbacks from DriverWorkspaceScreen, not a second controller.

- [ ] **Step 1: Write tests** `priced_review_has_no_editable_fare`, `location_label_requires_coordinate_match`, `no_catalog_service_blocks_priced_submit`, `changed_policy_requires_explicit_resubmit`, `changed_offer_requires_fresh_confirmation`, `preassignment_plate_absent`, `pending_offer_allows_other_requests`, `counteroffer_confirmation_matches_sent_minor_units`, `expiry_shows_terminal_outcome`. Reuse real controller state and fakes; assertion examples:
  ```dart
  expect(selectedRevisionSent, displayedRevision);
  expect(selectedFareShown, displayedFare);
  expect(commandCountAfterStaleConfirmation, 0);
  expect(find.text(preassignmentPlate), findsNothing);
  expect(find.text('Pickup · saved map pin'), findsOneWidget);
  ```
  The stale-confirmation case refreshes the offer while its dialog is open, and requires reopening/confirming current data. Include failed Places lookup retaining the other endpoint, failed route/price retry, unavailable Driver, no offers and expired request. Pending offer must leave another request review accessible.
- [ ] **Step 2: Run** `flutter test test/booking_marketplace_presentation_test.dart`; expect absent dedicated pages and old hierarchy assertions to fail.
- [ ] **Step 3: Implement** home/location selection/set-on-map/review/waiting/offer comparison and Driver readiness/list/review/counteroffer/pending surfaces using Tasks 1–3. Set-on-map edits one explicit field. Priced fare is a summary; any genuinely supported manual legacy mode keeps its existing gate. Local offer highlight is not assignment. Revalidate displayed revision/fare/selectability before sending the existing selection command; show refresh recovery rather than replacing the fare silently. Do not add sorting/rating, ETA or preassignment plate.
- [ ] **Step 4: Run** `flutter test test/booking_marketplace_presentation_test.dart test/rider_place_search_test.dart test/rider_service_catalog_test.dart test/rider_fare_controller_test.dart test/rider_route_preview_test.dart test/rider_request_submission_test.dart test/ride_flow_test.dart test/operating_selection_test.dart`; require pass. Check Rider and Driver screens at normal/large text and verify draft preservation on back.
- [ ] **Step 5: Commit** with message `feat(mobile): redesign booking and marketplace task screens`.

### Task 5: Active rides, external navigation and cash settlement

**Files:** Modify both task screen consumers and `mobile/lib/features/ride_flow/ride_flow_panels.dart`; create `mobile/lib/features/ride_flow/presentation/rider_trip_task.dart`, `driver_trip_task.dart`, `cash_collection_task.dart`; retain existing driver_route_controller.dart and driving_navigation.dart behavior; extend `mobile/test/ride_flow_test.dart`, `driver_route_map_test.dart`, `driving_navigation_test.dart`; create `mobile/test/active_trip_presentation_test.dart`.

**Interfaces:** `RiderTripTask({required String rideId})`, `DriverTripTask({required TripSnapshot trip, required VoidCallback? onNavigate, required bool navigating, required Widget routeSummary})`, `CashCollectionTask({required TripSnapshot trip})`. DriverTripTask uses existing driverTripControllerProvider commands; DriverWorkspaceScreen retains native-launch/camera ownership. CashCollectionTask calls existing confirmCashCollected only after exact-amount confirmation. No duplicate lifecycle observer or command owner.

- [ ] **Step 1: Write tests** `navigation_and_lifecycle_targets_are_separate`, `navigate_never_starts_or_completes`, `route_failure_keeps_navigation_and_legal_controls`, `maps_failure_and_return_do_not_duplicate_pending_command`, `in_progress_says_full_trip_route`, `terminal_hides_navigation`, `stale_driver_marker_is_unavailable`, `completed_unsettled_survives_recreation`, `cash_failure_remains_due_no_auto_online`. Assert:
  ```dart
  expect(startCallsAfterNavigate, 0);
  expect(completeCallsAfterNavigate, 0);
  expect(find.text('Full trip route'), findsOneWidget);
  expect(find.text('Cash due'), findsOneWidget); // failed confirmation/reload
  expect(goOnlineCallsAfterCollection, 0);
  ```
  Include delayed Start/Complete while Maps returns, cash confirmation succeeded remotely but response failed (reload must determine truth), missing route/operation fields, each permitted cancellation stage and Rider inability to Start/Complete/confirm cash.
- [ ] **Step 2: Run** `flutter test test/active_trip_presentation_test.dart`; expect old mixed panel priorities or absent new surfaces to fail.
- [ ] **Step 3: Implement** assigned/in-progress/completed/cancelled tasks across both capabilities. Render exact agreed fare from immutable operation context, truthful saved-pin/location fallbacks, real plate only after assignment, stage-specific estimates and separate legal controls. Start confirms Rider on board; Complete confirms destination reached. Cash task remains visible/recoverable until authoritative collection; navigation errors stay contextual and do not advance stage.
- [ ] **Step 4: Run** `flutter test test/active_trip_presentation_test.dart test/ride_flow_test.dart test/driver_route_map_test.dart test/driving_navigation_test.dart test/ride_flow_controller_split_test.dart test/ride_flow_polling_loop_test.dart`; require pass. Device-check Maps target/return/failure and action reachability before offering this PR for merge.
- [ ] **Step 5: Commit** with message `feat(mobile): redesign active trip and cash collection tasks`.

### Task 6: Supporting forms, management, histories and receipts

**Files:** Modify `mobile/lib/features/authentication/presentation/login_screen.dart`, `verification_screen.dart`, `mobile/lib/features/driver_workspace/presentation/driver_workspace_screen.dart`, `driver_readonly_surfaces.dart`, `vehicle_service_application_section.dart`, `mobile/lib/features/capabilities/presentation/account_screen.dart`, `mobile/lib/features/ride_flow/presentation/ride_history_screen.dart`; create `mobile/lib/features/driver_workspace/presentation/driver_setup_screen.dart`, `mobile/lib/features/ride_flow/presentation/ride_receipt_screen.dart`; extend `mobile/test/auth_repository_test.dart`, `driver_vehicle_application_test.dart`, `driver_readonly_surfaces_test.dart`, `app_routing_test.dart`; create `mobile/test/supporting_surfaces_test.dart`.

**Interfaces:** DriverSetupScreen extracts the current setup/review form using retained driverOnboardingControllerProvider and current review/submit callbacks. `RideReceiptScreen({RiderRideSnapshot? riderRide, DriverTripHistoryItem? driverTrip})` asserts exactly one source; receives an already-authorized history item, never fabricates missing context. History remains capability-specific using existing repository list methods. Login/verification and Driver management keep current constructor/provider contracts.

- [ ] **Step 1: Write tests** `auth_rules_are_existing_not_mockup_rules`, `forms_scroll_with_keyboard_and_2x_text`, `back_keeps_in_session_application_draft`, `pending_vehicle_cannot_operate`, `approved_profile_is_read_only`, `history_is_capability_scoped`, `receipt_missing_context_is_unavailable`, `receipt_cannot_mark_cash_collected`. Assert no unsupported Edit/chat/call/rating/payment controls, no measured driven-distance claim, and no writes from rendering/back/receipt navigation. Test server rejection/eligibility changed and empty/loading/error/retry lists.
- [ ] **Step 2: Run** `flutter test test/supporting_surfaces_test.dart`; expect absent focused surfaces or old layout assertions to fail.
- [ ] **Step 3: Implement** shared form sections, focused onboarding/precheck/review/status, read-only details, vehicles/services management, history rows and immutable receipts. Preserve current password/autofill/resend/verification contracts; do not introduce generated code-length/password rules. Preserve offline-only operating selection and approved-information boundaries. Remove obsolete embedded history/management/setup blocks after their consumers migrate.
- [ ] **Step 4: Run** `flutter test test/supporting_surfaces_test.dart test/auth_repository_test.dart test/session_controller_test.dart test/driver_vehicle_application_test.dart test/driver_readonly_surfaces_test.dart test/driver_onboarding_repository_test.dart test/app_routing_test.dart`; require pass. Inspect keyboard-visible forms and actual immutable receipts in both capabilities.
- [ ] **Step 5: Commit** with message `feat(mobile): unify supporting forms management and receipts`.

### Task 7: Cross-flow acceptance, cleanup and exact-head readiness

**Files:** Create `mobile/test/redesign_acceptance_test.dart`, `docs/ux/2026-10-08-mobile-ui-ux-redesign-validation.md`; modify audit/spec/plan checkboxes and obsolete mobile instructions only within redesign scope. Remove replaced presentation paths only after confirming no consumers remain with `rg`; preserve repositories/controllers/native integration.

**Interfaces:** No new production API. Use existing fakes and the real signed-in shell to run full presentation journeys. Validation document records exact tested commit, commands/results, screenshots, device/OS/build, two-client scenarios and any unresolved limitations; distinguishes automated checks from actual device evidence.

- [ ] **Step 1: Write acceptance tests** `rider_driver_complete_cash_journey`, `cancel_and_expire_recover`, `history_account_maps_return_preserve_active_task`, `restart_after_failed_cash_confirmation`. Exercise request→offer→selection→assigned→start→complete→cash due→collection and assert final fare equals the accepted offer, collection is separate and no auto-online occurred. Record missing coverage as a failure, not a presumed pass.
- [ ] **Step 2: Run** `flutter test test/redesign_acceptance_test.dart`; expect failures only for discovered integration gaps. Do not manufacture a failing test for behavior already implemented by earlier tasks; preserve successful regression coverage.
- [ ] **Step 3: Fix** uncovered presentation/integration gaps and remove replaced UI with no remaining consumers. Record revised ADR, accurate supported data and missing-address limitation. Produce real Flutter screenshots for normal, short, landscape, keyboard and large-text conditions, including failure/empty/pending states.
- [ ] **Step 4: Run exact-head readiness.** From `mobile/`: `flutter pub get`, `dart run build_runner build --delete-conflicting-outputs`, verify generated files introduce no uncommitted difference, `flutter analyze`, `flutter test`, and Android debug/release builds with the existing environment's API/Maps build configuration. Run changed-Dart formatting and current `.github/workflows/readiness.yml` guards. Require both existing CI jobs green at the final SHA; backend CI executes `go vet ./...` and `go test -p 1 ./...` with its Postgres fixture. No backend refactor is authorized by a failing unrelated baseline check.
- [ ] **Step 5: Record Android acceptance** at 360x640, 390x844, 360x500/short viewport, landscape, keyboard-visible and 1.0x/1.5x/2.0x text. With two independent clients, verify complete cash ride, both legal cancellation paths, no-service policy, changed-offer confirmation, Maps targets/return/unavailable, restart cash recovery, capability/tab changes, logout/account isolation, stale-location handling and map attribution/endpoints. Existing pre-redesign device results do not count as redesigned-build evidence. If a device condition cannot be simulated, list it as unverified.
- [ ] **Step 6: Commit** with message `test(mobile): verify redesigned journeys and document acceptance`. Any code change after tested SHA requires rerunning affected checks and updating evidence. Request final branch review; offer merge only after exact-head readiness and owner device acceptance, retaining branches.

## Coverage and execution handoff

Spec sections 1–2 map to Task 1 and all consumer tasks; navigation/lifecycle to Task 2; adaptive map/ADR to Task 3; booking/marketplace to Task 4; trip/route/settlement to Task 5; support/receipts to Task 6; definition of done to Task 7. Every required failure branch is owned by its feature task, with cross-flow recovery checked in Task 7.

The approved spec is complete. This implementation plan is pending owner review and execution-method selection. Recommend native execution because the shell, panel and feature consumers share interfaces and benefit from one implementer retaining context; use an independent whole-branch review before merge. Subagent-driven execution is available if separate implementer/reviewer gates per task are preferred.
