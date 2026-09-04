import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:latlong2/latlong.dart';

import '../../../core/dashboard/ride_dashboard_scaffold.dart';
import '../../../core/maps/ride_map.dart';
import '../../../core/providers.dart';
import '../../../core/theme/app_theme.dart';
import '../domain/ride_request.dart';

class RiderRequestScreen extends ConsumerStatefulWidget {
  const RiderRequestScreen({super.key});

  @override
  ConsumerState<RiderRequestScreen> createState() => _RiderRequestScreenState();
}

class _RiderRequestScreenState extends ConsumerState<RiderRequestScreen> {
  final _fare = TextEditingController();
  bool _selectingPickup = true;

  @override
  void dispose() {
    _fare.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    final amount = double.tryParse(_fare.text.trim());
    if (amount == null || amount <= 0) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Enter a valid proposed fare.')),
      );
      return;
    }
    await ref
        .read(riderRequestControllerProvider)
        .submit(amountMinor: (amount * 100).round(), currency: 'PKR');
  }

  @override
  Widget build(BuildContext context) {
    final controller = ref.watch(riderRequestControllerProvider);
    final state = controller.state;
    final active = state.active;
    final tiles = ref.watch(mapTilesProvider);
    final markers = active == null
        ? _markersFor(state.pickup, state.destination)
        : _markersFor(active.pickup, active.destination);

    return RideDashboardScaffold(
      maxPanelHeightFactor: active == null ? 0.70 : 0.58,
      map: RideMap(
        tiles: tiles,
        markers: markers,
        onTap: active == null ? _handleMapTap : null,
      ),
      floatingStatus: DashboardStatusCard(
        icon: active == null ? Icons.map_outlined : Icons.local_taxi,
        title: active == null ? 'Ride dashboard' : _statusTitle(active),
        message: active == null
            ? 'Tap the map to choose pickup and destination.'
            : 'Refresh to pick up assignment or trip status changes.',
      ),
      panel: state.loading && state.requests.isEmpty
          ? const _LoadingPanel()
          : active == null
          ? _RequestRidePanel(
              fare: _fare,
              state: state,
              selectingPickup: _selectingPickup,
              onSelectionChanged: (value) =>
                  setState(() => _selectingPickup = value),
              onUseCurrentPickup: controller.useCurrentPickup,
              onSubmit: _submit,
            )
          : _ActiveRequestPanel(request: active),
    );
  }

  void _handleMapTap(LatLng point) {
    final controller = ref.read(riderRequestControllerProvider);
    final selected = GeoPoint(
      latitude: point.latitude,
      longitude: point.longitude,
    );
    if (_selectingPickup) {
      controller.setPickup(selected);
      setState(() => _selectingPickup = false);
    } else {
      controller.setDestination(selected);
    }
  }

  List<RideMapMarker> _markersFor(GeoPoint? pickup, GeoPoint? destination) => [
    if (pickup != null)
      RideMapMarker(
        point: _latLng(pickup),
        icon: Icons.my_location,
        color: AppColors.success,
        label: 'Pickup',
      ),
    if (destination != null)
      RideMapMarker(
        point: _latLng(destination),
        icon: Icons.flag,
        color: AppColors.danger,
        label: 'Destination',
      ),
  ];

  LatLng _latLng(GeoPoint point) => LatLng(point.latitude, point.longitude);

  String _statusTitle(RideRequest request) {
    final status = request.trip?.status ?? request.status;
    return switch (status) {
      'requested' => 'Looking for Driver offers',
      'assigned' => 'Driver assigned',
      'in_progress' => 'Trip in progress',
      'completed' => 'Trip completed',
      'cancelled' => 'Trip cancelled',
      _ => 'Ride request updated',
    };
  }
}

class _LoadingPanel extends StatelessWidget {
  const _LoadingPanel();

  @override
  Widget build(BuildContext context) => const Padding(
    padding: EdgeInsets.all(AppSpacing.xl),
    child: Center(child: CircularProgressIndicator()),
  );
}

