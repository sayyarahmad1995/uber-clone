import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/providers.dart';
import 'domain/marketplace_request.dart';
import 'domain/ride_snapshot.dart';
import 'domain/ride_execution.dart';
import 'domain/trip.dart';
import 'application/driver_marketplace_controller.dart';
import 'application/driver_trip_controller.dart';
import 'application/rider_active_ride_controller.dart';

String fareText(dynamic fare) {
  if (fare is RideFare) {
    return '${fare.currency} ${(fare.amountMinor / 100).toStringAsFixed(2)}';
  }
  if (fare is TripOperationFare) {
    return '${fare.currency} ${(fare.amountMinor / 100).toStringAsFixed(2)}';
  }
  return 'Fare unavailable';
}

String pointText(RideLocation? point) => point == null
    ? 'Unavailable'
    : '${point.latitude.toStringAsFixed(5)}, ${point.longitude.toStringAsFixed(5)}';

String riderOfferDistanceText(int? distance) {
  final value = distance;
  return value == null
      ? 'Pickup distance unavailable'
      : '${(value / 1000).toStringAsFixed(1)} km to pickup · straight-line distance';
}

String statusText(String? status) => switch (status) {
  'assigned' => 'Driver assigned',
  'in_progress' => 'Trip in progress',
  'completed' => 'Trip completed',
  'cancelled' => 'Trip cancelled',
  _ => 'Waiting for Driver offers',
};

String settlementText(SettlementSnapshot? settlement) {
  if (settlement == null) return 'Settlement: Not recorded';
  if (settlement.status == 'unsettled') return 'Settlement: Cash due';
  if (settlement.status == 'cash_collected') {
    final collectedAt = settlement.cashCollectedAt;
    return collectedAt == null
        ? 'Settlement: Cash collected'
        : 'Settlement: Cash collected · $collectedAt';
  }
  return 'Settlement: ${settlement.status}';
}

bool settlementIsCashDue(SettlementSnapshot settlement) =>
    settlement.status == 'unsettled';

String tripAgreedFareText(TripSnapshot trip) {
  final fare = trip.operationContext?.fare;
  return fare == null ? 'Fare unavailable' : fareText(fare);
}

String cashCollectionMessage(TripSnapshot trip) {
  final agreedFare = tripAgreedFareText(trip);
  return agreedFare == 'Fare unavailable'
      ? 'Confirm that you collected the agreed cash fare from the Rider.'
      : 'Confirm that you collected $agreedFare in cash from the Rider.';
}

class OperationDetails extends StatelessWidget {
  const OperationDetails(this.operation, {super.key});
  final TripOperationContext? operation;

  @override
  Widget build(BuildContext context) {
    final value = operation;
    if (value == null) {
      return const Text(
        'Vehicle and service details unavailable for this historical trip.',
      );
    }
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(value.driverName),
        Text(
          '${value.make} ${value.model} ${value.modelYear} · ${value.color}',
        ),
        Text('License plate: ${value.licensePlate}'),
        Text('Service: ${value.serviceName}'),
        Text('Agreed fare: ${fareText(value.fare)}'),
      ],
    );
  }
}

class RideFlowPanel extends ConsumerStatefulWidget {
  const RideFlowPanel.rider({required String this.rideId, super.key})
    : driver = false;

  const RideFlowPanel.driver({super.key}) : rideId = null, driver = true;

  final String? rideId;
  final bool driver;

  @override
  ConsumerState<RideFlowPanel> createState() => _RideFlowPanelState();
}

