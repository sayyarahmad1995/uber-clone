import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../core/providers.dart';
import 'ride_flow_controller.dart';
import 'ride_flow_repository.dart';

String fareText(dynamic fare) {
  if (fare is! Map || fare['amount_minor'] is! num) return 'Fare unavailable';
  return '${fare['currency']} ${((fare['amount_minor'] as num) / 100).toStringAsFixed(2)}';
}
String pointText(dynamic point) => point is Map
  ? '${(point['latitude'] as num).toStringAsFixed(5)}, ${(point['longitude'] as num).toStringAsFixed(5)}' : 'Unavailable';
String statusText(String? status) => switch (status) {
  'assigned' => 'Driver assigned', 'in_progress' => 'Trip in progress',
  'completed' => 'Trip completed', 'cancelled' => 'Trip cancelled',
  _ => 'Waiting for Driver offers',
};

class OperationDetails extends StatelessWidget {
  const OperationDetails(this.operation, {super.key});
  final dynamic operation;
  @override
  Widget build(BuildContext context) {
    final value = operation;
    if (value is! Map) return const Text('Vehicle and service details unavailable for this historical trip.');
    return Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
      if (value['driver_name'] != null) Text('${value['driver_name']}'),
      Text('${value['make'] ?? ''} ${value['model'] ?? ''} ${value['model_year'] ?? ''} · ${value['color'] ?? ''}'),
      if (value['license_plate'] != null) Text('License plate: ${value['license_plate']}'),
      Text('Service: ${value['service_name'] ?? value['service_code']}'),
      Text('Agreed fare: ${fareText(value['fare'])}'),
    ]);
  }
}

class RideFlowPanel extends ConsumerStatefulWidget {
  const RideFlowPanel({super.key, this.rideId});
  final String? rideId;
  @override
  ConsumerState<RideFlowPanel> createState() => _RideFlowPanelState();
}

