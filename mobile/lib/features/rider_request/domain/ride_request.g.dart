// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'ride_request.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_GeoPoint _$GeoPointFromJson(Map<String, dynamic> json) => _GeoPoint(
  latitude: (json['latitude'] as num).toDouble(),
  longitude: (json['longitude'] as num).toDouble(),
);

Map<String, dynamic> _$GeoPointToJson(_GeoPoint instance) => <String, dynamic>{
  'latitude': instance.latitude,
  'longitude': instance.longitude,
};

_Money _$MoneyFromJson(Map<String, dynamic> json) => _Money(
  amountMinor: (json['amount_minor'] as num).toInt(),
  currency: json['currency'] as String,
);

Map<String, dynamic> _$MoneyToJson(_Money instance) => <String, dynamic>{
  'amount_minor': instance.amountMinor,
  'currency': instance.currency,
};

_SettlementSnapshot _$SettlementSnapshotFromJson(Map<String, dynamic> json) =>
    _SettlementSnapshot(
      status: json['status'] as String,
      method: json['method'] as String?,
      cashCollectedAt: json['cash_collected_at'] == null
          ? null
          : DateTime.parse(json['cash_collected_at'] as String),
    );

Map<String, dynamic> _$SettlementSnapshotToJson(_SettlementSnapshot instance) =>
    <String, dynamic>{
      'status': instance.status,
      'method': instance.method,
      'cash_collected_at': instance.cashCollectedAt?.toIso8601String(),
    };

_TripSnapshot _$TripSnapshotFromJson(Map<String, dynamic> json) =>
    _TripSnapshot(
      status: json['status'] as String,
      settlement: json['settlement'] == null
          ? null
          : SettlementSnapshot.fromJson(
              json['settlement'] as Map<String, dynamic>,
            ),
    );

Map<String, dynamic> _$TripSnapshotToJson(_TripSnapshot instance) =>
    <String, dynamic>{
      'status': instance.status,
      'settlement': instance.settlement,
    };

_RideRequest _$RideRequestFromJson(Map<String, dynamic> json) => _RideRequest(
  id: json['id'] as String,
  pickup: GeoPoint.fromJson(json['pickup'] as Map<String, dynamic>),
  destination: GeoPoint.fromJson(json['destination'] as Map<String, dynamic>),
  proposedFare: json['proposed_fare'] == null
      ? null
      : Money.fromJson(json['proposed_fare'] as Map<String, dynamic>),
  status: json['status'] as String,
  createdAt: DateTime.parse(json['created_at'] as String),
  trip: json['trip'] == null
      ? null
      : TripSnapshot.fromJson(json['trip'] as Map<String, dynamic>),
);

Map<String, dynamic> _$RideRequestToJson(_RideRequest instance) =>
    <String, dynamic>{
      'id': instance.id,
      'pickup': instance.pickup,
      'destination': instance.destination,
      'proposed_fare': instance.proposedFare,
      'status': instance.status,
      'created_at': instance.createdAt.toIso8601String(),
      'trip': instance.trip,
    };

_RideRequestList _$RideRequestListFromJson(Map<String, dynamic> json) =>
    _RideRequestList(
      rideRequests: (json['ride_requests'] as List<dynamic>)
          .map((e) => RideRequest.fromJson(e as Map<String, dynamic>))
          .toList(),
    );

Map<String, dynamic> _$RideRequestListToJson(_RideRequestList instance) =>
    <String, dynamic>{'ride_requests': instance.rideRequests};
