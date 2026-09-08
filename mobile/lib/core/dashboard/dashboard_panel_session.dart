import 'package:flutter/foundation.dart';

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
