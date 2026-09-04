import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/models/account.dart';
import '../../../core/providers.dart';
import '../../rider_request/presentation/rider_request_screen.dart';

class CapabilityHomeScreen extends ConsumerWidget {
  const CapabilityHomeScreen({super.key, required this.capability});
  final Capability capability;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final controller = ref.watch(sessionControllerProvider);
    final account = controller.state.account!;
    final canDrive = account.capabilities.contains(Capability.driver);
    return Scaffold(
      appBar: AppBar(
        title: Text(capability == Capability.rider ? 'Rider' : 'Driver'),
        actions: [
          IconButton(
            tooltip: 'Log out',
            onPressed: controller.state.busy ? null : controller.logout,
            icon: const Icon(Icons.logout),
          ),
        ],
      ),
      body: capability == Capability.rider
          ? const RiderRequestScreen()
          : _DriverPlaceholder(accountID: account.id),
      bottomNavigationBar: canDrive
          ? SafeArea(
              minimum: const EdgeInsets.all(12),
              child: SegmentedButton<Capability>(
                segments: const [
                  ButtonSegment(
                    value: Capability.rider,
                    label: Text('Rider'),
                    icon: Icon(Icons.person),
                  ),
                  ButtonSegment(
                    value: Capability.driver,
                    label: Text('Driver'),
                    icon: Icon(Icons.local_taxi),
                  ),
                ],
                selected: {capability},
                onSelectionChanged: (selection) async {
                  final next = selection.single;
                  await controller.selectCapability(next);
                  if (context.mounted) context.go('/${next.name}');
                },
              ),
            )
          : null,
    );
  }
}

class _DriverPlaceholder extends StatelessWidget {
  const _DriverPlaceholder({required this.accountID});
  final String accountID;

  @override
  Widget build(BuildContext context) => Center(
    child: Padding(
      padding: const EdgeInsets.all(24),
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          const Icon(Icons.local_taxi, size: 72),
          const SizedBox(height: 24),
          Text(
            'Driver workspace',
            style: Theme.of(context).textTheme.headlineSmall,
          ),
          const SizedBox(height: 8),
          Text('Account $accountID'),
        ],
      ),
    ),
  );
}