class _RequestRidePanel extends StatelessWidget {
  const _RequestRidePanel({
    required this.fare,
    required this.state,
    required this.selectingPickup,
    required this.onSelectionChanged,
    required this.onUseCurrentPickup,
    required this.onSubmit,
  });

  final TextEditingController fare;
  final RiderRequestState state;
  final bool selectingPickup;
  final ValueChanged<bool> onSelectionChanged;
  final Future<void> Function() onUseCurrentPickup;
  final Future<void> Function() onSubmit;

  @override
  Widget build(BuildContext context) => ListView(
    shrinkWrap: true,
    padding: const EdgeInsets.all(AppSpacing.md),
    children: [
      Text(
        'Where are you going?',
        style: Theme.of(context).textTheme.headlineSmall,
      ),
      const SizedBox(height: AppSpacing.xs),
      const Text(
        'Choose pickup and destination, then propose the fare you want to pay.',
      ),
      const SizedBox(height: AppSpacing.sm),
      SegmentedButton<bool>(
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
      const SizedBox(height: AppSpacing.sm),
      OutlinedButton.icon(
        onPressed: state.locating ? null : () => onUseCurrentPickup(),
        icon: state.locating
            ? const SizedBox.square(
                dimension: 18,
                child: CircularProgressIndicator(strokeWidth: 2),
              )
            : const Icon(Icons.gps_fixed),
        label: const Text('Use current location for pickup'),
      ),
      const SizedBox(height: AppSpacing.xs),
      _PointSummary(label: 'Pickup', point: state.pickup),
      _PointSummary(label: 'Destination', point: state.destination),
      const SizedBox(height: AppSpacing.sm),
      TextField(
        key: const Key('fareField'),
        controller: fare,
        keyboardType: const TextInputType.numberWithOptions(decimal: true),
        decoration: const InputDecoration(
          labelText: 'Your proposed fare',
          prefixText: 'PKR ',
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
      FilledButton.icon(
        key: const Key('requestRideButton'),
        onPressed: state.submitting ? null : () => onSubmit(),
        icon: const Icon(Icons.local_taxi),
        label: Text(state.submitting ? 'Requesting…' : 'Request ride'),
      ),
    ],
  );
}

class _PointSummary extends StatelessWidget {
  const _PointSummary({required this.label, required this.point});

  final String label;
  final GeoPoint? point;

  @override
  Widget build(BuildContext context) => ListTile(
    dense: true,
    contentPadding: EdgeInsets.zero,
    leading: Icon(label == 'Pickup' ? Icons.my_location : Icons.flag),
    title: Text(label),
    subtitle: Text(
      point == null
          ? 'Not selected'
          : '${point!.latitude.toStringAsFixed(5)}, ${point!.longitude.toStringAsFixed(5)}',
    ),
  );
}

class _ActiveRequestPanel extends ConsumerWidget {
  const _ActiveRequestPanel({required this.request});

  final RideRequest request;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final state = ref.watch(riderRequestControllerProvider).state;
    final status = request.trip?.status ?? request.status;
    return ListView(
      padding: const EdgeInsets.all(AppSpacing.md),
      shrinkWrap: true,
      children: [
        Text(
          _statusTitle(status),
          style: Theme.of(context).textTheme.headlineSmall,
        ),
        const SizedBox(height: AppSpacing.xs),
        Text('Request ${request.id}'),
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
                    '${request.proposedFare.currency} '
                    '${(request.proposedFare.amountMinor / 100).toStringAsFixed(2)}',
                  ),
                ),
              ],
            ),
          ),
        ),
        const SizedBox(height: AppSpacing.sm),
        OutlinedButton.icon(
          onPressed: state.loading
              ? null
              : ref.read(riderRequestControllerProvider).refreshActive,
          icon: const Icon(Icons.refresh),
          label: const Text('Refresh status'),
        ),
        if (status != 'completed' && status != 'cancelled') ...[
          const SizedBox(height: AppSpacing.xs),
          TextButton(
            onPressed: state.submitting
                ? null
                : () => _confirmCancellation(context, ref),
            child: Text(state.submitting ? 'Cancelling…' : 'Cancel request'),
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
