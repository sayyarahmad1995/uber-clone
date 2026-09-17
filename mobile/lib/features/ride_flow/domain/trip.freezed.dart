// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'trip.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$TripOperationFare {

@JsonKey(name: 'amount_minor') int get amountMinor; String get currency;
/// Create a copy of TripOperationFare
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$TripOperationFareCopyWith<TripOperationFare> get copyWith => _$TripOperationFareCopyWithImpl<TripOperationFare>(this as TripOperationFare, _$identity);

  /// Serializes this TripOperationFare to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as TripOperationFare;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is TripOperationFare&&(identical(other.amountMinor, _this.amountMinor) || other.amountMinor == _this.amountMinor)&&(identical(other.currency, _this.currency) || other.currency == _this.currency));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as TripOperationFare;
  return Object.hash(runtimeType,_this.amountMinor,_this.currency);
}

@override
String toString() {
  final _this = this as TripOperationFare;
  return 'TripOperationFare(amountMinor: ${_this.amountMinor}, currency: ${_this.currency})';
}


}

/// @nodoc
abstract mixin class $TripOperationFareCopyWith<$Res>  {
  factory $TripOperationFareCopyWith(TripOperationFare value, $Res Function(TripOperationFare) _then) = _$TripOperationFareCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'amount_minor') int amountMinor, String currency
});




}
/// @nodoc
class _$TripOperationFareCopyWithImpl<$Res>
    implements $TripOperationFareCopyWith<$Res> {
  _$TripOperationFareCopyWithImpl(this._self, this._then);

  final TripOperationFare _self;
  final $Res Function(TripOperationFare) _then;

/// Create a copy of TripOperationFare
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? amountMinor = null,Object? currency = null,}) {
  return _then(TripOperationFare(
amountMinor: null == amountMinor ? _self.amountMinor : amountMinor // ignore: cast_nullable_to_non_nullable
as int,currency: null == currency ? _self.currency : currency // ignore: cast_nullable_to_non_nullable
as String,
  ));
}

}


/// Adds pattern-matching-related methods to [TripOperationFare].
extension TripOperationFarePatterns on TripOperationFare {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _TripOperationFare value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _TripOperationFare() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _TripOperationFare value)  $default,){
final _that = this;
switch (_that) {
case _TripOperationFare():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _TripOperationFare value)?  $default,){
final _that = this;
switch (_that) {
case _TripOperationFare() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(name: 'amount_minor')  int amountMinor,  String currency)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _TripOperationFare() when $default != null:
return $default(_that.amountMinor,_that.currency);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(name: 'amount_minor')  int amountMinor,  String currency)  $default,) {final _that = this;
switch (_that) {
case _TripOperationFare():
return $default(_that.amountMinor,_that.currency);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(name: 'amount_minor')  int amountMinor,  String currency)?  $default,) {final _that = this;
switch (_that) {
case _TripOperationFare() when $default != null:
return $default(_that.amountMinor,_that.currency);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _TripOperationFare implements TripOperationFare {
  const _TripOperationFare({@JsonKey(name: 'amount_minor') required this.amountMinor, required this.currency});
  factory _TripOperationFare.fromJson(Map<String, dynamic> json) => _$TripOperationFareFromJson(json);

@override@JsonKey(name: 'amount_minor') final  int amountMinor;
@override final  String currency;

/// Create a copy of TripOperationFare
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$TripOperationFareCopyWith<_TripOperationFare> get copyWith => __$TripOperationFareCopyWithImpl<_TripOperationFare>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$TripOperationFareToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _TripOperationFare&&(identical(other.amountMinor, amountMinor) || other.amountMinor == amountMinor)&&(identical(other.currency, currency) || other.currency == currency));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,amountMinor,currency);
}

@override
String toString() {
    return 'TripOperationFare(amountMinor: $amountMinor, currency: $currency)';
}


}

/// @nodoc
abstract mixin class _$TripOperationFareCopyWith<$Res> implements $TripOperationFareCopyWith<$Res> {
  factory _$TripOperationFareCopyWith(_TripOperationFare value, $Res Function(_TripOperationFare) _then) = __$TripOperationFareCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'amount_minor') int amountMinor, String currency
});




}
/// @nodoc
class __$TripOperationFareCopyWithImpl<$Res>
    implements _$TripOperationFareCopyWith<$Res> {
  __$TripOperationFareCopyWithImpl(this._self, this._then);

  final _TripOperationFare _self;
  final $Res Function(_TripOperationFare) _then;

/// Create a copy of TripOperationFare
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? amountMinor = null,Object? currency = null,}) {
  return _then(_TripOperationFare(
amountMinor: null == amountMinor ? _self.amountMinor : amountMinor // ignore: cast_nullable_to_non_nullable
as int,currency: null == currency ? _self.currency : currency // ignore: cast_nullable_to_non_nullable
as String,
  ));
}


}


/// @nodoc
mixin _$TripOperationContext {

@JsonKey(name: 'vehicle_id') String get vehicleId;@JsonKey(name: 'service_code') String get serviceCode;@JsonKey(name: 'service_name') String get serviceName;@JsonKey(name: 'driver_name') String get driverName; String get make; String get model;@JsonKey(name: 'model_year') int get modelYear; String get color;@JsonKey(name: 'license_plate') String get licensePlate; TripOperationFare get fare;
/// Create a copy of TripOperationContext
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$TripOperationContextCopyWith<TripOperationContext> get copyWith => _$TripOperationContextCopyWithImpl<TripOperationContext>(this as TripOperationContext, _$identity);

  /// Serializes this TripOperationContext to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as TripOperationContext;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is TripOperationContext&&(identical(other.vehicleId, _this.vehicleId) || other.vehicleId == _this.vehicleId)&&(identical(other.serviceCode, _this.serviceCode) || other.serviceCode == _this.serviceCode)&&(identical(other.serviceName, _this.serviceName) || other.serviceName == _this.serviceName)&&(identical(other.driverName, _this.driverName) || other.driverName == _this.driverName)&&(identical(other.make, _this.make) || other.make == _this.make)&&(identical(other.model, _this.model) || other.model == _this.model)&&(identical(other.modelYear, _this.modelYear) || other.modelYear == _this.modelYear)&&(identical(other.color, _this.color) || other.color == _this.color)&&(identical(other.licensePlate, _this.licensePlate) || other.licensePlate == _this.licensePlate)&&(identical(other.fare, _this.fare) || other.fare == _this.fare));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as TripOperationContext;
  return Object.hash(runtimeType,_this.vehicleId,_this.serviceCode,_this.serviceName,_this.driverName,_this.make,_this.model,_this.modelYear,_this.color,_this.licensePlate,_this.fare);
}

