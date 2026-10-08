import 'package:flutter/foundation.dart';

import '../domain/route_preview.dart';

/// Proposal ownership belongs to one account and one booking selection.
class RiderFareController extends ChangeNotifier {
  String? selectionKey;
  String text = '';
  bool userEdited = false;
  String? version;
  bool _disposed = false;

  void invalidate(String key) {
    if (_disposed || selectionKey == key) return;
    selectionKey = key;
    text = '';
    userEdited = false;
    version = null;
    notifyListeners();
  }

  void applyPreview(String key, RoutePreview preview) {
    if (_disposed || key != selectionKey) return;
    version = preview.pricingPolicyVersion;
    final fare = preview.suggestedFare;
    if (!userEdited && fare != null) {
      text =
          '${fare.amountMinor ~/ 100}.${(fare.amountMinor % 100).toString().padLeft(2, '0')}';
    }
    notifyListeners();
  }

  void edit(String value) {
    if (_disposed) return;
    text = value;
    userEdited = true;
    notifyListeners();
  }

  @override
  void dispose() {
    _disposed = true;
    super.dispose();
  }
}
