import 'package:uber_clone/core/network/api_exception.dart';
import 'package:uber_clone/features/ride_flow/ride_flow_repository.dart';
import 'package:uber_clone/features/driver_workspace/domain/operating_state.dart';
import 'package:uber_clone/features/driver_workspace/domain/registered_vehicle.dart';
import 'package:uber_clone/core/models/account.dart';
import 'package:uber_clone/core/session/session_store.dart';
import 'package:uber_clone/features/authentication/data/auth_repository.dart';
import 'package:uber_clone/features/driver_workspace/data/driver_onboarding_repository.dart';
import 'package:uber_clone/features/driver_workspace/data/driver_repository.dart';
import 'package:uber_clone/features/driver_workspace/domain/driver_onboarding.dart';
import 'package:uber_clone/features/driver_workspace/domain/driver_profile.dart';
import 'package:uber_clone/features/rider_request/data/device_location.dart';
import 'package:uber_clone/features/rider_request/data/ride_request_repository.dart';
import 'package:uber_clone/features/rider_request/domain/ride_request.dart';

const riderAccount = Account(id: 'user-1', capabilities: [Capability.rider]);
const bothCapabilities = Account(
  id: 'user-1',
  capabilities: [Capability.rider, Capability.driver],
);

class FakeAuthRepository implements AuthRepository {
  FakeAuthRepository({this.account, this.loginError});
  final Account? account;
  final Object? loginError;
  @override
  Future<Account?> restore() async => account;
  @override
  Future<Account> login(String identifier, String password) async {
    if (loginError != null) throw loginError!;
    return account!;
  }
  @override
  Future<String> register(String identifier, String password) async =>
      'challenge';
  @override
  Future<String> startVerification(String email) async => 'challenge';
  @override
  Future<void> completeVerification(String verificationId, String code) async {}
  @override
  Future<void> logout() async {}
  @override
  Future<Account> enableDriver() async => bothCapabilities;
}

class MemoryCapabilityStore implements CapabilityStore {
  MemoryCapabilityStore([this.value]);
  Capability? value;
  @override
  Future<Capability?> read() async => value;
  @override
  Future<void> save(Capability capability) async => value = capability;
  @override
  Future<void> clear() async => value = null;
}

final requestedRide = RideRequest(
  id: 'ride-1',
  pickup: const GeoPoint(latitude: 24.8607, longitude: 67.0011),
  destination: const GeoPoint(latitude: 24.9056, longitude: 67.0822),
  proposedFare: const Money(amountMinor: 70000, currency: 'PKR'),
  status: 'requested',
  createdAt: DateTime.utc(2026, 9, 4),
);

class FakeRideRequestRepository implements RideRequestRepository {
  FakeRideRequestRepository({List<RideRequest>? requests})
    : requests = requests ?? [];
  List<RideRequest> requests;
  GeoPoint? submittedPickup;
  GeoPoint? submittedDestination;
  Money? submittedFare;

  @override
  Future<List<RideRequest>> list() async => requests;

  @override
  Future<RideRequest> create({
    required GeoPoint pickup,
    required GeoPoint destination,
    required Money proposedFare,
    String serviceCode = 'economy',
  }) async {
    submittedPickup = pickup;
    submittedDestination = destination;
    submittedFare = proposedFare;
    requests = [requestedRide];
    return requestedRide;
  }

  @override
  Future<RideRequest> get(String id) async => requests.first;

  @override
  Future<void> cancel(String id) async {
    requests = requests
        .map(
          (request) => request.id == id
              ? request.copyWith(status: 'cancelled')
              : request,
        )
        .toList();
  }
}

class FakeDeviceLocation implements DeviceLocation {
  const FakeDeviceLocation([
    this.point = const GeoPoint(latitude: 24.86, longitude: 67.01),
  ]);
  final GeoPoint point;
  @override
  Future<GeoPoint> current() async => point;
}

const driverVehicle = DriverVehicle(
  make: 'Toyota',
  model: 'Corolla',
  modelYear: 2024,
  color: 'White',
  licensePlate: 'ABC-123',
);
const driverProfile = DriverProfile(
  userId: 'user-1',
  displayName: 'Test Driver',
  status: 'active',
  isOnline: false,
  vehicle: driverVehicle,
);

