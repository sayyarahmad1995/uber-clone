import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:latlong2/latlong.dart';

import '../../../core/dashboard/ride_dashboard_scaffold.dart';
import '../../../core/maps/ride_map.dart';
import '../../../core/providers.dart';
import '../../../core/theme/app_theme.dart';
import '../domain/driver_onboarding.dart';
import '../domain/driver_profile.dart';

class DriverWorkspaceScreen extends ConsumerStatefulWidget {
  const DriverWorkspaceScreen({super.key, required this.accountID});
  final String accountID;
  @override
  ConsumerState<DriverWorkspaceScreen> createState() =>
      _DriverWorkspaceScreenState();
}

class _DriverWorkspaceScreenState extends ConsumerState<DriverWorkspaceScreen> {
  bool _reapplying = false;

  @override
  Widget build(BuildContext context) {
    final driver = ref.watch(driverControllerProvider);
    final profile = driver.profile;
    final location = driver.location;
    final onboarding = driver.loaded && profile == null
        ? ref.watch(driverOnboardingControllerProvider)
        : null;
    final application = onboarding?.application;
    final loading = !driver.loaded ||
        (profile == null && onboarding != null && !onboarding.loaded);
    final setup = !loading &&
        profile == null &&
        (application == null || _reapplying);

    return RideDashboardScaffold(
      panelIdentity: loading
          ? 'driver-loading'
          : profile != null
          ? 'driver-readiness'
          : setup
          ? 'driver-onboarding-form'
          : 'driver-onboarding-${application!.status}',
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
        message: _statusMessage(
          loading: loading,
          profile: profile,
          application: application,
        ),
      ),
      panelBuilder: (context, scrollController, scrollEnabled) {
        final physics = scrollEnabled
            ? const ClampingScrollPhysics()
            : const NeverScrollableScrollPhysics();

        if (loading) {
          return _LoadingPanel(
            scrollController: scrollController,
            physics: physics,
            busy: driver.busy || (onboarding?.busy ?? false),
            error: driver.error ?? onboarding?.error,
            onRetry: () {
              driver.load();
              onboarding?.load();
            },
          );
        }

        if (profile != null) {
          return _DriverReadinessPanel(
            profile: profile,
            location: location,
            busy: driver.busy,
            error: driver.error,
            scrollController: scrollController,
            physics: physics,
            onAvailabilityChanged: driver.setOnline,
            onPublishLocation: driver.publishLocation,
            onRefresh: driver.load,
          );
        }

        if (!setup && application != null && onboarding != null) {
          return _DriverApplicationStatusPanel(
            application: application,
            busy: onboarding.busy,
            error: onboarding.error,
            scrollController: scrollController,
            physics: physics,
            onRefresh: onboarding.load,
            onReapply: application.isRejected
                ? () => setState(() => _reapplying = true)
                : null,
          );
        }

        return _DriverSetupForm(
          key: ValueKey('setup-${widget.accountID}'),
          services: onboarding!.services,
          busy: onboarding.busy,
          error: onboarding.error,
          scrollController: scrollController,
          physics: physics,
          onReview: (name, service, vehicle) => _reviewAndSubmit(
            onboarding: onboarding,
            name: name,
            service: service,
            vehicle: vehicle,
          ),
          onCancel: _reapplying
              ? () => setState(() => _reapplying = false)
              : null,
        );
      },
    );
  }

  String _statusMessage({
    required bool loading,
    required DriverProfile? profile,
    required DriverOnboardingApplication? application,
  }) {
    if (loading) return 'Loading Driver status.';
    if (profile != null) {
      return profile.isOnline
          ? 'You are online. Keep your location updated.'
          : 'You are offline.';
    }
    if (_reapplying || application == null) {
      return 'Choose a service and submit your vehicle for review.';
    }
    if (application.isPending) return 'Your Driver application is under review.';
    if (application.isRejected) return 'Your Driver application was rejected.';
    return 'Your Driver application is approved.';
  }

  Future<void> _reviewAndSubmit({
    required dynamic onboarding,
    required String name,
    required DriverServiceOption service,
    required DriverVehicle vehicle,
  }) async {
    final precheck = await onboarding.precheck(
      displayName: name,
      serviceCode: service.code,
      vehicle: vehicle,
    );
    if (!mounted || precheck == null) return;
    if (!precheck.eligible) {
      await showDialog<void>(
        context: context,
        builder: (context) => AlertDialog(
          title: const Text('Vehicle not eligible'),
          content: Text(
            precheck.reasons.isEmpty
                ? 'This vehicle does not meet the selected service requirements.'
                : precheck.reasons.join('\n'),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context),
              child: const Text('Change application'),
            ),
          ],
        ),
      );
      return;
    }

    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Review application'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('Driver: ${name.trim()}'),
            Text('Service: ${service.displayName}'),
            Text(
              'Vehicle: ${vehicle.make} ${vehicle.model} ${vehicle.modelYear ?? ''}',
            ),
            Text('Color: ${vehicle.color}'),
            Text('License plate: ${vehicle.licensePlate}'),
            const SizedBox(height: AppSpacing.sm),
            const Text(
              'Submitting sends these details for review. They do not become approved information immediately.',
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Back'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            child: const Text('Submit for review'),
          ),
        ],
      ),
    );
    if (confirmed != true || !mounted) return;

    await onboarding.submit(
      displayName: name,
      serviceCode: service.code,
      vehicle: vehicle,
    );
    if (mounted && onboarding.error == null) {
      setState(() => _reapplying = false);
    }
  }
}

