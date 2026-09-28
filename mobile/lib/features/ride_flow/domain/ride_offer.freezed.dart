// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'ride_offer.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$RideOffer {

@JsonKey(name: 'ride_request_id') String get rideRequestId;@JsonKey(name: 'driver_user_id') String get driverUserId; RideFare get fare; String get status;@JsonKey(name: 'created_at') DateTime get createdAt;@JsonKey(name: 'updated_at') DateTime get updatedAt;@JsonKey(name: 'expires_at') DateTime get expiresAt;@JsonKey(name: 'decided_at') DateTime? get decidedAt;
/// Create a copy of RideOffer
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$RideOfferCopyWith<RideOffer> get copyWith => _$RideOfferCopyWithImpl<RideOffer>(this as RideOffer, _$identity);

  /// Serializes this RideOffer to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as RideOffer;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is RideOffer&&(identical(other.rideRequestId, _this.rideRequestId) || other.rideRequestId == _this.rideRequestId)&&(identical(other.driverUserId, _this.driverUserId) || other.driverUserId == _this.driverUserId)&&(identical(other.fare, _this.fare) || other.fare == _this.fare)&&(identical(other.status, _this.status) || other.status == _this.status)&&(identical(other.createdAt, _this.createdAt) || other.createdAt == _this.createdAt)&&(identical(other.updatedAt, _this.updatedAt) || other.updatedAt == _this.updatedAt)&&(identical(other.expiresAt, _this.expiresAt) || other.expiresAt == _this.expiresAt)&&(identical(other.decidedAt, _this.decidedAt) || other.decidedAt == _this.decidedAt));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as RideOffer;
  return Object.hash(runtimeType,_this.rideRequestId,_this.driverUserId,_this.fare,_this.status,_this.createdAt,_this.updatedAt,_this.expiresAt,_this.decidedAt);
}

@override
String toString() {
  final _this = this as RideOffer;
  return 'RideOffer(rideRequestId: ${_this.rideRequestId}, driverUserId: ${_this.driverUserId}, fare: ${_this.fare}, status: ${_this.status}, createdAt: ${_this.createdAt}, updatedAt: ${_this.updatedAt}, expiresAt: ${_this.expiresAt}, decidedAt: ${_this.decidedAt})';
}


}

/// @nodoc
abstract mixin class $RideOfferCopyWith<$Res>  {
  factory $RideOfferCopyWith(RideOffer value, $Res Function(RideOffer) _then) = _$RideOfferCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'ride_request_id') String rideRequestId,@JsonKey(name: 'driver_user_id') String driverUserId, RideFare fare, String status,@JsonKey(name: 'created_at') DateTime createdAt,@JsonKey(name: 'updated_at') DateTime updatedAt,@JsonKey(name: 'expires_at') DateTime expiresAt,@JsonKey(name: 'decided_at') DateTime? decidedAt
});


$RideFareCopyWith<$Res> get fare;

}
/// @nodoc
class _$RideOfferCopyWithImpl<$Res>
    implements $RideOfferCopyWith<$Res> {
  _$RideOfferCopyWithImpl(this._self, this._then);

  final RideOffer _self;
  final $Res Function(RideOffer) _then;

/// Create a copy of RideOffer
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? rideRequestId = null,Object? driverUserId = null,Object? fare = null,Object? status = null,Object? createdAt = null,Object? updatedAt = null,Object? expiresAt = null,Object? decidedAt = freezed,}) {
  return _then(RideOffer(
rideRequestId: null == rideRequestId ? _self.rideRequestId : rideRequestId // ignore: cast_nullable_to_non_nullable
as String,driverUserId: null == driverUserId ? _self.driverUserId : driverUserId // ignore: cast_nullable_to_non_nullable
as String,fare: null == fare ? _self.fare : fare // ignore: cast_nullable_to_non_nullable
as RideFare,status: null == status ? _self.status : status // ignore: cast_nullable_to_non_nullable
as String,createdAt: null == createdAt ? _self.createdAt : createdAt // ignore: cast_nullable_to_non_nullable
as DateTime,updatedAt: null == updatedAt ? _self.updatedAt : updatedAt // ignore: cast_nullable_to_non_nullable
as DateTime,expiresAt: null == expiresAt ? _self.expiresAt : expiresAt // ignore: cast_nullable_to_non_nullable
as DateTime,decidedAt: freezed == decidedAt ? _self.decidedAt : decidedAt // ignore: cast_nullable_to_non_nullable
as DateTime?,
  ));
}
/// Create a copy of RideOffer
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$RideFareCopyWith<$Res> get fare {
  
  return $RideFareCopyWith<$Res>(_self.fare, (value) {
    return _then(_self.copyWith(fare: value));
  });
}
}


