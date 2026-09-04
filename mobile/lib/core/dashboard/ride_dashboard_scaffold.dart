import 'package:flutter/material.dart';

import '../theme/app_theme.dart';

typedef DashboardPanelBuilder =
    Widget Function(BuildContext context, ScrollController scrollController);

/// Shared map-first dashboard shell for Rider and Driver workflows.
///
/// The shell keeps the map as the spatial base layer and lets business slices
/// swap task panels without rebuilding each flow as an unrelated static page.
class RideDashboardScaffold extends StatelessWidget {
  const RideDashboardScaffold({
    super.key,
    required this.map,
    required this.panelBuilder,
    this.floatingStatus,
    this.mapControls,
    this.minPanelSize = 0.16,
    this.initialPanelSize = 0.34,
    this.maxPanelSize = 0.62,
  }) : assert(minPanelSize > 0),
       assert(minPanelSize <= initialPanelSize),
       assert(initialPanelSize <= maxPanelSize),
       assert(maxPanelSize <= 1);

  final Widget map;
  final DashboardPanelBuilder panelBuilder;
  final Widget? floatingStatus;
  final Widget? mapControls;
  final double minPanelSize;
  final double initialPanelSize;
  final double maxPanelSize;

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: (context, constraints) {
        final controlBottom =
            (constraints.maxHeight * minPanelSize) + AppSpacing.lg;

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
            if (mapControls != null)
              Positioned(
                right: AppSpacing.md,
                bottom: controlBottom,
                child: SafeArea(
                  top: false,
                  child: mapControls!,
                ),
              ),
            DraggableScrollableSheet(
              minChildSize: minPanelSize,
              initialChildSize: initialPanelSize,
              maxChildSize: maxPanelSize,
              snap: true,
              builder: (context, scrollController) {
                return SafeArea(
                  top: false,
                  minimum: const EdgeInsets.fromLTRB(
                    AppSpacing.md,
                    0,
                    AppSpacing.md,
                    AppSpacing.md,
                  ),
                  child: Align(
                    alignment: Alignment.bottomCenter,
                    child: ConstrainedBox(
                      constraints: const BoxConstraints(maxWidth: 640),
                      child: Material(
                        color: Theme.of(context).colorScheme.surface,
                        elevation: 8,
                        shadowColor: Colors.black26,
                        borderRadius: const BorderRadius.all(
                          Radius.circular(AppRadii.xl),
                        ),
                        clipBehavior: Clip.antiAlias,
                        child: panelBuilder(context, scrollController),
                      ),
                    ),
                  ),
                );
              },
            ),
          ],
        );
      },
    );
  }
}

class DashboardPanelHandle extends StatelessWidget {
  const DashboardPanelHandle({super.key});

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Container(
        width: 44,
        height: 4,
        margin: const EdgeInsets.only(bottom: AppSpacing.md),
        decoration: BoxDecoration(
          color: Theme.of(context).colorScheme.outlineVariant,
          borderRadius: const BorderRadius.all(Radius.circular(AppRadii.sm)),
        ),
      ),
    );
  }
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
