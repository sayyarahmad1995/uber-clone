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
abstract class TripSnapshot with _$TripSnapshot {
  const factory TripSnapshot({
    @JsonKey(name: 'ride_request_id') String? rideRequestId,
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
