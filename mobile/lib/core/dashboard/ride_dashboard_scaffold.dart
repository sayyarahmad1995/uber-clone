import 'package:flutter/gestures.dart';
import 'package:flutter/material.dart';

import '../theme/app_theme.dart';
import 'dashboard_panel_session.dart';

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
    required this.panelIdentity,
    this.floatingStatus,
    this.mapControls,
    this.minPanelSize = 0.16,
    this.initialPanelSize = 0.16,
    this.maxPanelSize = 0.60,
    this.onExpansionChanged,
  }) : assert(minPanelSize > 0),
       assert(minPanelSize <= initialPanelSize),
       assert(initialPanelSize <= maxPanelSize),
       assert(maxPanelSize <= 0.60);

  final Widget map;
  final DashboardPanelBuilder panelBuilder;
  final Object panelIdentity;
  final Widget? floatingStatus;
  final Widget? mapControls;
  final double minPanelSize;
  final double initialPanelSize;
  final double maxPanelSize;
  final ValueChanged<bool>? onExpansionChanged;

  @override
  State<RideDashboardScaffold> createState() => _RideDashboardScaffoldState();
}

class _RideDashboardScaffoldState extends State<RideDashboardScaffold>
    with SingleTickerProviderStateMixin {
  final _contentScrollController = ScrollController();
  late final AnimationController _snapController;
  late final Animation<double> _snapCurve;
  late double _panelSize;
  late double _dragStartSize;
  bool _isDraggingPanel = false;
  double? _snapFromSize;
  double? _snapToSize;
  late bool _committedExpanded;
  bool _expansionSettled = false;
  DashboardPanelSession? _panelSession;
  int? _contentPointer;
  int? _handlePointer;
  final Set<int> _controlPointers = {};
  bool _contentDragStartedAtTop = false;
  bool _contentDragStartedCollapsed = false;
  double _collapsePullDistance = 0;
  double _expandPullDistance = 0;
  Widget? _cachedPanelContent;
  bool? _cachedPanelScrollEnabled;

  static const _collapsePullThreshold = 56.0;
  static const _snapDuration = Duration(milliseconds: 220);

  bool get _isSnappingPanel => _snapFromSize != null && _snapToSize != null;

  @override
  void initState() {
    super.initState();
    _snapController = AnimationController(vsync: this, duration: _snapDuration)
      ..addStatusListener((status) {
        if (status == AnimationStatus.completed) {
          _finishPanelSnap();
        }
      });
    _snapCurve = CurvedAnimation(
      parent: _snapController,
      curve: Curves.easeOutCubic,
    );
    _panelSize = widget.initialPanelSize;
    _dragStartSize = _panelSize;
    _committedExpanded = _isPanelExpanded;
    _expansionSettled = _committedExpanded;
  }

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    final session = DashboardPanelSessionScope.maybeOf(context);
    if (!identical(_panelSession, session)) {
      _panelSession = session;
      if (session != null) {
        _applySessionExtent(session.expanded);
      }
      return;
    }
    if (session != null &&
        session.expanded != _committedExpanded &&
        !_isDraggingPanel &&
        !_isSnappingPanel) {
      _applySessionExtent(session.expanded);
    }
  }

  @override
  void didUpdateWidget(covariant RideDashboardScaffold oldWidget) {
    super.didUpdateWidget(oldWidget);
    _invalidatePanelContent();
    if (oldWidget.minPanelSize != widget.minPanelSize ||
        oldWidget.maxPanelSize != widget.maxPanelSize) {
      _clearPanelSnap();
      _panelSize = _committedExpanded
          ? widget.maxPanelSize
          : widget.minPanelSize;
    }
    if (oldWidget.panelIdentity != widget.panelIdentity) {
      _clearPanelSnap();
      _panelSize = _committedExpanded
          ? widget.maxPanelSize
          : widget.minPanelSize;
      _dragStartSize = _panelSize;
      _expansionSettled = _committedExpanded;
      _isDraggingPanel = false;
      _contentPointer = null;
      _handlePointer = null;
      _controlPointers.clear();
      _resetContentDrag();
      if (_contentScrollController.hasClients) {
        _contentScrollController.jumpTo(
          _contentScrollController.position.minScrollExtent,
        );
      }
    }
  }

  @override
  void dispose() {
    _snapController.dispose();
    _contentScrollController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: (context, constraints) {
        final dashboardHeight = constraints.maxHeight;
        final panelHeight = dashboardHeight * _panelSize;
        final layoutPanelHeight = dashboardHeight * _panelLayoutSize;
        final transitionDuration = _isDraggingPanel
            ? Duration.zero
            : _snapDuration;

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
                bottom: AppSpacing.lg,
                child: TweenAnimationBuilder<double>(
                  tween: Tween<double>(end: panelHeight),
                  duration: transitionDuration,
                  curve: Curves.easeOutCubic,
                  builder: (context, panelOffset, child) => Transform.translate(
                    offset: Offset(0, -panelOffset),
                    child: child,
                  ),
                  child: SafeArea(
                    top: false,
                    child: RepaintBoundary(child: widget.mapControls!),
                  ),
                ),
              ),
            Positioned(
              left: 0,
              right: 0,
              bottom: 0,
              height: layoutPanelHeight,
              child: AnimatedBuilder(
                animation: _snapController,
                builder: (context, child) => Transform.translate(
                  offset: Offset(0, _panelSnapOffset(dashboardHeight)),
                  child: child,
                ),
                child: _buildPanelSurface(context, dashboardHeight),
              ),
            ),
          ],
        );
      },
    );
  }

  double get _panelLayoutSize {
    if (!_isSnappingPanel) {
      return _panelSize;
    }
    return _snapFromSize! > _snapToSize! ? _snapFromSize! : _snapToSize!;
  }

  double _panelSnapOffset(double dashboardHeight) {
    if (!_isSnappingPanel) {
      return 0;
    }
    final fromSize = _snapFromSize!;
    final toSize = _snapToSize!;
    final progress = _snapCurve.value;
    if (toSize > fromSize) {
      return dashboardHeight * (toSize - fromSize) * (1 - progress);
    }
    return dashboardHeight * (fromSize - toSize) * progress;
  }

  Widget _buildPanelSurface(BuildContext context, double dashboardHeight) {
    return SizedBox.expand(
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
                      onPointerDown: (pointer) {
                        if (_contentPointer != null ||
                            _handlePointer != null ||
                            _isSnappingPanel) {
                          return false;
                        }
                        _handlePointer = pointer;
                        return true;
                      },
                      onPointerFinished: (pointer) {
                        if (_handlePointer == pointer) {
                          _handlePointer = null;
                        }
                      },
                      onDragStart: _startPanelDrag,
                      onDragUpdate: (delta) =>
                          _resizePanel(delta, dashboardHeight),
                      onDragEnd: _endPanelDrag,
                      onDragCancel: _cancelPanelDrag,
                    ),
                    Expanded(
                      child: Listener(
                        onPointerDown: (event) {
                          if (_contentPointer != null ||
                              _handlePointer != null ||
                              _isDraggingPanel ||
                              _isSnappingPanel ||
                              _controlPointers.contains(event.pointer)) {
                            return;
                          }
                          _contentPointer = event.pointer;
                          _startContentDrag();
                        },
                        onPointerMove: (event) {
                          if (_contentPointer != event.pointer) {
                            return;
                          }
                          _updateContentDrag(event.delta.dy, dashboardHeight);
                        },
                        onPointerUp: (event) {
                          _controlPointers.remove(event.pointer);
                          if (_contentPointer != event.pointer) {
                            return;
                          }
                          _contentPointer = null;
                          _releaseContentDrag();
                        },
                        onPointerCancel: (event) {
                          _controlPointers.remove(event.pointer);
                          if (_contentPointer != event.pointer) {
                            return;
                          }
                          _contentPointer = null;
                          _cancelContentDrag();
                        },
                        child: _panelContentFor(
                          _committedExpanded &&
                              _expansionSettled &&
                              !_isDraggingPanel &&
                              !_isSnappingPanel,
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
    );
  }

  Widget _panelContentFor(bool scrollEnabled) {
    if (_cachedPanelContent == null ||
        _cachedPanelScrollEnabled != scrollEnabled) {
      _cachedPanelScrollEnabled = scrollEnabled;
      _cachedPanelContent = _DashboardPanelContent(
        builder: widget.panelBuilder,
        scrollController: _contentScrollController,
        scrollEnabled: scrollEnabled,
      );
    }
    return _cachedPanelContent!;
  }

  void _invalidatePanelContent() {
    _cachedPanelContent = null;
    _cachedPanelScrollEnabled = null;
  }

  void _resizePanel(double verticalDelta, double dashboardHeight) {
    if (_handlePointer == null || dashboardHeight <= 0) {
      return;
    }
    _setPreviewPanelSize(
      (_panelSize - verticalDelta / dashboardHeight).clamp(
        widget.minPanelSize,
        widget.maxPanelSize,
      ),
    );
  }

  void _startPanelDrag() {
    if (_handlePointer == null || _isSnappingPanel) {
      return;
    }
    _resetContentDrag();
    setState(() {
      _isDraggingPanel = true;
      _dragStartSize = _panelSize;
    });
  }

  void _endPanelDrag(double velocity) {
    if (_handlePointer == null) {
      return;
    }
    final movement = _panelSize - _dragStartSize;
    final expand = velocity < -50 || (velocity.abs() <= 50 && movement > 0);
    _startPanelSnap(
      expand ? widget.maxPanelSize : widget.minPanelSize,
      committedExpanded: expand,
    );
  }

  void _cancelPanelDrag() {
    if (_handlePointer == null) {
      return;
    }
    final midpoint = (widget.minPanelSize + widget.maxPanelSize) / 2;
    final targetSize = _panelSize >= midpoint
        ? widget.maxPanelSize
        : widget.minPanelSize;
    _startPanelSnap(
      targetSize,
      committedExpanded: targetSize == widget.maxPanelSize,
    );
  }

  bool get _isPanelExpanded => (_panelSize - widget.maxPanelSize).abs() < 0.001;

  void _startContentDrag() {
    _contentDragStartedCollapsed = _isPanelCollapsed;
    _contentDragStartedAtTop =
        _committedExpanded &&
        _expansionSettled &&
        _contentScrollController.hasClients &&
        _contentScrollController.position.pixels <=
            _contentScrollController.position.minScrollExtent + 0.5;
    _collapsePullDistance = 0;
    _expandPullDistance = 0;
    if (_contentDragStartedCollapsed || _contentDragStartedAtTop) {
      _dragStartSize = _panelSize;
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
      _setPreviewPanelSize(
        (widget.minPanelSize + _expandPullDistance / dashboardHeight).clamp(
          widget.minPanelSize,
          widget.maxPanelSize,
        ),
      );
      return;
    }
    if (!_contentDragStartedAtTop) {
      return;
    }
    _collapsePullDistance = (_collapsePullDistance + verticalDelta).clamp(
      0.0,
      dashboardHeight * (widget.maxPanelSize - widget.minPanelSize),
    );
    _setPreviewPanelSize(
      (widget.maxPanelSize - _collapsePullDistance / dashboardHeight).clamp(
        widget.minPanelSize,
        widget.maxPanelSize,
      ),
    );
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
    final targetSize = shouldExpand
        ? widget.maxPanelSize
        : shouldCollapse
        ? widget.minPanelSize
        : returnSize;
    final targetExpanded = shouldExpand
        ? true
        : shouldCollapse
        ? false
        : _committedExpanded;
    _startPanelSnap(targetSize, committedExpanded: targetExpanded);
  }

  void _cancelContentDrag() {
    final hadBodyPanelDrag =
        _contentDragStartedCollapsed || _contentDragStartedAtTop;
    final returnSize = _dragStartSize;
    _resetContentDrag();
    if (hadBodyPanelDrag) {
      _startPanelSnap(returnSize, committedExpanded: _committedExpanded);
    }
  }

  void _startPanelSnap(double targetSize, {required bool committedExpanded}) {
    final fromSize = _panelSize;
    final previousExpanded = _committedExpanded;
    final shouldAnimate = (fromSize - targetSize).abs() >= 0.000001;

    setState(() {
      _isDraggingPanel = false;
      _committedExpanded = committedExpanded;
      _panelSize = targetSize;
      _expansionSettled = false;
      if (shouldAnimate) {
        _snapFromSize = fromSize;
        _snapToSize = targetSize;
      } else {
        _clearPanelSnap();
        _expansionSettled = committedExpanded && _isPanelExpanded;
      }
    });
    if (shouldAnimate) {
      _snapController.forward(from: 0);
    }
    _notifyExpansionChanged(previousExpanded);
  }

  void _finishPanelSnap() {
    if (!_isSnappingPanel || !mounted) {
      return;
    }
    setState(() {
      _snapFromSize = null;
      _snapToSize = null;
      _dragStartSize = _panelSize;
      _expansionSettled = _committedExpanded && _isPanelExpanded;
    });
  }

  void _clearPanelSnap() {
    _snapController.stop();
    _snapFromSize = null;
    _snapToSize = null;
  }

  void _resetContentDrag() {
    _contentDragStartedAtTop = false;
    _contentDragStartedCollapsed = false;
    _collapsePullDistance = 0;
    _expandPullDistance = 0;
  }

  bool get _isPanelCollapsed =>
      (_panelSize - widget.minPanelSize).abs() < 0.001;

  void _setPreviewPanelSize(double nextSize) {
    if ((_panelSize - nextSize).abs() < 0.000001) {
      return;
    }
    setState(() {
      _isDraggingPanel = true;
      _panelSize = nextSize;
    });
  }

  void _applySessionExtent(bool expanded) {
    _clearPanelSnap();
    _committedExpanded = expanded;
    _panelSize = expanded ? widget.maxPanelSize : widget.minPanelSize;
    _dragStartSize = _panelSize;
    _expansionSettled = expanded;
    _isDraggingPanel = false;
    _contentPointer = null;
    _handlePointer = null;
    _controlPointers.clear();
    _resetContentDrag();
  }

  void _notifyExpansionChanged(bool previousExpanded) {
    if (previousExpanded == _committedExpanded) return;
    _panelSession?.setExpanded(_committedExpanded);
    widget.onExpansionChanged?.call(_committedExpanded);
  }
}

class _DashboardPanelContent extends StatelessWidget {
  const _DashboardPanelContent({
    required this.builder,
    required this.scrollController,
    required this.scrollEnabled,
  });

  final DashboardPanelBuilder builder;
  final ScrollController scrollController;
  final bool scrollEnabled;

  @override
  Widget build(BuildContext context) =>
      builder(context, scrollController, scrollEnabled);
}

/// Marks interactive content whose pointers must never resize the dashboard.
/// Wrap buttons, input fields and custom controls in feature panels.
class DashboardPanelControl extends StatelessWidget {
  const DashboardPanelControl({super.key, required this.child});

  final Widget child;

  @override
  Widget build(BuildContext context) => Listener(
    behavior: HitTestBehavior.translucent,
    onPointerDown: (event) => context
        .findAncestorStateOfType<_RideDashboardScaffoldState>()
        ?._controlPointers
        .add(event.pointer),
    child: child,
  );
}

class _PanelDragHandle extends StatelessWidget {
  const _PanelDragHandle({
    required this.onPointerDown,
    required this.onPointerFinished,
    required this.onDragStart,
    required this.onDragUpdate,
    required this.onDragEnd,
    required this.onDragCancel,
  });

  final bool Function(int) onPointerDown;
  final ValueChanged<int> onPointerFinished;
  final VoidCallback onDragStart;
  final ValueChanged<double> onDragUpdate;
  final ValueChanged<double> onDragEnd;
  final VoidCallback onDragCancel;

  @override
  Widget build(BuildContext context) {
    return RawGestureDetector(
      key: const Key('dashboardPanelDragHandle'),
      behavior: HitTestBehavior.opaque,
      gestures: {
        _OwnedHandleDragRecognizer:
            GestureRecognizerFactoryWithHandlers<_OwnedHandleDragRecognizer>(
              _OwnedHandleDragRecognizer.new,
              (recognizer) => recognizer
                ..acquirePointer = onPointerDown
                ..releasePointer = onPointerFinished
                ..onStart = ((_) => onDragStart())
                ..onUpdate = ((details) => onDragUpdate(details.delta.dy))
                ..onEnd = ((details) => onDragEnd(details.primaryVelocity ?? 0))
                ..onCancel = onDragCancel,
            ),
      },
      child: const SizedBox(
        height: 44,
        width: double.infinity,
        child: DashboardPanelHandle(),
      ),
    );
  }
}

/// Only admitted pointer-down events enter Flutter's normal vertical recognizer.
/// A rejected pointer is never tracked, so later moves/up/cancel cannot acquire
/// ownership or invoke any drag callback after the original owner releases.
class _OwnedHandleDragRecognizer extends VerticalDragGestureRecognizer {
  late bool Function(int) acquirePointer;
  late ValueChanged<int> releasePointer;
  int? _pointer;

  @override
  void addPointer(PointerDownEvent event) {
    if (_pointer == null &&
        isPointerAllowed(event) &&
        acquirePointer(event.pointer)) {
      _pointer = event.pointer;
      super.addPointer(event);
    }
  }

  @override
  void didStopTrackingLastPointer(int pointer) {
    // End/cancel callbacks must run while this pointer still owns the panel.
    super.didStopTrackingLastPointer(pointer);
    releasePointer(pointer);
    _pointer = null;
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
