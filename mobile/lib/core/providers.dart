import '../features/driver_workspace/domain/registered_vehicle.dart';

import 'package:dio/dio.dart';

import '../features/ride_flow/ride_flow_repository.dart';
import '../features/ride_flow/application/driver_marketplace_controller.dart';
import '../features/ride_flow/application/driver_trip_controller.dart';
import '../features/ride_flow/application/rider_active_ride_controller.dart';

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
import '../features/driver_workspace/application/driver_route_controller.dart';
import '../features/driver_workspace/data/driver_route_repository.dart';
import '../features/driver_workspace/application/driver_presence_service.dart';
import '../features/driver_workspace/data/android_driver_presence_service.dart';
import '../features/driver_workspace/application/driver_onboarding_controller.dart';
import '../features/driver_workspace/data/driver_onboarding_repository.dart';
import '../features/driver_workspace/data/driver_repository.dart';
import '../features/driver_workspace/domain/driver_onboarding.dart';
import '../features/driver_workspace/domain/driver_profile.dart';
import '../features/driver_workspace/presentation/driver_readonly_surfaces.dart';
import '../features/rider_request/application/rider_place_search_controller.dart';
import '../features/rider_request/application/rider_route_preview_controller.dart';
import '../features/rider_request/application/rider_request_controller.dart';
import '../features/rider_request/application/rider_fare_controller.dart';
import '../features/rider_request/data/device_location.dart';
import '../features/rider_request/data/place_search_repository.dart';
import '../features/rider_request/data/ride_request_repository.dart';
import '../features/rider_request/data/route_preview_repository.dart';
import '../features/rider_request/data/ride_service_repository.dart';
import '../features/rider_request/domain/ride_service.dart';
import 'config/app_config.dart';
import 'dashboard/dashboard_panel_session.dart';
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
final riderServiceRepositoryProvider = Provider<RideServiceRepository>(
  (ref) => ApiRideServiceRepository(
    ref.watch(dioProvider),
    ref.watch(sessionStoreProvider),
  ),
);
final riderServicesProvider = FutureProvider.autoDispose<List<RideService>>((
  ref,
) {
  // A catalog from an old account must never survive an account switch.
  ref.watch(sessionControllerProvider.select((s) => s.state.account?.id));
  return ref.watch(riderServiceRepositoryProvider).list();
});
final rideFlowRepositoryProvider = Provider<RideFlowRepository>(
  (ref) => ApiRideFlowRepository(
    ref.watch(dioProvider),
    ref.watch(sessionStoreProvider),
  ),
);
final riderActiveRideControllerProvider = ChangeNotifierProvider.autoDispose
    .family<RiderActiveRideController, String>((ref, rideId) {
      ref.watch(
        sessionControllerProvider.select((state) => state.state.account?.id),
      );
      return RiderActiveRideController(
        ref.watch(rideFlowRepositoryProvider),
        rideId: rideId,
      );
    });

final driverMarketplaceControllerProvider =
    ChangeNotifierProvider.autoDispose<DriverMarketplaceController>((ref) {
      ref.watch(
        sessionControllerProvider.select((state) => state.state.account?.id),
      );
      return DriverMarketplaceController(ref.watch(rideFlowRepositoryProvider));
    });

final driverTripControllerProvider =
    ChangeNotifierProvider.autoDispose<DriverTripController>((ref) {
      ref.watch(
        sessionControllerProvider.select((state) => state.state.account?.id),
      );
      return DriverTripController(ref.watch(rideFlowRepositoryProvider));
    });
final riderRequestControllerProvider =
    ChangeNotifierProvider.autoDispose<RiderRequestController>((ref) {
      ref.watch(sessionControllerProvider.select((s) => s.state.account?.id));
      return RiderRequestController(
        ref.watch(rideRequestRepositoryProvider),
        ref.watch(deviceLocationProvider),
      );
    });
final placeSearchRepositoryProvider = Provider<PlaceSearchRepository>(
  (ref) => ApiPlaceSearchRepository(
    ref.watch(dioProvider),
    ref.watch(sessionStoreProvider),
  ),
);
final riderPlaceSearchControllerProvider =
    ChangeNotifierProvider.autoDispose<RiderPlaceSearchController>((ref) {
      ref.watch(sessionControllerProvider.select((s) => s.state.account?.id));
      return RiderPlaceSearchController(
        ref.watch(placeSearchRepositoryProvider),
        ref.watch(riderRequestControllerProvider.notifier),
        ref.watch(deviceLocationProvider),
      );
    });
final routePreviewRepositoryProvider = Provider<RoutePreviewRepository>(
  (ref) => ApiRoutePreviewRepository(
    ref.watch(dioProvider),
    ref.watch(sessionStoreProvider),
  ),
);
final riderRoutePreviewControllerProvider =
    ChangeNotifierProvider.autoDispose<RiderRoutePreviewController>((ref) {
      ref.watch(sessionControllerProvider.select((s) => s.state.account?.id));
      return RiderRoutePreviewController(
        ref.watch(routePreviewRepositoryProvider),
        ref.watch(riderRequestControllerProvider),
      );
    });
