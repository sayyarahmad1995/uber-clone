import 'package:freezed_annotation/freezed_annotation.dart';

part 'ride_execution.freezed.dart';
part 'ride_execution.g.dart';

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

/// The lifecycle meaning shared by Rider request and ride-flow snapshots.
abstract interface class TripLifecycleSnapshot {
  String get status;

  SettlementSnapshot? get settlement;
}

extension TripLifecycleSnapshotSemantics on TripLifecycleSnapshot {
  bool get isTerminal =>
      status == 'cancelled' ||
      (status == 'completed' && settlement?.status == 'cash_collected');
}
