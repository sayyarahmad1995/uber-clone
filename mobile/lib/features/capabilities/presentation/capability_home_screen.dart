import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/models/account.dart';
import '../../../core/providers.dart';
import '../../driver_workspace/presentation/driver_workspace_screen.dart';
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
      resizeToAvoidBottomInset: false,
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
          : DriverWorkspaceScreen(accountID: account.id),
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