final riderFareControllerProvider =
    ChangeNotifierProvider.autoDispose<RiderFareController>((ref) {
  ref.watch(sessionControllerProvider.select((s) => s.state.account?.id));
  final rider = ref.watch(riderRequestControllerProvider.notifier);
  final route = ref.watch(riderRoutePreviewControllerProvider.notifier);
  final fare = RiderFareController();
  void sync() {
    if (rider.state.active != null) return;
    final key = route.selectionKey ?? '';
    fare.invalidate(key);
    final preview = route.state.preview;
    if (preview != null) fare.applyPreview(key, preview);
  }
  sync();
  rider.addListener(sync);
  route.addListener(sync);
  ref.listen(riderServicesProvider, (_, next) {
    final services = next.asData?.value;
    if (services != null && rider.state.active == null &&
        !services.any((s) => s.code == rider.serviceCode)) {
      rider.selectService('');
    }
  });
  ref.onDispose(() {
    rider.removeListener(sync);
    route.removeListener(sync);
  });
  return fare;
});
final driverRouteRepositoryProvider =
    Provider.autoDispose<DriverRouteRepository>((ref) {
      final repository = ApiDriverRouteRepository(
        ref.watch(dioProvider),
        ref.watch(sessionStoreProvider),
      );
      ref.onDispose(repository.cancel);
      return repository;
    });
final driverRouteControllerProvider =
    ChangeNotifierProvider.autoDispose<DriverRouteController>((ref) {
      ref.watch(sessionControllerProvider.select((s) => s.state.account?.id));
      return DriverRouteController(
        ref.watch(driverRouteRepositoryProvider),
        ref.watch(driverTripControllerProvider.notifier),
        ref.watch(deviceLocationProvider),
      );
    });
final driverRepositoryProvider = Provider<DriverRepository>(
  (ref) => ApiDriverRepository(
    ref.watch(dioProvider),
    ref.watch(sessionStoreProvider),
  ),
);
final driverPresenceServiceProvider = Provider<DriverPresenceService>(
  (ref) => AndroidDriverPresenceService(),
);
final approvedDriverProfileProvider =
    FutureProvider.autoDispose<DriverProfile?>((ref) {
      final account = ref.watch(
        sessionControllerProvider.select((value) => value.state.account),
      );
      if (account == null ||
          !account.capabilities.contains(Capability.driver)) {
        return Future.value(null);
      }
      return ref.watch(driverRepositoryProvider).get();
    });
final driverVehiclesProvider =
    FutureProvider.autoDispose<List<RegisteredVehicle>>((ref) {
      // Reload on account changes so records cannot survive an account switch.
      final account = ref.watch(
        sessionControllerProvider.select((value) => value.state.account),
      );
      if (account == null) return Future.value(<RegisteredVehicle>[]);
      return ref.watch(driverRepositoryProvider).listVehicles();
    });
final driverOnboardingRepositoryProvider = Provider<DriverOnboardingRepository>(
  (ref) => ApiDriverOnboardingRepository(
    ref.watch(dioProvider),
    ref.watch(sessionStoreProvider),
  ),
);
final latestDriverOnboardingApplicationProvider =
    FutureProvider.autoDispose<DriverOnboardingApplication?>((ref) {
      final account = ref.watch(
        sessionControllerProvider.select((value) => value.state.account),
      );
      if (account == null ||
          !account.capabilities.contains(Capability.driver)) {
        return Future.value(null);
      }
      return ref.watch(driverOnboardingRepositoryProvider).getLatest();
    });
final driverControllerProvider =
    ChangeNotifierProvider.autoDispose<DriverController>((ref) {
      ref.watch(
        sessionControllerProvider.select((value) => value.state.account?.id),
      );
      return DriverController(
        ref.watch(driverRepositoryProvider),
        ref.watch(deviceLocationProvider),
        ref.watch(driverPresenceServiceProvider),
      );
    });
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
    ref.watch(driverPresenceServiceProvider),
    ref.watch(driverRepositoryProvider),
    ref.watch(rideFlowRepositoryProvider),
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
    redirect: (context, state) async {
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
      if (location == '/rider' &&
          session.state.capability == Capability.driver) {
        return '/driver';
      }
      if (location == '/driver/details' || location == '/driver/vehicles') {
        try {
          final profile = await ref.read(approvedDriverProfileProvider.future);
          if (profile == null) return '/driver';
        } catch (_) {
          return '/driver';
        }
      }
      return null;
    },
    routes: [
      GoRoute(path: '/splash', builder: (_, _) => const SplashScreen()),
      GoRoute(path: '/login', builder: (_, _) => const LoginScreen()),
      GoRoute(
        path: '/verify',
        redirect: (_, state) {
          final query = state.uri.queryParameters;
          return query['email'] != null && query['verification_id'] != null
              ? null
              : '/login';
        },
        builder: (_, state) {
          final query = state.uri.queryParameters;
          return VerificationScreen(
            request: VerificationRequest(
              email: query['email']!,
              verificationId: query['verification_id']!,
            ),
          );
        },
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
