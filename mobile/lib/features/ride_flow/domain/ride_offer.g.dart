// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'ride_offer.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_RideOffer _$RideOfferFromJson(Map<String, dynamic> json) => _RideOffer(
  rideRequestId: json['ride_request_id'] as String,
  driverUserId: json['driver_user_id'] as String,
  fare: RideFare.fromJson(json['fare'] as Map<String, dynamic>),
  status: json['status'] as String,
  createdAt: DateTime.parse(json['created_at'] as String),
  updatedAt: DateTime.parse(json['updated_at'] as String),
  expiresAt: DateTime.parse(json['expires_at'] as String),
  decidedAt: json['decided_at'] == null
      ? null
      : DateTime.parse(json['decided_at'] as String),
);

Map<String, dynamic> _$RideOfferToJson(_RideOffer instance) =>
    <String, dynamic>{
      'ride_request_id': instance.rideRequestId,
      'driver_user_id': instance.driverUserId,
      'fare': instance.fare,
      'status': instance.status,
      'created_at': instance.createdAt.toIso8601String(),
      'updated_at': instance.updatedAt.toIso8601String(),
      'expires_at': instance.expiresAt.toIso8601String(),
      'decided_at': instance.decidedAt?.toIso8601String(),
    };

_DriverSummary _$DriverSummaryFromJson(Map<String, dynamic> json) =>
    _DriverSummary(displayName: json['display_name'] as String);

Map<String, dynamic> _$DriverSummaryToJson(_DriverSummary instance) =>
    <String, dynamic>{'display_name': instance.displayName};

_VehicleSummary _$VehicleSummaryFromJson(Map<String, dynamic> json) =>
    _VehicleSummary(
      make: json['make'] as String,
      model: json['model'] as String,
      modelYear: (json['model_year'] as num?)?.toInt(),
      color: json['color'] as String,
    );

Map<String, dynamic> _$VehicleSummaryToJson(_VehicleSummary instance) =>
    <String, dynamic>{
      'make': instance.make,
      'model': instance.model,
      'model_year': instance.modelYear,
      'color': instance.color,
    };

_ServiceSummary _$ServiceSummaryFromJson(Map<String, dynamic> json) =>
    _ServiceSummary(
      code: json['code'] as String,
      displayName: json['display_name'] as String,
    );

Map<String, dynamic> _$ServiceSummaryToJson(_ServiceSummary instance) =>
    <String, dynamic>{
      'code': instance.code,
      'display_name': instance.displayName,
    };

_RiderOfferComparison _$RiderOfferComparisonFromJson(
  Map<String, dynamic> json,
) => _RiderOfferComparison(
  rideRequestId: json['ride_request_id'] as String,
  driverUserId: json['driver_user_id'] as String,
  fare: RideFare.fromJson(json['fare'] as Map<String, dynamic>),
  status: json['status'] as String,
  createdAt: DateTime.parse(json['created_at'] as String),
  updatedAt: DateTime.parse(json['updated_at'] as String),
  expiresAt: DateTime.parse(json['expires_at'] as String),
  decidedAt: json['decided_at'] == null
      ? null
      : DateTime.parse(json['decided_at'] as String),
  pickupDistanceMeters: (json['pickup_distance_meters'] as num?)?.toInt(),
  matchesProposedFare: json['matches_proposed_fare'] as bool,
  selectable: json['selectable'] as bool,
  driver: json['driver'] == null
      ? null
      : DriverSummary.fromJson(json['driver'] as Map<String, dynamic>),
  vehicle: json['vehicle'] == null
      ? null
      : VehicleSummary.fromJson(json['vehicle'] as Map<String, dynamic>),
  service: json['service'] == null
      ? null
      : ServiceSummary.fromJson(json['service'] as Map<String, dynamic>),
);

Map<String, dynamic> _$RiderOfferComparisonToJson(
  _RiderOfferComparison instance,
) => <String, dynamic>{
  'ride_request_id': instance.rideRequestId,
  'driver_user_id': instance.driverUserId,
  'fare': instance.fare,
  'status': instance.status,
  'created_at': instance.createdAt.toIso8601String(),
  'updated_at': instance.updatedAt.toIso8601String(),
  'expires_at': instance.expiresAt.toIso8601String(),
  'decided_at': instance.decidedAt?.toIso8601String(),
  'pickup_distance_meters': instance.pickupDistanceMeters,
  'matches_proposed_fare': instance.matchesProposedFare,
  'selectable': instance.selectable,
  'driver': instance.driver,
  'vehicle': instance.vehicle,
  'service': instance.service,
};