/// Adds pattern-matching-related methods to [RideOffer].
extension RideOfferPatterns on RideOffer {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _RideOffer value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _RideOffer() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _RideOffer value)  $default,){
final _that = this;
switch (_that) {
case _RideOffer():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _RideOffer value)?  $default,){
final _that = this;
switch (_that) {
case _RideOffer() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(name: 'ride_request_id')  String rideRequestId, @JsonKey(name: 'driver_user_id')  String driverUserId,  RideFare fare,  String status, @JsonKey(name: 'created_at')  DateTime createdAt, @JsonKey(name: 'updated_at')  DateTime updatedAt, @JsonKey(name: 'expires_at')  DateTime expiresAt, @JsonKey(name: 'decided_at')  DateTime? decidedAt)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _RideOffer() when $default != null:
return $default(_that.rideRequestId,_that.driverUserId,_that.fare,_that.status,_that.createdAt,_that.updatedAt,_that.expiresAt,_that.decidedAt);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(name: 'ride_request_id')  String rideRequestId, @JsonKey(name: 'driver_user_id')  String driverUserId,  RideFare fare,  String status, @JsonKey(name: 'created_at')  DateTime createdAt, @JsonKey(name: 'updated_at')  DateTime updatedAt, @JsonKey(name: 'expires_at')  DateTime expiresAt, @JsonKey(name: 'decided_at')  DateTime? decidedAt)  $default,) {final _that = this;
switch (_that) {
case _RideOffer():
return $default(_that.rideRequestId,_that.driverUserId,_that.fare,_that.status,_that.createdAt,_that.updatedAt,_that.expiresAt,_that.decidedAt);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(name: 'ride_request_id')  String rideRequestId, @JsonKey(name: 'driver_user_id')  String driverUserId,  RideFare fare,  String status, @JsonKey(name: 'created_at')  DateTime createdAt, @JsonKey(name: 'updated_at')  DateTime updatedAt, @JsonKey(name: 'expires_at')  DateTime expiresAt, @JsonKey(name: 'decided_at')  DateTime? decidedAt)?  $default,) {final _that = this;
switch (_that) {
case _RideOffer() when $default != null:
return $default(_that.rideRequestId,_that.driverUserId,_that.fare,_that.status,_that.createdAt,_that.updatedAt,_that.expiresAt,_that.decidedAt);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _RideOffer implements RideOffer {
  const _RideOffer({@JsonKey(name: 'ride_request_id') required this.rideRequestId, @JsonKey(name: 'driver_user_id') required this.driverUserId, required this.fare, required this.status, @JsonKey(name: 'created_at') required this.createdAt, @JsonKey(name: 'updated_at') required this.updatedAt, @JsonKey(name: 'expires_at') required this.expiresAt, @JsonKey(name: 'decided_at') this.decidedAt});
  factory _RideOffer.fromJson(Map<String, dynamic> json) => _$RideOfferFromJson(json);

@override@JsonKey(name: 'ride_request_id') final  String rideRequestId;
@override@JsonKey(name: 'driver_user_id') final  String driverUserId;
@override final  RideFare fare;
@override final  String status;
@override@JsonKey(name: 'created_at') final  DateTime createdAt;
@override@JsonKey(name: 'updated_at') final  DateTime updatedAt;
@override@JsonKey(name: 'expires_at') final  DateTime expiresAt;
@override@JsonKey(name: 'decided_at') final  DateTime? decidedAt;

/// Create a copy of RideOffer
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$RideOfferCopyWith<_RideOffer> get copyWith => __$RideOfferCopyWithImpl<_RideOffer>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$RideOfferToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _RideOffer&&(identical(other.rideRequestId, rideRequestId) || other.rideRequestId == rideRequestId)&&(identical(other.driverUserId, driverUserId) || other.driverUserId == driverUserId)&&(identical(other.fare, fare) || other.fare == fare)&&(identical(other.status, status) || other.status == status)&&(identical(other.createdAt, createdAt) || other.createdAt == createdAt)&&(identical(other.updatedAt, updatedAt) || other.updatedAt == updatedAt)&&(identical(other.expiresAt, expiresAt) || other.expiresAt == expiresAt)&&(identical(other.decidedAt, decidedAt) || other.decidedAt == decidedAt));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,rideRequestId,driverUserId,fare,status,createdAt,updatedAt,expiresAt,decidedAt);
}

@override
String toString() {
    return 'RideOffer(rideRequestId: $rideRequestId, driverUserId: $driverUserId, fare: $fare, status: $status, createdAt: $createdAt, updatedAt: $updatedAt, expiresAt: $expiresAt, decidedAt: $decidedAt)';
}


}

/// @nodoc
abstract mixin class _$RideOfferCopyWith<$Res> implements $RideOfferCopyWith<$Res> {
  factory _$RideOfferCopyWith(_RideOffer value, $Res Function(_RideOffer) _then) = __$RideOfferCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'ride_request_id') String rideRequestId,@JsonKey(name: 'driver_user_id') String driverUserId, RideFare fare, String status,@JsonKey(name: 'created_at') DateTime createdAt,@JsonKey(name: 'updated_at') DateTime updatedAt,@JsonKey(name: 'expires_at') DateTime expiresAt,@JsonKey(name: 'decided_at') DateTime? decidedAt
});


@override $RideFareCopyWith<$Res> get fare;

}
/// @nodoc
class __$RideOfferCopyWithImpl<$Res>
    implements _$RideOfferCopyWith<$Res> {
  __$RideOfferCopyWithImpl(this._self, this._then);

  final _RideOffer _self;
  final $Res Function(_RideOffer) _then;

/// Create a copy of RideOffer
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? rideRequestId = null,Object? driverUserId = null,Object? fare = null,Object? status = null,Object? createdAt = null,Object? updatedAt = null,Object? expiresAt = null,Object? decidedAt = freezed,}) {
  return _then(_RideOffer(
rideRequestId: null == rideRequestId ? _self.rideRequestId : rideRequestId // ignore: cast_nullable_to_non_nullable
as String,driverUserId: null == driverUserId ? _self.driverUserId : driverUserId // ignore: cast_nullable_to_non_nullable
as String,fare: null == fare ? _self.fare : fare // ignore: cast_nullable_to_non_nullable
as RideFare,status: null == status ? _self.status : status // ignore: cast_nullable_to_non_nullable
as String,createdAt: null == createdAt ? _self.createdAt : createdAt // ignore: cast_nullable_to_non_nullable
as DateTime,updatedAt: null == updatedAt ? _self.updatedAt : updatedAt // ignore: cast_nullable_to_non_nullable
as DateTime,expiresAt: null == expiresAt ? _self.expiresAt : expiresAt // ignore: cast_nullable_to_non_nullable
as DateTime,decidedAt: freezed == decidedAt ? _self.decidedAt : decidedAt // ignore: cast_nullable_to_non_nullable
as DateTime?,
  ));
}

/// Create a copy of RideOffer
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$RideFareCopyWith<$Res> get fare {
  
  return $RideFareCopyWith<$Res>(_self.fare, (value) {
    return _then(_self.copyWith(fare: value));
  });
}
}


