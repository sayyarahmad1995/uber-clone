import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/providers.dart';
import '../domain/driver_onboarding.dart';
import '../domain/driver_profile.dart';

class DriverDetailsScreen extends ConsumerWidget {
  const DriverDetailsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final driver = ref.watch(driverControllerProvider);
    final onboarding = ref.watch(driverOnboardingControllerProvider);
    final profile = driver.profile;
    final application = onboarding.application;
    final loading = !driver.loaded || (profile == null && !onboarding.loaded);
    final error = profile == null
        ? driver.error ?? onboarding.error
        : driver.error;

    return Scaffold(
      appBar: AppBar(title: const Text('Driver details')),
      body: _ReadOnlyBody(
        loading: loading,
        error: error,
        empty: profile == null && application == null,
        emptyTitle: 'No submitted Driver details yet',
        emptyMessage: 'Complete Driver onboarding from the Driver dashboard to create Driver details.',
        child: _DriverDetailsContent(
          profile: profile,
          application: application,
        ),
      ),
    );
  }
}

class DriverVehiclesScreen extends ConsumerWidget {
  const DriverVehiclesScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final records = ref.watch(driverVehiclesProvider);
    final onboarding = ref.watch(driverOnboardingControllerProvider);
    final application = onboarding.application;
    return Scaffold(
      appBar: AppBar(title: const Text('Vehicles')),
      body: records.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (error, stack) => Center(child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Text('Unable to load vehicles'),
            TextButton(onPressed: () => ref.invalidate(driverVehiclesProvider), child: const Text('Retry')),
          ],
        )),
        data: (vehicles) => _ReadOnlyBody(
          loading: vehicles.isEmpty && !onboarding.loaded,
          error: onboarding.error,
          empty: vehicles.isEmpty && application == null,
          emptyTitle: 'No vehicle submitted yet',
          emptyMessage: 'Complete Driver onboarding from the Driver dashboard to submit a vehicle.',
          child: Column(children: [
            for (final record in vehicles)
              Card(key: ValueKey(record.id), child: Column(children: [
                _ReadOnlyTile(icon: Icons.directions_car_outlined, label: 'Vehicle', value: '${record.vehicle.make} ${record.vehicle.model}'),
                if (record.vehicle.modelYear != null)
                  _ReadOnlyTile(icon: Icons.calendar_today_outlined, label: 'Model year', value: '${record.vehicle.modelYear}'),
                _ReadOnlyTile(icon: Icons.palette_outlined, label: 'Color', value: record.vehicle.color),
                _ReadOnlyTile(icon: Icons.pin_outlined, label: 'License plate', value: record.vehicle.licensePlate),
                if (record.enrollments.isEmpty)
                  const ListTile(title: Text('No approved service enrollments')),
                for (final enrollment in record.enrollments)
                  _ReadOnlyTile(icon: Icons.verified_outlined, label: 'Approved service', value: '${enrollment.displayName}${enrollment.serviceActive ? '' : ' (currently unavailable)'}'),
              ])),
            if (vehicles.isEmpty && application != null)
              _VehicleContent(vehicle: application.vehicle, profile: null, application: application),
          ]),
        ),
      ),
    );
  }
}

class _ReadOnlyBody extends StatelessWidget {
  const _ReadOnlyBody({
    required this.loading,
    required this.error,
    required this.empty,
    required this.emptyTitle,
    required this.emptyMessage,
    required this.child,
  });

  final bool loading;
  final String? error;
  final bool empty;
  final String emptyTitle;
  final String emptyMessage;
  final Widget child;

  @override
  Widget build(BuildContext context) {
    if (loading) {
      return const Center(child: CircularProgressIndicator());
    }
    if (error != null && empty) {
      return _MessageState(
        icon: Icons.error_outline,
        title: 'Unable to load Driver information',
        message: error!,
      );
    }
    if (empty) {
      return _MessageState(
        icon: Icons.info_outline,
        title: emptyTitle,
        message: emptyMessage,
      );
    }
    return ListView(
      padding: const EdgeInsets.all(16),
      children: [
        Card(
          child: Padding(
            padding: const EdgeInsets.all(16),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Icon(Icons.lock_outline),
                const SizedBox(width: 12),
                Expanded(
                  child: Text(
                    'Read-only. Approved Driver and vehicle information cannot be changed directly from this screen.',
                    style: Theme.of(context).textTheme.bodyMedium,
                  ),
                ),
              ],
            ),
          ),
        ),
        const SizedBox(height: 12),
        child,
      ],
    );
  }
}

class _MessageState extends StatelessWidget {
  const _MessageState({
    required this.icon,
    required this.title,
    required this.message,
  });

