import 'dart:async';

import 'package:flutter/material.dart';

import '../../ride_flow/ride_flow_panels.dart';

import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/dashboard/ride_dashboard_scaffold.dart';
import '../../../core/maps/ride_map.dart';
import '../../../core/providers.dart';
import '../../../core/theme/app_theme.dart';
import '../application/rider_place_search_controller.dart';
import '../application/rider_route_preview_controller.dart';
import '../application/rider_request_controller.dart';
import '../domain/place_search.dart';
import '../domain/ride_request.dart';
import '../domain/route_preview.dart';

class RiderRequestScreen extends ConsumerStatefulWidget {
  const RiderRequestScreen({super.key});

  @override
  ConsumerState<RiderRequestScreen> createState() => _RiderRequestScreenState();
}

class _RiderRequestScreenState extends ConsumerState<RiderRequestScreen>
    with WidgetsBindingObserver {
  final _fare = TextEditingController();
  final _pickupSearch = TextEditingController();
  final _destinationSearch = TextEditingController();
  final _mapController = RideMapController();
  bool _selectingPickup = true;
  bool _pinSelectionMode = false;
  String? _lastFittedRouteKey;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (mounted) {
        _focusCurrentLocation(showError: false);
      }
    });
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    _fare.dispose();
    _pickupSearch.dispose();
    _destinationSearch.dispose();
    super.dispose();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (state == AppLifecycleState.resumed) {
      ref.invalidate(riderServicesProvider);
    }
    final id = ref.read(riderRequestControllerProvider).state.active?.id;
    if (id == null) return;
    if (state == AppLifecycleState.resumed) {
      ref.read(riderActiveRideControllerProvider(id)).setForeground(true);
    }
    if ([
      AppLifecycleState.paused,
      AppLifecycleState.hidden,
      AppLifecycleState.detached,
    ].contains(state)) {
      ref.read(riderActiveRideControllerProvider(id)).setForeground(false);
    }
  }

  Future<void> _submit() async {
    final services = ref.read(riderServicesProvider).asData?.value;
    final selected = ref.read(riderRequestControllerProvider).serviceCode;
    if (services == null ||
        !services.any((service) => service.code == selected)) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Choose an available ride service.')),
      );
      return;
    }
    final service = services.firstWhere((service) => service.code == selected);
    final route = ref.read(riderRoutePreviewControllerProvider);
    final fare = ref.read(riderFareControllerProvider);
    if (service.pricingRequired &&
        (route.state.loading ||
            route.state.error != null ||
            route.state.preview?.suggestedFare == null ||
            fare.version == null ||
            fare.selectionKey != route.selectionKey ||
            fare.version != route.state.preview?.pricingPolicyVersion)) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Refresh the suggested fare before requesting.'),
        ),
      );
      return;
    }
    final amount = service.pricingRequired
        ? route.state.preview?.suggestedFare?.amountMinor
        : parseFareMinor(_fare.text);
    if (amount == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Enter a valid proposed fare.')),
      );
      return;
    }
    final controller = ref.read(riderRequestControllerProvider);
    final submitted = await controller.submit(
      amountMinor: amount,
      currency: 'PKR',
      pricingPolicyVersion: route.state.preview?.pricingPolicyVersion,
    );
    if (!mounted) return;
    if (!submitted &&
        (controller.state.submissionErrorCode == 'pricing_policy_changed' ||
            controller.state.submissionErrorCode == 'suggested_fare_changed')) {
      ref.invalidate(riderServicesProvider);
      await route.retry();
    }
  }

  Future<void> _focusCurrentLocation({bool showError = true}) async {
    try {
      final point = await ref.read(deviceLocationProvider).current();
      if (!mounted) {
        return;
      }
      _mapController.move(_latLng(point), 15);
    } catch (error) {
      if (!mounted || !showError) {
        return;
      }
      ScaffoldMessenger.of(context)
          .showSnackBar(SnackBar(content: Text(error.toString())));
    }
  }

  @override
  Widget build(BuildContext context) {
    // Keep the catalog available when the lazy service picker scrolls offscreen.
    ref.watch(riderServicesProvider);
    final fareState = ref.watch(riderFareControllerProvider);
    ref.listen(riderFareControllerProvider, (_, next) => _syncFare(next.text));
    _syncFare(fareState.text);
    final controller = ref.watch(riderRequestControllerProvider);
    final state = controller.state;
    final active = state.active;
    final placeState = ref.watch(riderPlaceSearchControllerProvider).state;
    final routeState = ref.watch(riderRoutePreviewControllerProvider).state;
    final routeMatchesRide =
        active == null ||
        (active.pickup == state.pickup &&
            active.destination == state.destination);
    final routePreview = routeMatchesRide ? routeState.preview : null;
    _scheduleRouteFit(routePreview);
    final driverLocation = active == null
        ? null
        : freshDriverLocation(
            ref.watch(riderActiveRideControllerProvider(active.id)).location,
          );
    final markers = active == null
        ? _editableMarkersFor(state.pickup, state.destination)
        : _markersFor(active.pickup, active.destination);

    return RideDashboardScaffold(
      panelIdentity: state.loading && state.requests.isEmpty
          ? 'rider-loading'
          : active == null
          ? 'rider-request-form'
          : 'rider-active-${active.id}',
      minPanelSize: 0.18,
      initialPanelSize: 0.18,
      maxPanelSize: 0.60,
      map: RideMap(
        mapController: _mapController,
        markers: [
          ...markers,
          if (driverLocation != null)
            RideMapMarker(
              point: RideMapPoint(
                driverLocation.latitude,
                driverLocation.longitude,
              ),
              color: AppColors.success,
              label: 'Driver',
            ),
        ],
        polylines: routePreview != null
            ? [
                RideMapPolyline(
                  points: routePreview.points
                      .map(_latLng)
                      .toList(growable: false),
                  color: Theme.of(context).colorScheme.primary,
                  width: 6,
                ),
              ]
            : const [],
        onTap: active == null ? _handleMapTap : null,
        onPlaceTap: active == null ? _handlePlaceTap : null,
        showCenterPin: active == null && _pinSelectionMode,
        centerPinColor: _selectingPickup ? AppColors.success : AppColors.danger,
      ),
      mapControls: _RiderMapControls(
        onFocus: _focusCurrentLocation,
        pinSelectionMode: active == null && _pinSelectionMode,
        selectingPickup: _selectingPickup,
        resolving: active == null && placeState.field(_selectedField).resolving,
        onConfirmPin: _confirmPinSelection,
      ),
      floatingStatus: DashboardStatusCard(
        icon: active == null ? Icons.map_outlined : Icons.local_taxi,
        title: active == null ? 'Ride dashboard' : 'Active ride request',
        message: active == null
            ? _pinSelectionMode
                  ? 'Move the map under the pin and confirm, or tap a labeled place to select it directly.'
                  : 'Search, tap a labeled place, tap the map, or use Set on map.'
            : 'Status updates appear in the ride panel below.',
      ),
      panelBuilder: (context, scrollController, scrollEnabled) {
        return _buildPanel(
          scrollController: scrollController,
          scrollEnabled: scrollEnabled,
          state: state,
          active: active,
          controller: controller,
          placeState: placeState,
          routeState: routeState,
        );
      },
    );
  }

  void _syncFare(String text) {
    if (_fare.text == text) return;
    _fare.value = TextEditingValue(
      text: text,
      selection: TextSelection.collapsed(offset: text.length),
    );
  }

  Widget _buildPanel({
    required ScrollController scrollController,
    required bool scrollEnabled,
    required RiderRequestState state,
    required RideRequest? active,
    required RiderRequestController controller,
    required RiderPlaceSearchState placeState,
    required RiderRoutePreviewState routeState,
  }) {
    if (state.loading && state.requests.isEmpty) {
      return _LoadingPanel(
        scrollController: scrollController,
        scrollEnabled: scrollEnabled,
      );
    }
    if (active == null) {
      return _RequestRidePanel(
        scrollController: scrollController,
        scrollEnabled: scrollEnabled,
        fare: _fare,
        pricingRequired: ref.read(riderServicesProvider).asData?.value.any((service) => service.code == controller.serviceCode && service.pricingRequired) ?? false,
        onFareChanged: ref.read(riderFareControllerProvider).edit,
        pickupSearch: _pickupSearch,
        destinationSearch: _destinationSearch,
        state: state,
        placeState: placeState,
        routeState: routeState,
        selectingPickup: _selectingPickup,
        pinSelectionMode: _pinSelectionMode,
        onSelectionChanged: _selectField,
        onSetOnMap: _startPinSelection,
        onSearch: _searchPlaces,
        onSuggestionSelected: _selectSuggestion,
        onUseCurrentPickup: _useCurrentPickup,
        onRetryRoute: ref.read(riderRoutePreviewControllerProvider).retry,
        onSubmit: _submit,
      );
    }
    return _ActiveRequestPanel(
      scrollController: scrollController,
      scrollEnabled: scrollEnabled,
      request: active,
    );
  }

  RiderPlaceField get _selectedField =>
      _selectingPickup ? RiderPlaceField.pickup : RiderPlaceField.destination;

  void _selectField(bool pickup) {
    if (_selectingPickup == pickup && !_pinSelectionMode) return;
    setState(() {
      _selectingPickup = pickup;
      _pinSelectionMode = false;
    });
  }

  void _startPinSelection(RiderPlaceField field) {
    FocusScope.of(context).unfocus();
    setState(() {
      _selectingPickup = field == RiderPlaceField.pickup;
      _pinSelectionMode = true;
    });
  }

  Future<void> _confirmPinSelection() async {
    if (!_pinSelectionMode) return;
    final field = _selectedField;
    if (ref
        .read(riderPlaceSearchControllerProvider)
        .state
        .field(field)
        .resolving) {
      return;
    }
    final point = _mapController.center;
    if (point == null || !point.isValid) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Move the map before confirming the pin.'),
        ),
      );
      return;
    }
    final resolved = await _reconcilePoint(field, point);
    if (!mounted || resolved == null) return;
    if (!mounted) return;
    setState(() {
      _pinSelectionMode = false;
      if (field == RiderPlaceField.pickup) {
        _selectingPickup = false;
      }
    });
  }

  void _handleMapTap(RideMapPoint point) {
    if (_pinSelectionMode) {
      unawaited(_mapController.move(point, 16));
      return;
    }
    final field = _selectedField;
    unawaited(_reconcilePoint(field, point));
    if (_selectingPickup) {
      setState(() => _selectingPickup = false);
    }
  }

  Future<void> _handlePlaceTap(RideMapPlace place) async {
    final field = _selectedField;
    if (ref
        .read(riderPlaceSearchControllerProvider)
        .state
        .field(field)
        .resolving) {
      return;
    }

    final selected = await ref
        .read(riderPlaceSearchControllerProvider)
        .selectPlaceId(field, place.placeId);
    if (!mounted || selected == null) return;

    _searchController(field).text = selected.label;
    FocusScope.of(context).unfocus();
    setState(() {
      _pinSelectionMode = false;
      if (field == RiderPlaceField.pickup) {
        _selectingPickup = false;
      }
    });
  }

  void _searchPlaces(RiderPlaceField field, String input) {
    if ((field == RiderPlaceField.pickup) != _selectingPickup ||
        _pinSelectionMode) {
      setState(() {
        _selectingPickup = field == RiderPlaceField.pickup;
        _pinSelectionMode = false;
      });
    }
    ref.read(riderPlaceSearchControllerProvider).search(field, input);
  }

  Future<void> _selectSuggestion(
    RiderPlaceField field,
    PlaceSuggestion suggestion,
  ) async {
    if (_pinSelectionMode) {
      setState(() => _pinSelectionMode = false);
    }
    final selected = await ref
        .read(riderPlaceSearchControllerProvider)
        .select(field, suggestion);
    if (!mounted || selected == null) return;
    _searchController(field).text = selected.label;
    FocusScope.of(context).unfocus();
    await _mapController.move(_latLng(selected.point), 15);
  }

  Future<void> _useCurrentPickup() async {
    if (_pinSelectionMode) {
      setState(() => _pinSelectionMode = false);
    }
    final rider = ref.read(riderRequestControllerProvider);
    await rider.useCurrentPickup();
    if (!mounted) return;
    final point = rider.state.pickup;
    if (point == null) return;
    await _reconcilePoint(
      RiderPlaceField.pickup,
      _latLng(point),
      moveCamera: true,
    );
  }

  Future<PlaceSelection?> _reconcilePoint(
    RiderPlaceField field,
    RideMapPoint mapPoint, {
    bool moveCamera = false,
  }) async {
    final point = GeoPoint(
      latitude: mapPoint.latitude,
      longitude: mapPoint.longitude,
    );
    final resolved = await ref
        .read(riderPlaceSearchControllerProvider)
        .reconcilePin(field, point);
    if (!mounted) return null;
    _searchController(field).text = resolved?.label ?? _formatPoint(point);
    if (moveCamera) {
      await _mapController.move(
        resolved == null ? mapPoint : _latLng(resolved.point),
        15,
      );
    }
    return resolved;
  }

  TextEditingController _searchController(RiderPlaceField field) =>
      field == RiderPlaceField.pickup ? _pickupSearch : _destinationSearch;

  String _formatPoint(GeoPoint point) =>
      '${point.latitude.toStringAsFixed(5)}, '
      '${point.longitude.toStringAsFixed(5)}';

  List<RideMapMarker> _editableMarkersFor(
    GeoPoint? pickup,
    GeoPoint? destination,
  ) => [
    if (pickup != null)
      RideMapMarker(
        point: _latLng(pickup),
        color: AppColors.success,
        label: 'Pickup',
        onDragEnd: (point) =>
            unawaited(_reconcilePoint(RiderPlaceField.pickup, point)),
      ),
    if (destination != null)
      RideMapMarker(
        point: _latLng(destination),
        color: AppColors.danger,
        label: 'Destination',
        onDragEnd: (point) =>
            unawaited(_reconcilePoint(RiderPlaceField.destination, point)),
      ),
  ];

  List<RideMapMarker> _markersFor(GeoPoint? pickup, GeoPoint? destination) => [
    if (pickup != null)
      RideMapMarker(
        point: _latLng(pickup),
        color: AppColors.success,
        label: 'Pickup',
      ),
    if (destination != null)
      RideMapMarker(
        point: _latLng(destination),
        color: AppColors.danger,
        label: 'Destination',
      ),
  ];

  void _scheduleRouteFit(RoutePreview? preview) {
    if (preview == null) {
      _lastFittedRouteKey = null;
      return;
    }
    final key = preview.encodedPolyline;
    if (_lastFittedRouteKey == key) {
      return;
    }
    _lastFittedRouteKey = key;
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted) return;
      unawaited(_mapController.fit(preview.points.map(_latLng), padding: 48));
    });
  }

  RideMapPoint _latLng(GeoPoint point) =>
      RideMapPoint(point.latitude, point.longitude);
}

