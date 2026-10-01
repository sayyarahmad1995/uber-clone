import 'dart:async';

import 'package:flutter/material.dart';

import '../../ride_flow/ride_flow_panels.dart';

import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/dashboard/ride_dashboard_scaffold.dart';
import '../../../core/maps/ride_map.dart';
import '../../../core/providers.dart';
import '../../../core/theme/app_theme.dart';
import '../application/rider_place_search_controller.dart';
import '../application/rider_request_controller.dart';
import '../domain/place_search.dart';
import '../domain/ride_request.dart';

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
    final amount = parseFareMinor(_fare.text);
    if (amount == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Enter a valid proposed fare.')),
      );
      return;
    }
    await ref
        .read(riderRequestControllerProvider)
        .submit(amountMinor: amount, currency: 'PKR');
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
    final controller = ref.watch(riderRequestControllerProvider);
    final state = controller.state;
    final active = state.active;
    final placeState = ref.watch(riderPlaceSearchControllerProvider).state;
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
        onTap: active == null ? _handleMapTap : null,
      ),
      mapControls: _MapFocusButton(onPressed: _focusCurrentLocation),
      floatingStatus: DashboardStatusCard(
        icon: active == null ? Icons.map_outlined : Icons.local_taxi,
        title: active == null ? 'Ride dashboard' : 'Active ride request',
        message: active == null
            ? 'Tap the map to choose pickup and destination.'
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
        );
      },
    );
  }

  Widget _buildPanel({
    required ScrollController scrollController,
    required bool scrollEnabled,
    required RiderRequestState state,
    required RideRequest? active,
    required RiderRequestController controller,
    required RiderPlaceSearchState placeState,
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
        pickupSearch: _pickupSearch,
        destinationSearch: _destinationSearch,
        state: state,
        placeState: placeState,
        selectingPickup: _selectingPickup,
        onSelectionChanged: (value) => setState(() => _selectingPickup = value),
        onSearch: _searchPlaces,
        onSuggestionSelected: _selectSuggestion,
        onUseCurrentPickup: _useCurrentPickup,
        onSubmit: _submit,
      );
    }
    return _ActiveRequestPanel(
      scrollController: scrollController,
      scrollEnabled: scrollEnabled,
      request: active,
    );
  }

  void _handleMapTap(RideMapPoint point) {
    final field = _selectingPickup
        ? RiderPlaceField.pickup
        : RiderPlaceField.destination;
    unawaited(_reconcilePoint(field, point));
    if (_selectingPickup) {
      setState(() => _selectingPickup = false);
    }
  }

  void _searchPlaces(RiderPlaceField field, String input) {
    if ((field == RiderPlaceField.pickup) != _selectingPickup) {
      setState(() => _selectingPickup = field == RiderPlaceField.pickup);
    }
    ref.read(riderPlaceSearchControllerProvider).search(field, input);
  }

  Future<void> _selectSuggestion(
    RiderPlaceField field,
    PlaceSuggestion suggestion,
  ) async {
    final selected = await ref
        .read(riderPlaceSearchControllerProvider)
        .select(field, suggestion);
    if (!mounted || selected == null) return;
    _searchController(field).text = selected.label;
    FocusScope.of(context).unfocus();
    await _mapController.move(_latLng(selected.point), 15);
  }

  Future<void> _useCurrentPickup() async {
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

  Future<void> _reconcilePoint(
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
    if (!mounted) return;
    _searchController(field).text = resolved?.label ?? _formatPoint(point);
    if (moveCamera) {
      await _mapController.move(mapPoint, 15);
    }
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

  RideMapPoint _latLng(GeoPoint point) =>
      RideMapPoint(point.latitude, point.longitude);
}

class _MapFocusButton extends StatelessWidget {
  const _MapFocusButton({required this.onPressed});

  final Future<void> Function() onPressed;

  @override
  Widget build(BuildContext context) {
    return FloatingActionButton.small(
      heroTag: 'rider-current-location',
      tooltip: 'Center map on your location',
      onPressed: () => onPressed(),
      child: const Icon(Icons.my_location),
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
    required this.pickupSearch,
    required this.destinationSearch,
    required this.state,
    required this.placeState,
    required this.selectingPickup,
    required this.onSelectionChanged,
    required this.onSearch,
    required this.onSuggestionSelected,
    required this.onUseCurrentPickup,
    required this.onSubmit,
  });

  final ScrollController scrollController;
  final bool scrollEnabled;
  final TextEditingController fare;
  final TextEditingController pickupSearch;
  final TextEditingController destinationSearch;
  final RiderRequestState state;
  final RiderPlaceSearchState placeState;
  final bool selectingPickup;
  final ValueChanged<bool> onSelectionChanged;
  final void Function(RiderPlaceField field, String input) onSearch;
  final Future<void> Function(RiderPlaceField field, PlaceSuggestion suggestion)
  onSuggestionSelected;
  final Future<void> Function() onUseCurrentPickup;
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
          'Choose pickup and destination, then propose the fare you want to pay.',
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
          onFocus: () => onSelectionChanged(true),
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
          onFocus: () => onSelectionChanged(false),
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
        DashboardPanelControl(
          child: TextField(
            key: const Key('fareField'),
            controller: fare,
            keyboardType: const TextInputType.numberWithOptions(decimal: true),
            decoration: const InputDecoration(
              labelText: 'Your proposed fare',
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
    required this.onFocus,
    required this.onChanged,
    required this.onSuggestionSelected,
  });

  final RiderPlaceField field;
  final TextEditingController controller;
  final PlaceFieldSearchState state;
  final bool selected;
  final VoidCallback onFocus;
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
              suffixIcon: state.searching || state.resolving
                  ? const Padding(
                      padding: EdgeInsets.all(AppSpacing.sm),
                      child: SizedBox.square(
                        dimension: 18,
                        child: CircularProgressIndicator(strokeWidth: 2),
                      ),
                    )
                  : selected
                  ? const Icon(Icons.edit_location_alt_outlined)
                  : null,
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
