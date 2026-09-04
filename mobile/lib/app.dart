import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'core/providers.dart';
import 'core/theme/app_theme.dart';

class UberCloneApp extends ConsumerWidget {
  const UberCloneApp({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) => MaterialApp.router(
    title: 'Uber Clone',
    debugShowCheckedModeBanner: false,
    theme: AppTheme.light,
    routerConfig: ref.watch(routerProvider),
  );
}