class _RiderMapControls extends StatelessWidget {
  const _RiderMapControls({
    required this.onFocus,
    required this.pinSelectionMode,
    required this.selectingPickup,
    required this.resolving,
    required this.onConfirmPin,
  });

  final Future<void> Function() onFocus;
  final bool pinSelectionMode;
  final bool selectingPickup;
  final bool resolving;
  final Future<void> Function() onConfirmPin;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.end,
      mainAxisSize: MainAxisSize.min,
      children: [
        if (pinSelectionMode) ...[
          FloatingActionButton.extended(
            key: const Key('confirmPinButton'),
            heroTag: 'rider-confirm-pin',
            onPressed: resolving ? null : () => onConfirmPin(),
            icon: resolving
                ? const SizedBox.square(
                    dimension: 16,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                : const Icon(Icons.check),
            label: Text(
              resolving
                  ? 'Resolving location...'
                  : 'Confirm ${selectingPickup ? 'pickup' : 'destination'}',
            ),
          ),
          const SizedBox(height: AppSpacing.xs),
        ],
        FloatingActionButton.small(
          heroTag: 'rider-current-location',
          tooltip: 'Center map on your location',
          onPressed: () => onFocus(),
          child: const Icon(Icons.my_location),
        ),
      ],
    );
  }
}

