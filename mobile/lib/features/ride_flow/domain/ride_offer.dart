import 'package:freezed_annotation/freezed_annotation.dart';

import 'marketplace_request.dart';

part 'ride_offer.freezed.dart';
part 'ride_offer.g.dart';

@freezed
abstract class RideOffer with _$RideOffer {
  const factory RideOffer({
    @JsonKey(name: 'ride_request_id') required String rideRequestId,
    @JsonKey(name: 'driver_user_id') required String driverUserId,
    required RideFare fare,
    required String status,
    @JsonKey(name: 'created_at') required DateTime createdAt,
    @JsonKey(name: 'updated_at') required DateTime updatedAt,
    @JsonKey(name: 'expires_at') required DateTime expiresAt,
    @JsonKey(name: 'decided_at') DateTime? decidedAt,
  }) = _RideOffer;

  factory RideOffer.fromJson(Map<String, dynamic> json) =>
      _$RideOfferFromJson(json);
}

@freezed
abstract class DriverSummary with _$DriverSummary {
  const factory DriverSummary({
    @JsonKey(name: 'display_name') required String displayName,
  }) = _DriverSummary;

  factory DriverSummary.fromJson(Map<String, dynamic> json) =>
      _$DriverSummaryFromJson(json);
}

@freezed
abstract class VehicleSummary with _$VehicleSummary {
  const factory VehicleSummary({
    required String make,
    required String model,
    @JsonKey(name: 'model_year') int? modelYear,
    required String color,
  }) = _VehicleSummary;

  factory VehicleSummary.fromJson(Map<String, dynamic> json) =>
      _$VehicleSummaryFromJson(json);
}

@freezed
abstract class ServiceSummary with _$ServiceSummary {
  const factory ServiceSummary({
    required String code,
    @JsonKey(name: 'display_name') required String displayName,
  }) = _ServiceSummary;

  factory ServiceSummary.fromJson(Map<String, dynamic> json) =>
      _$ServiceSummaryFromJson(json);
}

@freezed
abstract class RiderOfferComparison with _$RiderOfferComparison {
  const factory RiderOfferComparison({
    @JsonKey(name: 'ride_request_id') required String rideRequestId,
    @JsonKey(name: 'driver_user_id') required String driverUserId,
    required RideFare fare,
    required String status,
    @JsonKey(name: 'created_at') required DateTime createdAt,
    @JsonKey(name: 'updated_at') required DateTime updatedAt,
    @JsonKey(name: 'expires_at') required DateTime expiresAt,
    @JsonKey(name: 'decided_at') DateTime? decidedAt,
    @JsonKey(name: 'pickup_distance_meters') int? pickupDistanceMeters,
    @JsonKey(name: 'matches_proposed_fare') required bool matchesProposedFare,
    required bool selectable,
    DriverSummary? driver,
    VehicleSummary? vehicle,
    ServiceSummary? service,
  }) = _RiderOfferComparison;

  factory RiderOfferComparison.fromJson(Map<String, dynamic> json) =>
      _$RiderOfferComparisonFromJson(json);
}