@override
String toString() {
  final _this = this as TripOperationContext;
  return 'TripOperationContext(vehicleId: ${_this.vehicleId}, serviceCode: ${_this.serviceCode}, serviceName: ${_this.serviceName}, driverName: ${_this.driverName}, make: ${_this.make}, model: ${_this.model}, modelYear: ${_this.modelYear}, color: ${_this.color}, licensePlate: ${_this.licensePlate}, fare: ${_this.fare})';
}


}

/// @nodoc
abstract mixin class $TripOperationContextCopyWith<$Res>  {
  factory $TripOperationContextCopyWith(TripOperationContext value, $Res Function(TripOperationContext) _then) = _$TripOperationContextCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'vehicle_id') String vehicleId,@JsonKey(name: 'service_code') String serviceCode,@JsonKey(name: 'service_name') String serviceName,@JsonKey(name: 'driver_name') String driverName, String make, String model,@JsonKey(name: 'model_year') int modelYear, String color,@JsonKey(name: 'license_plate') String licensePlate, TripOperationFare fare
});


$TripOperationFareCopyWith<$Res> get fare;

}
/// @nodoc
class _$TripOperationContextCopyWithImpl<$Res>
    implements $TripOperationContextCopyWith<$Res> {
  _$TripOperationContextCopyWithImpl(this._self, this._then);

  final TripOperationContext _self;
  final $Res Function(TripOperationContext) _then;

/// Create a copy of TripOperationContext
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? vehicleId = null,Object? serviceCode = null,Object? serviceName = null,Object? driverName = null,Object? make = null,Object? model = null,Object? modelYear = null,Object? color = null,Object? licensePlate = null,Object? fare = null,}) {
  return _then(TripOperationContext(
vehicleId: null == vehicleId ? _self.vehicleId : vehicleId // ignore: cast_nullable_to_non_nullable
as String,serviceCode: null == serviceCode ? _self.serviceCode : serviceCode // ignore: cast_nullable_to_non_nullable
as String,serviceName: null == serviceName ? _self.serviceName : serviceName // ignore: cast_nullable_to_non_nullable
as String,driverName: null == driverName ? _self.driverName : driverName // ignore: cast_nullable_to_non_nullable
as String,make: null == make ? _self.make : make // ignore: cast_nullable_to_non_nullable
as String,model: null == model ? _self.model : model // ignore: cast_nullable_to_non_nullable
as String,modelYear: null == modelYear ? _self.modelYear : modelYear // ignore: cast_nullable_to_non_nullable
as int,color: null == color ? _self.color : color // ignore: cast_nullable_to_non_nullable
as String,licensePlate: null == licensePlate ? _self.licensePlate : licensePlate // ignore: cast_nullable_to_non_nullable
as String,fare: null == fare ? _self.fare : fare // ignore: cast_nullable_to_non_nullable
as TripOperationFare,
  ));
}
/// Create a copy of TripOperationContext
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$TripOperationFareCopyWith<$Res> get fare {

  return $TripOperationFareCopyWith<$Res>(_self.fare, (value) {
    return _then(_self.copyWith(fare: value));
  });
}
}


/// Adds pattern-matching-related methods to [TripOperationContext].
extension TripOperationContextPatterns on TripOperationContext {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _TripOperationContext value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _TripOperationContext() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _TripOperationContext value)  $default,){
final _that = this;
switch (_that) {
case _TripOperationContext():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _TripOperationContext value)?  $default,){
final _that = this;
switch (_that) {
case _TripOperationContext() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(name: 'vehicle_id')  String vehicleId, @JsonKey(name: 'service_code')  String serviceCode, @JsonKey(name: 'service_name')  String serviceName, @JsonKey(name: 'driver_name')  String driverName,  String make,  String model, @JsonKey(name: 'model_year')  int modelYear,  String color, @JsonKey(name: 'license_plate')  String licensePlate,  TripOperationFare fare)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _TripOperationContext() when $default != null:
return $default(_that.vehicleId,_that.serviceCode,_that.serviceName,_that.driverName,_that.make,_that.model,_that.modelYear,_that.color,_that.licensePlate,_that.fare);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(name: 'vehicle_id')  String vehicleId, @JsonKey(name: 'service_code')  String serviceCode, @JsonKey(name: 'service_name')  String serviceName, @JsonKey(name: 'driver_name')  String driverName,  String make,  String model, @JsonKey(name: 'model_year')  int modelYear,  String color, @JsonKey(name: 'license_plate')  String licensePlate,  TripOperationFare fare)  $default,) {final _that = this;
switch (_that) {
case _TripOperationContext():
return $default(_that.vehicleId,_that.serviceCode,_that.serviceName,_that.driverName,_that.make,_that.model,_that.modelYear,_that.color,_that.licensePlate,_that.fare);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(name: 'vehicle_id')  String vehicleId, @JsonKey(name: 'service_code')  String serviceCode, @JsonKey(name: 'service_name')  String serviceName, @JsonKey(name: 'driver_name')  String driverName,  String make,  String model, @JsonKey(name: 'model_year')  int modelYear,  String color, @JsonKey(name: 'license_plate')  String licensePlate,  TripOperationFare fare)?  $default,) {final _that = this;
switch (_that) {
case _TripOperationContext() when $default != null:
return $default(_that.vehicleId,_that.serviceCode,_that.serviceName,_that.driverName,_that.make,_that.model,_that.modelYear,_that.color,_that.licensePlate,_that.fare);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _TripOperationContext implements TripOperationContext {
  const _TripOperationContext({@JsonKey(name: 'vehicle_id') required this.vehicleId, @JsonKey(name: 'service_code') required this.serviceCode, @JsonKey(name: 'service_name') required this.serviceName, @JsonKey(name: 'driver_name') required this.driverName, required this.make, required this.model, @JsonKey(name: 'model_year') required this.modelYear, required this.color, @JsonKey(name: 'license_plate') required this.licensePlate, required this.fare});
  factory _TripOperationContext.fromJson(Map<String, dynamic> json) => _$TripOperationContextFromJson(json);

@override@JsonKey(name: 'vehicle_id') final  String vehicleId;
@override@JsonKey(name: 'service_code') final  String serviceCode;
@override@JsonKey(name: 'service_name') final  String serviceName;
@override@JsonKey(name: 'driver_name') final  String driverName;
@override final  String make;
@override final  String model;
@override@JsonKey(name: 'model_year') final  int modelYear;
@override final  String color;
@override@JsonKey(name: 'license_plate') final  String licensePlate;
@override final  TripOperationFare fare;

/// Create a copy of TripOperationContext
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$TripOperationContextCopyWith<_TripOperationContext> get copyWith => __$TripOperationContextCopyWithImpl<_TripOperationContext>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$TripOperationContextToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _TripOperationContext&&(identical(other.vehicleId, vehicleId) || other.vehicleId == vehicleId)&&(identical(other.serviceCode, serviceCode) || other.serviceCode == serviceCode)&&(identical(other.serviceName, serviceName) || other.serviceName == serviceName)&&(identical(other.driverName, driverName) || other.driverName == driverName)&&(identical(other.make, make) || other.make == make)&&(identical(other.model, model) || other.model == model)&&(identical(other.modelYear, modelYear) || other.modelYear == modelYear)&&(identical(other.color, color) || other.color == color)&&(identical(other.licensePlate, licensePlate) || other.licensePlate == licensePlate)&&(identical(other.fare, fare) || other.fare == fare));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,vehicleId,serviceCode,serviceName,driverName,make,model,modelYear,color,licensePlate,fare);
}