/// @nodoc
mixin _$DriverSummary {

@JsonKey(name: 'display_name') String get displayName;
/// Create a copy of DriverSummary
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$DriverSummaryCopyWith<DriverSummary> get copyWith => _$DriverSummaryCopyWithImpl<DriverSummary>(this as DriverSummary, _$identity);

  /// Serializes this DriverSummary to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as DriverSummary;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is DriverSummary&&(identical(other.displayName, _this.displayName) || other.displayName == _this.displayName));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as DriverSummary;
  return Object.hash(runtimeType,_this.displayName);
}

@override
String toString() {
  final _this = this as DriverSummary;
  return 'DriverSummary(displayName: ${_this.displayName})';
}


}

/// @nodoc
abstract mixin class $DriverSummaryCopyWith<$Res>  {
  factory $DriverSummaryCopyWith(DriverSummary value, $Res Function(DriverSummary) _then) = _$DriverSummaryCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'display_name') String displayName
});




}
/// @nodoc
class _$DriverSummaryCopyWithImpl<$Res>
    implements $DriverSummaryCopyWith<$Res> {
  _$DriverSummaryCopyWithImpl(this._self, this._then);

  final DriverSummary _self;
  final $Res Function(DriverSummary) _then;

/// Create a copy of DriverSummary
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? displayName = null,}) {
  return _then(DriverSummary(
displayName: null == displayName ? _self.displayName : displayName // ignore: cast_nullable_to_non_nullable
as String,
  ));
}

}


/// Adds pattern-matching-related methods to [DriverSummary].
extension DriverSummaryPatterns on DriverSummary {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _DriverSummary value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _DriverSummary() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _DriverSummary value)  $default,){
final _that = this;
switch (_that) {
case _DriverSummary():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _DriverSummary value)?  $default,){
final _that = this;
switch (_that) {
case _DriverSummary() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(name: 'display_name')  String displayName)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _DriverSummary() when $default != null:
return $default(_that.displayName);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(name: 'display_name')  String displayName)  $default,) {final _that = this;
switch (_that) {
case _DriverSummary():
return $default(_that.displayName);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(name: 'display_name')  String displayName)?  $default,) {final _that = this;
switch (_that) {
case _DriverSummary() when $default != null:
return $default(_that.displayName);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _DriverSummary implements DriverSummary {
  const _DriverSummary({@JsonKey(name: 'display_name') required this.displayName});
  factory _DriverSummary.fromJson(Map<String, dynamic> json) => _$DriverSummaryFromJson(json);

@override@JsonKey(name: 'display_name') final  String displayName;

/// Create a copy of DriverSummary
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$DriverSummaryCopyWith<_DriverSummary> get copyWith => __$DriverSummaryCopyWithImpl<_DriverSummary>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$DriverSummaryToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _DriverSummary&&(identical(other.displayName, displayName) || other.displayName == displayName));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,displayName);
}

@override
String toString() {
    return 'DriverSummary(displayName: $displayName)';
}


}

/// @nodoc
abstract mixin class _$DriverSummaryCopyWith<$Res> implements $DriverSummaryCopyWith<$Res> {
  factory _$DriverSummaryCopyWith(_DriverSummary value, $Res Function(_DriverSummary) _then) = __$DriverSummaryCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'display_name') String displayName
});




}
/// @nodoc
class __$DriverSummaryCopyWithImpl<$Res>
    implements _$DriverSummaryCopyWith<$Res> {
  __$DriverSummaryCopyWithImpl(this._self, this._then);

  final _DriverSummary _self;
  final $Res Function(_DriverSummary) _then;

/// Create a copy of DriverSummary
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? displayName = null,}) {
  return _then(_DriverSummary(
displayName: null == displayName ? _self.displayName : displayName // ignore: cast_nullable_to_non_nullable
as String,
  ));
}


}


/// @nodoc
mixin _$VehicleSummary {

 String get make; String get model;@JsonKey(name: 'model_year') int? get modelYear; String get color;
/// Create a copy of VehicleSummary
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$VehicleSummaryCopyWith<VehicleSummary> get copyWith => _$VehicleSummaryCopyWithImpl<VehicleSummary>(this as VehicleSummary, _$identity);

  /// Serializes this VehicleSummary to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as VehicleSummary;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is VehicleSummary&&(identical(other.make, _this.make) || other.make == _this.make)&&(identical(other.model, _this.model) || other.model == _this.model)&&(identical(other.modelYear, _this.modelYear) || other.modelYear == _this.modelYear)&&(identical(other.color, _this.color) || other.color == _this.color));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as VehicleSummary;
  return Object.hash(runtimeType,_this.make,_this.model,_this.modelYear,_this.color);
}

@override
String toString() {
  final _this = this as VehicleSummary;
  return 'VehicleSummary(make: ${_this.make}, model: ${_this.model}, modelYear: ${_this.modelYear}, color: ${_this.color})';
}


}

/// @nodoc
abstract mixin class $VehicleSummaryCopyWith<$Res>  {
  factory $VehicleSummaryCopyWith(VehicleSummary value, $Res Function(VehicleSummary) _then) = _$VehicleSummaryCopyWithImpl;
@useResult
$Res call({
 String make, String model,@JsonKey(name: 'model_year') int? modelYear, String color
});




}
/// @nodoc
class _$VehicleSummaryCopyWithImpl<$Res>
    implements $VehicleSummaryCopyWith<$Res> {
  _$VehicleSummaryCopyWithImpl(this._self, this._then);

  final VehicleSummary _self;
  final $Res Function(VehicleSummary) _then;

/// Create a copy of VehicleSummary
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? make = null,Object? model = null,Object? modelYear = freezed,Object? color = null,}) {
  return _then(VehicleSummary(
make: null == make ? _self.make : make // ignore: cast_nullable_to_non_nullable
as String,model: null == model ? _self.model : model // ignore: cast_nullable_to_non_nullable
as String,modelYear: freezed == modelYear ? _self.modelYear : modelYear // ignore: cast_nullable_to_non_nullable
as int?,color: null == color ? _self.color : color // ignore: cast_nullable_to_non_nullable
as String,
  ));
}

}


