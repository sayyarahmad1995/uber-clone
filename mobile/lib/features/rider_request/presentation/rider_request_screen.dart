import 'package:flutter/material.dart';
import 'package:flutter_map/flutter_map.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:latlong2/latlong.dart';

import '../../../core/providers.dart';
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
    if (state.loading && state.requests.isEmpty) {
      return const Center(child: CircularProgressIndicator());
    }
    final active = state.active;
    if (active != null) {
      return _ActiveRequest(request: active);
    }
    final tiles = ref.watch(mapTilesProvider);
    return ListView(
      padding: const EdgeInsets.all(16),
      children: [
        Text(
          'Where are you going?',
          style: Theme.of(context).textTheme.headlineSmall,
        ),
        const SizedBox(height: 8),
        const Text(
          'Tap the map to choose each point, or use your current location for pickup.',
        ),
        const SizedBox(height: 12),
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
          selected: {_selectingPickup},
          onSelectionChanged: (value) =>
              setState(() => _selectingPickup = value.single),
        ),
        const SizedBox(height: 12),
        SizedBox(
          height: 330,
          child: ClipRRect(
            borderRadius: BorderRadius.circular(16),
            child: FlutterMap(
              options: MapOptions(
                initialCenter: const LatLng(24.8607, 67.0011),
                initialZoom: 12,
                onTap: (_, point) {
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
                },
              ),
              children: [
                TileLayer(
                  urlTemplate: tiles.urlTemplate,
                  userAgentPackageName: tiles.userAgentPackageName,
                ),
                MarkerLayer(
                  markers: [
                    if (state.pickup != null)
                      _marker(state.pickup!, Icons.my_location, Colors.green),
                    if (state.destination != null)
                      _marker(state.destination!, Icons.flag, Colors.red),
                  ],
                ),
                RichAttributionWidget(
                  attributions: [TextSourceAttribution(tiles.attribution)],
                ),
              ],
            ),
          ),
        ),
        const SizedBox(height: 8),
        OutlinedButton.icon(
          onPressed: state.locating ? null : controller.useCurrentPickup,
          icon: state.locating
              ? const SizedBox.square(
                  dimension: 18,
                  child: CircularProgressIndicator(strokeWidth: 2),
                )
              : const Icon(Icons.gps_fixed),
          label: const Text('Use my current location for pickup'),
        ),
        const SizedBox(height: 8),
        _PointSummary(label: 'Pickup', point: state.pickup),
        _PointSummary(label: 'Destination', point: state.destination),
        const SizedBox(height: 12),
        TextField(
          key: const Key('fareField'),
          controller: _fare,
          keyboardType: const TextInputType.numberWithOptions(decimal: true),
          decoration: const InputDecoration(
            labelText: 'Your proposed fare',
            prefixText: 'PKR ',
            border: OutlineInputBorder(),
          ),
        ),
        if (state.error != null) ...[
          const SizedBox(height: 12),
          Text(
            state.error!,
            style: TextStyle(color: Theme.of(context).colorScheme.error),
          ),
        ],
        const SizedBox(height: 16),
        FilledButton.icon(
          key: const Key('requestRideButton'),
          onPressed: state.submitting ? null : _submit,
          icon: const Icon(Icons.local_taxi),
          label: Text(state.submitting ? 'Requesting…' : 'Request ride'),
        ),
      ],
    );
  }

  Marker _marker(GeoPoint point, IconData icon, Color color) => Marker(
    point: LatLng(point.latitude, point.longitude),
    width: 48,
    height: 48,
    child: Icon(icon, color: color, size: 38),
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

class _ActiveRequest extends ConsumerWidget {
  const _ActiveRequest({required this.request});
  final RideRequest request;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final state = ref.watch(riderRequestControllerProvider).state;
    final status = request.trip?.status ?? request.status;
    return RefreshIndicator(
      onRefresh: ref.read(riderRequestControllerProvider).refreshActive,
      child: ListView(
        physics: const AlwaysScrollableScrollPhysics(),
        padding: const EdgeInsets.all(24),
        children: [
          const SizedBox(height: 60),
          const Icon(Icons.local_taxi, size: 72),
          const SizedBox(height: 24),
          Text(
            _statusTitle(status),
            textAlign: TextAlign.center,
            style: Theme.of(context).textTheme.headlineSmall,
          ),
          const SizedBox(height: 8),
          Text('Request ${request.id}', textAlign: TextAlign.center),
          const SizedBox(height: 24),
          Card(
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                children: [
                  _PointSummary(label: 'Pickup', point: request.pickup),
                  _PointSummary(
                    label: 'Destination',
                    point: request.destination,
                  ),
                  ListTile(
                    contentPadding: EdgeInsets.zero,
                    leading: const Icon(Icons.payments_outlined),
                    title: const Text('Your proposed fare'),
                    subtitle: Text(
                      '${request.proposedFare.currency} ${(request.proposedFare.amountMinor / 100).toStringAsFixed(2)}',
                    ),
                  ),
                ],
              ),
            ),
          ),
          const SizedBox(height: 12),
          OutlinedButton.icon(
            onPressed: state.loading
                ? null
                : ref.read(riderRequestControllerProvider).refreshActive,
            icon: const Icon(Icons.refresh),
            label: const Text('Refresh status'),
          ),
          if (status != 'completed' && status != 'cancelled') ...[
            const SizedBox(height: 8),
            TextButton(
              onPressed: state.submitting
                  ? null
                  : () => _confirmCancellation(context, ref),
              child: Text(state.submitting ? 'Cancelling…' : 'Cancel request'),
            ),
          ],
          if (state.error != null) ...[
            const SizedBox(height: 12),
            Text(
              state.error!,
              style: TextStyle(color: Theme.of(context).colorScheme.error),
            ),
          ],
        ],
      ),
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
