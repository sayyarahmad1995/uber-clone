import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/dashboard/ride_dashboard_scaffold.dart';
import '../../../core/maps/ride_map.dart';
import '../../../core/providers.dart';
import '../../../core/theme/app_theme.dart';

/// Driver dashboard entry point.
///
/// This intentionally remains lightweight: it gives the Driver slice a real
/// dashboard surface without pretending onboarding, availability, marketplace,
/// or trip controls are implemented yet.
class DriverWorkspaceScreen extends ConsumerWidget {
  const DriverWorkspaceScreen({super.key, required this.accountID});

  final String accountID;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final tiles = ref.watch(mapTilesProvider);
    return RideDashboardScaffold(
      minPanelSize: 0.16,
      initialPanelSize: 0.16,
      maxPanelSize: 0.60,
      map: RideMap(tiles: tiles),
      floatingStatus: const DashboardStatusCard(
        icon: Icons.local_taxi,
        title: 'Driver dashboard',
        message: 'Driver onboarding and marketplace panels will attach here.',
      ),
      panelBuilder: (context, scrollController) => _DriverFoundationPanel(
        accountID: accountID,
        scrollController: scrollController,
      ),
    );
  }
}

class _DriverFoundationPanel extends StatelessWidget {
  const _DriverFoundationPanel({
    required this.accountID,
    required this.scrollController,
  });

  final String accountID;
  final ScrollController scrollController;

  @override
  Widget build(BuildContext context) {
    return ListView(
      controller: scrollController,
      padding: const EdgeInsets.all(AppSpacing.md),
      children: [
        Text(
          'Driver workspace',
          style: Theme.of(context).textTheme.headlineSmall,
        ),
        const SizedBox(height: AppSpacing.xs),
        Text('Account $accountID'),
        const SizedBox(height: AppSpacing.md),
        const _NextPanelStep(
          icon: Icons.badge_outlined,
          title: 'Onboarding panel',
          message: 'Collect Driver and vehicle profile details.',
        ),
        const _NextPanelStep(
          icon: Icons.power_settings_new,
          title: 'Availability panel',
          message: 'Let the Driver go online and offline.',
        ),
        const _NextPanelStep(
          icon: Icons.near_me_outlined,
          title: 'Location panel',
          message: 'Publish current location before marketplace discovery.',
        ),
        const _NextPanelStep(
          icon: Icons.format_list_bulleted,
          title: 'Marketplace panel',
          message: 'Show eligible Rider requests and offer actions.',
        ),
      ],
    );
  }
}

class _NextPanelStep extends StatelessWidget {
  const _NextPanelStep({
    required this.icon,
    required this.title,
    required this.message,
  });

  final IconData icon;
  final String title;
  final String message;

  @override
  Widget build(BuildContext context) => ListTile(
    contentPadding: EdgeInsets.zero,
    leading: Icon(icon),
    title: Text(title),
    subtitle: Text(message),
  );
}