/// Adds pattern-matching-related methods to [VehicleSummary].
extension VehicleSummaryPatterns on VehicleSummary {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _VehicleSummary value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _VehicleSummary() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _VehicleSummary value)  $default,){
final _that = this;
switch (_that) {
case _VehicleSummary():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _VehicleSummary value)?  $default,){
final _that = this;
switch (_that) {
case _VehicleSummary() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( String make,  String model, @JsonKey(name: 'model_year')  int? modelYear,  String color)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _VehicleSummary() when $default != null:
return $default(_that.make,_that.model,_that.modelYear,_that.color);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( String make,  String model, @JsonKey(name: 'model_year')  int? modelYear,  String color)  $default,) {final _that = this;
switch (_that) {
case _VehicleSummary():
return $default(_that.make,_that.model,_that.modelYear,_that.color);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( String make,  String model, @JsonKey(name: 'model_year')  int? modelYear,  String color)?  $default,) {final _that = this;
switch (_that) {
case _VehicleSummary() when $default != null:
return $default(_that.make,_that.model,_that.modelYear,_that.color);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _VehicleSummary implements VehicleSummary {
  const _VehicleSummary({required this.make, required this.model, @JsonKey(name: 'model_year') this.modelYear, required this.color});
  factory _VehicleSummary.fromJson(Map<String, dynamic> json) => _$VehicleSummaryFromJson(json);

@override final  String make;
@override final  String model;
@override@JsonKey(name: 'model_year') final  int? modelYear;
@override final  String color;

/// Create a copy of VehicleSummary
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$VehicleSummaryCopyWith<_VehicleSummary> get copyWith => __$VehicleSummaryCopyWithImpl<_VehicleSummary>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$VehicleSummaryToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _VehicleSummary&&(identical(other.make, make) || other.make == make)&&(identical(other.model, model) || other.model == model)&&(identical(other.modelYear, modelYear) || other.modelYear == modelYear)&&(identical(other.color, color) || other.color == color));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,make,model,modelYear,color);
}

@override
String toString() {
    return 'VehicleSummary(make: $make, model: $model, modelYear: $modelYear, color: $color)';
}


}

/// @nodoc
abstract mixin class _$VehicleSummaryCopyWith<$Res> implements $VehicleSummaryCopyWith<$Res> {
  factory _$VehicleSummaryCopyWith(_VehicleSummary value, $Res Function(_VehicleSummary) _then) = __$VehicleSummaryCopyWithImpl;
@override @useResult
$Res call({
 String make, String model,@JsonKey(name: 'model_year') int? modelYear, String color
});




}
/// @nodoc
class __$VehicleSummaryCopyWithImpl<$Res>
    implements _$VehicleSummaryCopyWith<$Res> {
  __$VehicleSummaryCopyWithImpl(this._self, this._then);

  final _VehicleSummary _self;
  final $Res Function(_VehicleSummary) _then;

/// Create a copy of VehicleSummary
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? make = null,Object? model = null,Object? modelYear = freezed,Object? color = null,}) {
  return _then(_VehicleSummary(
make: null == make ? _self.make : make // ignore: cast_nullable_to_non_nullable
as String,model: null == model ? _self.model : model // ignore: cast_nullable_to_non_nullable
as String,modelYear: freezed == modelYear ? _self.modelYear : modelYear // ignore: cast_nullable_to_non_nullable
as int?,color: null == color ? _self.color : color // ignore: cast_nullable_to_non_nullable
as String,
  ));
}


}


/// @nodoc
mixin _$ServiceSummary {

 String get code;@JsonKey(name: 'display_name') String get displayName;
/// Create a copy of ServiceSummary
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$ServiceSummaryCopyWith<ServiceSummary> get copyWith => _$ServiceSummaryCopyWithImpl<ServiceSummary>(this as ServiceSummary, _$identity);

  /// Serializes this ServiceSummary to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as ServiceSummary;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is ServiceSummary&&(identical(other.code, _this.code) || other.code == _this.code)&&(identical(other.displayName, _this.displayName) || other.displayName == _this.displayName));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as ServiceSummary;
  return Object.hash(runtimeType,_this.code,_this.displayName);
}

@override
String toString() {
  final _this = this as ServiceSummary;
  return 'ServiceSummary(code: ${_this.code}, displayName: ${_this.displayName})';
}


}

/// @nodoc
abstract mixin class $ServiceSummaryCopyWith<$Res>  {
  factory $ServiceSummaryCopyWith(ServiceSummary value, $Res Function(ServiceSummary) _then) = _$ServiceSummaryCopyWithImpl;
@useResult
$Res call({
 String code,@JsonKey(name: 'display_name') String displayName
});




}
/// @nodoc
class _$ServiceSummaryCopyWithImpl<$Res>
    implements $ServiceSummaryCopyWith<$Res> {
  _$ServiceSummaryCopyWithImpl(this._self, this._then);

  final ServiceSummary _self;
  final $Res Function(ServiceSummary) _then;

/// Create a copy of ServiceSummary
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? code = null,Object? displayName = null,}) {
  return _then(ServiceSummary(
code: null == code ? _self.code : code // ignore: cast_nullable_to_non_nullable
as String,displayName: null == displayName ? _self.displayName : displayName // ignore: cast_nullable_to_non_nullable
as String,
  ));
}

}