class _RideFlowPanelState extends ConsumerState<RideFlowPanel> with WidgetsBindingObserver {
  String get flowKey => widget.rideId ?? 'driver';
  bool _syncing = false;
  @override
  void initState() { super.initState(); WidgetsBinding.instance.addObserver(this); }
  @override
  void dispose() { WidgetsBinding.instance.removeObserver(this); super.dispose(); }
  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (state == AppLifecycleState.resumed) ref.read(rideFlowControllerProvider(flowKey)).setForeground(true);
    if ([AppLifecycleState.paused, AppLifecycleState.hidden, AppLifecycleState.detached].contains(state)) {
      ref.read(rideFlowControllerProvider(flowKey)).setForeground(false);
    }
  }
  @override
  Widget build(BuildContext context) {
    final flow = ref.watch(rideFlowControllerProvider(flowKey));
    ref.listen(rideFlowControllerProvider(flowKey), (_, next) {
      if (!next.loaded || next.busy) return;
      if (widget.rideId == null) {
        ref.read(driverControllerProvider).setActiveTrip(next.current != null);
      } else if (!_syncing) {
        final request = ref.read(riderRequestControllerProvider).state.active;
        if (request?.id == widget.rideId && (request?.trip?.status ?? request?.status) != next.status) {
          _syncing = true;
          ref.read(riderRequestControllerProvider).refreshActive().whenComplete(() => _syncing = false);
        }
      }
    });
    return Column(crossAxisAlignment: CrossAxisAlignment.stretch, children: [
      const Divider(),
      Text(widget.rideId == null ? 'Trips and requests' : statusText(flow.status), style: Theme.of(context).textTheme.titleLarge),
      if (!flow.loaded && flow.busy) const LinearProgressIndicator(),
      if (flow.error != null) Text(flow.error!, style: TextStyle(color: Theme.of(context).colorScheme.error)),
      TextButton.icon(onPressed: flow.busy ? null : flow.refresh,
        icon: const Icon(Icons.refresh), label: const Text('Refresh rides')),
      if (widget.rideId == null) ..._driver(flow) else ..._rider(flow),
    ]);
  }

  List<Widget> _driver(RideFlowController flow) {
    final trip = flow.current;
    if (trip != null) {
      final id = trip['ride_request_id'];
      return [
        Text(statusText(trip['status'] as String?)),
        OperationDetails(trip['operation_context']),
        Text('Pickup: ${pointText(trip['pickup'])}'),
        Text('Destination: ${pointText(trip['destination'])}'),
        if (trip['status'] == 'assigned') FilledButton(onPressed: flow.busy ? null : () =>
          _confirm(flow, '/v1/driver/ride-requests/$id/start', 'Start trip?', 'Confirm that the Rider is on board.', 'Start trip'), child: const Text('Start trip')),
        if (trip['status'] == 'in_progress') FilledButton(onPressed: flow.busy ? null : () =>
          _confirm(flow, '/v1/driver/ride-requests/$id/complete', 'Complete trip?', 'Confirm that you have reached the destination.', 'Complete trip'), child: const Text('Complete trip')),
        TextButton(onPressed: flow.busy ? null : () =>
          _confirm(flow, '/v1/driver/ride-requests/$id/cancel', 'Cancel trip?', 'This ends the trip for both you and the Rider.', 'Cancel trip'), child: const Text('Cancel trip')),
      ];
    }
    final online = ref.watch(driverControllerProvider).profile?.isOnline == true;
    return [
      if (!online) const Text('Go online to receive requests for your selected service.'),
      if (online && flow.loaded && flow.requests.isEmpty) const Text('No requests available. This list updates automatically.'),
      if (online) for (final request in flow.requests) Card(child: Padding(
        padding: const EdgeInsets.all(12), child: Column(crossAxisAlignment: CrossAxisAlignment.stretch, children: [
          Text('Rider fare: ${fareText(request['proposed_fare'])}'),
          Text('Pickup: ${pointText(request['pickup'])}'),
          Text('Destination: ${pointText(request['destination'])}'),
          if (request['pickup_distance_meters'] is num) Text('${((request['pickup_distance_meters'] as num)/1000).toStringAsFixed(1)} km to pickup · straight-line distance'),
          if (request['own_offer'] is Map) Text('Your response: ${fareText(request['own_offer']['fare'])} · ${request['own_offer']['status']}'),
          const Text('The Rider chooses a Driver. Sending a response does not assign this trip.'),
          FilledButton(onPressed: flow.busy ? null : () => flow.act('/v1/driver/ride-requests/${request['id']}/accept'), child: const Text('Accept Rider fare')),
          OutlinedButton(onPressed: flow.busy ? null : () => _counter(flow, request), child: const Text('Propose another fare')),
        ]))),
      ExpansionTile(title: const Text('Trip history'), children: [
        if (flow.history.isEmpty) const ListTile(title: Text('No completed or cancelled trips.')),
        for (final item in flow.history) ListTile(title: Text(statusText(item['status'] as String?)), subtitle: Column(
          crossAxisAlignment: CrossAxisAlignment.start, children: [
            Text('Assigned: ${item['assigned_at']}'), OperationDetails(item['operation_context']),
          ])),
      ]),
    ];
  }

  List<Widget> _rider(RideFlowController flow) {
    final trip = flow.current?['trip'];
    if (trip is Map) {
      return [
        OperationDetails(trip['operation_context']),
        if (['assigned','in_progress'].contains(trip['status'])) Text(
          freshDriverLocation(flow.location) == null ? 'Driver location is currently unavailable.' : 'Driver location updated: ${flow.location!['updated_at']}'),
      ];
    }
    return [
      if (flow.loaded && flow.offers.isEmpty) const Text('No offers yet. Driver responses will appear here.'),
      for (final offer in flow.offers) Card(child: Padding(padding: const EdgeInsets.all(12),
        child: Column(crossAxisAlignment: CrossAxisAlignment.stretch, children: [
          Text('${offer['driver']?['display_name'] ?? 'Driver'}', style: Theme.of(context).textTheme.titleMedium),
          if (offer['vehicle'] is Map) Text('${offer['vehicle']['make']} ${offer['vehicle']['model']} ${offer['vehicle']['model_year'] ?? ''} · ${offer['vehicle']['color']}'),
          Text('${fareText(offer['fare'])}${offer['matches_proposed_fare'] == true ? ' · Your fare' : ''}'),
          if (offer['pickup_distance_meters'] is num) Text('${((offer['pickup_distance_meters'] as num)/1000).toStringAsFixed(1)} km to pickup · straight-line distance'),
          if (offer['selectable'] != true) Text(offer['status'] == 'pending' ? 'Driver currently unavailable' : '${offer['status']}'),
          FilledButton(onPressed: flow.busy || offer['selectable'] != true ? null : () =>
            _confirm(flow, '/v1/ride-requests/${widget.rideId}/offers/${offer['driver_user_id']}/accept', 'Choose this Driver?', 'Agree to ${fareText(offer['fare'])} for this ride.', 'Choose Driver', data: {'updated_at': offer['updated_at']}), child: const Text('Choose Driver')),
          if (offer['status'] == 'pending') TextButton(onPressed: flow.busy ? null : () => flow.act(
            '/v1/ride-requests/${widget.rideId}/offers/${offer['driver_user_id']}/reject'), child: const Text('Reject offer')),
        ]))),
    ];
  }

  Future<void> _confirm(RideFlowController flow, String path, String title, String message, String action, {Json? data}) async {
    final confirmed = await showDialog<bool>(context: context, builder: (context) => AlertDialog(
      title: Text(title), content: Text(message), actions: [
        TextButton(onPressed: () => Navigator.pop(context, false), child: const Text('Back')),
        FilledButton(onPressed: () => Navigator.pop(context, true), child: Text(action)),
      ]));
    if (confirmed == true && mounted) await flow.act(path, data: data);
  }

  Future<void> _counter(RideFlowController flow, Json request) async {
    final field = TextEditingController(text: ((request['proposed_fare']['amount_minor'] as num)/100).toStringAsFixed(2));
    final proposed = request['proposed_fare']['amount_minor'] as int;
    String? error;
    final amount = await showDialog<int>(context: context, builder: (context) => StatefulBuilder(builder: (context, update) => AlertDialog(
      title: const Text('Propose another fare'), content: TextField(controller: field,
        keyboardType: const TextInputType.numberWithOptions(decimal: true),
        decoration: InputDecoration(labelText: '${request['proposed_fare']['currency']} · 90%–130% of Rider fare', errorText: error)),
      actions: [TextButton(onPressed: () => Navigator.pop(context), child: const Text('Back')),
        FilledButton(onPressed: () {
          final value = parseFareMinor(field.text);
          if (value == null || value < (proposed * 90 + 99) ~/ 100 || value > proposed * 130 ~/ 100) {
            update(() => error = 'Enter a fare within the allowed range.'); return;
          }
          Navigator.pop(context, value);
        }, child: const Text('Send offer'))])));
    // Dispose after the dialog transition has released its TextField.
    await Future<void>.delayed(const Duration(milliseconds: 300));
    field.dispose();
    if (amount != null && mounted) await flow.act('/v1/driver/ride-requests/${request['id']}/offer', data: {'amount_minor': amount}, put: true);
  }
}