class _LoadingPanel extends StatelessWidget {
  const _LoadingPanel({
    required this.scrollController,
    required this.physics,
    required this.busy,
    required this.error,
    required this.onRetry,
  });

  final ScrollController scrollController;
  final ScrollPhysics physics;
  final bool busy;
  final String? error;
  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) => ListView(
    controller: scrollController,
    physics: physics,
    padding: const EdgeInsets.all(AppSpacing.md),
    children: [
      Text(busy ? 'Loading Driver status…' : 'Unable to load Driver status'),
      if (error != null) Text(error!),
      DashboardPanelControl(
        child: TextButton(
          onPressed: busy ? null : onRetry,
          child: const Text('Retry'),
        ),
      ),
    ],
  );
}

class _DriverApplicationStatusPanel extends StatelessWidget {
  const _DriverApplicationStatusPanel({
    required this.application,
    required this.busy,
    required this.error,
    required this.scrollController,
    required this.physics,
    required this.onRefresh,
    this.onReapply,
  });

  final DriverOnboardingApplication application;
  final bool busy;
  final String? error;
  final ScrollController scrollController;
  final ScrollPhysics physics;
  final VoidCallback onRefresh;
  final VoidCallback? onReapply;

  @override
  Widget build(BuildContext context) {
    final title = application.isPending
        ? 'Application under review'
        : application.isRejected
        ? 'Application rejected'
        : 'Application approved';
    return ListView(
      controller: scrollController,
      physics: physics,
      padding: const EdgeInsets.all(AppSpacing.md),
      children: [
        Text(title, style: Theme.of(context).textTheme.headlineSmall),
        const SizedBox(height: AppSpacing.sm),
        Text('Service: ${application.service.displayName}'),
        Text(
          '${application.vehicle.make} ${application.vehicle.model} ${application.vehicle.modelYear ?? ''} • ${application.vehicle.color}',
        ),
        Text(application.vehicle.licensePlate),
        Text('Submitted ${application.submittedAt.toLocal()}'),
        if (application.isPending) ...[
          const SizedBox(height: AppSpacing.sm),
          const Text(
            'Your submitted details remain pending until the review decision is recorded.',
          ),
        ],
        if (application.isRejected) ...[
          const SizedBox(height: AppSpacing.sm),
          Text(
            application.rejectionReason ??
                'The selected vehicle/service application was not approved.',
            style: TextStyle(color: Theme.of(context).colorScheme.error),
          ),
        ],
        if (application.isApproved) ...[
          const SizedBox(height: AppSpacing.sm),
          const Text(
            'Approval is recorded. Operational Driver activation is handled by the approved vehicle/service foundation.',
          ),
        ],
        if (error != null)
          Text(
            error!,
            style: TextStyle(color: Theme.of(context).colorScheme.error),
          ),
        DashboardPanelControl(
          child: TextButton(
            onPressed: busy ? null : onRefresh,
            child: const Text('Refresh application'),
          ),
        ),
        if (onReapply != null)
          DashboardPanelControl(
            child: FilledButton(
              onPressed: busy ? null : onReapply,
              child: const Text('Apply again'),
            ),
          ),
      ],
    );
  }
}

class _DriverReadinessPanel extends StatelessWidget {
  const _DriverReadinessPanel({
    required this.profile,
    required this.location,
    required this.busy,
    required this.error,
    required this.scrollController,
    required this.physics,
    required this.onAvailabilityChanged,
    required this.onPublishLocation,
    required this.onRefresh,
  });

  final DriverProfile profile;
  final PublishedDriverLocation? location;
  final bool busy;
  final String? error;
  final ScrollController scrollController;
  final ScrollPhysics physics;
  final Future<void> Function(bool) onAvailabilityChanged;
  final Future<void> Function() onPublishLocation;
  final Future<void> Function() onRefresh;