class _RideFlowPanelState extends ConsumerState<RideFlowPanel>
    with WidgetsBindingObserver {
  bool _syncing = false;

  bool get _driver => widget.driver;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    super.dispose();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    final foreground = state == AppLifecycleState.resumed;
    final background = [
      AppLifecycleState.paused,
      AppLifecycleState.hidden,
      AppLifecycleState.detached,
    ].contains(state);

    if (!foreground && !background) return;

    if (_driver) {
      ref.read(driverMarketplaceControllerProvider).setForeground(foreground);
      ref.read(driverTripControllerProvider).setForeground(foreground);
      return;
    }

    ref
        .read(riderActiveRideControllerProvider(widget.rideId!))
        .setForeground(foreground);
  }

  @override
  Widget build(BuildContext context) {
    return _driver ? _buildDriver(context) : _buildRider(context);
  }

  Widget _buildDriver(BuildContext context) {
    final marketplace = ref.watch(driverMarketplaceControllerProvider);
    final trips = ref.watch(driverTripControllerProvider);

    ref.listen(driverTripControllerProvider, (_, next) {
      if (!next.loaded || next.busy) return;
      ref.read(driverControllerProvider).setActiveTrip(next.trip != null);
    });

    final busy = marketplace.busy || trips.busy;
    final loaded = marketplace.loaded && trips.loaded;
    final error = trips.error ?? marketplace.error;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        const Divider(),
        Text(
          'Trips and requests',
          style: Theme.of(context).textTheme.titleLarge,
        ),
        if (!loaded && busy) const LinearProgressIndicator(),
        if (error != null)
          Text(
            error,
            style: TextStyle(color: Theme.of(context).colorScheme.error),
          ),
        TextButton.icon(
          onPressed: busy
              ? null
              : () async {
                  await Future.wait([marketplace.refresh(), trips.refresh()]);
                },
          icon: const Icon(Icons.refresh),
          label: const Text('Refresh rides'),
        ),
        ..._driverContent(marketplace, trips),
      ],
    );
  }

  Widget _buildRider(BuildContext context) {
    final rideId = widget.rideId!;
    final flow = ref.watch(riderActiveRideControllerProvider(rideId));

    ref.listen(riderActiveRideControllerProvider(rideId), (_, next) {
      if (!next.loaded || next.busy || _syncing) return;

      final request = ref.read(riderRequestControllerProvider).state.active;
      if (request?.id == rideId &&
          (request?.trip?.status ?? request?.status) != next.status) {
        _syncing = true;
        ref
            .read(riderRequestControllerProvider)
            .refreshActive()
            .whenComplete(() => _syncing = false);
      }
    });

    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        const Divider(),
        Text(
          statusText(flow.status),
          style: Theme.of(context).textTheme.titleLarge,
        ),
        if (!flow.loaded && flow.busy) const LinearProgressIndicator(),
        if (flow.error != null)
          Text(
            flow.error!,
            style: TextStyle(color: Theme.of(context).colorScheme.error),
          ),
        TextButton.icon(
          onPressed: flow.busy ? null : flow.refresh,
          icon: const Icon(Icons.refresh),
          label: const Text('Refresh rides'),
        ),
        ..._riderContent(flow),
      ],
    );
  }

  List<Widget> _driverContent(
    DriverMarketplaceController marketplace,
    DriverTripController trips,
  ) {
    final trip = trips.trip;
    if (trip != null) {
      final id = trip.rideRequestId!;
      final cashDue =
          trip.status == 'completed' && settlementIsCashDue(trip.settlement);

      return [
        Text(statusText(trip.status)),
        OperationDetails(trip.operationContext),
        Text(settlementText(trip.settlement)),
        Text('Pickup: ${pointText(trip.pickup)}'),
        Text('Destination: ${pointText(trip.destination)}'),
        if (trip.status == 'assigned')
          FilledButton(
            onPressed: trips.busy
                ? null
                : () => _confirm(
                    'Start trip?',
                    'Confirm that the Rider is on board.',
                    'Start trip',
                    () => trips.startTrip(id),
                  ),
            child: const Text('Start trip'),
          ),
        if (trip.status == 'in_progress')
          FilledButton(
            onPressed: trips.busy
                ? null
                : () => _confirm(
                    'Complete trip?',
                    'Confirm that you have reached the destination.',
                    'Complete trip',
                    () => trips.completeTrip(id),
                  ),
            child: const Text('Complete trip'),
          ),
        if (cashDue) const Text('Collect the agreed cash fare from the Rider.'),
        if (cashDue)
          FilledButton(
            onPressed: trips.busy
                ? null
                : () => _confirm(
                    'Confirm cash collected?',
                    cashCollectionMessage(trip),
                    'Confirm cash collected',
                    () => trips.confirmCashCollected(id),
                  ),
            child: const Text('Confirm cash collected'),
          ),
        if (['assigned', 'in_progress'].contains(trip.status))
          TextButton(
            onPressed: trips.busy
                ? null
                : () => _confirm(
                    'Cancel trip?',
                    'This ends the trip for both you and the Rider.',
                    'Cancel trip',
                    () => trips.cancelTrip(id),
                  ),
            child: const Text('Cancel trip'),
          ),
        ..._tripHistory(trips),
      ];
    }

    final online =
        ref.watch(driverControllerProvider).profile?.isOnline == true;

    return [
      if (!online)
        const Text('Go online to receive requests for your selected service.'),
      if (online && marketplace.loaded && marketplace.requests.isEmpty)
        const Text('No requests available. This list updates automatically.'),
      if (online)
        for (final request in marketplace.requests)
          Card(
            child: Padding(
              padding: const EdgeInsets.all(12),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  Text('Rider fare: ${fareText(request.proposedFare)}'),
                  Text('Pickup: ${pointText(request.pickup)}'),
                  Text('Destination: ${pointText(request.destination)}'),
                  Text(
                    '${(request.pickupDistanceMeters / 1000).toStringAsFixed(1)} km to pickup · straight-line distance',
                  ),
                  if (request.ownOffer != null)
                    Text(
                      'Your response: ${fareText(request.ownOffer!.fare)} · ${request.ownOffer!.status}',
                    ),
                  const Text(
                    'The Rider chooses a Driver. Sending a response does not assign this trip.',
                  ),
                  FilledButton(
                    onPressed: marketplace.busy
                        ? null
                        : () => marketplace.acceptProposedFare(request.id),
                    child: const Text('Accept Rider fare'),
                  ),
                  OutlinedButton(
                    onPressed: marketplace.busy
                        ? null
                        : () => _counter(marketplace, request),
                    child: const Text('Propose another fare'),
                  ),
                  if (request.ownOffer == null)
                    OutlinedButton(
                      onPressed: marketplace.busy
                          ? null
                          : () => _confirm(
                              'Decline request?',
                              'This removes the request only from your marketplace. The Rider can still receive responses from other Drivers.',
                              'Decline',
                              () => marketplace.declineRideRequest(request.id),
                            ),
                      child: const Text('Decline'),
                    ),
                ],
              ),
            ),
          ),
      ..._tripHistory(trips),
    ];
  }

  List<Widget> _tripHistory(DriverTripController trips) {
    return [
      ExpansionTile(
        title: const Text('Trip history'),
        children: [
          if (trips.history.isEmpty)
            const ListTile(title: Text('No completed or cancelled trips.')),
          for (final item in trips.history)
            ListTile(
              title: Text(statusText(item.status)),
              subtitle: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text('Assigned: ${item.assignedAt}'),
                  OperationDetails(item.operationContext),
                  Text(settlementText(item.settlement)),
                ],
              ),
            ),
        ],
      ),
    ];
  }

  List<Widget> _riderContent(RiderActiveRideController flow) {
    final trip = flow.riderRide?.trip;
    if (trip != null) {
      return [
        OperationDetails(trip.operationContext),
        Text(settlementText(trip.settlement)),
        if (['assigned', 'in_progress'].contains(trip.status))
          Text(
            freshDriverLocation(flow.location) == null
                ? 'Driver location is currently unavailable.'
                : 'Driver location updated: ${flow.location!.updatedAt}',
          ),
      ];
    }

    return [
      if (flow.loaded && flow.offers.isEmpty)
        const Text('No offers yet. Driver responses will appear here.'),
      for (final offer in flow.offers)
        Card(
          child: Padding(
            padding: const EdgeInsets.all(12),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                Text(
                  offer.driver?.displayName ?? 'Driver',
                  style: Theme.of(context).textTheme.titleMedium,
                ),
                if (offer.vehicle != null)
                  Text(
                    '${offer.vehicle!.make} ${offer.vehicle!.model} ${offer.vehicle!.modelYear ?? ''} · ${offer.vehicle!.color}',
                  ),
                Text(
                  '${fareText(offer.fare)}${offer.matchesProposedFare ? ' · Your fare' : ''}',
                ),
                Text(riderOfferDistanceText(offer.pickupDistanceMeters)),
                if (!offer.selectable)
                  Text(
                    offer.status == 'pending'
                        ? 'Driver currently unavailable'
                        : offer.status,
                  ),
                FilledButton(
                  onPressed: flow.busy || !offer.selectable
                      ? null
                      : () => _confirm(
                          'Choose this Driver?',
                          'Agree to ${fareText(offer.fare)} for this ride.',
                          'Choose Driver',
                          () => flow.selectOffer(
                            offer.driverUserId,
                            offer.updatedAt,
                          ),
                        ),
                  child: const Text('Choose Driver'),
                ),
                if (offer.status == 'pending')
                  OutlinedButton(
                    onPressed: flow.busy
                        ? null
                        : () => _confirm(
                            'Decline offer?',
                            'Only this Driver offer will be declined. Your ride request and other Driver offers remain open.',
                            'Decline',
                            () => flow.declineOffer(offer.driverUserId),
                          ),
                    child: const Text('Decline'),
                  ),
              ],
            ),
          ),
        ),
    ];
  }

  Future<void> _confirm(
    String title,
    String message,
    String action,
    Future<void> Function() command,
  ) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(title),
        content: Text(message),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Back'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            child: Text(action),
          ),
        ],
      ),
    );

    if (confirmed == true && mounted) await command();
  }

  Future<void> _counter(
    DriverMarketplaceController marketplace,
    MarketplaceRequest request,
  ) async {
    final field = TextEditingController(
      text: (request.proposedFare.amountMinor / 100).toStringAsFixed(2),
    );
    final proposed = request.proposedFare.amountMinor;
    String? error;

    final amount = await showDialog<int>(
      context: context,
      builder: (context) => StatefulBuilder(
        builder: (context, update) => AlertDialog(
          title: const Text('Propose another fare'),
          content: TextField(
            controller: field,
            keyboardType: const TextInputType.numberWithOptions(decimal: true),
            decoration: InputDecoration(
              labelText:
                  '${request.proposedFare.currency} · 90%–130% of Rider fare',
              errorText: error,
            ),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context),
              child: const Text('Back'),
            ),
            FilledButton(
              onPressed: () {
                final value = parseFareMinor(field.text);
                if (value == null ||
                    value < (proposed * 90 + 99) ~/ 100 ||
                    value > proposed * 130 ~/ 100) {
                  update(
                    () => error = 'Enter a fare within the allowed range.',
                  );
                  return;
                }
                Navigator.pop(context, value);
              },
              child: const Text('Send offer'),
            ),
          ],
        ),
      ),
    );

    await Future<void>.delayed(const Duration(milliseconds: 300));
    field.dispose();

    if (amount != null && mounted) {
      await marketplace.submitOffer(request.id, amount);
    }
  }
}

