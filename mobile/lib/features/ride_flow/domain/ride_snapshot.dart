import 'package:freezed_annotation/freezed_annotation.dart';

import 'marketplace_request.dart';
import 'ride_offer.dart';
import 'trip.dart';

part 'ride_snapshot.freezed.dart';
part 'ride_snapshot.g.dart';

@freezed
abstract class RiderRideSnapshot with _$RiderRideSnapshot {
  const factory RiderRideSnapshot({
    @JsonKey(name: 'service_code') required String serviceCode,
    required String id,
    required RideLocation pickup,
    required RideLocation destination,
    @JsonKey(name: 'proposed_fare') required RideFare proposedFare,
    required String status,
    @JsonKey(name: 'created_at') required DateTime createdAt,
    @JsonKey(name: 'expires_at') required DateTime expiresAt,
    TripSnapshot? trip,
  }) = _RiderRideSnapshot;

  factory RiderRideSnapshot.fromJson(Map<String, dynamic> json) => _$RiderRideSnapshotFromJson(json);
}

@freezed
abstract class DriverLocationSnapshot with _$DriverLocationSnapshot {
  const factory DriverLocationSnapshot({
    @JsonKey(name: 'ride_request_id') required String rideRequestId,
    required double latitude,
    required double longitude,
    @JsonKey(name: 'updated_at') required DateTime updatedAt,
  }) = _DriverLocationSnapshot;

  factory DriverLocationSnapshot.fromJson(Map<String, dynamic> json) => _$DriverLocationSnapshotFromJson(json);
}

class RideSnapshot {
  const RideSnapshot({required this.ride, required this.offers, this.driverLocation});

  final RiderRideSnapshot ride;
  final List<RiderOfferComparison> offers;
  final DriverLocationSnapshot? driverLocation;
}
