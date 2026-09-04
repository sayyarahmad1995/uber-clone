import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'core/providers.dart';

class UberCloneApp extends ConsumerWidget {
  const UberCloneApp({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) => MaterialApp.router(
    title: 'Uber Clone',
    debugShowCheckedModeBanner: false,
    theme: ThemeData(
      colorScheme: ColorScheme.fromSeed(seedColor: const Color(0xFF111111)),
      useMaterial3: true,
    ),
    routerConfig: ref.watch(routerProvider),
  );
}
