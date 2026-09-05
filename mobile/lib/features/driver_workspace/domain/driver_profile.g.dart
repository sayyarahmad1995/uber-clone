// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'driver_profile.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_DriverVehicle _$DriverVehicleFromJson(Map<String, dynamic> json) =>
    _DriverVehicle(
      make: json['make'] as String,
      model: json['model'] as String,
      modelYear: (json['model_year'] as num?)?.toInt(),
      color: json['color'] as String,
      licensePlate: json['license_plate'] as String,
    );

Map<String, dynamic> _$DriverVehicleToJson(_DriverVehicle instance) =>
    <String, dynamic>{
      'make': instance.make,
      'model': instance.model,
      'model_year': instance.modelYear,
      'color': instance.color,
      'license_plate': instance.licensePlate,
    };

_DriverProfile _$DriverProfileFromJson(Map<String, dynamic> json) =>
    _DriverProfile(
      userId: json['user_id'] as String,
      displayName: json['display_name'] as String?,
      status: json['status'] as String,
      isOnline: json['is_online'] as bool,
      vehicle: DriverVehicle.fromJson(json['vehicle'] as Map<String, dynamic>),
    );

Map<String, dynamic> _$DriverProfileToJson(_DriverProfile instance) =>
    <String, dynamic>{
      'user_id': instance.userId,
      'display_name': instance.displayName,
      'status': instance.status,
      'is_online': instance.isOnline,
      'vehicle': instance.vehicle,
    };

_PublishedDriverLocation _$PublishedDriverLocationFromJson(
  Map<String, dynamic> json,
) => _PublishedDriverLocation(
  latitude: (json['latitude'] as num).toDouble(),
  longitude: (json['longitude'] as num).toDouble(),
  updatedAt: DateTime.parse(json['updated_at'] as String),
);

Map<String, dynamic> _$PublishedDriverLocationToJson(
  _PublishedDriverLocation instance,
) => <String, dynamic>{
  'latitude': instance.latitude,
  'longitude': instance.longitude,
  'updated_at': instance.updatedAt.toIso8601String(),
};