class _LoadingPanel extends StatelessWidget {
  const _LoadingPanel({
    required this.scrollController,
    required this.scrollEnabled,
  });

  final ScrollController scrollController;
  final bool scrollEnabled;

  @override
  Widget build(BuildContext context) {
    return ListView(
      controller: scrollController,
      physics: scrollEnabled
          ? const ClampingScrollPhysics()
          : const NeverScrollableScrollPhysics(),
      padding: const EdgeInsets.all(AppSpacing.md),
      children: const [
        SizedBox(height: 96, child: Center(child: CircularProgressIndicator())),
      ],
    );
  }
}

class _RequestRidePanel extends StatelessWidget {
  const _RequestRidePanel({
    required this.scrollController,
    required this.scrollEnabled,
    required this.fare,
    required this.pricingRequired,
    required this.onFareChanged,
    required this.pickupSearch,
    required this.destinationSearch,
    required this.state,
    required this.placeState,
    required this.routeState,
    required this.selectingPickup,
    required this.pinSelectionMode,
    required this.onSelectionChanged,
    required this.onSetOnMap,
    required this.onSearch,
    required this.onSuggestionSelected,
    required this.onUseCurrentPickup,
    required this.onRetryRoute,
    required this.onSubmit,
  });

  final ScrollController scrollController;
  final bool scrollEnabled;
  final TextEditingController fare;
  final bool pricingRequired;
  final ValueChanged<String> onFareChanged;
  final TextEditingController pickupSearch;
  final TextEditingController destinationSearch;
  final RiderRequestState state;
  final RiderPlaceSearchState placeState;
  final RiderRoutePreviewState routeState;
  final bool selectingPickup;
  final bool pinSelectionMode;
  final ValueChanged<bool> onSelectionChanged;
  final ValueChanged<RiderPlaceField> onSetOnMap;
  final void Function(RiderPlaceField field, String input) onSearch;
  final Future<void> Function(RiderPlaceField field, PlaceSuggestion suggestion)
  onSuggestionSelected;
  final Future<void> Function() onUseCurrentPickup;
  final Future<void> Function() onRetryRoute;
  final Future<void> Function() onSubmit;

