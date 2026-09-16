// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'ride_snapshot.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_RiderRideSnapshot _$RiderRideSnapshotFromJson(Map<String, dynamic> json) =>
    _RiderRideSnapshot(
      serviceCode: json['service_code'] as String,
      id: json['id'] as String,
      pickup: RideLocation.fromJson(json['pickup'] as Map<String, dynamic>),
      destination: RideLocation.fromJson(
        json['destination'] as Map<String, dynamic>,
      ),
      proposedFare: RideFare.fromJson(
        json['proposed_fare'] as Map<String, dynamic>,
      ),
      status: json['status'] as String,
      createdAt: DateTime.parse(json['created_at'] as String),
      expiresAt: DateTime.parse(json['expires_at'] as String),
      trip: json['trip'] == null
          ? null
          : TripSnapshot.fromJson(json['trip'] as Map<String, dynamic>),
    );

Map<String, dynamic> _$RiderRideSnapshotToJson(_RiderRideSnapshot instance) =>
    <String, dynamic>{
      'service_code': instance.serviceCode,
      'id': instance.id,
      'pickup': instance.pickup,
      'destination': instance.destination,
      'proposed_fare': instance.proposedFare,
      'status': instance.status,
      'created_at': instance.createdAt.toIso8601String(),
      'expires_at': instance.expiresAt.toIso8601String(),
      'trip': instance.trip,
    };

_DriverLocationSnapshot _$DriverLocationSnapshotFromJson(
  Map<String, dynamic> json,
) => _DriverLocationSnapshot(
  rideRequestId: json['ride_request_id'] as String,
  latitude: (json['latitude'] as num).toDouble(),
  longitude: (json['longitude'] as num).toDouble(),
  updatedAt: DateTime.parse(json['updated_at'] as String),
);

Map<String, dynamic> _$DriverLocationSnapshotToJson(
  _DriverLocationSnapshot instance,
) => <String, dynamic>{
  'ride_request_id': instance.rideRequestId,
  'latitude': instance.latitude,
  'longitude': instance.longitude,
  'updated_at': instance.updatedAt.toIso8601String(),
};