@override
String toString() {
    return 'TripOperationContext(vehicleId: $vehicleId, serviceCode: $serviceCode, serviceName: $serviceName, driverName: $driverName, make: $make, model: $model, modelYear: $modelYear, color: $color, licensePlate: $licensePlate, fare: $fare)';
}


}

/// @nodoc
abstract mixin class _$TripOperationContextCopyWith<$Res> implements $TripOperationContextCopyWith<$Res> {
  factory _$TripOperationContextCopyWith(_TripOperationContext value, $Res Function(_TripOperationContext) _then) = __$TripOperationContextCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'vehicle_id') String vehicleId,@JsonKey(name: 'service_code') String serviceCode,@JsonKey(name: 'service_name') String serviceName,@JsonKey(name: 'driver_name') String driverName, String make, String model,@JsonKey(name: 'model_year') int modelYear, String color,@JsonKey(name: 'license_plate') String licensePlate, TripOperationFare fare
});


@override $TripOperationFareCopyWith<$Res> get fare;

}
/// @nodoc
class __$TripOperationContextCopyWithImpl<$Res>
    implements _$TripOperationContextCopyWith<$Res> {
  __$TripOperationContextCopyWithImpl(this._self, this._then);

  final _TripOperationContext _self;
  final $Res Function(_TripOperationContext) _then;

/// Create a copy of TripOperationContext
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? vehicleId = null,Object? serviceCode = null,Object? serviceName = null,Object? driverName = null,Object? make = null,Object? model = null,Object? modelYear = null,Object? color = null,Object? licensePlate = null,Object? fare = null,}) {
  return _then(_TripOperationContext(
vehicleId: null == vehicleId ? _self.vehicleId : vehicleId // ignore: cast_nullable_to_non_nullable
as String,serviceCode: null == serviceCode ? _self.serviceCode : serviceCode // ignore: cast_nullable_to_non_nullable
as String,serviceName: null == serviceName ? _self.serviceName : serviceName // ignore: cast_nullable_to_non_nullable
as String,driverName: null == driverName ? _self.driverName : driverName // ignore: cast_nullable_to_non_nullable
as String,make: null == make ? _self.make : make // ignore: cast_nullable_to_non_nullable
as String,model: null == model ? _self.model : model // ignore: cast_nullable_to_non_nullable
as String,modelYear: null == modelYear ? _self.modelYear : modelYear // ignore: cast_nullable_to_non_nullable
as int,color: null == color ? _self.color : color // ignore: cast_nullable_to_non_nullable
as String,licensePlate: null == licensePlate ? _self.licensePlate : licensePlate // ignore: cast_nullable_to_non_nullable
as String,fare: null == fare ? _self.fare : fare // ignore: cast_nullable_to_non_nullable
as TripOperationFare,
  ));
}

/// Create a copy of TripOperationContext
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$TripOperationFareCopyWith<$Res> get fare {

  return $TripOperationFareCopyWith<$Res>(_self.fare, (value) {
    return _then(_self.copyWith(fare: value));
  });
}
}


/// @nodoc
mixin _$TripSnapshot {

@JsonKey(name: 'operation_context') TripOperationContext? get operationContext;@JsonKey(name: 'ride_request_id') String? get rideRequestId;@JsonKey(name: 'rider_user_id') String? get riderUserId;@JsonKey(name: 'driver_user_id') String? get driverUserId; RideLocation? get pickup; RideLocation? get destination; String get status;@JsonKey(name: 'assigned_at') DateTime get assignedAt;@JsonKey(name: 'started_at') DateTime? get startedAt;@JsonKey(name: 'completed_at') DateTime? get completedAt;@JsonKey(name: 'cancelled_at') DateTime? get cancelledAt; SettlementSnapshot get settlement;
/// Create a copy of TripSnapshot
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$TripSnapshotCopyWith<TripSnapshot> get copyWith => _$TripSnapshotCopyWithImpl<TripSnapshot>(this as TripSnapshot, _$identity);

  /// Serializes this TripSnapshot to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as TripSnapshot;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is TripSnapshot&&(identical(other.operationContext, _this.operationContext) || other.operationContext == _this.operationContext)&&(identical(other.rideRequestId, _this.rideRequestId) || other.rideRequestId == _this.rideRequestId)&&(identical(other.riderUserId, _this.riderUserId) || other.riderUserId == _this.riderUserId)&&(identical(other.driverUserId, _this.driverUserId) || other.driverUserId == _this.driverUserId)&&(identical(other.pickup, _this.pickup) || other.pickup == _this.pickup)&&(identical(other.destination, _this.destination) || other.destination == _this.destination)&&(identical(other.status, _this.status) || other.status == _this.status)&&(identical(other.assignedAt, _this.assignedAt) || other.assignedAt == _this.assignedAt)&&(identical(other.startedAt, _this.startedAt) || other.startedAt == _this.startedAt)&&(identical(other.completedAt, _this.completedAt) || other.completedAt == _this.completedAt)&&(identical(other.cancelledAt, _this.cancelledAt) || other.cancelledAt == _this.cancelledAt)&&(identical(other.settlement, _this.settlement) || other.settlement == _this.settlement));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as TripSnapshot;
  return Object.hash(runtimeType,_this.operationContext,_this.rideRequestId,_this.riderUserId,_this.driverUserId,_this.pickup,_this.destination,_this.status,_this.assignedAt,_this.startedAt,_this.completedAt,_this.cancelledAt,_this.settlement);
}

