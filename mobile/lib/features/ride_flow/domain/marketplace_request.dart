import 'package:freezed_annotation/freezed_annotation.dart';

import 'ride_offer.dart';

part 'marketplace_request.freezed.dart';
part 'marketplace_request.g.dart';

@freezed
abstract class RideLocation with _$RideLocation {
  const factory RideLocation({
    required double latitude,
    required double longitude,
  }) = _RideLocation;

  factory RideLocation.fromJson(Map<String, dynamic> json) =>
      _$RideLocationFromJson(json);
}

@freezed
abstract class RideFare with _$RideFare {
  const factory RideFare({
    @JsonKey(name: 'amount_minor') required int amountMinor,
    required String currency,
  }) = _RideFare;

  factory RideFare.fromJson(Map<String, dynamic> json) => _$RideFareFromJson(json);
}

@freezed
abstract class MarketplaceRequest with _$MarketplaceRequest {
  const factory MarketplaceRequest({
    required String id,
    required RideLocation pickup,
    required RideLocation destination,
    @JsonKey(name: 'proposed_fare') required RideFare proposedFare,
    @JsonKey(name: 'created_at') required DateTime createdAt,
    @JsonKey(name: 'expires_at') required DateTime expiresAt,
    @JsonKey(name: 'response_deadline') required DateTime responseDeadline,
    @JsonKey(name: 'own_offer') RideOffer? ownOffer,
    @JsonKey(name: 'pickup_distance_meters') required int pickupDistanceMeters,
  }) = _MarketplaceRequest;

  factory MarketplaceRequest.fromJson(Map<String, dynamic> json) =>
      _$MarketplaceRequestFromJson(json);
}
