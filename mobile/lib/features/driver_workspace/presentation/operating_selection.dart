import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/dashboard/ride_dashboard_scaffold.dart';
import '../../../core/providers.dart';
import '../../../core/theme/app_theme.dart';
import '../domain/operating_state.dart';
import '../domain/registered_vehicle.dart';

class OperatingSelectionControl extends ConsumerWidget {
  const OperatingSelectionControl({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final driver = ref.watch(driverControllerProvider);
    final operation = driver.operation;
    final choices = _choices(driver.vehicles);
    final activeChoices = choices
        .where((choice) => choice.service.serviceActive)
        .toList(growable: false);
    final inactiveChoices = choices
        .where((choice) => !choice.service.serviceActive)
        .toList(growable: false);
    final selected = _selectedChoice(activeChoices, operation);
    final canChangeSelection =
        operation?.canChange == true &&
        driver.profile?.isOnline != true &&
        !driver.hasActiveTrip;
    final selectionLocked =
        operation != null && activeChoices.isNotEmpty && !canChangeSelection;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Text(
          'Operating vehicle and service',
          style: Theme.of(context).textTheme.titleMedium,
        ),
        const SizedBox(height: AppSpacing.xs),
        if (selected != null) _SelectedOperatingContext(choice: selected),
        if (selected == null && activeChoices.isNotEmpty)
          const Text('Choose an approved combination before going online.'),
        if (activeChoices.isEmpty)
          const Text(
            'No approved active vehicle/service combination is available.',
          ),
        if (activeChoices.isNotEmpty) ...[
          const SizedBox(height: AppSpacing.sm),
          Text(
            'Approved operating options',
            style: Theme.of(context).textTheme.labelLarge,
          ),
          const SizedBox(height: AppSpacing.xs),
          for (final choice in activeChoices)
            DashboardPanelControl(
              child: _OperatingChoiceTile(
                choice: choice,
                selected: choice.matches(operation),
                enabled: !driver.busy && canChangeSelection,
                onSelect: () {
                  driver.selectOperation(
                    choice.record.id,
                    choice.service.serviceCode,
                  );
                },
              ),
            ),
        ],
        if (inactiveChoices.isNotEmpty) ...[
          const SizedBox(height: AppSpacing.md),
          Text(
            'Unavailable enrollments',
            style: Theme.of(context).textTheme.labelLarge,
          ),
          const SizedBox(height: AppSpacing.xs),
          for (final choice in inactiveChoices)
            _UnavailableChoiceTile(choice: choice),
        ],
        if (selectionLocked) ...[
          const SizedBox(height: AppSpacing.sm),
          const Text(
            'Go offline and finish any active trip before changing selection.',
          ),
        ],
      ],
    );
  }
}

class _OperatingChoiceTile extends StatelessWidget {
  const _OperatingChoiceTile({
    required this.choice,
    required this.selected,
    required this.enabled,
    required this.onSelect,
  });

  final _OperatingChoice choice;
  final bool selected;
  final bool enabled;
  final VoidCallback onSelect;

  @override
  Widget build(BuildContext context) {
    return ListTile(
      key: ValueKey('operating-option-${choice.key}'),
      contentPadding: EdgeInsets.zero,
      enabled: selected || enabled,
      leading: Icon(
        selected ? Icons.check_circle : Icons.radio_button_unchecked,
      ),
      title: Text(choice.vehicleLabel),
      subtitle: Text(choice.service.displayName),
      trailing: selected ? const Text('Selected') : null,
      onTap: selected || !enabled ? null : onSelect,
    );
  }
}

class _UnavailableChoiceTile extends StatelessWidget {
  const _UnavailableChoiceTile({required this.choice});

  final _OperatingChoice choice;

  @override
  Widget build(BuildContext context) {
    return ListTile(
      key: ValueKey('operating-unavailable-${choice.key}'),
      contentPadding: EdgeInsets.zero,
      enabled: false,
      leading: const Icon(Icons.block),
      title: Text(choice.vehicleLabel),
      subtitle: Text('${choice.service.displayName} (currently unavailable)'),
    );
  }
}

class _SelectedOperatingContext extends StatelessWidget {
  const _SelectedOperatingContext({required this.choice});

  final _OperatingChoice choice;

  @override
  Widget build(BuildContext context) {
    return DecoratedBox(
      key: const ValueKey('operating-current-selection'),
      decoration: BoxDecoration(
        border: Border.all(color: Theme.of(context).dividerColor),
        borderRadius: BorderRadius.circular(AppRadii.md),
      ),
      child: Padding(
        padding: const EdgeInsets.all(AppSpacing.sm),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'Current selection',
              style: Theme.of(context).textTheme.labelLarge,
            ),
            Text(choice.vehicleLabel),
            Text('Service: ${choice.service.displayName}'),
          ],
        ),
      ),
    );
  }
}

class _OperatingChoice {
  const _OperatingChoice({required this.record, required this.service});

  final RegisteredVehicle record;
  final ApprovedServiceEnrollment service;

  String get key => '${record.id}-${service.serviceCode}';

  String get vehicleLabel {
    final vehicle = record.vehicle;
    final year = vehicle.modelYear == null ? '' : ' ${vehicle.modelYear}';
    return '${vehicle.make} ${vehicle.model}$year • ${vehicle.licensePlate}';
  }

  bool matches(OperatingState? operation) {
    return operation?.valid == true &&
        operation?.vehicleId == record.id &&
        operation?.serviceCode == service.serviceCode;
  }
}

List<_OperatingChoice> _choices(List<RegisteredVehicle> vehicles) {
  return [
    for (final record in vehicles)
      for (final service in record.enrollments)
        _OperatingChoice(record: record, service: service),
  ];
}

_OperatingChoice? _selectedChoice(
  List<_OperatingChoice> choices,
  OperatingState? operation,
) {
  for (final choice in choices) {
    if (choice.matches(operation)) return choice;
  }
  return null;
}