  @override
  Widget build(BuildContext context) {
    return ListView(
      controller: scrollController,
      physics: scrollEnabled
          ? const ClampingScrollPhysics()
          : const NeverScrollableScrollPhysics(),
      padding: const EdgeInsets.all(AppSpacing.md),
      children: [
        Text(
          'Where are you going?',
          style: Theme.of(context).textTheme.headlineSmall,
        ),
        const RiderRideHistory(),
        const SizedBox(height: AppSpacing.xs),
        const Text(
          'Choose pickup and destination. Select a Driver offer to agree the final fare.',
        ),
        const SizedBox(height: AppSpacing.sm),
        const RiderServicePicker(),
        DashboardPanelControl(
          child: SegmentedButton<bool>(
            segments: const [
              ButtonSegment(
                value: true,
                label: Text('Pickup'),
                icon: Icon(Icons.my_location),
              ),
              ButtonSegment(
                value: false,
                label: Text('Destination'),
                icon: Icon(Icons.flag),
              ),
            ],
            selected: {selectingPickup},
            onSelectionChanged: (value) => onSelectionChanged(value.single),
          ),
        ),
        const SizedBox(height: AppSpacing.sm),
        _PlaceSearchField(
          field: RiderPlaceField.pickup,
          controller: pickupSearch,
          state: placeState.pickup,
          selected: selectingPickup,
          pinSelectionActive: pinSelectionMode && selectingPickup,
          onFocus: () => onSelectionChanged(true),
          onSetOnMap: () => onSetOnMap(RiderPlaceField.pickup),
          onChanged: (value) => onSearch(RiderPlaceField.pickup, value),
          onSuggestionSelected: (suggestion) =>
              onSuggestionSelected(RiderPlaceField.pickup, suggestion),
        ),
        const SizedBox(height: AppSpacing.xs),
        _PlaceSearchField(
          field: RiderPlaceField.destination,
          controller: destinationSearch,
          state: placeState.destination,
          selected: !selectingPickup,
          pinSelectionActive: pinSelectionMode && !selectingPickup,
          onFocus: () => onSelectionChanged(false),
          onSetOnMap: () => onSetOnMap(RiderPlaceField.destination),
          onChanged: (value) => onSearch(RiderPlaceField.destination, value),
          onSuggestionSelected: (suggestion) =>
              onSuggestionSelected(RiderPlaceField.destination, suggestion),
        ),
        const SizedBox(height: AppSpacing.sm),
        DashboardPanelControl(
          child: OutlinedButton.icon(
            onPressed: state.locating ? null : () => onUseCurrentPickup(),
            icon: state.locating
                ? const SizedBox.square(
                    dimension: 18,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                : const Icon(Icons.gps_fixed),
            label: const Text('Use current location for pickup'),
          ),
        ),
        const SizedBox(height: AppSpacing.xs),
        _PointSummary(
          label: 'Pickup',
          point: state.pickup,
          address: placeState.pickup.label,
        ),
        _PointSummary(
          label: 'Destination',
          point: state.destination,
          address: placeState.destination.label,
        ),
        const SizedBox(height: AppSpacing.sm),
        _RoutePreviewCard(
          state: routeState,
          hasEndpoints: state.pickup != null && state.destination != null,
          onRetry: onRetryRoute,
        ),
        const SizedBox(height: AppSpacing.sm),
        DashboardPanelControl(
          child: TextField(
            key: const Key('fareField'),
            controller: fare,
            readOnly: pricingRequired,
            onChanged: pricingRequired ? null : onFareChanged,
            keyboardType: const TextInputType.numberWithOptions(decimal: true),
            decoration: InputDecoration(
              labelText: pricingRequired ? 'Suggested fare' : 'Your proposed fare',
              prefixText: 'PKR ',
            ),
          ),
        ),
        if (state.error != null) ...[
          const SizedBox(height: AppSpacing.sm),
          Text(
            state.error!,
            style: TextStyle(color: Theme.of(context).colorScheme.error),
          ),
        ],
        const SizedBox(height: AppSpacing.md),
        DashboardPanelControl(
          child: FilledButton.icon(
            key: const Key('requestRideButton'),
            onPressed: state.submitting ? null : () => onSubmit(),
            icon: const Icon(Icons.local_taxi),
            label: Text(state.submitting ? 'Requesting...' : 'Request ride'),
          ),
        ),
      ],
    );
  }
}

class _PlaceSearchField extends StatelessWidget {
  const _PlaceSearchField({
    required this.field,
    required this.controller,
    required this.state,
    required this.selected,
    required this.pinSelectionActive,
    required this.onFocus,
    required this.onSetOnMap,
    required this.onChanged,
    required this.onSuggestionSelected,
  });

