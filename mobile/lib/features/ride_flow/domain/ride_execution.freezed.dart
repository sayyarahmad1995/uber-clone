// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'ride_execution.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$SettlementSnapshot {

 String get status; String? get method;@JsonKey(name: 'cash_collected_at') DateTime? get cashCollectedAt;
/// Create a copy of SettlementSnapshot
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$SettlementSnapshotCopyWith<SettlementSnapshot> get copyWith => _$SettlementSnapshotCopyWithImpl<SettlementSnapshot>(this as SettlementSnapshot, _$identity);

  /// Serializes this SettlementSnapshot to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as SettlementSnapshot;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is SettlementSnapshot&&(identical(other.status, _this.status) || other.status == _this.status)&&(identical(other.method, _this.method) || other.method == _this.method)&&(identical(other.cashCollectedAt, _this.cashCollectedAt) || other.cashCollectedAt == _this.cashCollectedAt));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as SettlementSnapshot;
  return Object.hash(runtimeType,_this.status,_this.method,_this.cashCollectedAt);
}

@override
String toString() {
  final _this = this as SettlementSnapshot;
  return 'SettlementSnapshot(status: ${_this.status}, method: ${_this.method}, cashCollectedAt: ${_this.cashCollectedAt})';
}


}

/// @nodoc
abstract mixin class $SettlementSnapshotCopyWith<$Res>  {
  factory $SettlementSnapshotCopyWith(SettlementSnapshot value, $Res Function(SettlementSnapshot) _then) = _$SettlementSnapshotCopyWithImpl;
@useResult
$Res call({
 String status, String? method,@JsonKey(name: 'cash_collected_at') DateTime? cashCollectedAt
});




}
/// @nodoc
class _$SettlementSnapshotCopyWithImpl<$Res>
    implements $SettlementSnapshotCopyWith<$Res> {
  _$SettlementSnapshotCopyWithImpl(this._self, this._then);

  final SettlementSnapshot _self;
  final $Res Function(SettlementSnapshot) _then;

/// Create a copy of SettlementSnapshot
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? status = null,Object? method = freezed,Object? cashCollectedAt = freezed,}) {
  return _then(SettlementSnapshot(
status: null == status ? _self.status : status // ignore: cast_nullable_to_non_nullable
as String,method: freezed == method ? _self.method : method // ignore: cast_nullable_to_non_nullable
as String?,cashCollectedAt: freezed == cashCollectedAt ? _self.cashCollectedAt : cashCollectedAt // ignore: cast_nullable_to_non_nullable
as DateTime?,
  ));
}

}


/// Adds pattern-matching-related methods to [SettlementSnapshot].
extension SettlementSnapshotPatterns on SettlementSnapshot {
/// A variant of `map` that fallback to returning `orElse`.
///
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case final Subclass value:
///     return ...;
///   case _:
///     return orElse();
/// }
/// ```

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _SettlementSnapshot value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _SettlementSnapshot() when $default != null:
return $default(_that);case _:
  return orElse();

}
}
/// A `switch`-like method, using callbacks.
///
/// Callbacks receives the raw object, upcasted.
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case final Subclass value:
///     return ...;
///   case final Subclass2 value:
///     return ...;
/// }
/// ```

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _SettlementSnapshot value)  $default,){
final _that = this;
switch (_that) {
case _SettlementSnapshot():
return $default(_that);case _:
  throw StateError('Unexpected subclass');

}
}
/// A variant of `map` that fallback to returning `null`.
///
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case final Subclass value:
///     return ...;
///   case _:
///     return null;
/// }
/// ```

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _SettlementSnapshot value)?  $default,){
final _that = this;
switch (_that) {
case _SettlementSnapshot() when $default != null:
return $default(_that);case _:
  return null;

}
}
/// A variant of `when` that fallback to an `orElse` callback.
///
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case Subclass(:final field):
///     return ...;
///   case _:
///     return orElse();
/// }
/// ```

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( String status,  String? method, @JsonKey(name: 'cash_collected_at')  DateTime? cashCollectedAt)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _SettlementSnapshot() when $default != null:
return $default(_that.status,_that.method,_that.cashCollectedAt);case _:
  return orElse();

}
}
/// A `switch`-like method, using callbacks.
///
/// As opposed to `map`, this offers destructuring.
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case Subclass(:final field):
///     return ...;
///   case Subclass2(:final field2):
///     return ...;
/// }
/// ```

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( String status,  String? method, @JsonKey(name: 'cash_collected_at')  DateTime? cashCollectedAt)  $default,) {final _that = this;
switch (_that) {
case _SettlementSnapshot():
return $default(_that.status,_that.method,_that.cashCollectedAt);case _:
  throw StateError('Unexpected subclass');

}
}
/// A variant of `when` that fallback to returning `null`
///
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case Subclass(:final field):
///     return ...;
///   case _:
///     return null;
/// }
/// ```

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( String status,  String? method, @JsonKey(name: 'cash_collected_at')  DateTime? cashCollectedAt)?  $default,) {final _that = this;
switch (_that) {
case _SettlementSnapshot() when $default != null:
return $default(_that.status,_that.method,_that.cashCollectedAt);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _SettlementSnapshot implements SettlementSnapshot {
  const _SettlementSnapshot({required this.status, this.method, @JsonKey(name: 'cash_collected_at') this.cashCollectedAt});
  factory _SettlementSnapshot.fromJson(Map<String, dynamic> json) => _$SettlementSnapshotFromJson(json);

@override final  String status;
@override final  String? method;
@override@JsonKey(name: 'cash_collected_at') final  DateTime? cashCollectedAt;

/// Create a copy of SettlementSnapshot
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$SettlementSnapshotCopyWith<_SettlementSnapshot> get copyWith => __$SettlementSnapshotCopyWithImpl<_SettlementSnapshot>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$SettlementSnapshotToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _SettlementSnapshot&&(identical(other.status, status) || other.status == status)&&(identical(other.method, method) || other.method == method)&&(identical(other.cashCollectedAt, cashCollectedAt) || other.cashCollectedAt == cashCollectedAt));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,status,method,cashCollectedAt);
}

@override
String toString() {
    return 'SettlementSnapshot(status: $status, method: $method, cashCollectedAt: $cashCollectedAt)';
}


}

/// @nodoc
abstract mixin class _$SettlementSnapshotCopyWith<$Res> implements $SettlementSnapshotCopyWith<$Res> {
  factory _$SettlementSnapshotCopyWith(_SettlementSnapshot value, $Res Function(_SettlementSnapshot) _then) = __$SettlementSnapshotCopyWithImpl;
@override @useResult
$Res call({
 String status, String? method,@JsonKey(name: 'cash_collected_at') DateTime? cashCollectedAt
});




}
/// @nodoc
class __$SettlementSnapshotCopyWithImpl<$Res>
    implements _$SettlementSnapshotCopyWith<$Res> {
  __$SettlementSnapshotCopyWithImpl(this._self, this._then);

  final _SettlementSnapshot _self;
  final $Res Function(_SettlementSnapshot) _then;

/// Create a copy of SettlementSnapshot
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? status = null,Object? method = freezed,Object? cashCollectedAt = freezed,}) {
  return _then(_SettlementSnapshot(
status: null == status ? _self.status : status // ignore: cast_nullable_to_non_nullable
as String,method: freezed == method ? _self.method : method // ignore: cast_nullable_to_non_nullable
as String?,cashCollectedAt: freezed == cashCollectedAt ? _self.cashCollectedAt : cashCollectedAt // ignore: cast_nullable_to_non_nullable
as DateTime?,
  ));
}


}

// dart format on
