import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:latlong2/latlong.dart';

import '../../../core/dashboard/ride_dashboard_scaffold.dart';
import '../../../core/maps/ride_map.dart';
import '../../../core/providers.dart';
import '../../../core/theme/app_theme.dart';
import '../domain/driver_profile.dart';

class DriverWorkspaceScreen extends ConsumerStatefulWidget {
  const DriverWorkspaceScreen({super.key, required this.accountID});
  final String accountID;
  @override
  ConsumerState<DriverWorkspaceScreen> createState() =>
      _DriverWorkspaceScreenState();
}

class _DriverWorkspaceScreenState extends ConsumerState<DriverWorkspaceScreen> {
  bool _editing = false;
  @override
  Widget build(BuildContext context) {
    final driver = ref.watch(driverControllerProvider);
    final profile = driver.profile;
    final location = driver.location;
    final setup = driver.loaded && (profile == null || _editing);
    return RideDashboardScaffold(
      panelIdentity: !driver.loaded
          ? 'driver-loading'
          : setup
          ? 'driver-setup'
          : 'driver-readiness',
      map: RideMap(
        tiles: ref.watch(mapTilesProvider),
        markers: location == null
            ? const []
            : [
                RideMapMarker(
                  point: LatLng(location.latitude, location.longitude),
                  icon: Icons.local_taxi,
                  color: AppColors.success,
                  label: 'Your published location',
                ),
              ],
      ),
      floatingStatus: DashboardStatusCard(
        icon: Icons.local_taxi,
        title: 'Driver dashboard',
        message: profile == null
            ? 'Set up your Driver and vehicle details.'
            : profile.isOnline
            ? 'You are online. Keep your location updated.'
            : 'You are offline.',
      ),
      panelBuilder: (context, scrollController, scrollEnabled) {
        final physics = scrollEnabled
            ? const ClampingScrollPhysics()
            : const NeverScrollableScrollPhysics();
        if (!driver.loaded) {
          return ListView(
            controller: scrollController,
            physics: physics,
            padding: const EdgeInsets.all(AppSpacing.md),
            children: [
              Text(
                driver.busy
                    ? 'Loading Driver profile…'
                    : 'Unable to load Driver profile',
              ),
              if (driver.error != null) Text(driver.error!),
              DashboardPanelControl(
                child: TextButton(
                  onPressed: driver.busy ? null : driver.load,
                  child: const Text('Retry'),
                ),
              ),
            ],
          );
        }
        if (setup) {
          return _DriverSetupForm(
            key: ValueKey('setup-${widget.accountID}'),
            profile: profile,
            busy: driver.busy,
            error: driver.error,
            scrollController: scrollController,
            physics: physics,
            onCancel: profile == null
                ? null
                : () => setState(() => _editing = false),
            onSave: (name, vehicle) async {
              await driver.onboard(name, vehicle);
              if (mounted && driver.error == null) {
                setState(() => _editing = false);
              }
            },
          );
        }
        return ListView(
          controller: scrollController,
          physics: physics,
          padding: const EdgeInsets.all(AppSpacing.md),
          children: [
            Text(
              profile!.isOnline ? 'You are online' : 'Ready to go online',
              style: Theme.of(context).textTheme.headlineSmall,
            ),
            const SizedBox(height: AppSpacing.sm),
            Text(
              profile.displayName?.isNotEmpty == true
                  ? profile.displayName!
                  : 'Driver profile',
            ),
            Text(
              '${profile.vehicle.make} ${profile.vehicle.model} ${profile.vehicle.modelYear ?? ''} • ${profile.vehicle.color}',
            ),
            Text(profile.vehicle.licensePlate),
            const SizedBox(height: AppSpacing.md),
            DashboardPanelControl(
              child: FilledButton.icon(
                onPressed: driver.busy
                    ? null
                    : () => driver.setOnline(!profile.isOnline),
                icon: const Icon(Icons.power_settings_new),
                label: Text(profile.isOnline ? 'Go offline' : 'Go online'),
              ),
            ),
            const Text(
              'Going online publishes your current location. Update it within two minutes for marketplace eligibility.',
            ),
            const SizedBox(height: AppSpacing.sm),
            Text(
              location == null
                  ? 'No location published during this visit.'
                  : 'Location published at ${location.updatedAt.toLocal()}',
            ),
            DashboardPanelControl(
              child: OutlinedButton.icon(
                onPressed: driver.busy ? null : driver.publishLocation,
                icon: const Icon(Icons.my_location),
                label: const Text('Update current location'),
              ),
            ),
            const Text(
              'Location updates are manual. Request discovery and offers are coming in the next slice.',
            ),
            if (driver.busy) const LinearProgressIndicator(),
            if (driver.error != null)
              Text(
                driver.error!,
                style: TextStyle(color: Theme.of(context).colorScheme.error),
              ),
            DashboardPanelControl(
              child: TextButton(
                onPressed: driver.busy ? null : driver.load,
                child: const Text('Refresh status'),
              ),
            ),
            DashboardPanelControl(
              child: TextButton(
                onPressed: driver.busy
                    ? null
                    : () => setState(() => _editing = true),
                child: const Text('Edit Driver details'),
              ),
            ),
          ],
        );
      },
    );
  }
}