const economyService = DriverServiceOption(
  code: 'economy',
  displayName: 'Economy',
  description: 'Vehicle eligibility is confirmed during review.',
);
const comfortService = DriverServiceOption(
  code: 'comfort',
  displayName: 'Comfort',
  description:
      'Vehicle condition and service eligibility are confirmed during review.',
  impliedServiceCode: 'economy',
);

class FakeDriverRepository implements DriverRepository {
  FakeDriverRepository({this.profile, this.vehicles = const []});
  List<RegisteredVehicle> vehicles;
  @override
  Future<List<RegisteredVehicle>> listVehicles() async => vehicles;
  OperatingState operation = const OperatingState(vehicleId:'test-vehicle', serviceCode:'economy', valid:true);
  @override
  Future<OperatingState> operatingState() async => operation;
  @override
  Future<OperatingState> selectOperation(String vehicleId, String serviceCode) async => operation = OperatingState(vehicleId:vehicleId,serviceCode:serviceCode,valid:true);
  DriverProfile? profile;
  final calls = <String>[];
  bool failPublish = false;
  bool failAvailability = false;
  @override
  Future<DriverProfile?> get() async => profile;
  @override
  Future<DriverProfile> onboard(String name, DriverVehicle vehicle) async {
    calls.add('onboard');
    return profile = driverProfile.copyWith(
      displayName: name,
      vehicle: vehicle,
    );
  }

  @override
  Future<DriverProfile> setOnline(bool online) async {
    calls.add('online=$online');
    if (failAvailability) throw Exception('Availability failed');
    return profile = profile!.copyWith(isOnline: online);
  }

  @override
  Future<PublishedDriverLocation> publishLocation(GeoPoint point) async {
    calls.add('location');
    if (failPublish) throw Exception('Location publish failed');
    return PublishedDriverLocation(
      latitude: point.latitude,
      longitude: point.longitude,
      updatedAt: DateTime.utc(2026, 9, 5),
    );
  }
}

class FakeDriverOnboardingRepository implements DriverOnboardingRepository {
  FakeDriverOnboardingRepository({
    List<DriverServiceOption>? services,
    this.application,
  }) : services = services ?? const [economyService, comfortService];

  final List<DriverServiceOption> services;
  DriverOnboardingApplication? application;
  final calls = <String>[];
  bool precheckEligible = true;
  List<String> precheckReasons = const [];

  @override
  Future<List<DriverServiceOption>> listServices() async => services;

  @override
  Future<DriverOnboardingApplication?> getLatest() async => application;

  @override
  Future<DriverOnboardingPrecheck> precheck({
    required String displayName,
    required String serviceCode,
    required DriverVehicle vehicle,
  }) async {
    calls.add('precheck:$serviceCode');
    final service = services.firstWhere((item) => item.code == serviceCode);
    return DriverOnboardingPrecheck(
      eligible: precheckEligible,
      reasons: precheckReasons,
      service: service,
    );
  }

  @override
  Future<DriverOnboardingApplication> submit({
    required String displayName,
    required String serviceCode,
    required DriverVehicle vehicle,
  }) async {
    calls.add('submit:$serviceCode');
    final service = services.firstWhere((item) => item.code == serviceCode);
    return application = DriverOnboardingApplication(
      id: 'application-1',
      displayName: displayName,
      status: 'pending',
      service: service,
      vehicle: vehicle,
      submittedAt: DateTime.utc(2026, 9, 7),
    );
  }
}

// Existing dashboard tests isolate networking; flow behavior has dedicated tests.
class FakeRideFlowRepository implements RideFlowRepository {
  @override
  Future<Json> get(String path) async {
    if (path == '/v1/driver/trip') throw const ApiException('not_found','No active trip',statusCode:404);
    if (path.endsWith('/offers')) return {'offers': <Json>[]};
    if (path.endsWith('/trips')) return {'trips': <Json>[]};
    if (path.contains('/marketplace/')) return {'ride_requests': <Json>[]};
    return {'id': path.split('/').last, 'status':'requested', 'trip':null};
  }
  @override
  Future<void> act(String path, {Json? data, bool put=false}) async {}
}