@override
String toString() {
  final _this = this as TripSnapshot;
  return 'TripSnapshot(operationContext: ${_this.operationContext}, rideRequestId: ${_this.rideRequestId}, riderUserId: ${_this.riderUserId}, driverUserId: ${_this.driverUserId}, pickup: ${_this.pickup}, destination: ${_this.destination}, status: ${_this.status}, assignedAt: ${_this.assignedAt}, startedAt: ${_this.startedAt}, completedAt: ${_this.completedAt}, cancelledAt: ${_this.cancelledAt}, settlement: ${_this.settlement})';
}


}

/// @nodoc
abstract mixin class $TripSnapshotCopyWith<$Res>  {
  factory $TripSnapshotCopyWith(TripSnapshot value, $Res Function(TripSnapshot) _then) = _$TripSnapshotCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'operation_context') TripOperationContext? operationContext,@JsonKey(name: 'ride_request_id') String? rideRequestId,@JsonKey(name: 'rider_user_id') String? riderUserId,@JsonKey(name: 'driver_user_id') String? driverUserId, RideLocation? pickup, RideLocation? destination, String status,@JsonKey(name: 'assigned_at') DateTime assignedAt,@JsonKey(name: 'started_at') DateTime? startedAt,@JsonKey(name: 'completed_at') DateTime? completedAt,@JsonKey(name: 'cancelled_at') DateTime? cancelledAt, SettlementSnapshot settlement
});


$TripOperationContextCopyWith<$Res>? get operationContext;$RideLocationCopyWith<$Res>? get pickup;$RideLocationCopyWith<$Res>? get destination;$SettlementSnapshotCopyWith<$Res> get settlement;

}
/// @nodoc
class _$TripSnapshotCopyWithImpl<$Res>
    implements $TripSnapshotCopyWith<$Res> {
  _$TripSnapshotCopyWithImpl(this._self, this._then);

  final TripSnapshot _self;
  final $Res Function(TripSnapshot) _then;

/// Create a copy of TripSnapshot
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? operationContext = freezed,Object? rideRequestId = freezed,Object? riderUserId = freezed,Object? driverUserId = freezed,Object? pickup = freezed,Object? destination = freezed,Object? status = null,Object? assignedAt = null,Object? startedAt = freezed,Object? completedAt = freezed,Object? cancelledAt = freezed,Object? settlement = null,}) {
  return _then(TripSnapshot(
operationContext: freezed == operationContext ? _self.operationContext : operationContext // ignore: cast_nullable_to_non_nullable
as TripOperationContext?,rideRequestId: freezed == rideRequestId ? _self.rideRequestId : rideRequestId // ignore: cast_nullable_to_non_nullable
as String?,riderUserId: freezed == riderUserId ? _self.riderUserId : riderUserId // ignore: cast_nullable_to_non_nullable
as String?,driverUserId: freezed == driverUserId ? _self.driverUserId : driverUserId // ignore: cast_nullable_to_non_nullable
as String?,pickup: freezed == pickup ? _self.pickup : pickup // ignore: cast_nullable_to_non_nullable
as RideLocation?,destination: freezed == destination ? _self.destination : destination // ignore: cast_nullable_to_non_nullable
as RideLocation?,status: null == status ? _self.status : status // ignore: cast_nullable_to_non_nullable
as String,assignedAt: null == assignedAt ? _self.assignedAt : assignedAt // ignore: cast_nullable_to_non_nullable
as DateTime,startedAt: freezed == startedAt ? _self.startedAt : startedAt // ignore: cast_nullable_to_non_nullable
as DateTime?,completedAt: freezed == completedAt ? _self.completedAt : completedAt // ignore: cast_nullable_to_non_nullable
as DateTime?,cancelledAt: freezed == cancelledAt ? _self.cancelledAt : cancelledAt // ignore: cast_nullable_to_non_nullable
as DateTime?,settlement: null == settlement ? _self.settlement : settlement // ignore: cast_nullable_to_non_nullable
as SettlementSnapshot,
  ));
}
/// Create a copy of TripSnapshot
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$TripOperationContextCopyWith<$Res>? get operationContext {
    if (_self.operationContext == null) {
    return null;
  }

  return $TripOperationContextCopyWith<$Res>(_self.operationContext!, (value) {
    return _then(_self.copyWith(operationContext: value));
  });
}/// Create a copy of TripSnapshot
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$RideLocationCopyWith<$Res>? get pickup {
    if (_self.pickup == null) {
    return null;
  }

  return $RideLocationCopyWith<$Res>(_self.pickup!, (value) {
    return _then(_self.copyWith(pickup: value));
  });
}/// Create a copy of TripSnapshot
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$RideLocationCopyWith<$Res>? get destination {
    if (_self.destination == null) {
    return null;
  }

  return $RideLocationCopyWith<$Res>(_self.destination!, (value) {
    return _then(_self.copyWith(destination: value));
  });
}/// Create a copy of TripSnapshot
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$SettlementSnapshotCopyWith<$Res> get settlement {

  return $SettlementSnapshotCopyWith<$Res>(_self.settlement, (value) {
    return _then(_self.copyWith(settlement: value));
  });
}
}