class _DriverSetupForm extends StatefulWidget {
  const _DriverSetupForm({
    super.key,
    required this.profile,
    required this.busy,
    required this.error,
    required this.scrollController,
    required this.physics,
    required this.onSave,
    this.onCancel,
  });
  final DriverProfile? profile;
  final bool busy;
  final String? error;
  final ScrollController scrollController;
  final ScrollPhysics physics;
  final Future<void> Function(String, DriverVehicle) onSave;
  final VoidCallback? onCancel;
  @override
  State<_DriverSetupForm> createState() => _DriverSetupFormState();
}

class _DriverSetupFormState extends State<_DriverSetupForm> {
  final _form = GlobalKey<FormState>();
  late final List<TextEditingController> _fields;
  static const _labels = [
    'Display name',
    'Vehicle make',
    'Vehicle model',
    'Model year',
    'Vehicle color',
    'License plate',
  ];
  @override
  void initState() {
    super.initState();
    final p = widget.profile;
    _fields = [
      p?.displayName,
      p?.vehicle.make,
      p?.vehicle.model,
      p?.vehicle.modelYear?.toString(),
      p?.vehicle.color,
      p?.vehicle.licensePlate,
    ].map((value) => TextEditingController(text: value ?? '')).toList();
  }

  @override
  void dispose() {
    for (final field in _fields) {
      field.dispose();
    }
    super.dispose();
  }

  @override
  Widget build(BuildContext context) => Form(
    key: _form,
    child: ListView(
      controller: widget.scrollController,
      physics: widget.physics,
      padding: const EdgeInsets.all(AppSpacing.md),
      children: [
        Text('Driver setup', style: Theme.of(context).textTheme.headlineSmall),
        const Text('Add your Driver and vehicle details.'),
        for (var index = 0; index < _labels.length; index++)
          Padding(
            padding: const EdgeInsets.only(top: AppSpacing.sm),
            child: DashboardPanelControl(
              child: TextFormField(
                key: ValueKey('driver-field-$index'),
                controller: _fields[index],
                enabled: !widget.busy,
                decoration: InputDecoration(labelText: _labels[index]),
                keyboardType: index == 3
                    ? TextInputType.number
                    : TextInputType.text,
                validator: (value) {
                  if (value == null || value.trim().isEmpty) return 'Required';
                  if (index == 3) {
                    final year = int.tryParse(value.trim());
                    if (year == null ||
                        year < 1886 ||
                      year > DateTime.now().toUtc().year + 1) {
                      return 'Enter a valid model year';
                    }
                  }
                  return null;
                },
              ),
            ),
          ),
        if (widget.error != null)
          Text(
            widget.error!,
            style: TextStyle(color: Theme.of(context).colorScheme.error),
          ),
        DashboardPanelControl(
          child: FilledButton(
            onPressed: widget.busy
                ? null
                : () {
                    if (!_form.currentState!.validate()) return;
                    final values = _fields
                        .map((field) => field.text.trim())
                        .toList();
                    widget.onSave(
                      values[0],
                      DriverVehicle(
                        make: values[1],
                        model: values[2],
                        modelYear: int.parse(values[3]),
                        color: values[4],
                        licensePlate: values[5],
                      ),
                    );
                  },
            child: Text(widget.busy ? 'Saving…' : 'Save Driver details'),
          ),
        ),
        if (widget.onCancel != null)
          DashboardPanelControl(
            child: TextButton(
              onPressed: widget.busy ? null : widget.onCancel,
              child: const Text('Cancel editing'),
            ),
          ),
      ],
    ),
  );
}