int? parseFareMinor(String text) {
  final match = RegExp(r'^(\d+)(?:\.(\d{1,2}))?$').firstMatch(text.trim());
  if (match == null) return null;
  final whole = int.tryParse(match.group(1)!);
  if (whole == null || whole > 10000000000) return null;
  final value =
      whole * 100 + int.parse((match.group(2) ?? '').padRight(2, '0'));
  return value > 0 && value <= 1000000000000 ? value : null;
}

DriverLocationSnapshot? freshDriverLocation(DriverLocationSnapshot? value) {
  if (value == null) return null;
  final age = DateTime.now().toUtc().difference(value.updatedAt.toUtc());
  return age.isNegative || age > const Duration(minutes: 2) ? null : value;
}

class RiderServicePicker extends ConsumerWidget {
  const RiderServicePicker({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final controller = ref.watch(riderRequestControllerProvider);
    return DropdownButtonFormField<String>(
      initialValue: controller.serviceCode,
      decoration: const InputDecoration(labelText: 'Ride service'),
      items: const [
        DropdownMenuItem(value: 'economy', child: Text('Economy')),
        DropdownMenuItem(value: 'comfort', child: Text('Comfort')),
      ],
      onChanged: controller.state.submitting
          ? null
          : (value) {
              if (value != null) controller.selectService(value);
            },
    );
  }
}

class RiderRideHistory extends ConsumerStatefulWidget {
  const RiderRideHistory({super.key});

