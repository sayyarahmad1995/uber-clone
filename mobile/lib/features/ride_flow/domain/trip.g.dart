// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'trip.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

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

_TripOperationFare _$TripOperationFareFromJson(Map<String, dynamic> json) =>
    _TripOperationFare(
      amountMinor: (json['amount_minor'] as num).toInt(),
      currency: json['currency'] as String,
    );

Map<String, dynamic> _$TripOperationFareToJson(_TripOperationFare instance) =>
    <String, dynamic>{
      'amount_minor': instance.amountMinor,
      'currency': instance.currency,
    };

_TripOperationContext _$TripOperationContextFromJson(
  Map<String, dynamic> json,
) => _TripOperationContext(
  vehicleId: json['vehicle_id'] as String,
  serviceCode: json['service_code'] as String,
  serviceName: json['service_name'] as String,
  driverName: json['driver_name'] as String,
  make: json['make'] as String,
  model: json['model'] as String,
  modelYear: (json['model_year'] as num).toInt(),
  color: json['color'] as String,
  licensePlate: json['license_plate'] as String,
  fare: TripOperationFare.fromJson(json['fare'] as Map<String, dynamic>),
);

Map<String, dynamic> _$TripOperationContextToJson(
  _TripOperationContext instance,
) => <String, dynamic>{
  'vehicle_id': instance.vehicleId,
  'service_code': instance.serviceCode,
  'service_name': instance.serviceName,
  'driver_name': instance.driverName,
  'make': instance.make,
  'model': instance.model,
  'model_year': instance.modelYear,
  'color': instance.color,
  'license_plate': instance.licensePlate,
  'fare': instance.fare,
};

_TripSnapshot _$TripSnapshotFromJson(Map<String, dynamic> json) =>
    _TripSnapshot(
      operationContext: json['operation_context'] == null
          ? null
          : TripOperationContext.fromJson(
              json['operation_context'] as Map<String, dynamic>,
            ),
      rideRequestId: json['ride_request_id'] as String?,
      riderUserId: json['rider_user_id'] as String?,
      driverUserId: json['driver_user_id'] as String?,
      pickup: json['pickup'] == null
          ? null
          : RideLocation.fromJson(json['pickup'] as Map<String, dynamic>),
      destination: json['destination'] == null
          ? null
          : RideLocation.fromJson(json['destination'] as Map<String, dynamic>),
      status: json['status'] as String,
      assignedAt: DateTime.parse(json['assigned_at'] as String),
      startedAt: json['started_at'] == null
          ? null
          : DateTime.parse(json['started_at'] as String),
      completedAt: json['completed_at'] == null
          ? null
          : DateTime.parse(json['completed_at'] as String),
      cancelledAt: json['cancelled_at'] == null
          ? null
          : DateTime.parse(json['cancelled_at'] as String),
      settlement: SettlementSnapshot.fromJson(
        json['settlement'] as Map<String, dynamic>,
      ),
    );

Map<String, dynamic> _$TripSnapshotToJson(_TripSnapshot instance) =>
    <String, dynamic>{
      'operation_context': instance.operationContext,
      'ride_request_id': instance.rideRequestId,
      'rider_user_id': instance.riderUserId,
      'driver_user_id': instance.driverUserId,
      'pickup': instance.pickup,
      'destination': instance.destination,
      'status': instance.status,
      'assigned_at': instance.assignedAt.toIso8601String(),
      'started_at': instance.startedAt?.toIso8601String(),
      'completed_at': instance.completedAt?.toIso8601String(),
      'cancelled_at': instance.cancelledAt?.toIso8601String(),
      'settlement': instance.settlement,
    };

_DriverTripHistoryItem _$DriverTripHistoryItemFromJson(
  Map<String, dynamic> json,
) => _DriverTripHistoryItem(
  operationContext: json['operation_context'] == null
      ? null
      : TripOperationContext.fromJson(
          json['operation_context'] as Map<String, dynamic>,
        ),
  rideRequestId: json['ride_request_id'] as String,
  pickup: RideLocation.fromJson(json['pickup'] as Map<String, dynamic>),
  destination: RideLocation.fromJson(
    json['destination'] as Map<String, dynamic>,
  ),
  status: json['status'] as String,
  assignedAt: DateTime.parse(json['assigned_at'] as String),
  startedAt: json['started_at'] == null
      ? null
      : DateTime.parse(json['started_at'] as String),
  completedAt: json['completed_at'] == null
      ? null
      : DateTime.parse(json['completed_at'] as String),
  cancelledAt: json['cancelled_at'] == null
      ? null
      : DateTime.parse(json['cancelled_at'] as String),
  settlement: SettlementSnapshot.fromJson(
    json['settlement'] as Map<String, dynamic>,
  ),
);

Map<String, dynamic> _$DriverTripHistoryItemToJson(
  _DriverTripHistoryItem instance,
) => <String, dynamic>{
  'operation_context': instance.operationContext,
  'ride_request_id': instance.rideRequestId,
  'pickup': instance.pickup,
  'destination': instance.destination,
  'status': instance.status,
  'assigned_at': instance.assignedAt.toIso8601String(),
  'started_at': instance.startedAt?.toIso8601String(),
  'completed_at': instance.completedAt?.toIso8601String(),
  'cancelled_at': instance.cancelledAt?.toIso8601String(),
  'settlement': instance.settlement,
};