/// Adds pattern-matching-related methods to [TripSnapshot].
extension TripSnapshotPatterns on TripSnapshot {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _TripSnapshot value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _TripSnapshot() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _TripSnapshot value)  $default,){
final _that = this;
switch (_that) {
case _TripSnapshot():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _TripSnapshot value)?  $default,){
final _that = this;
switch (_that) {
case _TripSnapshot() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(name: 'operation_context')  TripOperationContext? operationContext, @JsonKey(name: 'ride_request_id')  String? rideRequestId, @JsonKey(name: 'rider_user_id')  String? riderUserId, @JsonKey(name: 'driver_user_id')  String? driverUserId,  RideLocation? pickup,  RideLocation? destination,  String status, @JsonKey(name: 'assigned_at')  DateTime assignedAt, @JsonKey(name: 'started_at')  DateTime? startedAt, @JsonKey(name: 'completed_at')  DateTime? completedAt, @JsonKey(name: 'cancelled_at')  DateTime? cancelledAt,  SettlementSnapshot settlement)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _TripSnapshot() when $default != null:
return $default(_that.operationContext,_that.rideRequestId,_that.riderUserId,_that.driverUserId,_that.pickup,_that.destination,_that.status,_that.assignedAt,_that.startedAt,_that.completedAt,_that.cancelledAt,_that.settlement);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(name: 'operation_context')  TripOperationContext? operationContext, @JsonKey(name: 'ride_request_id')  String? rideRequestId, @JsonKey(name: 'rider_user_id')  String? riderUserId, @JsonKey(name: 'driver_user_id')  String? driverUserId,  RideLocation? pickup,  RideLocation? destination,  String status, @JsonKey(name: 'assigned_at')  DateTime assignedAt, @JsonKey(name: 'started_at')  DateTime? startedAt, @JsonKey(name: 'completed_at')  DateTime? completedAt, @JsonKey(name: 'cancelled_at')  DateTime? cancelledAt,  SettlementSnapshot settlement)  $default,) {final _that = this;
switch (_that) {
case _TripSnapshot():
return $default(_that.operationContext,_that.rideRequestId,_that.riderUserId,_that.driverUserId,_that.pickup,_that.destination,_that.status,_that.assignedAt,_that.startedAt,_that.completedAt,_that.cancelledAt,_that.settlement);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(name: 'operation_context')  TripOperationContext? operationContext, @JsonKey(name: 'ride_request_id')  String? rideRequestId, @JsonKey(name: 'rider_user_id')  String? riderUserId, @JsonKey(name: 'driver_user_id')  String? driverUserId,  RideLocation? pickup,  RideLocation? destination,  String status, @JsonKey(name: 'assigned_at')  DateTime assignedAt, @JsonKey(name: 'started_at')  DateTime? startedAt, @JsonKey(name: 'completed_at')  DateTime? completedAt, @JsonKey(name: 'cancelled_at')  DateTime? cancelledAt,  SettlementSnapshot settlement)?  $default,) {final _that = this;
switch (_that) {
case _TripSnapshot() when $default != null:
return $default(_that.operationContext,_that.rideRequestId,_that.riderUserId,_that.driverUserId,_that.pickup,_that.destination,_that.status,_that.assignedAt,_that.startedAt,_that.completedAt,_that.cancelledAt,_that.settlement);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _TripSnapshot implements TripSnapshot {
  const _TripSnapshot({@JsonKey(name: 'operation_context') this.operationContext, @JsonKey(name: 'ride_request_id') this.rideRequestId, @JsonKey(name: 'rider_user_id') this.riderUserId, @JsonKey(name: 'driver_user_id') this.driverUserId, this.pickup, this.destination, required this.status, @JsonKey(name: 'assigned_at') required this.assignedAt, @JsonKey(name: 'started_at') this.startedAt, @JsonKey(name: 'completed_at') this.completedAt, @JsonKey(name: 'cancelled_at') this.cancelledAt, required this.settlement});
  factory _TripSnapshot.fromJson(Map<String, dynamic> json) => _$TripSnapshotFromJson(json);

@override@JsonKey(name: 'operation_context') final  TripOperationContext? operationContext;
@override@JsonKey(name: 'ride_request_id') final  String? rideRequestId;
@override@JsonKey(name: 'rider_user_id') final  String? riderUserId;
@override@JsonKey(name: 'driver_user_id') final  String? driverUserId;
@override final  RideLocation? pickup;
@override final  RideLocation? destination;
@override final  String status;
@override@JsonKey(name: 'assigned_at') final  DateTime assignedAt;
@override@JsonKey(name: 'started_at') final  DateTime? startedAt;
@override@JsonKey(name: 'completed_at') final  DateTime? completedAt;
@override@JsonKey(name: 'cancelled_at') final  DateTime? cancelledAt;
@override final  SettlementSnapshot settlement;

/// Create a copy of TripSnapshot
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$TripSnapshotCopyWith<_TripSnapshot> get copyWith => __$TripSnapshotCopyWithImpl<_TripSnapshot>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$TripSnapshotToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _TripSnapshot&&(identical(other.operationContext, operationContext) || other.operationContext == operationContext)&&(identical(other.rideRequestId, rideRequestId) || other.rideRequestId == rideRequestId)&&(identical(other.riderUserId, riderUserId) || other.riderUserId == riderUserId)&&(identical(other.driverUserId, driverUserId) || other.driverUserId == driverUserId)&&(identical(other.pickup, pickup) || other.pickup == pickup)&&(identical(other.destination, destination) || other.destination == destination)&&(identical(other.status, status) || other.status == status)&&(identical(other.assignedAt, assignedAt) || other.assignedAt == assignedAt)&&(identical(other.startedAt, startedAt) || other.startedAt == startedAt)&&(identical(other.completedAt, completedAt) || other.completedAt == completedAt)&&(identical(other.cancelledAt, cancelledAt) || other.cancelledAt == cancelledAt)&&(identical(other.settlement, settlement) || other.settlement == settlement));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,operationContext,rideRequestId,riderUserId,driverUserId,pickup,destination,status,assignedAt,startedAt,completedAt,cancelledAt,settlement);
}

@override
String toString() {
    return 'TripSnapshot(operationContext: $operationContext, rideRequestId: $rideRequestId, riderUserId: $riderUserId, driverUserId: $driverUserId, pickup: $pickup, destination: $destination, status: $status, assignedAt: $assignedAt, startedAt: $startedAt, completedAt: $completedAt, cancelledAt: $cancelledAt, settlement: $settlement)';
}


}

/// @nodoc
abstract mixin class _$TripSnapshotCopyWith<$Res> implements $TripSnapshotCopyWith<$Res> {
  factory _$TripSnapshotCopyWith(_TripSnapshot value, $Res Function(_TripSnapshot) _then) = __$TripSnapshotCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'operation_context') TripOperationContext? operationContext,@JsonKey(name: 'ride_request_id') String? rideRequestId,@JsonKey(name: 'rider_user_id') String? riderUserId,@JsonKey(name: 'driver_user_id') String? driverUserId, RideLocation? pickup, RideLocation? destination, String status,@JsonKey(name: 'assigned_at') DateTime assignedAt,@JsonKey(name: 'started_at') DateTime? startedAt,@JsonKey(name: 'completed_at') DateTime? completedAt,@JsonKey(name: 'cancelled_at') DateTime? cancelledAt, SettlementSnapshot settlement
});