  final RiderPlaceField field;
  final TextEditingController controller;
  final PlaceFieldSearchState state;
  final bool selected;
  final bool pinSelectionActive;
  final VoidCallback onFocus;
  final VoidCallback onSetOnMap;
  final ValueChanged<String> onChanged;
  final ValueChanged<PlaceSuggestion> onSuggestionSelected;

  @override
  Widget build(BuildContext context) {
    final pickup = field == RiderPlaceField.pickup;
    final label = pickup ? 'Pickup search' : 'Destination search';
    return DashboardPanelControl(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          TextField(
            key: Key(pickup ? 'pickupSearchField' : 'destinationSearchField'),
            controller: controller,
            onTap: onFocus,
            onChanged: onChanged,
            textInputAction: TextInputAction.search,
            decoration: InputDecoration(
              labelText: label,
              hintText: pickup
                  ? 'Search pickup address or place'
                  : 'Search destination address or place',
              prefixIcon: Icon(pickup ? Icons.my_location : Icons.flag),
              suffixIcon: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  if (state.searching || state.resolving)
                    const Padding(
                      padding: EdgeInsets.all(AppSpacing.sm),
                      child: SizedBox.square(
                        dimension: 18,
                        child: CircularProgressIndicator(strokeWidth: 2),
                      ),
                    )
                  else if (selected)
                    const Icon(Icons.edit_location_alt_outlined),
                  IconButton(
                    key: Key(
                      pickup
                          ? 'pickupSetOnMapButton'
                          : 'destinationSetOnMapButton',
                    ),
                    tooltip: pickup
                        ? 'Set pickup on map'
                        : 'Set destination on map',
                    onPressed: onSetOnMap,
                    icon: Icon(
                      pinSelectionActive
                          ? Icons.location_pin
                          : Icons.pin_drop_outlined,
                    ),
                  ),
                ],
              ),
            ),
          ),
          if (state.suggestions.isNotEmpty)
            ...state.suggestions.map(
              (suggestion) => ListTile(
                dense: true,
                contentPadding: EdgeInsets.zero,
                leading: const Icon(Icons.location_on_outlined),
                title: Text(suggestion.label),
                onTap: () => onSuggestionSelected(suggestion),
              ),
            ),
          if (state.searched &&
              !state.searching &&
              state.suggestions.isEmpty &&
              state.error == null)
            const Padding(
              padding: EdgeInsets.only(top: AppSpacing.xs),
              child: Text('No matching places found.'),
            ),
          if (state.error != null)
            Padding(
              padding: const EdgeInsets.only(top: AppSpacing.xs),
              child: Text(
                state.error!,
                style: TextStyle(color: Theme.of(context).colorScheme.error),
              ),
            ),
        ],
      ),
    );
  }
}

