import 'package:flutter/material.dart';

import '../theme/app_theme.dart';

typedef DashboardPanelBuilder = Widget Function(
  BuildContext context,
  ScrollController scrollController,
);

/// Shared map-first dashboard shell for Rider and Driver workflows.
///
/// The shell keeps the map as the spatial base layer and lets business slices
/// swap task panels without rebuilding each flow as an unrelated static page.
class RideDashboardScaffold extends StatefulWidget {
  const RideDashboardScaffold({
    super.key,
    required this.map,
    required this.panelBuilder,
    this.floatingStatus,
    this.mapControls,
    this.minPanelSize = 0.16,
    this.initialPanelSize = 0.16,
    this.maxPanelSize = 0.60,
  }) : assert(minPanelSize > 0),
       assert(minPanelSize <= initialPanelSize),
       assert(initialPanelSize <= maxPanelSize),
       assert(maxPanelSize <= 0.60);

  final Widget map;
  final DashboardPanelBuilder panelBuilder;
  final Widget? floatingStatus;
  final Widget? mapControls;
  final double minPanelSize;
  final double initialPanelSize;
  final double maxPanelSize;

  @override
  State<RideDashboardScaffold> createState() => _RideDashboardScaffoldState();
}

class _RideDashboardScaffoldState extends State<RideDashboardScaffold> {
  final _contentScrollController = ScrollController();
  late double _panelSize;

  @override
  void initState() {
    super.initState();
    _panelSize = widget.initialPanelSize;
  }

  @override
  void didUpdateWidget(covariant RideDashboardScaffold oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.minPanelSize != widget.minPanelSize ||
        oldWidget.maxPanelSize != widget.maxPanelSize) {
      _panelSize = _panelSize.clamp(widget.minPanelSize, widget.maxPanelSize);
    }
  }

  @override
  void dispose() {
    _contentScrollController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: (context, constraints) {
        final minimumPanelHeight = constraints.maxHeight * widget.minPanelSize;
        final controlBottom = minimumPanelHeight + AppSpacing.lg;

        return Stack(
          children: [
            Positioned.fill(child: RepaintBoundary(child: widget.map)),
            if (widget.floatingStatus != null)
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
                      child: RepaintBoundary(child: widget.floatingStatus!),
                    ),
                  ),
                ),
              ),
            if (widget.mapControls != null)
              Positioned(
                right: AppSpacing.md,
                bottom: controlBottom,
                child: SafeArea(
                  top: false,
                  child: RepaintBoundary(child: widget.mapControls!),
                ),
              ),
            Positioned(
              left: 0,
              right: 0,
              bottom: 0,
              height: constraints.maxHeight * _panelSize,
              child: SizedBox.expand(
                key: const Key('dashboardPanel'),
                child: SafeArea(
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
                        child: RepaintBoundary(
                          child: Column(
                            children: [
                              _PanelDragHandle(
                                onDragUpdate: (delta) =>
                                    _resizePanel(delta, constraints.maxHeight),
                              ),
                              Expanded(
                                child: widget.panelBuilder(
                                  context,
                                  _contentScrollController,
                                ),
                              ),
                            ],
                          ),
                        ),
                      ),
                    ),
                  ),
                ),
              ),
            ),
          ],
        );
      },
    );
  }

  void _resizePanel(double verticalDelta, double dashboardHeight) {
    if (dashboardHeight <= 0) {
      return;
    }
    setState(() {
      _panelSize = (_panelSize - verticalDelta / dashboardHeight).clamp(
        widget.minPanelSize,
        widget.maxPanelSize,
      );
    });
  }
}

class _PanelDragHandle extends StatelessWidget {
  const _PanelDragHandle({required this.onDragUpdate});

  final ValueChanged<double> onDragUpdate;

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      key: const Key('dashboardPanelDragHandle'),
      behavior: HitTestBehavior.opaque,
      onVerticalDragUpdate: (details) => onDragUpdate(details.delta.dy),
      child: const SizedBox(
        height: 44,
        width: double.infinity,
        child: DashboardPanelHandle(),
      ),
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
  Widget build(BuildContext context) {
    return Material(
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
}