int? parseFareMinor(String text) {
  final match = RegExp(r'^(\d+)(?:\.(\d{1,2}))?$').firstMatch(text.trim());
  if (match == null) return null;
  final whole = int.tryParse(match.group(1)!);
  if (whole == null || whole > 10000000000) return null;
  final value = whole * 100 + int.parse((match.group(2) ?? '').padRight(2, '0'));
  return value > 0 && value <= 1000000000000 ? value : null;
}

Json? freshDriverLocation(Json? value) {
  final time = DateTime.tryParse('${value?['updated_at']}');
  if (time == null) return null;
  final age = DateTime.now().toUtc().difference(time.toUtc());
  return age.isNegative || age > const Duration(minutes: 2) ? null : value;
}

class RiderServicePicker extends ConsumerWidget {
  const RiderServicePicker({super.key});
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final controller = ref.watch(riderRequestControllerProvider);
    return DropdownButtonFormField<String>(initialValue: controller.serviceCode,
      decoration: const InputDecoration(labelText: 'Ride service'),
      items: const [DropdownMenuItem(value: 'economy', child: Text('Economy')), DropdownMenuItem(value: 'comfort', child: Text('Comfort'))],
      onChanged: controller.state.submitting ? null : (value) { if (value != null) controller.selectService(value); });
  }
}

class RiderRideHistory extends ConsumerStatefulWidget {
  const RiderRideHistory({super.key});
  @override
  ConsumerState<RiderRideHistory> createState() => _RiderRideHistoryState();
}
class _RiderRideHistoryState extends ConsumerState<RiderRideHistory> {
  List<Json>? _rides;
  String? _error;
  bool _loading = false;
  Future<void> _load() async {
    if (_loading) {
      return;
    }
    setState(() { _loading = true; _error = null; });
    try {
      final data = await ref.read(rideFlowRepositoryProvider).get('/v1/ride-requests');
      if (mounted) {
        setState(() { _rides = (data['ride_requests'] as List? ?? []).cast<Json>()
          .where((ride) => ride['status'] == 'cancelled' || ['completed','cancelled'].contains(ride['trip']?['status'])).toList(); });
      }
    } catch (e) {
      if (mounted) {
        setState(() => _error = '$e');
      }
    } finally {
      if (mounted) {
        setState(() => _loading = false);
      }
    }
  }
  @override
  Widget build(BuildContext context) => ExpansionTile(title: const Text('Recent rides'),
    onExpansionChanged: (open) {
      if (open) {
        _load();
      }
    }, children: [
      if (_loading) const LinearProgressIndicator(),
      if (_error != null) Text(_error!),
      if (_rides?.isEmpty == true) const Text('No completed or cancelled rides.'),
      for (final ride in _rides ?? <Json>[]) ListTile(title: Text(statusText((ride['trip']?['status'] ?? ride['status']) as String?)),
        subtitle: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
          Text('${ride['created_at']}'),
          if (ride['trip'] != null) OperationDetails(ride['trip']['operation_context']),
        ])),
    ]);
}