@override $TripOperationContextCopyWith<$Res>? get operationContext;@override $RideLocationCopyWith<$Res>? get pickup;@override $RideLocationCopyWith<$Res>? get destination;@override $SettlementSnapshotCopyWith<$Res> get settlement;

}
/// @nodoc
class __$TripSnapshotCopyWithImpl<$Res>
    implements _$TripSnapshotCopyWith<$Res> {
  __$TripSnapshotCopyWithImpl(this._self, this._then);

  final _TripSnapshot _self;
  final $Res Function(_TripSnapshot) _then;

/// Create a copy of TripSnapshot
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? operationContext = freezed,Object? rideRequestId = freezed,Object? riderUserId = freezed,Object? driverUserId = freezed,Object? pickup = freezed,Object? destination = freezed,Object? status = null,Object? assignedAt = null,Object? startedAt = freezed,Object? completedAt = freezed,Object? cancelledAt = freezed,Object? settlement = null,}) {
  return _then(_TripSnapshot(
operationContext: freezed == operationContext ? _self.operationContext : operationContext // ignore: cast_nullable_to_non_nullable
as TripOperationContext?,rideRequestId: freezed == rideRequestId ? _self.rideRequestId : rideRequestId // ignore: cast_nullable_to_non_nullable
as String?,riderUserId: freezed == riderUserId ? _self.riderUserId : riderUserId // ignore: cast_nullable_to_non_nullable
as String?,driverUserId: freezed == driverUserId ? _self.driverUserId : driverUserId // ignore: cast_nullable_to_non_nullable
as String?,pickup: freezed == pickup ? _self.pickup : pickup // ignore: cast_nullable_to_non_nullable
as RideLocation?,destination: freezed == destination ? _self.destination : destination // ignore: cast_nullable_to_non_nullable
as RideLocation?,status: null == status ? _self.status : status // ignore: cast_nullable_to_non_nullable
as String,assignedAt: null == assignedAt ? _self.assignedAt : assignedAt // ignore: cast_nullable_to_non_nullable
as DateTime,startedAt: freezed == startedAt ? _self.startedAt : startedAt // ignore: cast_nullable_to_non_nullable
as DateTime?,completedAt: freezed == completedAt ? _self.completedAt : completedAt // ignore: cast_nullable_to_non_nullable
as DateTime?,cancelledAt: freezed == cancelledAt ? _self.cancelledAt : cancelledAt // ignore: cast_nullable_to_non_nullable
as DateTime?,settlement: null == settlement ? _self.settlement : settlement // ignore: cast_nullable_to_non_nullable
as SettlementSnapshot,
  ));
}

/// Create a copy of TripSnapshot
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$TripOperationContextCopyWith<$Res>? get operationContext {
    if (_self.operationContext == null) {
    return null;
  }

  return $TripOperationContextCopyWith<$Res>(_self.operationContext!, (value) {
    return _then(_self.copyWith(operationContext: value));
  });
}/// Create a copy of TripSnapshot
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$RideLocationCopyWith<$Res>? get pickup {
    if (_self.pickup == null) {
    return null;
  }

  return $RideLocationCopyWith<$Res>(_self.pickup!, (value) {
    return _then(_self.copyWith(pickup: value));
  });
}/// Create a copy of TripSnapshot
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$RideLocationCopyWith<$Res>? get destination {
    if (_self.destination == null) {
    return null;
  }

  return $RideLocationCopyWith<$Res>(_self.destination!, (value) {
    return _then(_self.copyWith(destination: value));
  });
}/// Create a copy of TripSnapshot
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$SettlementSnapshotCopyWith<$Res> get settlement {

  return $SettlementSnapshotCopyWith<$Res>(_self.settlement, (value) {
    return _then(_self.copyWith(settlement: value));
  });
}
}


/// @nodoc
mixin _$DriverTripHistoryItem {

@JsonKey(name: 'operation_context') TripOperationContext? get operationContext;@JsonKey(name: 'ride_request_id') String get rideRequestId; RideLocation get pickup; RideLocation get destination; String get status;@JsonKey(name: 'assigned_at') DateTime get assignedAt;@JsonKey(name: 'started_at') DateTime? get startedAt;@JsonKey(name: 'completed_at') DateTime? get completedAt;@JsonKey(name: 'cancelled_at') DateTime? get cancelledAt; SettlementSnapshot get settlement;
/// Create a copy of DriverTripHistoryItem
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$DriverTripHistoryItemCopyWith<DriverTripHistoryItem> get copyWith => _$DriverTripHistoryItemCopyWithImpl<DriverTripHistoryItem>(this as DriverTripHistoryItem, _$identity);

  /// Serializes this DriverTripHistoryItem to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as DriverTripHistoryItem;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is DriverTripHistoryItem&&(identical(other.operationContext, _this.operationContext) || other.operationContext == _this.operationContext)&&(identical(other.rideRequestId, _this.rideRequestId) || other.rideRequestId == _this.rideRequestId)&&(identical(other.pickup, _this.pickup) || other.pickup == _this.pickup)&&(identical(other.destination, _this.destination) || other.destination == _this.destination)&&(identical(other.status, _this.status) || other.status == _this.status)&&(identical(other.assignedAt, _this.assignedAt) || other.assignedAt == _this.assignedAt)&&(identical(other.startedAt, _this.startedAt) || other.startedAt == _this.startedAt)&&(identical(other.completedAt, _this.completedAt) || other.completedAt == _this.completedAt)&&(identical(other.cancelledAt, _this.cancelledAt) || other.cancelledAt == _this.cancelledAt)&&(identical(other.settlement, _this.settlement) || other.settlement == _this.settlement));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as DriverTripHistoryItem;
  return Object.hash(runtimeType,_this.operationContext,_this.rideRequestId,_this.pickup,_this.destination,_this.status,_this.assignedAt,_this.startedAt,_this.completedAt,_this.cancelledAt,_this.settlement);
}

@override
String toString() {
  final _this = this as DriverTripHistoryItem;
  return 'DriverTripHistoryItem(operationContext: ${_this.operationContext}, rideRequestId: ${_this.rideRequestId}, pickup: ${_this.pickup}, destination: ${_this.destination}, status: ${_this.status}, assignedAt: ${_this.assignedAt}, startedAt: ${_this.startedAt}, completedAt: ${_this.completedAt}, cancelledAt: ${_this.cancelledAt}, settlement: ${_this.settlement})';
}


}

/// @nodoc
abstract mixin class $DriverTripHistoryItemCopyWith<$Res>  {
  factory $DriverTripHistoryItemCopyWith(DriverTripHistoryItem value, $Res Function(DriverTripHistoryItem) _then) = _$DriverTripHistoryItemCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'operation_context') TripOperationContext? operationContext,@JsonKey(name: 'ride_request_id') String rideRequestId, RideLocation pickup, RideLocation destination, String status,@JsonKey(name: 'assigned_at') DateTime assignedAt,@JsonKey(name: 'started_at') DateTime? startedAt,@JsonKey(name: 'completed_at') DateTime? completedAt,@JsonKey(name: 'cancelled_at') DateTime? cancelledAt, SettlementSnapshot settlement
});