class _RoutePreviewCard extends StatelessWidget {
  const _RoutePreviewCard({
    required this.state,
    required this.hasEndpoints,
    required this.onRetry,
  });

  final RiderRoutePreviewState state;
  final bool hasEndpoints;
  final Future<void> Function() onRetry;

  @override
  Widget build(BuildContext context) {
    if (!hasEndpoints) {
      return const Text(
        'Select pickup and destination to calculate the driving route.',
      );
    }
    if (state.loading) {
      return const Card(
        child: Padding(
          padding: EdgeInsets.all(AppSpacing.md),
          child: Row(
            children: [
              SizedBox.square(
                dimension: 18,
                child: CircularProgressIndicator(strokeWidth: 2),
              ),
              SizedBox(width: AppSpacing.sm),
              Expanded(child: Text('Calculating traffic-aware route...')),
            ],
          ),
        ),
      );
    }
    if (state.error != null) {
      return Card(
        child: Padding(
          padding: const EdgeInsets.all(AppSpacing.md),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Text(
                state.error!,
                style: TextStyle(color: Theme.of(context).colorScheme.error),
              ),
              const SizedBox(height: AppSpacing.xs),
              DashboardPanelControl(
                child: OutlinedButton.icon(
                  key: const Key('retryRoutePreviewButton'),
                  onPressed: () => onRetry(),
                  icon: const Icon(Icons.refresh),
                  label: const Text('Retry route'),
                ),
              ),
            ],
          ),
        ),
      );
    }

    final preview = state.preview;
    if (preview == null) {
      return const SizedBox.shrink();
    }

    return Card(
      key: const Key('routePreviewSummary'),
      child: Padding(
        padding: const EdgeInsets.symmetric(vertical: AppSpacing.xs),
        child: ListTile(
          leading: const Icon(Icons.route),
          title: Text(
            '${_formatRouteDuration(preview.durationSeconds)} · '
            '${_formatRouteDistance(preview.distanceMeters)}',
          ),
          subtitle: const Text('Google recommended · current traffic'),
        ),
      ),
    );
  }
}