  @override
  Widget build(BuildContext context) => ListView(
    controller: scrollController,
    physics: physics,
    padding: const EdgeInsets.all(AppSpacing.md),
    children: [
      Text(
        profile.isOnline ? 'You are online' : 'Ready to go online',
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
          onPressed: busy
              ? null
              : () => onAvailabilityChanged(!profile.isOnline),
          icon: const Icon(Icons.power_settings_new),
          label: Text(profile.isOnline ? 'Go offline' : 'Go online'),
        ),
      ),
      const Text(
        'This is the legacy operational vehicle path. Vehicle/service selection will replace it before marketplace expansion.',
      ),
      const SizedBox(height: AppSpacing.sm),
      Text(
        location == null
            ? 'No location published during this visit.'
            : 'Location published at ${location!.updatedAt.toLocal()}',
      ),
      DashboardPanelControl(
        child: OutlinedButton.icon(
          onPressed: busy ? null : onPublishLocation,
          icon: const Icon(Icons.my_location),
          label: const Text('Update current location'),
        ),
      ),
      if (busy) const LinearProgressIndicator(),
      if (error != null)
        Text(
          error!,
          style: TextStyle(color: Theme.of(context).colorScheme.error),
        ),
      DashboardPanelControl(
        child: TextButton(
          onPressed: busy ? null : onRefresh,
          child: const Text('Refresh status'),
        ),
      ),
    ],
  );
}

class _DriverSetupForm extends StatefulWidget {
  const _DriverSetupForm({
    super.key,
    required this.services,
    required this.busy,
    required this.error,
    required this.scrollController,
    required this.physics,
    required this.onReview,
    this.onCancel,
  });

  final List<DriverServiceOption> services;
  final bool busy;
  final String? error;
  final ScrollController scrollController;
  final ScrollPhysics physics;
  final Future<void> Function(
    String,
    DriverServiceOption,
    DriverVehicle,
  ) onReview;
  final VoidCallback? onCancel;

  @override
  State<_DriverSetupForm> createState() => _DriverSetupFormState();
}

class _DriverSetupFormState extends State<_DriverSetupForm> {
  final _form = GlobalKey<FormState>();
  late final List<TextEditingController> _fields;
  String? _serviceCode;

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
    _fields = List.generate(6, (_) => TextEditingController());
    if (widget.services.isNotEmpty) {
      _serviceCode = widget.services.first.code;
    }
  }

  @override
  void didUpdateWidget(covariant _DriverSetupForm oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (_serviceCode == null && widget.services.isNotEmpty) {
      setState(() => _serviceCode = widget.services.first.code);
    }
  }

  @override
  void dispose() {
    for (final field in _fields) {
      field.dispose();
    }
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final selectedService = widget.services
        .where((service) => service.code == _serviceCode)
        .firstOrNull;
    return Form(
      key: _form,
      child: ListView(
        controller: widget.scrollController,
        physics: widget.physics,
        padding: const EdgeInsets.all(AppSpacing.md),
        children: [
          Text('Become a Driver', style: Theme.of(context).textTheme.headlineSmall),
          const Text('Choose one service for your first vehicle application.'),
          const SizedBox(height: AppSpacing.sm),
          DashboardPanelControl(
            child: DropdownButtonFormField<String>(
              key: const Key('driver-service-field'),
              value: _serviceCode,
              decoration: const InputDecoration(labelText: 'Service'),
              items: widget.services
                  .map(
                    (service) => DropdownMenuItem(
                      value: service.code,
                      child: Text(service.displayName),
                    ),
                  )
                  .toList(growable: false),
              onChanged: widget.busy
                  ? null
                  : (value) => setState(() => _serviceCode = value),
              validator: (value) => value == null ? 'Select a service' : null,
            ),
          ),
          if (selectedService != null) ...[
            const SizedBox(height: AppSpacing.sm),
            Text(selectedService.description),
            if (selectedService.minimumModelYear != null)
              Text(
                'Minimum model year: ${selectedService.minimumModelYear}',
              ),
          ],
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
              onPressed: widget.busy || selectedService == null
                  ? null
                  : () {
                      if (!_form.currentState!.validate()) return;
                      final values = _fields
                          .map((field) => field.text.trim())
                          .toList();
                      widget.onReview(
                        values[0],
                        selectedService,
                        DriverVehicle(
                          make: values[1],
                          model: values[2],
                          modelYear: int.parse(values[3]),
                          color: values[4],
                          licensePlate: values[5],
                        ),
                      );
                    },
              child: Text(widget.busy ? 'Checking…' : 'Review application'),
            ),
          ),
          if (widget.onCancel != null)
            DashboardPanelControl(
              child: TextButton(
                onPressed: widget.busy ? null : widget.onCancel,
                child: const Text('Cancel new application'),
              ),
            ),
        ],
      ),
    );
  }
}