/// Adds pattern-matching-related methods to [ServiceSummary].
extension ServiceSummaryPatterns on ServiceSummary {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _ServiceSummary value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _ServiceSummary() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _ServiceSummary value)  $default,){
final _that = this;
switch (_that) {
case _ServiceSummary():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _ServiceSummary value)?  $default,){
final _that = this;
switch (_that) {
case _ServiceSummary() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( String code, @JsonKey(name: 'display_name')  String displayName)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _ServiceSummary() when $default != null:
return $default(_that.code,_that.displayName);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( String code, @JsonKey(name: 'display_name')  String displayName)  $default,) {final _that = this;
switch (_that) {
case _ServiceSummary():
return $default(_that.code,_that.displayName);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( String code, @JsonKey(name: 'display_name')  String displayName)?  $default,) {final _that = this;
switch (_that) {
case _ServiceSummary() when $default != null:
return $default(_that.code,_that.displayName);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _ServiceSummary implements ServiceSummary {
  const _ServiceSummary({required this.code, @JsonKey(name: 'display_name') required this.displayName});
  factory _ServiceSummary.fromJson(Map<String, dynamic> json) => _$ServiceSummaryFromJson(json);

@override final  String code;
@override@JsonKey(name: 'display_name') final  String displayName;

/// Create a copy of ServiceSummary
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$ServiceSummaryCopyWith<_ServiceSummary> get copyWith => __$ServiceSummaryCopyWithImpl<_ServiceSummary>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$ServiceSummaryToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _ServiceSummary&&(identical(other.code, code) || other.code == code)&&(identical(other.displayName, displayName) || other.displayName == displayName));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,code,displayName);
}

@override
String toString() {
    return 'ServiceSummary(code: $code, displayName: $displayName)';
}


}

/// @nodoc
abstract mixin class _$ServiceSummaryCopyWith<$Res> implements $ServiceSummaryCopyWith<$Res> {
  factory _$ServiceSummaryCopyWith(_ServiceSummary value, $Res Function(_ServiceSummary) _then) = __$ServiceSummaryCopyWithImpl;
@override @useResult
$Res call({
 String code,@JsonKey(name: 'display_name') String displayName
});




}
/// @nodoc
class __$ServiceSummaryCopyWithImpl<$Res>
    implements _$ServiceSummaryCopyWith<$Res> {
  __$ServiceSummaryCopyWithImpl(this._self, this._then);

  final _ServiceSummary _self;
  final $Res Function(_ServiceSummary) _then;

/// Create a copy of ServiceSummary
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? code = null,Object? displayName = null,}) {
  return _then(_ServiceSummary(
code: null == code ? _self.code : code // ignore: cast_nullable_to_non_nullable
as String,displayName: null == displayName ? _self.displayName : displayName // ignore: cast_nullable_to_non_nullable
as String,
  ));
}


}


/// @nodoc
mixin _$RiderOfferComparison {

@JsonKey(name: 'ride_request_id') String get rideRequestId;@JsonKey(name: 'driver_user_id') String get driverUserId; RideFare get fare; String get status;@JsonKey(name: 'created_at') DateTime get createdAt;@JsonKey(name: 'updated_at') DateTime get updatedAt;@JsonKey(name: 'expires_at') DateTime get expiresAt;@JsonKey(name: 'decided_at') DateTime? get decidedAt;@JsonKey(name: 'pickup_distance_meters') int? get pickupDistanceMeters;@JsonKey(name: 'matches_proposed_fare') bool get matchesProposedFare; bool get selectable; DriverSummary? get driver; VehicleSummary? get vehicle; ServiceSummary? get service;
/// Create a copy of RiderOfferComparison
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$RiderOfferComparisonCopyWith<RiderOfferComparison> get copyWith => _$RiderOfferComparisonCopyWithImpl<RiderOfferComparison>(this as RiderOfferComparison, _$identity);

  /// Serializes this RiderOfferComparison to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as RiderOfferComparison;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is RiderOfferComparison&&(identical(other.rideRequestId, _this.rideRequestId) || other.rideRequestId == _this.rideRequestId)&&(identical(other.driverUserId, _this.driverUserId) || other.driverUserId == _this.driverUserId)&&(identical(other.fare, _this.fare) || other.fare == _this.fare)&&(identical(other.status, _this.status) || other.status == _this.status)&&(identical(other.createdAt, _this.createdAt) || other.createdAt == _this.createdAt)&&(identical(other.updatedAt, _this.updatedAt) || other.updatedAt == _this.updatedAt)&&(identical(other.expiresAt, _this.expiresAt) || other.expiresAt == _this.expiresAt)&&(identical(other.decidedAt, _this.decidedAt) || other.decidedAt == _this.decidedAt)&&(identical(other.pickupDistanceMeters, _this.pickupDistanceMeters) || other.pickupDistanceMeters == _this.pickupDistanceMeters)&&(identical(other.matchesProposedFare, _this.matchesProposedFare) || other.matchesProposedFare == _this.matchesProposedFare)&&(identical(other.selectable, _this.selectable) || other.selectable == _this.selectable)&&(identical(other.driver, _this.driver) || other.driver == _this.driver)&&(identical(other.vehicle, _this.vehicle) || other.vehicle == _this.vehicle)&&(identical(other.service, _this.service) || other.service == _this.service));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as RiderOfferComparison;
  return Object.hash(runtimeType,_this.rideRequestId,_this.driverUserId,_this.fare,_this.status,_this.createdAt,_this.updatedAt,_this.expiresAt,_this.decidedAt,_this.pickupDistanceMeters,_this.matchesProposedFare,_this.selectable,_this.driver,_this.vehicle,_this.service);
}

@override
String toString() {
  final _this = this as RiderOfferComparison;
  return 'RiderOfferComparison(rideRequestId: ${_this.rideRequestId}, driverUserId: ${_this.driverUserId}, fare: ${_this.fare}, status: ${_this.status}, createdAt: ${_this.createdAt}, updatedAt: ${_this.updatedAt}, expiresAt: ${_this.expiresAt}, decidedAt: ${_this.decidedAt}, pickupDistanceMeters: ${_this.pickupDistanceMeters}, matchesProposedFare: ${_this.matchesProposedFare}, selectable: ${_this.selectable}, driver: ${_this.driver}, vehicle: ${_this.vehicle}, service: ${_this.service})';
}


}

