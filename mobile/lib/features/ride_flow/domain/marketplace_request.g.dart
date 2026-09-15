// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'marketplace_request.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_RideLocation _$RideLocationFromJson(Map<String, dynamic> json) =>
    _RideLocation(
      latitude: (json['latitude'] as num).toDouble(),
      longitude: (json['longitude'] as num).toDouble(),
    );

Map<String, dynamic> _$RideLocationToJson(_RideLocation instance) =>
    <String, dynamic>{
      'latitude': instance.latitude,
      'longitude': instance.longitude,
    };

_RideFare _$RideFareFromJson(Map<String, dynamic> json) => _RideFare(
  amountMinor: (json['amount_minor'] as num).toInt(),
  currency: json['currency'] as String,
);

Map<String, dynamic> _$RideFareToJson(_RideFare instance) => <String, dynamic>{
  'amount_minor': instance.amountMinor,
  'currency': instance.currency,
};

_MarketplaceRequest _$MarketplaceRequestFromJson(Map<String, dynamic> json) =>
    _MarketplaceRequest(
      id: json['id'] as String,
      pickup: RideLocation.fromJson(json['pickup'] as Map<String, dynamic>),
      destination: RideLocation.fromJson(
        json['destination'] as Map<String, dynamic>,
      ),
      proposedFare: RideFare.fromJson(
        json['proposed_fare'] as Map<String, dynamic>,
      ),
      createdAt: DateTime.parse(json['created_at'] as String),
      expiresAt: DateTime.parse(json['expires_at'] as String),
      responseDeadline: DateTime.parse(json['response_deadline'] as String),
      ownOffer: json['own_offer'] == null
          ? null
          : RideOffer.fromJson(json['own_offer'] as Map<String, dynamic>),
      pickupDistanceMeters: (json['pickup_distance_meters'] as num).toInt(),
    );

Map<String, dynamic> _$MarketplaceRequestToJson(_MarketplaceRequest instance) =>
    <String, dynamic>{
      'id': instance.id,
      'pickup': instance.pickup,
      'destination': instance.destination,
      'proposed_fare': instance.proposedFare,
      'created_at': instance.createdAt.toIso8601String(),
      'expires_at': instance.expiresAt.toIso8601String(),
      'response_deadline': instance.responseDeadline.toIso8601String(),
      'own_offer': instance.ownOffer,
      'pickup_distance_meters': instance.pickupDistanceMeters,
    };