  @override
  ConsumerState<RiderRideHistory> createState() => _RiderRideHistoryState();
}

class _RiderRideHistoryState extends ConsumerState<RiderRideHistory> {
  List<RiderRideSnapshot>? _rides;
  String? _error;
  bool _loading = false;

  Future<void> _load() async {
    if (_loading) return;
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final data = await ref.read(rideFlowRepositoryProvider).listRiderRides();
      if (mounted) {
        setState(() {
          _rides = data
              .where(
                (ride) =>
                    ride.status == 'cancelled' ||
                    ['completed', 'cancelled'].contains(ride.trip?.status),
              )
              .toList();
        });
      }
    } catch (e) {
      if (mounted) setState(() => _error = '$e');
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) => ExpansionTile(
    title: const Text('Recent rides'),
    onExpansionChanged: (open) {
      if (open) _load();
    },
    children: [
      if (_loading) const LinearProgressIndicator(),
      if (_error != null) Text(_error!),
      if (_rides?.isEmpty == true)
        const Text('No completed or cancelled rides.'),
      for (final ride in _rides ?? <RiderRideSnapshot>[])
        ListTile(
          title: Text(statusText(ride.trip?.status ?? ride.status)),
          subtitle: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text('${ride.createdAt}'),
              if (ride.trip != null) ...[
                OperationDetails(ride.trip!.operationContext),
                Text(settlementText(ride.trip!.settlement)),
              ],
            ],
          ),
        ),
    ],
  );
}