  final IconData icon;
  final String title;
  final String message;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(icon, size: 40),
            const SizedBox(height: 12),
            Text(title, style: Theme.of(context).textTheme.titleMedium),
            const SizedBox(height: 8),
            Text(message, textAlign: TextAlign.center),
          ],
        ),
      ),
    );
  }
}

class _DriverDetailsContent extends StatelessWidget {
  const _DriverDetailsContent({
    required this.profile,
    required this.application,
  });

  final DriverProfile? profile;
  final DriverOnboardingApplication? application;

  @override
  Widget build(BuildContext context) {
    final name = profile?.displayName ?? application?.displayName;
    final status = profile?.status ?? application?.status;
    return Card(
      child: Column(
        children: [
          _ReadOnlyTile(
            icon: Icons.badge_outlined,
            label: 'Display name',
            value: name ?? 'Not available',
          ),
          _ReadOnlyTile(
            icon: Icons.verified_outlined,
            label: 'Driver status',
            value: _statusLabel(status),
          ),
          if (profile != null)
            _ReadOnlyTile(
              icon: Icons.power_settings_new,
              label: 'Availability',
              value: profile!.isOnline ? 'Online' : 'Offline',
            ),
          if (application != null)
            _ReadOnlyTile(
              icon: Icons.local_taxi_outlined,
              label: 'Selected service',
              value: application!.service.displayName,
            ),
          if (application?.rejectionReason != null)
            _ReadOnlyTile(
              icon: Icons.report_outlined,
              label: 'Review reason',
              value: application!.rejectionReason!,
            ),
          if (application != null)
            _ReadOnlyTile(
              icon: Icons.schedule_outlined,
              label: 'Submitted',
              value: _dateLabel(application!.submittedAt),
            ),
          if (application?.decidedAt != null)
            _ReadOnlyTile(
              icon: Icons.event_available_outlined,
              label: 'Decided',
              value: _dateLabel(application!.decidedAt!),
            ),
        ],
      ),
    );
  }
}

class _VehicleContent extends StatelessWidget {
  const _VehicleContent({
    required this.vehicle,
    required this.profile,
    required this.application,
  });

  final DriverVehicle vehicle;
  final DriverProfile? profile;
  final DriverOnboardingApplication? application;

  @override
  Widget build(BuildContext context) {
    final contextLabel = profile != null
        ? 'Current Driver profile'
        : 'Driver onboarding application';
    final status = profile?.status ?? application?.status;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Text('Submitted vehicle', style: Theme.of(context).textTheme.titleMedium),
        const SizedBox(height: 8),
        Card(
          child: Column(
            children: [
              _ReadOnlyTile(
                icon: Icons.directions_car_outlined,
                label: 'Source',
                value: contextLabel,
              ),
              _ReadOnlyTile(
                icon: Icons.verified_outlined,
                label: 'Status',
                value: _statusLabel(status),
              ),
              _ReadOnlyTile(
                icon: Icons.factory_outlined,
                label: 'Make',
                value: vehicle.make,
              ),
              _ReadOnlyTile(
                icon: Icons.directions_car_filled,
                label: 'Model',
                value: vehicle.model,
              ),
              if (vehicle.modelYear != null)
                _ReadOnlyTile(
                  icon: Icons.calendar_today_outlined,
                  label: 'Model year',
                  value: '${vehicle.modelYear}',
                ),
              _ReadOnlyTile(
                icon: Icons.palette_outlined,
                label: 'Color',
                value: vehicle.color,
              ),
              _ReadOnlyTile(
                icon: Icons.pin_outlined,
                label: 'License plate',
                value: vehicle.licensePlate,
              ),
              if (application != null)
                _ReadOnlyTile(
                  icon: Icons.local_taxi_outlined,
                  label: 'Selected service',
                  value: application!.service.displayName,
                ),
            ],
          ),
        ),
        const SizedBox(height: 8),
        Text(
          'This is your submitted application snapshot. Approved service enrollments are shown on registered vehicle records.',
          style: Theme.of(context).textTheme.bodySmall,
        ),
      ],
    );
  }
}

class _ReadOnlyTile extends StatelessWidget {
  const _ReadOnlyTile({
    required this.icon,
    required this.label,
    required this.value,
  });

  final IconData icon;
  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return ListTile(
      leading: Icon(icon),
      title: Text(label),
      subtitle: Text(value),
    );
  }
}

String _statusLabel(String? status) {
  if (status == null || status.isEmpty) return 'Not available';
  return status
      .split('_')
      .map(
        (part) => part.isEmpty
            ? part
            : '${part.substring(0, 1).toUpperCase()}${part.substring(1)}',
      )
      .join(' ');
}

String _dateLabel(DateTime value) {
  final local = value.toLocal();
  final month = local.month.toString().padLeft(2, '0');
  final day = local.day.toString().padLeft(2, '0');
  return '${local.year}-$month-$day';
}
