import 'package:flutter/material.dart';

import '../theme/app_theme.dart';

/// Shared map-first dashboard shell for Rider and Driver workflows.
///
/// The shell keeps the map as the spatial base layer and lets business slices
/// swap task panels without rebuilding each flow as an unrelated static page.
class RideDashboardScaffold extends StatelessWidget {
  const RideDashboardScaffold({
    super.key,
    required this.map,
    required this.panel,
    this.floatingStatus,
    this.maxPanelHeightFactor = 0.62,
  });

  final Widget map;
  final Widget panel;
  final Widget? floatingStatus;
  final double maxPanelHeightFactor;

  @override
  Widget build(BuildContext context) => LayoutBuilder(
    builder: (context, constraints) {
      final panelMaxHeight = constraints.maxHeight * maxPanelHeightFactor;
      return Stack(
        children: [
          Positioned.fill(child: map),
          if (floatingStatus != null)
            Positioned(
              top: AppSpacing.md,
              left: AppSpacing.md,
              right: AppSpacing.md,
              child: SafeArea(
                bottom: false,
                child: Align(
                  alignment: Alignment.topCenter,
                  child: ConstrainedBox(
                    constraints: const BoxConstraints(maxWidth: 640),
                    child: floatingStatus!,
                  ),
                ),
              ),
            ),
          Align(
            alignment: Alignment.bottomCenter,
            child: SafeArea(
              top: false,
              minimum: const EdgeInsets.fromLTRB(
                AppSpacing.md,
                0,
                AppSpacing.md,
                AppSpacing.md,
              ),
              child: ConstrainedBox(
                constraints: BoxConstraints(
                  maxWidth: 640,
                  maxHeight: panelMaxHeight,
                ),
                child: Material(
                  color: Theme.of(context).colorScheme.surface,
                  elevation: 8,
                  shadowColor: Colors.black26,
                  borderRadius: const BorderRadius.all(
                    Radius.circular(AppRadii.xl),
                  ),
                  clipBehavior: Clip.antiAlias,
                  child: panel,
                ),
              ),
            ),
          ),
        ],
      );
    },
  );
}

class DashboardStatusCard extends StatelessWidget {
  const DashboardStatusCard({
    super.key,
    required this.icon,
    required this.title,
    required this.message,
  });

  final IconData icon;
  final String title;
  final String message;

  @override
  Widget build(BuildContext context) => Material(
    color: Theme.of(context).colorScheme.surface,
    elevation: 4,
    shadowColor: Colors.black12,
    borderRadius: const BorderRadius.all(Radius.circular(AppRadii.lg)),
    child: Padding(
      padding: const EdgeInsets.all(AppSpacing.md),
      child: Row(
        children: [
          Icon(icon),
          const SizedBox(width: AppSpacing.sm),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              mainAxisSize: MainAxisSize.min,
              children: [
                Text(title, style: Theme.of(context).textTheme.titleMedium),
                const SizedBox(height: AppSpacing.xxs),
                Text(message, style: Theme.of(context).textTheme.bodySmall),
              ],
            ),
          ),
        ],
      ),
    ),
  );
}