/// @nodoc
abstract mixin class $RiderOfferComparisonCopyWith<$Res>  {
  factory $RiderOfferComparisonCopyWith(RiderOfferComparison value, $Res Function(RiderOfferComparison) _then) = _$RiderOfferComparisonCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'ride_request_id') String rideRequestId,@JsonKey(name: 'driver_user_id') String driverUserId, RideFare fare, String status,@JsonKey(name: 'created_at') DateTime createdAt,@JsonKey(name: 'updated_at') DateTime updatedAt,@JsonKey(name: 'expires_at') DateTime expiresAt,@JsonKey(name: 'decided_at') DateTime? decidedAt,@JsonKey(name: 'pickup_distance_meters') int? pickupDistanceMeters,@JsonKey(name: 'matches_proposed_fare') bool matchesProposedFare, bool selectable, DriverSummary? driver, VehicleSummary? vehicle, ServiceSummary? service
});


$RideFareCopyWith<$Res> get fare;$DriverSummaryCopyWith<$Res>? get driver;$VehicleSummaryCopyWith<$Res>? get vehicle;$ServiceSummaryCopyWith<$Res>? get service;

}
/// @nodoc
class _$RiderOfferComparisonCopyWithImpl<$Res>
    implements $RiderOfferComparisonCopyWith<$Res> {
  _$RiderOfferComparisonCopyWithImpl(this._self, this._then);

  final RiderOfferComparison _self;
  final $Res Function(RiderOfferComparison) _then;

/// Create a copy of RiderOfferComparison
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? rideRequestId = null,Object? driverUserId = null,Object? fare = null,Object? status = null,Object? createdAt = null,Object? updatedAt = null,Object? expiresAt = null,Object? decidedAt = freezed,Object? pickupDistanceMeters = freezed,Object? matchesProposedFare = null,Object? selectable = null,Object? driver = freezed,Object? vehicle = freezed,Object? service = freezed,}) {
  return _then(RiderOfferComparison(
rideRequestId: null == rideRequestId ? _self.rideRequestId : rideRequestId // ignore: cast_nullable_to_non_nullable
as String,driverUserId: null == driverUserId ? _self.driverUserId : driverUserId // ignore: cast_nullable_to_non_nullable
as String,fare: null == fare ? _self.fare : fare // ignore: cast_nullable_to_non_nullable
as RideFare,status: null == status ? _self.status : status // ignore: cast_nullable_to_non_nullable
as String,createdAt: null == createdAt ? _self.createdAt : createdAt // ignore: cast_nullable_to_non_nullable
as DateTime,updatedAt: null == updatedAt ? _self.updatedAt : updatedAt // ignore: cast_nullable_to_non_nullable
as DateTime,expiresAt: null == expiresAt ? _self.expiresAt : expiresAt // ignore: cast_nullable_to_non_nullable
as DateTime,decidedAt: freezed == decidedAt ? _self.decidedAt : decidedAt // ignore: cast_nullable_to_non_nullable
as DateTime?,pickupDistanceMeters: freezed == pickupDistanceMeters ? _self.pickupDistanceMeters : pickupDistanceMeters // ignore: cast_nullable_to_non_nullable
as int?,matchesProposedFare: null == matchesProposedFare ? _self.matchesProposedFare : matchesProposedFare // ignore: cast_nullable_to_non_nullable
as bool,selectable: null == selectable ? _self.selectable : selectable // ignore: cast_nullable_to_non_nullable
as bool,driver: freezed == driver ? _self.driver : driver // ignore: cast_nullable_to_non_nullable
as DriverSummary?,vehicle: freezed == vehicle ? _self.vehicle : vehicle // ignore: cast_nullable_to_non_nullable
as VehicleSummary?,service: freezed == service ? _self.service : service // ignore: cast_nullable_to_non_nullable
as ServiceSummary?,
  ));
}
/// Create a copy of RiderOfferComparison
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$RideFareCopyWith<$Res> get fare {
  
  return $RideFareCopyWith<$Res>(_self.fare, (value) {
    return _then(_self.copyWith(fare: value));
  });
}/// Create a copy of RiderOfferComparison
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$DriverSummaryCopyWith<$Res>? get driver {
    if (_self.driver == null) {
    return null;
  }

  return $DriverSummaryCopyWith<$Res>(_self.driver!, (value) {
    return _then(_self.copyWith(driver: value));
  });
}/// Create a copy of RiderOfferComparison
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$VehicleSummaryCopyWith<$Res>? get vehicle {
    if (_self.vehicle == null) {
    return null;
  }

  return $VehicleSummaryCopyWith<$Res>(_self.vehicle!, (value) {
    return _then(_self.copyWith(vehicle: value));
  });
}/// Create a copy of RiderOfferComparison
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$ServiceSummaryCopyWith<$Res>? get service {
    if (_self.service == null) {
    return null;
  }

  return $ServiceSummaryCopyWith<$Res>(_self.service!, (value) {
    return _then(_self.copyWith(service: value));
  });
}
}


