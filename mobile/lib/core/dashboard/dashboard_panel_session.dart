import 'package:flutter/foundation.dart';
import 'package:flutter/widgets.dart';

class DashboardPanelSession extends ChangeNotifier {
  bool _expanded = false;

  bool get expanded => _expanded;

  void setExpanded(bool expanded) {
    if (_expanded == expanded) return;
    _expanded = expanded;
    notifyListeners();
  }

  void reset() => setExpanded(false);
}

class DashboardPanelSessionScope extends InheritedNotifier<DashboardPanelSession> {
  const DashboardPanelSessionScope({
    super.key,
    required DashboardPanelSession session,
    required super.child,
  }) : super(notifier: session);

  static DashboardPanelSession? maybeOf(BuildContext context) => context
      .dependOnInheritedWidgetOfExactType<DashboardPanelSessionScope>()
      ?.notifier;

  static DashboardPanelSession of(BuildContext context) {
    final session = maybeOf(context);
    assert(session != null, 'DashboardPanelSessionScope is required.');
    return session!;
  }
}
