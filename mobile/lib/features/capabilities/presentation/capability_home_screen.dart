import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/dashboard/dashboard_panel_session.dart';
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
    final panelSession = ref.watch(dashboardPanelSessionProvider);
    final account = controller.state.account;
    if (account == null) {
      return const SizedBox.shrink();
    }
    final canDrive = account.capabilities.contains(Capability.driver);

    Future<void> selectCapability(Capability next) async {
      Navigator.of(context).pop();
      if (next == capability) return;
      await controller.selectCapability(next);
      if (context.mounted) context.go('/${next.name}');
    }

    Future<void> enableDriver() async {
      Navigator.of(context).pop();
      final enabled = await controller.enableDriver();
      if (!context.mounted) return;
      if (enabled) {
        await controller.selectCapability(Capability.driver);
        if (context.mounted) context.go('/driver');
        return;
      }
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(
            controller.state.error ?? 'Unable to enable Driver access.',
          ),
        ),
      );
    }

    void openDestination(String location) {
      context.push<void>(location);
    }

    Future<void> logout() async {
      Navigator.of(context).pop();
      await controller.logout();
      panelSession.reset();
    }

    return Scaffold(
      drawer: Drawer(
        child: SafeArea(
          child: ListView(
            padding: EdgeInsets.zero,
            children: [
              DrawerHeader(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  mainAxisAlignment: MainAxisAlignment.end,
                  children: [
                    const Icon(Icons.account_circle_outlined, size: 36),
                    const SizedBox(height: 12),
                    Text(
                      capability == Capability.rider ? 'Rider' : 'Driver',
                      style: Theme.of(context).textTheme.titleLarge,
                    ),
                    const SizedBox(height: 4),
                    Text(
                      'Account menu',
                      style: Theme.of(context).textTheme.bodyMedium,
                    ),
                  ],
                ),
              ),
              ListTile(
                key: const Key('drawerRider'),
                leading: const Icon(Icons.person_outline),
                title: const Text('Rider'),
                selected: capability == Capability.rider,
                onTap: controller.state.busy
                    ? null
                    : () => selectCapability(Capability.rider),
              ),
              if (canDrive)
                ListTile(
                  key: const Key('drawerDriver'),
                  leading: const Icon(Icons.local_taxi_outlined),
                  title: const Text('Driver'),
                  selected: capability == Capability.driver,
                  onTap: controller.state.busy
                      ? null
                      : () => selectCapability(Capability.driver),
                )
              else
                ListTile(
                  key: const Key('drawerBecomeDriver'),
                  leading: const Icon(Icons.add_circle_outline),
                  title: const Text('Become a Driver'),
                  onTap: controller.state.busy ? null : enableDriver,
                ),
              if (canDrive) ...[
                const Divider(),
                ListTile(
                  key: const Key('drawerDriverDetails'),
                  leading: const Icon(Icons.badge_outlined),
                  title: const Text('Driver details'),
                  onTap: controller.state.busy
                      ? null
                      : () => openDestination('/driver/details'),
                ),
                ListTile(
                  key: const Key('drawerVehicles'),
                  leading: const Icon(Icons.directions_car_outlined),
                  title: const Text('Vehicles'),
                  onTap: controller.state.busy
                      ? null
                      : () => openDestination('/driver/vehicles'),
                ),
              ],
              const Divider(),
              ListTile(
                key: const Key('drawerLogout'),
                leading: const Icon(Icons.logout),
                title: const Text('Log out'),
                onTap: controller.state.busy ? null : logout,
              ),
            ],
          ),
        ),
      ),
      appBar: AppBar(
        leading: Builder(
          builder: (context) => IconButton(
            key: const Key('capabilityMenuButton'),
            tooltip: 'Open menu',
            onPressed: Scaffold.of(context).openDrawer,
            icon: const Icon(Icons.menu),
          ),
        ),
        title: Text(capability == Capability.rider ? 'Rider' : 'Driver'),
        notificationPredicate: (_) => false,
      ),
      body: DashboardPanelSessionScope(
        session: panelSession,
        child: capability == Capability.rider
            ? const RiderRequestScreen()
            : DriverWorkspaceScreen(accountID: account.id),
      ),
    );
  }
}
