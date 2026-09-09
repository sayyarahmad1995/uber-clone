import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:go_router/go_router.dart';

import '../features/authentication/application/session_controller.dart';
import '../features/authentication/data/auth_repository.dart';
import '../features/authentication/presentation/login_screen.dart';
import '../features/authentication/presentation/verification_screen.dart';
import '../features/capabilities/presentation/capability_home_screen.dart';
import '../features/capabilities/presentation/splash_screen.dart';
import '../features/driver_workspace/application/driver_controller.dart';
import '../features/driver_workspace/application/driver_onboarding_controller.dart';
import '../features/driver_workspace/data/driver_onboarding_repository.dart';
import '../features/driver_workspace/data/driver_repository.dart';
import '../features/driver_workspace/presentation/driver_readonly_surfaces.dart';
import '../features/rider_request/application/rider_request_controller.dart';
import '../features/rider_request/data/device_location.dart';
import '../features/rider_request/data/ride_request_repository.dart';
import 'config/app_config.dart';
import 'dashboard/dashboard_panel_session.dart';
import 'maps/map_tiles.dart';
import 'models/account.dart';
import 'session/session_store.dart';

const _driverDetailTransitionDuration = Duration(milliseconds: 250);

final dioProvider = Provider<Dio>(
  (ref) => Dio(
    BaseOptions(
      baseUrl: AppConfig.apiBaseUrl,
      connectTimeout: const Duration(seconds: 10),
      receiveTimeout: const Duration(seconds: 10),
      headers: {'Accept': 'application/json'},
    ),
  ),
);

final sessionStoreProvider = Provider<SessionStore>(
  (ref) => SecureSessionStore(const FlutterSecureStorage()),
);
final capabilityStoreProvider = Provider<CapabilityStore>(
  (ref) => PreferencesCapabilityStore(),
);
final authRepositoryProvider = Provider<AuthRepository>(
  (ref) => ApiAuthRepository(
    ref.watch(dioProvider),
    ref.watch(sessionStoreProvider),
  ),
);
final mapTilesProvider = Provider<MapTiles>(
  (ref) => const OpenStreetMapTiles(),
);
final dashboardPanelSessionProvider =
    ChangeNotifierProvider<DashboardPanelSession>(
      (ref) => DashboardPanelSession(),
    );
final deviceLocationProvider = Provider<DeviceLocation>(
  (ref) => GeolocatorDeviceLocation(),
);
final rideRequestRepositoryProvider = Provider<RideRequestRepository>(
  (ref) => ApiRideRequestRepository(
    ref.watch(dioProvider),
    ref.watch(sessionStoreProvider),
  ),
);
final riderRequestControllerProvider =
    ChangeNotifierProvider.autoDispose<RiderRequestController>(
      (ref) => RiderRequestController(
        ref.watch(rideRequestRepositoryProvider),
        ref.watch(deviceLocationProvider),
      ),
    );
final driverRepositoryProvider = Provider<DriverRepository>(
  (ref) => ApiDriverRepository(
    ref.watch(dioProvider),
    ref.watch(sessionStoreProvider),
  ),
);
final driverOnboardingRepositoryProvider = Provider<DriverOnboardingRepository>(
  (ref) => ApiDriverOnboardingRepository(
    ref.watch(dioProvider),
    ref.watch(sessionStoreProvider),
  ),
);
final driverControllerProvider =
    ChangeNotifierProvider.autoDispose<DriverController>(
      (ref) => DriverController(
        ref.watch(driverRepositoryProvider),
        ref.watch(deviceLocationProvider),
      ),
    );
final driverOnboardingControllerProvider =
    ChangeNotifierProvider.autoDispose<DriverOnboardingController>(
      (ref) => DriverOnboardingController(
        ref.watch(driverOnboardingRepositoryProvider),
      ),
    );
final sessionControllerProvider = ChangeNotifierProvider<SessionController>(
  (ref) => SessionController(
    ref.watch(authRepositoryProvider),
    ref.watch(capabilityStoreProvider),
  ),
);

CustomTransitionPage<void> _driverDetailPage({
  required LocalKey key,
  required Widget child,
}) {
  return CustomTransitionPage<void>(
    key: key,
    opaque: false,
    transitionDuration: _driverDetailTransitionDuration,
    reverseTransitionDuration: _driverDetailTransitionDuration,
    transitionsBuilder: (_, animation, _, child) {
      final position =
          Tween<Offset>(begin: const Offset(-1, 0), end: Offset.zero).animate(
            CurvedAnimation(
              parent: animation,
              curve: Curves.easeOutCubic,
              reverseCurve: Curves.easeInCubic,
            ),
          );
      return SlideTransition(position: position, child: child);
    },
    child: child,
  );
}

final routerProvider = Provider<GoRouter>((ref) {
  final session = ref.watch(sessionControllerProvider.notifier);
  return GoRouter(
    initialLocation: '/splash',
    refreshListenable: session,
    redirect: (context, state) {
      final status = session.state.status;
      final location = state.matchedLocation;
      if (status == SessionStatus.bootstrapping) {
        return location == '/splash' ? null : '/splash';
      }
      if (status == SessionStatus.signedOut) {
        return location == '/login' || location == '/verify' ? null : '/login';
      }
      final hasDriver =
          session.state.account?.capabilities.contains(Capability.driver) ??
          false;
      if (location == '/splash' ||
          location == '/login' ||
          location == '/verify') {
        return session.state.capability == Capability.driver && hasDriver
            ? '/driver'
            : '/rider';
      }
      if ((location == '/driver' || location.startsWith('/driver/')) &&
          !hasDriver) {
        return '/rider';
      }
      return null;
    },
    routes: [
      GoRoute(path: '/splash', builder: (_, _) => const SplashScreen()),
      GoRoute(path: '/login', builder: (_, _) => const LoginScreen()),
      GoRoute(
        path: '/verify',
        redirect: (_, state) => state.extra is String ? null : '/login',
        builder: (_, state) =>
            VerificationScreen(verificationId: state.extra! as String),
      ),
      GoRoute(
        path: '/rider',
        builder: (_, _) =>
            const CapabilityHomeScreen(capability: Capability.rider),
      ),
      GoRoute(
        path: '/driver',
        builder: (_, _) =>
            const CapabilityHomeScreen(capability: Capability.driver),
      ),
      GoRoute(
        path: '/driver/details',
        pageBuilder: (_, state) => _driverDetailPage(
          key: state.pageKey,
          child: const DriverDetailsScreen(),
        ),
      ),
      GoRoute(
        path: '/driver/vehicles',
        pageBuilder: (_, state) => _driverDetailPage(
          key: state.pageKey,
          child: const DriverVehiclesScreen(),
        ),
      ),
    ],
  );
});