$TripOperationContextCopyWith<$Res>? get operationContext;$RideLocationCopyWith<$Res> get pickup;$RideLocationCopyWith<$Res> get destination;$SettlementSnapshotCopyWith<$Res> get settlement;

}
/// @nodoc
class _$DriverTripHistoryItemCopyWithImpl<$Res>
    implements $DriverTripHistoryItemCopyWith<$Res> {
  _$DriverTripHistoryItemCopyWithImpl(this._self, this._then);

  final DriverTripHistoryItem _self;
  final $Res Function(DriverTripHistoryItem) _then;

/// Create a copy of DriverTripHistoryItem
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? operationContext = freezed,Object? rideRequestId = null,Object? pickup = null,Object? destination = null,Object? status = null,Object? assignedAt = null,Object? startedAt = freezed,Object? completedAt = freezed,Object? cancelledAt = freezed,Object? settlement = null,}) {
  return _then(DriverTripHistoryItem(
operationContext: freezed == operationContext ? _self.operationContext : operationContext // ignore: cast_nullable_to_non_nullable
as TripOperationContext?,rideRequestId: null == rideRequestId ? _self.rideRequestId : rideRequestId // ignore: cast_nullable_to_non_nullable
as String,pickup: null == pickup ? _self.pickup : pickup // ignore: cast_nullable_to_non_nullable
as RideLocation,destination: null == destination ? _self.destination : destination // ignore: cast_nullable_to_non_nullable
as RideLocation,status: null == status ? _self.status : status // ignore: cast_nullable_to_non_nullable
as String,assignedAt: null == assignedAt ? _self.assignedAt : assignedAt // ignore: cast_nullable_to_non_nullable
as DateTime,startedAt: freezed == startedAt ? _self.startedAt : startedAt // ignore: cast_nullable_to_non_nullable
as DateTime?,completedAt: freezed == completedAt ? _self.completedAt : completedAt // ignore: cast_nullable_to_non_nullable
as DateTime?,cancelledAt: freezed == cancelledAt ? _self.cancelledAt : cancelledAt // ignore: cast_nullable_to_non_nullable
as DateTime?,settlement: null == settlement ? _self.settlement : settlement // ignore: cast_nullable_to_non_nullable
as SettlementSnapshot,
  ));
}
/// Create a copy of DriverTripHistoryItem
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$TripOperationContextCopyWith<$Res>? get operationContext {
    if (_self.operationContext == null) {
    return null;
  }

  return $TripOperationContextCopyWith<$Res>(_self.operationContext!, (value) {
    return _then(_self.copyWith(operationContext: value));
  });
}/// Create a copy of DriverTripHistoryItem
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$RideLocationCopyWith<$Res> get pickup {

  return $RideLocationCopyWith<$Res>(_self.pickup, (value) {
    return _then(_self.copyWith(pickup: value));
  });
}/// Create a copy of DriverTripHistoryItem
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$RideLocationCopyWith<$Res> get destination {

  return $RideLocationCopyWith<$Res>(_self.destination, (value) {
    return _then(_self.copyWith(destination: value));
  });
}/// Create a copy of DriverTripHistoryItem
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$SettlementSnapshotCopyWith<$Res> get settlement {

  return $SettlementSnapshotCopyWith<$Res>(_self.settlement, (value) {
    return _then(_self.copyWith(settlement: value));
  });
}
}


/// Adds pattern-matching-related methods to [DriverTripHistoryItem].
extension DriverTripHistoryItemPatterns on DriverTripHistoryItem {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _DriverTripHistoryItem value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _DriverTripHistoryItem() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _DriverTripHistoryItem value)  $default,){
final _that = this;
switch (_that) {
case _DriverTripHistoryItem():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _DriverTripHistoryItem value)?  $default,){
final _that = this;
switch (_that) {
case _DriverTripHistoryItem() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(name: 'operation_context')  TripOperationContext? operationContext, @JsonKey(name: 'ride_request_id')  String rideRequestId,  RideLocation pickup,  RideLocation destination,  String status, @JsonKey(name: 'assigned_at')  DateTime assignedAt, @JsonKey(name: 'started_at')  DateTime? startedAt, @JsonKey(name: 'completed_at')  DateTime? completedAt, @JsonKey(name: 'cancelled_at')  DateTime? cancelledAt,  SettlementSnapshot settlement)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _DriverTripHistoryItem() when $default != null:
return $default(_that.operationContext,_that.rideRequestId,_that.pickup,_that.destination,_that.status,_that.assignedAt,_that.startedAt,_that.completedAt,_that.cancelledAt,_that.settlement);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(name: 'operation_context')  TripOperationContext? operationContext, @JsonKey(name: 'ride_request_id')  String rideRequestId,  RideLocation pickup,  RideLocation destination,  String status, @JsonKey(name: 'assigned_at')  DateTime assignedAt, @JsonKey(name: 'started_at')  DateTime? startedAt, @JsonKey(name: 'completed_at')  DateTime? completedAt, @JsonKey(name: 'cancelled_at')  DateTime? cancelledAt,  SettlementSnapshot settlement)  $default,) {final _that = this;
switch (_that) {
case _DriverTripHistoryItem():
return $default(_that.operationContext,_that.rideRequestId,_that.pickup,_that.destination,_that.status,_that.assignedAt,_that.startedAt,_that.completedAt,_that.cancelledAt,_that.settlement);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(name: 'operation_context')  TripOperationContext? operationContext, @JsonKey(name: 'ride_request_id')  String rideRequestId,  RideLocation pickup,  RideLocation destination,  String status, @JsonKey(name: 'assigned_at')  DateTime assignedAt, @JsonKey(name: 'started_at')  DateTime? startedAt, @JsonKey(name: 'completed_at')  DateTime? completedAt, @JsonKey(name: 'cancelled_at')  DateTime? cancelledAt,  SettlementSnapshot settlement)?  $default,) {final _that = this;
switch (_that) {
case _DriverTripHistoryItem() when $default != null:
return $default(_that.operationContext,_that.rideRequestId,_that.pickup,_that.destination,_that.status,_that.assignedAt,_that.startedAt,_that.completedAt,_that.cancelledAt,_that.settlement);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _DriverTripHistoryItem implements DriverTripHistoryItem {
  const _DriverTripHistoryItem({@JsonKey(name: 'operation_context') this.operationContext, @JsonKey(name: 'ride_request_id') required this.rideRequestId, required this.pickup, required this.destination, required this.status, @JsonKey(name: 'assigned_at') required this.assignedAt, @JsonKey(name: 'started_at') this.startedAt, @JsonKey(name: 'completed_at') this.completedAt, @JsonKey(name: 'cancelled_at') this.cancelledAt, required this.settlement});
  factory _DriverTripHistoryItem.fromJson(Map<String, dynamic> json) => _$DriverTripHistoryItemFromJson(json);

@override@JsonKey(name: 'operation_context') final  TripOperationContext? operationContext;
@override@JsonKey(name: 'ride_request_id') final  String rideRequestId;
@override final  RideLocation pickup;
@override final  RideLocation destination;
@override final  String status;
@override@JsonKey(name: 'assigned_at') final  DateTime assignedAt;
@override@JsonKey(name: 'started_at') final  DateTime? startedAt;
@override@JsonKey(name: 'completed_at') final  DateTime? completedAt;
@override@JsonKey(name: 'cancelled_at') final  DateTime? cancelledAt;
@override final  SettlementSnapshot settlement;

/// Create a copy of DriverTripHistoryItem
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$DriverTripHistoryItemCopyWith<_DriverTripHistoryItem> get copyWith => __$DriverTripHistoryItemCopyWithImpl<_DriverTripHistoryItem>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$DriverTripHistoryItemToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _DriverTripHistoryItem&&(identical(other.operationContext, operationContext) || other.operationContext == operationContext)&&(identical(other.rideRequestId, rideRequestId) || other.rideRequestId == rideRequestId)&&(identical(other.pickup, pickup) || other.pickup == pickup)&&(identical(other.destination, destination) || other.destination == destination)&&(identical(other.status, status) || other.status == status)&&(identical(other.assignedAt, assignedAt) || other.assignedAt == assignedAt)&&(identical(other.startedAt, startedAt) || other.startedAt == startedAt)&&(identical(other.completedAt, completedAt) || other.completedAt == completedAt)&&(identical(other.cancelledAt, cancelledAt) || other.cancelledAt == cancelledAt)&&(identical(other.settlement, settlement) || other.settlement == settlement));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,operationContext,rideRequestId,pickup,destination,status,assignedAt,startedAt,completedAt,cancelledAt,settlement);
}