String _formatRouteDistance(int meters) {
  if (meters < 1000) {
    return '$meters m';
  }
  final kilometers = meters / 1000;
  return '${kilometers.toStringAsFixed(kilometers < 10 ? 1 : 0)} km';
}

String _formatRouteDuration(int seconds) {
  final minutes = (seconds + 59) ~/ 60;
  if (minutes < 60) {
    return '$minutes min';
  }
  final hours = minutes ~/ 60;
  final remainder = minutes % 60;
  return remainder == 0 ? '$hours hr' : '$hours hr $remainder min';
}

class _PointSummary extends StatelessWidget {
  const _PointSummary({required this.label, required this.point, this.address});

  final String label;
  final GeoPoint? point;
  final String? address;

  @override
  Widget build(BuildContext context) {
    final value = point;
    final coordinates = value == null
        ? null
        : '${value.latitude.toStringAsFixed(5)}, '
              '${value.longitude.toStringAsFixed(5)}';
    return ListTile(
      dense: true,
      contentPadding: EdgeInsets.zero,
      leading: Icon(label == 'Pickup' ? Icons.my_location : Icons.flag),
      title: Text(label),
      subtitle: Text(
        value == null
            ? 'Not selected'
            : address == null || address!.trim().isEmpty
            ? coordinates!
            : '$address\n$coordinates',
      ),
    );
  }
}