/// Adds pattern-matching-related methods to [RiderOfferComparison].
extension RiderOfferComparisonPatterns on RiderOfferComparison {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _RiderOfferComparison value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _RiderOfferComparison() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _RiderOfferComparison value)  $default,){
final _that = this;
switch (_that) {
case _RiderOfferComparison():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _RiderOfferComparison value)?  $default,){
final _that = this;
switch (_that) {
case _RiderOfferComparison() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(name: 'ride_request_id')  String rideRequestId, @JsonKey(name: 'driver_user_id')  String driverUserId,  RideFare fare,  String status, @JsonKey(name: 'created_at')  DateTime createdAt, @JsonKey(name: 'updated_at')  DateTime updatedAt, @JsonKey(name: 'expires_at')  DateTime expiresAt, @JsonKey(name: 'decided_at')  DateTime? decidedAt, @JsonKey(name: 'pickup_distance_meters')  int? pickupDistanceMeters, @JsonKey(name: 'matches_proposed_fare')  bool matchesProposedFare,  bool selectable,  DriverSummary? driver,  VehicleSummary? vehicle,  ServiceSummary? service)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _RiderOfferComparison() when $default != null:
return $default(_that.rideRequestId,_that.driverUserId,_that.fare,_that.status,_that.createdAt,_that.updatedAt,_that.expiresAt,_that.decidedAt,_that.pickupDistanceMeters,_that.matchesProposedFare,_that.selectable,_that.driver,_that.vehicle,_that.service);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(name: 'ride_request_id')  String rideRequestId, @JsonKey(name: 'driver_user_id')  String driverUserId,  RideFare fare,  String status, @JsonKey(name: 'created_at')  DateTime createdAt, @JsonKey(name: 'updated_at')  DateTime updatedAt, @JsonKey(name: 'expires_at')  DateTime expiresAt, @JsonKey(name: 'decided_at')  DateTime? decidedAt, @JsonKey(name: 'pickup_distance_meters')  int? pickupDistanceMeters, @JsonKey(name: 'matches_proposed_fare')  bool matchesProposedFare,  bool selectable,  DriverSummary? driver,  VehicleSummary? vehicle,  ServiceSummary? service)  $default,) {final _that = this;
switch (_that) {
case _RiderOfferComparison():
return $default(_that.rideRequestId,_that.driverUserId,_that.fare,_that.status,_that.createdAt,_that.updatedAt,_that.expiresAt,_that.decidedAt,_that.pickupDistanceMeters,_that.matchesProposedFare,_that.selectable,_that.driver,_that.vehicle,_that.service);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(name: 'ride_request_id')  String rideRequestId, @JsonKey(name: 'driver_user_id')  String driverUserId,  RideFare fare,  String status, @JsonKey(name: 'created_at')  DateTime createdAt, @JsonKey(name: 'updated_at')  DateTime updatedAt, @JsonKey(name: 'expires_at')  DateTime expiresAt, @JsonKey(name: 'decided_at')  DateTime? decidedAt, @JsonKey(name: 'pickup_distance_meters')  int? pickupDistanceMeters, @JsonKey(name: 'matches_proposed_fare')  bool matchesProposedFare,  bool selectable,  DriverSummary? driver,  VehicleSummary? vehicle,  ServiceSummary? service)?  $default,) {final _that = this;
switch (_that) {
case _RiderOfferComparison() when $default != null:
return $default(_that.rideRequestId,_that.driverUserId,_that.fare,_that.status,_that.createdAt,_that.updatedAt,_that.expiresAt,_that.decidedAt,_that.pickupDistanceMeters,_that.matchesProposedFare,_that.selectable,_that.driver,_that.vehicle,_that.service);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _RiderOfferComparison implements RiderOfferComparison {
  const _RiderOfferComparison({@JsonKey(name: 'ride_request_id') required this.rideRequestId, @JsonKey(name: 'driver_user_id') required this.driverUserId, required this.fare, required this.status, @JsonKey(name: 'created_at') required this.createdAt, @JsonKey(name: 'updated_at') required this.updatedAt, @JsonKey(name: 'expires_at') required this.expiresAt, @JsonKey(name: 'decided_at') this.decidedAt, @JsonKey(name: 'pickup_distance_meters') this.pickupDistanceMeters, @JsonKey(name: 'matches_proposed_fare') required this.matchesProposedFare, required this.selectable, this.driver, this.vehicle, this.service});
  factory _RiderOfferComparison.fromJson(Map<String, dynamic> json) => _$RiderOfferComparisonFromJson(json);

@override@JsonKey(name: 'ride_request_id') final  String rideRequestId;
@override@JsonKey(name: 'driver_user_id') final  String driverUserId;
@override final  RideFare fare;
@override final  String status;
@override@JsonKey(name: 'created_at') final  DateTime createdAt;
@override@JsonKey(name: 'updated_at') final  DateTime updatedAt;
@override@JsonKey(name: 'expires_at') final  DateTime expiresAt;
@override@JsonKey(name: 'decided_at') final  DateTime? decidedAt;
@override@JsonKey(name: 'pickup_distance_meters') final  int? pickupDistanceMeters;
@override@JsonKey(name: 'matches_proposed_fare') final  bool matchesProposedFare;
@override final  bool selectable;
@override final  DriverSummary? driver;
@override final  VehicleSummary? vehicle;
@override final  ServiceSummary? service;

/// Create a copy of RiderOfferComparison
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$RiderOfferComparisonCopyWith<_RiderOfferComparison> get copyWith => __$RiderOfferComparisonCopyWithImpl<_RiderOfferComparison>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$RiderOfferComparisonToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _RiderOfferComparison&&(identical(other.rideRequestId, rideRequestId) || other.rideRequestId == rideRequestId)&&(identical(other.driverUserId, driverUserId) || other.driverUserId == driverUserId)&&(identical(other.fare, fare) || other.fare == fare)&&(identical(other.status, status) || other.status == status)&&(identical(other.createdAt, createdAt) || other.createdAt == createdAt)&&(identical(other.updatedAt, updatedAt) || other.updatedAt == updatedAt)&&(identical(other.expiresAt, expiresAt) || other.expiresAt == expiresAt)&&(identical(other.decidedAt, decidedAt) || other.decidedAt == decidedAt)&&(identical(other.pickupDistanceMeters, pickupDistanceMeters) || other.pickupDistanceMeters == pickupDistanceMeters)&&(identical(other.matchesProposedFare, matchesProposedFare) || other.matchesProposedFare == matchesProposedFare)&&(identical(other.selectable, selectable) || other.selectable == selectable)&&(identical(other.driver, driver) || other.driver == driver)&&(identical(other.vehicle, vehicle) || other.vehicle == vehicle)&&(identical(other.service, service) || other.service == service));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,rideRequestId,driverUserId,fare,status,createdAt,updatedAt,expiresAt,decidedAt,pickupDistanceMeters,matchesProposedFare,selectable,driver,vehicle,service);
}

