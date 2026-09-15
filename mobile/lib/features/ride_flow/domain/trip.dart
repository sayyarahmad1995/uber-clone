import 'package:freezed_annotation/freezed_annotation.dart';

import 'marketplace_request.dart';

part 'trip.freezed.dart';
part 'trip.g.dart';

@freezed
abstract class SettlementSnapshot with _$SettlementSnapshot {
  const factory SettlementSnapshot({
    required String status,
    String? method,
    @JsonKey(name: 'cash_collected_at') DateTime? cashCollectedAt,
  }) = _SettlementSnapshot;

  factory SettlementSnapshot.fromJson(Map<String, dynamic> json) =>
      _$SettlementSnapshotFromJson(json);
}

@freezed
abstract class TripOperationFare with _$TripOperationFare {
  const factory TripOperationFare({
    @JsonKey(name: 'amount_minor') required int amountMinor,
    required String currency,
  }) = _TripOperationFare;

  factory TripOperationFare.fromJson(Map<String, dynamic> json) =>
      _$TripOperationFareFromJson(json);
}

@freezed
abstract class TripOperationContext with _$TripOperationContext {
  const factory TripOperationContext({
    @JsonKey(name: 'vehicle_id') required String vehicleId,
    @JsonKey(name: 'service_code') required String serviceCode,
    @JsonKey(name: 'service_name') required String serviceName,
    @JsonKey(name: 'driver_name') required String driverName,
    required String make,
    required String model,
    @JsonKey(name: 'model_year') required int modelYear,
    required String color,
    @JsonKey(name: 'license_plate') required String licensePlate,
    required TripOperationFare fare,
  }) = _TripOperationContext;

  factory TripOperationContext.fromJson(Map<String, dynamic> json) =>
      _$TripOperationContextFromJson(json);
}

@freezed
abstract class TripSnapshot with _$TripSnapshot {
  const factory TripSnapshot({
    @JsonKey(name: 'operation_context') TripOperationContext? operationContext,
    @JsonKey(name: 'ride_request_id') String? rideRequestId,
    @JsonKey(name: 'rider_user_id') String? riderUserId,
    @JsonKey(name: 'driver_user_id') String? driverUserId,
    RideLocation? pickup,
    RideLocation? destination,
    required String status,
    @JsonKey(name: 'assigned_at') required DateTime assignedAt,
    @JsonKey(name: 'started_at') DateTime? startedAt,
    @JsonKey(name: 'completed_at') DateTime? completedAt,
    @JsonKey(name: 'cancelled_at') DateTime? cancelledAt,
    required SettlementSnapshot settlement,
  }) = _TripSnapshot;

  factory TripSnapshot.fromJson(Map<String, dynamic> json) =>
      _$TripSnapshotFromJson(json);
}

@freezed
abstract class DriverTripHistoryItem with _$DriverTripHistoryItem {
  const factory DriverTripHistoryItem({
    @JsonKey(name: 'operation_context') TripOperationContext? operationContext,
    @JsonKey(name: 'ride_request_id') required String rideRequestId,
    required RideLocation pickup,
    required RideLocation destination,
    required String status,
    @JsonKey(name: 'assigned_at') required DateTime assignedAt,
    @JsonKey(name: 'started_at') DateTime? startedAt,
    @JsonKey(name: 'completed_at') DateTime? completedAt,
    @JsonKey(name: 'cancelled_at') DateTime? cancelledAt,
    required SettlementSnapshot settlement,
  }) = _DriverTripHistoryItem;

  factory DriverTripHistoryItem.fromJson(Map<String, dynamic> json) =>
      _$DriverTripHistoryItemFromJson(json);
}
