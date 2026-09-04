import 'package:flutter/material.dart';

import '../theme/app_theme.dart';

typedef DashboardPanelBuilder = Widget Function(
  BuildContext context,
  ScrollController scrollController,
  bool scrollEnabled,
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
  late double _dragStartSize;
  bool _isDraggingPanel = false;
  bool _contentDragStartedAtTop = false;
  bool _contentDragStartedCollapsed = false;
  double _collapsePullDistance = 0;
  double _expandPullDistance = 0;

  static const _collapsePullThreshold = 56.0;

  @override
  void initState() {
    super.initState();
    _panelSize = widget.initialPanelSize;
    _dragStartSize = _panelSize;
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
            AnimatedPositioned(
              left: 0,
              right: 0,
              bottom: 0,
              height: constraints.maxHeight * _panelSize,
              duration: _isDraggingPanel
                  ? Duration.zero
                  : const Duration(milliseconds: 220),
              curve: Curves.easeOutCubic,
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
                                onDragStart: _startPanelDrag,
                                onDragUpdate: (delta) =>
                                    _resizePanel(delta, constraints.maxHeight),
                                onDragEnd: _endPanelDrag,
                                onDragCancel: _cancelPanelDrag,
                              ),
                              Expanded(
                                child: Listener(
                                  onPointerDown: (_) => _startContentDrag(),
                                  onPointerMove: (event) => _updateContentDrag(
                                    event.delta.dy,
                                    constraints.maxHeight,
                                  ),
                                  onPointerUp: (_) => _releaseContentDrag(),
                                  onPointerCancel: (_) => _cancelContentDrag(),
                                  child: widget.panelBuilder(
                                    context,
                                    _contentScrollController,
                                    _isPanelExpanded,
                                  ),
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

  void _startPanelDrag() {
    setState(() {
      _isDraggingPanel = true;
      _dragStartSize = _panelSize;
    });
  }

  void _endPanelDrag(double velocity) {
    final movement = _panelSize - _dragStartSize;
    final expand = velocity < -50 || (velocity.abs() <= 50 && movement > 0);
    setState(() {
      _isDraggingPanel = false;
      _panelSize = expand ? widget.maxPanelSize : widget.minPanelSize;
    });
  }

  void _cancelPanelDrag() {
    final midpoint = (widget.minPanelSize + widget.maxPanelSize) / 2;
    setState(() {
      _isDraggingPanel = false;
      _panelSize = _panelSize >= midpoint
          ? widget.maxPanelSize
          : widget.minPanelSize;
    });
  }

  bool get _isPanelExpanded => (_panelSize - widget.maxPanelSize).abs() < 0.001;

  void _startContentDrag() {
    _contentDragStartedCollapsed = _isPanelCollapsed;
    _contentDragStartedAtTop =
        _isPanelExpanded &&
        _contentScrollController.hasClients &&
        _contentScrollController.position.pixels <=
            _contentScrollController.position.minScrollExtent + 0.5;
    _collapsePullDistance = 0;
    _expandPullDistance = 0;
    if (_contentDragStartedCollapsed || _contentDragStartedAtTop) {
      setState(() {
        _isDraggingPanel = true;
        _dragStartSize = _panelSize;
      });
    }
  }

  void _updateContentDrag(double verticalDelta, double dashboardHeight) {
    if (dashboardHeight <= 0) {
      return;
    }
    if (_contentDragStartedCollapsed) {
      _expandPullDistance = (_expandPullDistance - verticalDelta).clamp(
        0.0,
        dashboardHeight * (widget.maxPanelSize - widget.minPanelSize),
      );
      setState(() {
        _panelSize =
            (widget.minPanelSize + _expandPullDistance / dashboardHeight).clamp(
              widget.minPanelSize,
              widget.maxPanelSize,
            );
      });
      return;
    }
    if (!_contentDragStartedAtTop) {
      return;
    }
    _collapsePullDistance = (_collapsePullDistance + verticalDelta).clamp(
      0.0,
      dashboardHeight * (widget.maxPanelSize - widget.minPanelSize),
    );
    setState(() {
      _panelSize =
          (widget.maxPanelSize - _collapsePullDistance / dashboardHeight).clamp(
            widget.minPanelSize,
            widget.maxPanelSize,
          );
    });
  }

  void _releaseContentDrag() {
    final shouldExpand =
        _contentDragStartedCollapsed &&
        _expandPullDistance >= _collapsePullThreshold;
    final shouldCollapse =
        _contentDragStartedAtTop &&
        _collapsePullDistance >= _collapsePullThreshold;
    final hadBodyPanelDrag =
        _contentDragStartedCollapsed || _contentDragStartedAtTop;
    final returnSize = _dragStartSize;
    _resetContentDrag();
    if (!hadBodyPanelDrag) {
      return;
    }
    if ((shouldExpand || shouldCollapse) &&
        _contentScrollController.hasClients) {
      _contentScrollController.jumpTo(
        _contentScrollController.position.minScrollExtent,
      );
    }
    setState(() {
      _isDraggingPanel = false;
      _panelSize = shouldExpand
          ? widget.maxPanelSize
          : shouldCollapse
          ? widget.minPanelSize
          : returnSize;
    });
  }

  void _cancelContentDrag() {
    final hadBodyPanelDrag =
        _contentDragStartedCollapsed || _contentDragStartedAtTop;
    final returnSize = _dragStartSize;
    _resetContentDrag();
    if (hadBodyPanelDrag) {
      setState(() {
        _isDraggingPanel = false;
        _panelSize = returnSize;
      });
    }
  }

  void _resetContentDrag() {
    _contentDragStartedAtTop = false;
    _contentDragStartedCollapsed = false;
    _collapsePullDistance = 0;
    _expandPullDistance = 0;
  }

  bool get _isPanelCollapsed =>
      (_panelSize - widget.minPanelSize).abs() < 0.001;
}

class _PanelDragHandle extends StatelessWidget {
  const _PanelDragHandle({
    required this.onDragStart,
    required this.onDragUpdate,
    required this.onDragEnd,
    required this.onDragCancel,
  });

  final VoidCallback onDragStart;
  final ValueChanged<double> onDragUpdate;
  final ValueChanged<double> onDragEnd;
  final VoidCallback onDragCancel;

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      key: const Key('dashboardPanelDragHandle'),
      behavior: HitTestBehavior.opaque,
      onVerticalDragStart: (_) => onDragStart(),
      onVerticalDragUpdate: (details) => onDragUpdate(details.delta.dy),
      onVerticalDragEnd: (details) => onDragEnd(details.primaryVelocity ?? 0),
      onVerticalDragCancel: onDragCancel,
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
