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

class DashboardPanelSessionScope extends InheritedWidget {
  const DashboardPanelSessionScope({
    super.key,
    required this.session,
    required super.child,
  });

  final DashboardPanelSession session;

  static DashboardPanelSession? maybeOf(BuildContext context) => context
      .dependOnInheritedWidgetOfExactType<DashboardPanelSessionScope>()
      ?.session;

  static DashboardPanelSession of(BuildContext context) {
    final session = maybeOf(context);
    assert(session != null, 'DashboardPanelSessionScope is required.');
    return session!;
  }

  @override
  bool updateShouldNotify(DashboardPanelSessionScope oldWidget) =>
      !identical(session, oldWidget.session);
}