class _ActiveRequestPanel extends ConsumerWidget {
  const _ActiveRequestPanel({
    required this.scrollController,
    required this.scrollEnabled,
    required this.request,
  });

  final ScrollController scrollController;
  final bool scrollEnabled;
  final RideRequest request;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final state = ref.watch(riderRequestControllerProvider).state;
    final status = request.trip?.status ?? request.status;
    return ListView(
      controller: scrollController,
      physics: scrollEnabled
          ? const ClampingScrollPhysics()
          : const NeverScrollableScrollPhysics(),
      padding: const EdgeInsets.all(AppSpacing.md),
      children: [
        Text(
          _statusTitle(status),
          style: Theme.of(context).textTheme.headlineSmall,
        ),
        const SizedBox(height: AppSpacing.xs),
        RideFlowPanel.rider(rideId: request.id),
        const SizedBox(height: AppSpacing.md),
        Card(
          child: Padding(
            padding: const EdgeInsets.all(AppSpacing.md),
            child: Column(
              children: [
                _PointSummary(label: 'Pickup', point: request.pickup),
                _PointSummary(label: 'Destination', point: request.destination),
                ListTile(
                  contentPadding: EdgeInsets.zero,
                  leading: const Icon(Icons.payments_outlined),
                  title: const Text('Your proposed fare'),
                  subtitle: Text(
                    request.proposedFare == null
                        ? 'Fare unavailable'
                        : '${request.proposedFare!.currency} '
                              '${(request.proposedFare!.amountMinor / 100).toStringAsFixed(2)}',
                  ),
                ),
              ],
            ),
          ),
        ),
        const SizedBox(height: AppSpacing.sm),
        DashboardPanelControl(
          child: OutlinedButton.icon(
            onPressed: state.loading
                ? null
                : ref.read(riderRequestControllerProvider).refreshActive,
            icon: const Icon(Icons.refresh),
            label: const Text('Refresh status'),
          ),
        ),
        if (status != 'completed' && status != 'cancelled') ...[
          const SizedBox(height: AppSpacing.xs),
          DashboardPanelControl(
            child: TextButton(
              onPressed: state.submitting
                  ? null
                  : () => _confirmCancellation(context, ref),
              child: Text(
                state.submitting ? 'Cancelling...' : 'Cancel request',
              ),
            ),
          ),
        ],
        if (state.error != null) ...[
          const SizedBox(height: AppSpacing.sm),
          Text(
            state.error!,
            style: TextStyle(color: Theme.of(context).colorScheme.error),
          ),
        ],
      ],
    );
  }

  String _statusTitle(String status) => switch (status) {
    'requested' => 'Looking for Driver offers',
    'assigned' => 'Driver assigned',
    'in_progress' => 'Trip in progress',
    'completed' => 'Trip completed',
    'cancelled' => 'Trip cancelled',
    _ => 'Ride request updated',
  };

  Future<void> _confirmCancellation(BuildContext context, WidgetRef ref) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Cancel this request?'),
        content: const Text(
          'Drivers will no longer be able to offer on this ride.',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Keep request'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            child: const Text('Cancel request'),
          ),
        ],
      ),
    );
    if (confirmed == true) {
      await ref.read(riderRequestControllerProvider).cancelActive();
    }
  }
}
