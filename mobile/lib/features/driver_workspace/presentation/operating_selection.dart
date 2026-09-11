import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/providers.dart';
import '../../../core/dashboard/ride_dashboard_scaffold.dart';

class OperatingSelectionControl extends ConsumerWidget {
  const OperatingSelectionControl({super.key});
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final driver = ref.watch(driverControllerProvider);
    final operation = driver.operation;
    final choices = [
      for (final record in driver.vehicles)
        for (final service in record.enrollments)
          if (service.serviceActive) (record: record, service: service),
    ];
    return Column(crossAxisAlignment: CrossAxisAlignment.stretch, children: [
      Text('Operating vehicle and service', style: Theme.of(context).textTheme.titleMedium),
      if (choices.isEmpty) const Text('No approved vehicle and service are available. Contact support for review.'),
      if (operation?.valid != true && choices.isNotEmpty) const Text('Choose an approved combination before going online.'),
      for (final choice in choices)
        DashboardPanelControl(child: ListTile(
          key: ValueKey('select-${choice.record.id}-${choice.service.serviceCode}'),
          title: Text('${choice.record.vehicle.make} ${choice.record.vehicle.model} • ${choice.record.vehicle.licensePlate}'),
          subtitle: Text(choice.service.displayName),
          trailing: operation?.vehicleId == choice.record.id && operation?.serviceCode == choice.service.serviceCode && operation?.valid == true
              ? const Icon(Icons.check_circle) : const Icon(Icons.radio_button_unchecked),
          onTap: driver.busy || operation?.canChange != true || driver.profile?.isOnline == true
              ? null : () => driver.selectOperation(choice.record.id, choice.service.serviceCode),
        )),
      if (operation?.canChange == false) const Text('Go offline and finish your active trip before changing selection.'),
    ]);
  }
}
