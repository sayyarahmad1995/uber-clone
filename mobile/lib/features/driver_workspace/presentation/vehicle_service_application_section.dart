import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/providers.dart';
import '../../../core/theme/app_theme.dart';
import '../application/driver_onboarding_controller.dart';
import '../domain/driver_onboarding.dart';
import '../domain/driver_profile.dart';

class VehicleServiceApplicationSection extends ConsumerStatefulWidget {
  const VehicleServiceApplicationSection({super.key});

  @override
  ConsumerState<VehicleServiceApplicationSection> createState() =>
      _VehicleServiceApplicationSectionState();
}

class _VehicleServiceApplicationSectionState
    extends ConsumerState<VehicleServiceApplicationSection> {
  final _form = GlobalKey<FormState>();
  final _fields = List.generate(5, (_) => TextEditingController());
  bool _adding = false;
  String? _serviceCode;

  static const _labels = [
    'Vehicle make',
    'Vehicle model',
    'Model year',
    'Vehicle color',
    'License plate',
  ];

  @override
  void dispose() {
    for (final field in _fields) {
      field.dispose();
    }
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final onboarding = ref.watch(driverOnboardingControllerProvider);
    final driver = ref.watch(driverControllerProvider);
    final selectedService = _selectedService(onboarding.services);
    _serviceCode ??= selectedService?.code;

    final latestAdditional = _latestAdditional(onboarding.application);
    final child = _adding
        ? _buildForm(
            context: context,
            onboarding: onboarding,
            profile: driver.profile,
            selectedService: selectedService,
          )
        : _buildSummary(
            context: context,
            onboarding: onboarding,
            latestAdditional: latestAdditional,
          );

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(AppSpacing.md),
        child: AnimatedSwitcher(
          duration: const Duration(milliseconds: 200),
          child: child,
        ),
      ),
    );
  }

  Widget _buildSummary({
    required BuildContext context,
    required DriverOnboardingController onboarding,
    required DriverOnboardingApplication? latestAdditional,
  }) {
    final pending = latestAdditional?.isPending == true;
    return Column(
      key: const ValueKey('additional-vehicle-summary'),
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Text(
          'Vehicle/service applications',
          style: Theme.of(context).textTheme.titleMedium,
        ),
        const SizedBox(height: AppSpacing.xs),
        Text(_summaryMessage(pending)),
        if (latestAdditional != null) ...[
          const SizedBox(height: AppSpacing.sm),
          _ApplicationSummary(application: latestAdditional),
        ],
        if (onboarding.error != null) ...[
          const SizedBox(height: AppSpacing.sm),
          Text(
            onboarding.error!,
            style: TextStyle(color: Theme.of(context).colorScheme.error),
          ),
        ],
        const SizedBox(height: AppSpacing.sm),
        Wrap(
          spacing: AppSpacing.sm,
          runSpacing: AppSpacing.xs,
          children: [
            FilledButton.icon(
              key: const Key('driverAddVehicleServiceApplicationButton'),
              onPressed: onboarding.busy || pending
                  ? null
                  : () => setState(() => _adding = true),
              icon: const Icon(Icons.add),
              label: const Text('Add vehicle/service'),
            ),
            TextButton(
              onPressed: onboarding.busy ? null : () => _refresh(onboarding),
              child: const Text('Refresh'),
            ),
          ],
        ),
      ],
    );
  }

  Widget _buildForm({
    required BuildContext context,
    required DriverOnboardingController onboarding,
    required DriverProfile? profile,
    required DriverServiceOption? selectedService,
  }) {
    return Form(
      key: _form,
      child: Column(
        key: const ValueKey('additional-vehicle-form'),
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Text(
            'Add vehicle/service',
            style: Theme.of(context).textTheme.titleMedium,
          ),
          const SizedBox(height: AppSpacing.xs),
          const Text(
            'Submit a new vehicle and one requested service for review. '
            'Approval will add a new approved operating option.',
          ),
          const SizedBox(height: AppSpacing.sm),
          _ServicePicker(
            services: onboarding.services,
            serviceCode: _serviceCode,
            busy: onboarding.busy,
            onChanged: (value) => setState(() => _serviceCode = value),
          ),
          if (selectedService != null) ...[
            const SizedBox(height: AppSpacing.xs),
            Text(selectedService.description),
            if (selectedService.minimumModelYear != null)
              Text('Minimum model year: ${selectedService.minimumModelYear}'),
          ],
          for (var index = 0; index < _labels.length; index++)
            Padding(
              padding: const EdgeInsets.only(top: AppSpacing.sm),
              child: TextFormField(
                key: ValueKey('additional-vehicle-field-$index'),
                controller: _fields[index],
                enabled: !onboarding.busy,
                decoration: InputDecoration(labelText: _labels[index]),
                keyboardType: index == 2
                    ? TextInputType.number
                    : TextInputType.text,
                validator: (value) => _validateField(index, value),
              ),
            ),
          if (onboarding.error != null) ...[
            const SizedBox(height: AppSpacing.sm),
            Text(
              onboarding.error!,
              style: TextStyle(color: Theme.of(context).colorScheme.error),
            ),
          ],
          const SizedBox(height: AppSpacing.sm),
          Wrap(
            spacing: AppSpacing.sm,
            runSpacing: AppSpacing.xs,
            children: [
              FilledButton(
                onPressed: onboarding.busy || selectedService == null
                    ? null
                    : () => _reviewAndSubmit(
                          onboarding: onboarding,
                          profile: profile,
                          service: selectedService,
                        ),
                child: Text(
                  onboarding.busy ? 'Checking…' : 'Review application',
                ),
              ),
              TextButton(
                onPressed: onboarding.busy
                    ? null
                    : () => setState(() => _adding = false),
                child: const Text('Cancel'),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Future<void> _reviewAndSubmit({
    required DriverOnboardingController onboarding,
    required DriverProfile? profile,
    required DriverServiceOption service,
  }) async {
    if (!_form.currentState!.validate()) return;

    final vehicle = _vehicleFromFields();
    final displayName = _displayName(profile);
    final precheck = await onboarding.precheck(
      displayName: displayName,
      serviceCode: service.code,
      vehicle: vehicle,
    );
    if (!mounted || precheck == null) return;

    if (!precheck.eligible) {
      await _showIneligibleDialog(precheck);
      return;
    }

    final confirmed = await _confirmSubmission(service, vehicle);
    if (confirmed != true || !mounted) return;

    await onboarding.submit(
      displayName: displayName,
      serviceCode: service.code,
      vehicle: vehicle,
    );
    if (!mounted || onboarding.error != null) return;
    setState(() => _adding = false);
  }

  Future<void> _showIneligibleDialog(DriverOnboardingPrecheck precheck) async {
    await showDialog<void>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Vehicle not eligible'),
        content: Text(_ineligibleMessage(precheck)),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Change application'),
          ),
        ],
      ),
    );
  }

  Future<bool?> _confirmSubmission(
    DriverServiceOption service,
    DriverVehicle vehicle,
  ) async {
    return showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Review vehicle/service application'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('Service: ${service.displayName}'),
            Text('Vehicle: ${_vehicleLabel(vehicle)}'),
            Text('Color: ${vehicle.color}'),
            Text('License plate: ${vehicle.licensePlate}'),
            const SizedBox(height: AppSpacing.sm),
            const Text(
              'Submitting sends this vehicle/service application for review. '
              'It will not change your approved Driver profile or online status.',
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
  }

  void _refresh(DriverOnboardingController onboarding) {
    onboarding.load();
    ref.invalidate(driverVehiclesProvider);
  }

  DriverOnboardingApplication? _latestAdditional(
    DriverOnboardingApplication? application,
  ) {
    if (application?.isAdditionalVehicleService != true) return null;
    return application;
  }

  DriverServiceOption? _selectedService(List<DriverServiceOption> services) {
    for (final service in services) {
      if (service.code == _serviceCode) return service;
    }
    return services.isEmpty ? null : services.first;
  }

  DriverVehicle _vehicleFromFields() {
    final values = _fields.map((field) => field.text.trim()).toList();
    return DriverVehicle(
      make: values[0],
      model: values[1],
      modelYear: int.parse(values[2]),
      color: values[3],
      licensePlate: values[4],
    );
  }

  String? _validateField(int index, String? value) {
    if (value == null || value.trim().isEmpty) {
      return 'Required';
    }
    if (index == 2) {
      final year = int.tryParse(value.trim());
      if (year == null ||
          year < 1886 ||
          year > DateTime.now().toUtc().year + 1) {
        return 'Enter a valid model year';
      }
    }
    return null;
  }

  String _displayName(DriverProfile? profile) {
    final name = profile?.displayName?.trim();
    return name == null || name.isEmpty ? 'Driver' : name;
  }

  String _summaryMessage(bool pending) {
    if (pending) {
      return 'Your latest vehicle/service application is under review.';
    }
    return 'Add another vehicle and service without changing your approved '
        'Driver profile.';
  }

  String _ineligibleMessage(DriverOnboardingPrecheck precheck) {
    if (precheck.reasons.isEmpty) {
      return 'This vehicle does not meet the selected service requirements.';
    }
    return precheck.reasons.join('\n');
  }
}

class _ServicePicker extends StatelessWidget {
  const _ServicePicker({
    required this.services,
    required this.serviceCode,
    required this.busy,
    required this.onChanged,
  });

  final List<DriverServiceOption> services;
  final String? serviceCode;
  final bool busy;
  final ValueChanged<String?> onChanged;

  @override
  Widget build(BuildContext context) {
    return DropdownButtonFormField<String>(
      key: const Key('additional-service-field'),
      initialValue: serviceCode,
      decoration: const InputDecoration(labelText: 'Requested service'),
      items: services
          .map(
            (service) => DropdownMenuItem<String>(
              value: service.code,
              child: Text(service.displayName),
            ),
          )
          .toList(growable: false),
      onChanged: busy || services.isEmpty ? null : onChanged,
      validator: (value) => value == null ? 'Select a service' : null,
    );
  }
}

class _ApplicationSummary extends StatelessWidget {
  const _ApplicationSummary({required this.application});

  final DriverOnboardingApplication application;

  @override
  Widget build(BuildContext context) {
    return DecoratedBox(
      decoration: BoxDecoration(
        border: Border.all(color: Theme.of(context).dividerColor),
        borderRadius: BorderRadius.circular(AppRadii.md),
      ),
      child: Padding(
        padding: const EdgeInsets.all(AppSpacing.sm),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(_latestApplicationLabel(application)),
            Text('Service: ${application.service.displayName}'),
            Text('${_vehicleLabel(application.vehicle)} • ${application.vehicle.color}'),
            Text('License plate: ${application.vehicle.licensePlate}'),
            if (application.rejectionReason != null)
              Text(
                application.rejectionReason!,
                style: TextStyle(color: Theme.of(context).colorScheme.error),
              ),
          ],
        ),
      ),
    );
  }
}

String _latestApplicationLabel(DriverOnboardingApplication application) {
  return 'Latest additional application: ${_applicationStatus(application)}';
}

String _applicationStatus(DriverOnboardingApplication application) {
  if (application.isPending) return 'Under review';
  if (application.isRejected) return 'Rejected';
  return 'Approved';
}

String _vehicleLabel(DriverVehicle vehicle) {
  final year = vehicle.modelYear == null ? '' : ' ${vehicle.modelYear}';
  return '${vehicle.make} ${vehicle.model}$year';
}
