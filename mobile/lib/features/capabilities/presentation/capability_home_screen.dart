import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/models/account.dart';
import '../../../core/providers.dart';

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
      body: Center(
        child: ConstrainedBox(
          constraints: const BoxConstraints(maxWidth: 480),
          child: Padding(
            padding: const EdgeInsets.all(24),
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                Icon(
                  capability == Capability.rider
                      ? Icons.person_pin_circle
                      : Icons.local_taxi,
                  size: 72,
                ),
                const SizedBox(height: 24),
                Text(
                  capability == Capability.rider
                      ? 'Ready to request a ride'
                      : 'Driver workspace',
                  textAlign: TextAlign.center,
                  style: Theme.of(context).textTheme.headlineSmall,
                ),
                const SizedBox(height: 8),
                Text(
                  'Account ${account.id}',
                  textAlign: TextAlign.center,
                  style: Theme.of(context).textTheme.bodySmall,
                ),
                if (canDrive) ...[
                  const SizedBox(height: 32),
                  SegmentedButton<Capability>(
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
                ],
              ],
            ),
          ),
        ),
      ),
    );
  }
}