@override
String toString() {
    return 'RiderOfferComparison(rideRequestId: $rideRequestId, driverUserId: $driverUserId, fare: $fare, status: $status, createdAt: $createdAt, updatedAt: $updatedAt, expiresAt: $expiresAt, decidedAt: $decidedAt, pickupDistanceMeters: $pickupDistanceMeters, matchesProposedFare: $matchesProposedFare, selectable: $selectable, driver: $driver, vehicle: $vehicle, service: $service)';
}


}

/// @nodoc
abstract mixin class _$RiderOfferComparisonCopyWith<$Res> implements $RiderOfferComparisonCopyWith<$Res> {
  factory _$RiderOfferComparisonCopyWith(_RiderOfferComparison value, $Res Function(_RiderOfferComparison) _then) = __$RiderOfferComparisonCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'ride_request_id') String rideRequestId,@JsonKey(name: 'driver_user_id') String driverUserId, RideFare fare, String status,@JsonKey(name: 'created_at') DateTime createdAt,@JsonKey(name: 'updated_at') DateTime updatedAt,@JsonKey(name: 'expires_at') DateTime expiresAt,@JsonKey(name: 'decided_at') DateTime? decidedAt,@JsonKey(name: 'pickup_distance_meters') int? pickupDistanceMeters,@JsonKey(name: 'matches_proposed_fare') bool matchesProposedFare, bool selectable, DriverSummary? driver, VehicleSummary? vehicle, ServiceSummary? service
});


@override $RideFareCopyWith<$Res> get fare;@override $DriverSummaryCopyWith<$Res>? get driver;@override $VehicleSummaryCopyWith<$Res>? get vehicle;@override $ServiceSummaryCopyWith<$Res>? get service;

}
/// @nodoc
class __$RiderOfferComparisonCopyWithImpl<$Res>
    implements _$RiderOfferComparisonCopyWith<$Res> {
  __$RiderOfferComparisonCopyWithImpl(this._self, this._then);

  final _RiderOfferComparison _self;
  final $Res Function(_RiderOfferComparison) _then;

/// Create a copy of RiderOfferComparison
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? rideRequestId = null,Object? driverUserId = null,Object? fare = null,Object? status = null,Object? createdAt = null,Object? updatedAt = null,Object? expiresAt = null,Object? decidedAt = freezed,Object? pickupDistanceMeters = freezed,Object? matchesProposedFare = null,Object? selectable = null,Object? driver = freezed,Object? vehicle = freezed,Object? service = freezed,}) {
  return _then(_RiderOfferComparison(
rideRequestId: null == rideRequestId ? _self.rideRequestId : rideRequestId // ignore: cast_nullable_to_non_nullable
as String,driverUserId: null == driverUserId ? _self.driverUserId : driverUserId // ignore: cast_nullable_to_non_nullable
as String,fare: null == fare ? _self.fare : fare // ignore: cast_nullable_to_non_nullable
as RideFare,status: null == status ? _self.status : status // ignore: cast_nullable_to_non_nullable
as String,createdAt: null == createdAt ? _self.createdAt : createdAt // ignore: cast_nullable_to_non_nullable
as DateTime,updatedAt: null == updatedAt ? _self.updatedAt : updatedAt // ignore: cast_nullable_to_non_nullable
as DateTime,expiresAt: null == expiresAt ? _self.expiresAt : expiresAt // ignore: cast_nullable_to_non_nullable
as DateTime,decidedAt: freezed == decidedAt ? _self.decidedAt : decidedAt // ignore: cast_nullable_to_non_nullable
as DateTime?,pickupDistanceMeters: freezed == pickupDistanceMeters ? _self.pickupDistanceMeters : pickupDistanceMeters // ignore: cast_nullable_to_non_nullable
as int?,matchesProposedFare: null == matchesProposedFare ? _self.matchesProposedFare : matchesProposedFare // ignore: cast_nullable_to_non_nullable
as bool,selectable: null == selectable ? _self.selectable : selectable // ignore: cast_nullable_to_non_nullable
as bool,driver: freezed == driver ? _self.driver : driver // ignore: cast_nullable_to_non_nullable
as DriverSummary?,vehicle: freezed == vehicle ? _self.vehicle : vehicle // ignore: cast_nullable_to_non_nullable
as VehicleSummary?,service: freezed == service ? _self.service : service // ignore: cast_nullable_to_non_nullable
as ServiceSummary?,
  ));
}

/// Create a copy of RiderOfferComparison
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$RideFareCopyWith<$Res> get fare {
  
  return $RideFareCopyWith<$Res>(_self.fare, (value) {
    return _then(_self.copyWith(fare: value));
  });
}/// Create a copy of RiderOfferComparison
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$DriverSummaryCopyWith<$Res>? get driver {
    if (_self.driver == null) {
    return null;
  }

  return $DriverSummaryCopyWith<$Res>(_self.driver!, (value) {
    return _then(_self.copyWith(driver: value));
  });
}/// Create a copy of RiderOfferComparison
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$VehicleSummaryCopyWith<$Res>? get vehicle {
    if (_self.vehicle == null) {
    return null;
  }

  return $VehicleSummaryCopyWith<$Res>(_self.vehicle!, (value) {
    return _then(_self.copyWith(vehicle: value));
  });
}/// Create a copy of RiderOfferComparison
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$ServiceSummaryCopyWith<$Res>? get service {
    if (_self.service == null) {
    return null;
  }

  return $ServiceSummaryCopyWith<$Res>(_self.service!, (value) {
    return _then(_self.copyWith(service: value));
  });
}
}

// dart format on
