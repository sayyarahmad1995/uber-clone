import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:uber_clone/core/theme/app_theme.dart';

void main() {
  test('app bar and status bar keep a stable dashboard background', () {
    final appBar = AppTheme.light.appBarTheme;

    expect(appBar.backgroundColor, AppColors.background);
    expect(appBar.surfaceTintColor, Colors.transparent);
    expect(appBar.elevation, 0);
    expect(appBar.scrolledUnderElevation, 0);
    expect(
      appBar.systemOverlayStyle,
      const SystemUiOverlayStyle(
        statusBarColor: AppColors.background,
        statusBarIconBrightness: Brightness.dark,
        statusBarBrightness: Brightness.light,
        systemStatusBarContrastEnforced: false,
      ),
    );
  });
}