@override
String toString() {
    return 'DriverTripHistoryItem(operationContext: $operationContext, rideRequestId: $rideRequestId, pickup: $pickup, destination: $destination, status: $status, assignedAt: $assignedAt, startedAt: $startedAt, completedAt: $completedAt, cancelledAt: $cancelledAt, settlement: $settlement)';
}


}

/// @nodoc
abstract mixin class _$DriverTripHistoryItemCopyWith<$Res> implements $DriverTripHistoryItemCopyWith<$Res> {
  factory _$DriverTripHistoryItemCopyWith(_DriverTripHistoryItem value, $Res Function(_DriverTripHistoryItem) _then) = __$DriverTripHistoryItemCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'operation_context') TripOperationContext? operationContext,@JsonKey(name: 'ride_request_id') String rideRequestId, RideLocation pickup, RideLocation destination, String status,@JsonKey(name: 'assigned_at') DateTime assignedAt,@JsonKey(name: 'started_at') DateTime? startedAt,@JsonKey(name: 'completed_at') DateTime? completedAt,@JsonKey(name: 'cancelled_at') DateTime? cancelledAt, SettlementSnapshot settlement
});


@override $TripOperationContextCopyWith<$Res>? get operationContext;@override $RideLocationCopyWith<$Res> get pickup;@override $RideLocationCopyWith<$Res> get destination;@override $SettlementSnapshotCopyWith<$Res> get settlement;

}
/// @nodoc
class __$DriverTripHistoryItemCopyWithImpl<$Res>
    implements _$DriverTripHistoryItemCopyWith<$Res> {
  __$DriverTripHistoryItemCopyWithImpl(this._self, this._then);

  final _DriverTripHistoryItem _self;
  final $Res Function(_DriverTripHistoryItem) _then;

/// Create a copy of DriverTripHistoryItem
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? operationContext = freezed,Object? rideRequestId = null,Object? pickup = null,Object? destination = null,Object? status = null,Object? assignedAt = null,Object? startedAt = freezed,Object? completedAt = freezed,Object? cancelledAt = freezed,Object? settlement = null,}) {
  return _then(_DriverTripHistoryItem(
operationContext: freezed == operationContext ? _self.operationContext : operationContext // ignore: cast_nullable_to_non_nullable
as TripOperationContext?,rideRequestId: null == rideRequestId ? _self.rideRequestId : rideRequestId // ignore: cast_nullable_to_non_nullable
as String,pickup: null == pickup ? _self.pickup : pickup // ignore: cast_nullable_to_non_nullable
as RideLocation,destination: null == destination ? _self.destination : destination // ignore: cast_nullable_to_non_nullable
as RideLocation,status: null == status ? _self.status : status // ignore: cast_nullable_to_non_nullable
as String,assignedAt: null == assignedAt ? _self.assignedAt : assignedAt // ignore: cast_nullable_to_non_nullable
as DateTime,startedAt: freezed == startedAt ? _self.startedAt : startedAt // ignore: cast_nullable_to_non_nullable
as DateTime?,completedAt: freezed == completedAt ? _self.completedAt : completedAt // ignore: cast_nullable_to_non_nullable
as DateTime?,cancelledAt: freezed == cancelledAt ? _self.cancelledAt : cancelledAt // ignore: cast_nullable_to_non_nullable
as DateTime?,settlement: null == settlement ? _self.settlement : settlement // ignore: cast_nullable_to_non_nullable
as SettlementSnapshot,
  ));
}

/// Create a copy of DriverTripHistoryItem
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$TripOperationContextCopyWith<$Res>? get operationContext {
    if (_self.operationContext == null) {
    return null;
  }

  return $TripOperationContextCopyWith<$Res>(_self.operationContext!, (value) {
    return _then(_self.copyWith(operationContext: value));
  });
}/// Create a copy of DriverTripHistoryItem
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$RideLocationCopyWith<$Res> get pickup {

  return $RideLocationCopyWith<$Res>(_self.pickup, (value) {
    return _then(_self.copyWith(pickup: value));
  });
}/// Create a copy of DriverTripHistoryItem
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$RideLocationCopyWith<$Res> get destination {

  return $RideLocationCopyWith<$Res>(_self.destination, (value) {
    return _then(_self.copyWith(destination: value));
  });
}/// Create a copy of DriverTripHistoryItem
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$SettlementSnapshotCopyWith<$Res> get settlement {

  return $SettlementSnapshotCopyWith<$Res>(_self.settlement, (value) {
    return _then(_self.copyWith(settlement: value));
  });
}
}

// dart format on
