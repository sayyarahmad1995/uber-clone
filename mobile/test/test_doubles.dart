import 'package:uber_clone/core/models/account.dart';
import 'package:uber_clone/core/session/session_store.dart';
import 'package:uber_clone/features/authentication/data/auth_repository.dart';
import 'package:uber_clone/features/rider_request/data/device_location.dart';
import 'package:uber_clone/features/rider_request/data/ride_request_repository.dart';
import 'package:uber_clone/features/rider_request/domain/ride_request.dart';

const riderAccount = Account(id: 'user-1', capabilities: [Capability.rider]);
const bothCapabilities = Account(
  id: 'user-1',
  capabilities: [Capability.rider, Capability.driver],
);

class FakeAuthRepository implements AuthRepository {
  FakeAuthRepository({this.account});
  final Account? account;
  @override
  Future<Account?> restore() async => account;
  @override
  Future<Account> login(String identifier, String password) async => account!;
  @override
  Future<String> register(String identifier, String password) async =>
      'challenge';
  @override
  Future<void> completeVerification(String verificationId, String code) async {}
  @override
  Future<void> logout() async {}
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
