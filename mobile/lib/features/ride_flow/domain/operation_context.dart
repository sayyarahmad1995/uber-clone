import 'package:freezed_annotation/freezed_annotation.dart';
part 'operation_context.freezed.dart';
part 'operation_context.g.dart';

@freezed
abstract class OperationFare with _$OperationFare {
  const factory OperationFare({required int amountMinor, required String currency}) = _OperationFare;
  factory OperationFare.fromJson(Map<String, dynamic> json) => _$OperationFareFromJson(json);
}
