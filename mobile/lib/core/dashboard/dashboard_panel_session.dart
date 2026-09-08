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

  static DashboardPanelSession of(BuildContext context) {
    final scope = context
        .dependOnInheritedWidgetOfExactType<DashboardPanelSessionScope>();
    assert(scope != null, 'DashboardPanelSessionScope is required.');
    return scope!.notifier!;
  }
}
