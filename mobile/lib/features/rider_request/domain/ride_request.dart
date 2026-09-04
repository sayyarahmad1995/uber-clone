import 'package:freezed_annotation/freezed_annotation.dart';

part 'ride_request.freezed.dart';
part 'ride_request.g.dart';

@freezed
abstract class GeoPoint with _$GeoPoint {
  const factory GeoPoint({
    required double latitude,
    required double longitude,
  }) = _GeoPoint;
  factory GeoPoint.fromJson(Map<String, dynamic> json) =>
      _$GeoPointFromJson(json);
}

@freezed
abstract class Money with _$Money {
  const factory Money({
    @JsonKey(name: 'amount_minor') required int amountMinor,
    required String currency,
  }) = _Money;
  factory Money.fromJson(Map<String, dynamic> json) => _$MoneyFromJson(json);
}

@freezed
abstract class TripSnapshot with _$TripSnapshot {
  const factory TripSnapshot({required String status}) = _TripSnapshot;
  factory TripSnapshot.fromJson(Map<String, dynamic> json) =>
      _$TripSnapshotFromJson(json);
}

@freezed
abstract class RideRequest with _$RideRequest {
  const factory RideRequest({
    required String id,
    required GeoPoint pickup,
    required GeoPoint destination,
    @JsonKey(name: 'proposed_fare') required Money proposedFare,
    required String status,
    @JsonKey(name: 'created_at') required DateTime createdAt,
    TripSnapshot? trip,
  }) = _RideRequest;
  factory RideRequest.fromJson(Map<String, dynamic> json) =>
      _$RideRequestFromJson(json);
}

@freezed
abstract class RideRequestList with _$RideRequestList {
  const factory RideRequestList({
    @JsonKey(name: 'ride_requests') required List<RideRequest> rideRequests,
  }) = _RideRequestList;
  factory RideRequestList.fromJson(Map<String, dynamic> json) =>
      _$RideRequestListFromJson(json);
}
