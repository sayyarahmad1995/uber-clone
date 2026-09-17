// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'ride_execution.dart';

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
