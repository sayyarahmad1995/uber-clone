import 'dart:io';

import 'package:flutter_test/flutter_test.dart';

void main() {
  test('Android declares a non-exported location foreground service', () {
    final manifest = File(
      'android/app/src/main/AndroidManifest.xml',
    ).readAsStringSync();

    expect(manifest, contains('android.permission.FOREGROUND_SERVICE'));
    expect(manifest, contains('android.permission.FOREGROUND_SERVICE_LOCATION'));
    expect(manifest, contains('android.permission.POST_NOTIFICATIONS'));
    expect(
      manifest,
      matches(
        RegExp(
          r'<service\s+[^>]*android:name="com\.pravera\.flutter_foreground_task\.service\.ForegroundService"[^>]*android:foregroundServiceType="location"[^>]*android:exported="false"',
          dotAll: true,
        ),
      ),
    );
  });
}
