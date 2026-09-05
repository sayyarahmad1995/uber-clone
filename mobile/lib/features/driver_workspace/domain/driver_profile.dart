import 'package:freezed_annotation/freezed_annotation.dart';

part 'driver_profile.freezed.dart';
part 'driver_profile.g.dart';

@freezed
abstract class DriverVehicle with _$DriverVehicle {
  const factory DriverVehicle({
    required String make,
    required String model,
    @JsonKey(name: 'model_year') int? modelYear,
    required String color,
    @JsonKey(name: 'license_plate') required String licensePlate,
  }) = _DriverVehicle;
  factory DriverVehicle.fromJson(Map<String, dynamic> json) =>
      _$DriverVehicleFromJson(json);
}

@freezed
abstract class DriverProfile with _$DriverProfile {
  const factory DriverProfile({
    @JsonKey(name: 'user_id') required String userId,
    @JsonKey(name: 'display_name') String? displayName,
    required String status,
    @JsonKey(name: 'is_online') required bool isOnline,
    required DriverVehicle vehicle,
  }) = _DriverProfile;
  factory DriverProfile.fromJson(Map<String, dynamic> json) =>
      _$DriverProfileFromJson(json);
}

@freezed
abstract class PublishedDriverLocation with _$PublishedDriverLocation {
  const factory PublishedDriverLocation({
    required double latitude,
    required double longitude,
    @JsonKey(name: 'updated_at') required DateTime updatedAt,
  }) = _PublishedDriverLocation;
  factory PublishedDriverLocation.fromJson(Map<String, dynamic> json) =>
      _$PublishedDriverLocationFromJson(json);
}
