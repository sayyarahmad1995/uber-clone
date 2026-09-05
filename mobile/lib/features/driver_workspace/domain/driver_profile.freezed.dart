// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'driver_profile.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$DriverVehicle {

 String get make; String get model;@JsonKey(name: 'model_year') int? get modelYear; String get color;@JsonKey(name: 'license_plate') String get licensePlate;
/// Create a copy of DriverVehicle
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$DriverVehicleCopyWith<DriverVehicle> get copyWith => _$DriverVehicleCopyWithImpl<DriverVehicle>(this as DriverVehicle, _$identity);

  /// Serializes this DriverVehicle to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as DriverVehicle;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is DriverVehicle&&(identical(other.make, _this.make) || other.make == _this.make)&&(identical(other.model, _this.model) || other.model == _this.model)&&(identical(other.modelYear, _this.modelYear) || other.modelYear == _this.modelYear)&&(identical(other.color, _this.color) || other.color == _this.color)&&(identical(other.licensePlate, _this.licensePlate) || other.licensePlate == _this.licensePlate));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as DriverVehicle;
  return Object.hash(runtimeType,_this.make,_this.model,_this.modelYear,_this.color,_this.licensePlate);
}

@override
String toString() {
  final _this = this as DriverVehicle;
  return 'DriverVehicle(make: ${_this.make}, model: ${_this.model}, modelYear: ${_this.modelYear}, color: ${_this.color}, licensePlate: ${_this.licensePlate})';
}


}

/// @nodoc
abstract mixin class $DriverVehicleCopyWith<$Res>  {
  factory $DriverVehicleCopyWith(DriverVehicle value, $Res Function(DriverVehicle) _then) = _$DriverVehicleCopyWithImpl;
@useResult
$Res call({
 String make, String model,@JsonKey(name: 'model_year') int? modelYear, String color,@JsonKey(name: 'license_plate') String licensePlate
});




}
/// @nodoc
class _$DriverVehicleCopyWithImpl<$Res>
    implements $DriverVehicleCopyWith<$Res> {
  _$DriverVehicleCopyWithImpl(this._self, this._then);

  final DriverVehicle _self;
  final $Res Function(DriverVehicle) _then;

/// Create a copy of DriverVehicle
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? make = null,Object? model = null,Object? modelYear = freezed,Object? color = null,Object? licensePlate = null,}) {
  return _then(DriverVehicle(
make: null == make ? _self.make : make // ignore: cast_nullable_to_non_nullable
as String,model: null == model ? _self.model : model // ignore: cast_nullable_to_non_nullable
as String,modelYear: freezed == modelYear ? _self.modelYear : modelYear // ignore: cast_nullable_to_non_nullable
as int?,color: null == color ? _self.color : color // ignore: cast_nullable_to_non_nullable
as String,licensePlate: null == licensePlate ? _self.licensePlate : licensePlate // ignore: cast_nullable_to_non_nullable
as String,
  ));
}

}


/// Adds pattern-matching-related methods to [DriverVehicle].
extension DriverVehiclePatterns on DriverVehicle {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _DriverVehicle value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _DriverVehicle() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _DriverVehicle value)  $default,){
final _that = this;
switch (_that) {
case _DriverVehicle():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _DriverVehicle value)?  $default,){
final _that = this;
switch (_that) {
case _DriverVehicle() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( String make,  String model, @JsonKey(name: 'model_year')  int? modelYear,  String color, @JsonKey(name: 'license_plate')  String licensePlate)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _DriverVehicle() when $default != null:
return $default(_that.make,_that.model,_that.modelYear,_that.color,_that.licensePlate);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( String make,  String model, @JsonKey(name: 'model_year')  int? modelYear,  String color, @JsonKey(name: 'license_plate')  String licensePlate)  $default,) {final _that = this;
switch (_that) {
case _DriverVehicle():
return $default(_that.make,_that.model,_that.modelYear,_that.color,_that.licensePlate);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( String make,  String model, @JsonKey(name: 'model_year')  int? modelYear,  String color, @JsonKey(name: 'license_plate')  String licensePlate)?  $default,) {final _that = this;
switch (_that) {
case _DriverVehicle() when $default != null:
return $default(_that.make,_that.model,_that.modelYear,_that.color,_that.licensePlate);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _DriverVehicle implements DriverVehicle {
  const _DriverVehicle({required this.make, required this.model, @JsonKey(name: 'model_year') this.modelYear, required this.color, @JsonKey(name: 'license_plate') required this.licensePlate});
  factory _DriverVehicle.fromJson(Map<String, dynamic> json) => _$DriverVehicleFromJson(json);

@override final  String make;
@override final  String model;
@override@JsonKey(name: 'model_year') final  int? modelYear;
@override final  String color;
@override@JsonKey(name: 'license_plate') final  String licensePlate;

/// Create a copy of DriverVehicle
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$DriverVehicleCopyWith<_DriverVehicle> get copyWith => __$DriverVehicleCopyWithImpl<_DriverVehicle>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$DriverVehicleToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _DriverVehicle&&(identical(other.make, make) || other.make == make)&&(identical(other.model, model) || other.model == model)&&(identical(other.modelYear, modelYear) || other.modelYear == modelYear)&&(identical(other.color, color) || other.color == color)&&(identical(other.licensePlate, licensePlate) || other.licensePlate == licensePlate));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,make,model,modelYear,color,licensePlate);
}

@override
String toString() {
    return 'DriverVehicle(make: $make, model: $model, modelYear: $modelYear, color: $color, licensePlate: $licensePlate)';
}


}

/// @nodoc
abstract mixin class _$DriverVehicleCopyWith<$Res> implements $DriverVehicleCopyWith<$Res> {
  factory _$DriverVehicleCopyWith(_DriverVehicle value, $Res Function(_DriverVehicle) _then) = __$DriverVehicleCopyWithImpl;
@override @useResult
$Res call({
 String make, String model,@JsonKey(name: 'model_year') int? modelYear, String color,@JsonKey(name: 'license_plate') String licensePlate
});




}
/// @nodoc
class __$DriverVehicleCopyWithImpl<$Res>
    implements _$DriverVehicleCopyWith<$Res> {
  __$DriverVehicleCopyWithImpl(this._self, this._then);

  final _DriverVehicle _self;
  final $Res Function(_DriverVehicle) _then;

/// Create a copy of DriverVehicle
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? make = null,Object? model = null,Object? modelYear = freezed,Object? color = null,Object? licensePlate = null,}) {
  return _then(_DriverVehicle(
make: null == make ? _self.make : make // ignore: cast_nullable_to_non_nullable
as String,model: null == model ? _self.model : model // ignore: cast_nullable_to_non_nullable
as String,modelYear: freezed == modelYear ? _self.modelYear : modelYear // ignore: cast_nullable_to_non_nullable
as int?,color: null == color ? _self.color : color // ignore: cast_nullable_to_non_nullable
as String,licensePlate: null == licensePlate ? _self.licensePlate : licensePlate // ignore: cast_nullable_to_non_nullable
as String,
  ));
}


}


/// @nodoc
mixin _$DriverProfile {

@JsonKey(name: 'user_id') String get userId;@JsonKey(name: 'display_name') String? get displayName; String get status;@JsonKey(name: 'is_online') bool get isOnline; DriverVehicle get vehicle;
/// Create a copy of DriverProfile
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$DriverProfileCopyWith<DriverProfile> get copyWith => _$DriverProfileCopyWithImpl<DriverProfile>(this as DriverProfile, _$identity);

  /// Serializes this DriverProfile to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as DriverProfile;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is DriverProfile&&(identical(other.userId, _this.userId) || other.userId == _this.userId)&&(identical(other.displayName, _this.displayName) || other.displayName == _this.displayName)&&(identical(other.status, _this.status) || other.status == _this.status)&&(identical(other.isOnline, _this.isOnline) || other.isOnline == _this.isOnline)&&(identical(other.vehicle, _this.vehicle) || other.vehicle == _this.vehicle));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as DriverProfile;
  return Object.hash(runtimeType,_this.userId,_this.displayName,_this.status,_this.isOnline,_this.vehicle);
}

@override
String toString() {
  final _this = this as DriverProfile;
  return 'DriverProfile(userId: ${_this.userId}, displayName: ${_this.displayName}, status: ${_this.status}, isOnline: ${_this.isOnline}, vehicle: ${_this.vehicle})';
}


}

/// @nodoc
abstract mixin class $DriverProfileCopyWith<$Res>  {
  factory $DriverProfileCopyWith(DriverProfile value, $Res Function(DriverProfile) _then) = _$DriverProfileCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'user_id') String userId,@JsonKey(name: 'display_name') String? displayName, String status,@JsonKey(name: 'is_online') bool isOnline, DriverVehicle vehicle
});


$DriverVehicleCopyWith<$Res> get vehicle;

}
/// @nodoc
class _$DriverProfileCopyWithImpl<$Res>
    implements $DriverProfileCopyWith<$Res> {
  _$DriverProfileCopyWithImpl(this._self, this._then);

  final DriverProfile _self;
  final $Res Function(DriverProfile) _then;

/// Create a copy of DriverProfile
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? userId = null,Object? displayName = freezed,Object? status = null,Object? isOnline = null,Object? vehicle = null,}) {
  return _then(DriverProfile(
userId: null == userId ? _self.userId : userId // ignore: cast_nullable_to_non_nullable
as String,displayName: freezed == displayName ? _self.displayName : displayName // ignore: cast_nullable_to_non_nullable
as String?,status: null == status ? _self.status : status // ignore: cast_nullable_to_non_nullable
as String,isOnline: null == isOnline ? _self.isOnline : isOnline // ignore: cast_nullable_to_non_nullable
as bool,vehicle: null == vehicle ? _self.vehicle : vehicle // ignore: cast_nullable_to_non_nullable
as DriverVehicle,
  ));
}
/// Create a copy of DriverProfile
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$DriverVehicleCopyWith<$Res> get vehicle {
  
  return $DriverVehicleCopyWith<$Res>(_self.vehicle, (value) {
    return _then(_self.copyWith(vehicle: value));
  });
}
}


/// Adds pattern-matching-related methods to [DriverProfile].
extension DriverProfilePatterns on DriverProfile {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _DriverProfile value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _DriverProfile() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _DriverProfile value)  $default,){
final _that = this;
switch (_that) {
case _DriverProfile():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _DriverProfile value)?  $default,){
final _that = this;
switch (_that) {
case _DriverProfile() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(name: 'user_id')  String userId, @JsonKey(name: 'display_name')  String? displayName,  String status, @JsonKey(name: 'is_online')  bool isOnline,  DriverVehicle vehicle)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _DriverProfile() when $default != null:
return $default(_that.userId,_that.displayName,_that.status,_that.isOnline,_that.vehicle);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(name: 'user_id')  String userId, @JsonKey(name: 'display_name')  String? displayName,  String status, @JsonKey(name: 'is_online')  bool isOnline,  DriverVehicle vehicle)  $default,) {final _that = this;
switch (_that) {
case _DriverProfile():
return $default(_that.userId,_that.displayName,_that.status,_that.isOnline,_that.vehicle);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(name: 'user_id')  String userId, @JsonKey(name: 'display_name')  String? displayName,  String status, @JsonKey(name: 'is_online')  bool isOnline,  DriverVehicle vehicle)?  $default,) {final _that = this;
switch (_that) {
case _DriverProfile() when $default != null:
return $default(_that.userId,_that.displayName,_that.status,_that.isOnline,_that.vehicle);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _DriverProfile implements DriverProfile {
  const _DriverProfile({@JsonKey(name: 'user_id') required this.userId, @JsonKey(name: 'display_name') this.displayName, required this.status, @JsonKey(name: 'is_online') required this.isOnline, required this.vehicle});
  factory _DriverProfile.fromJson(Map<String, dynamic> json) => _$DriverProfileFromJson(json);

@override@JsonKey(name: 'user_id') final  String userId;
@override@JsonKey(name: 'display_name') final  String? displayName;
@override final  String status;
@override@JsonKey(name: 'is_online') final  bool isOnline;
@override final  DriverVehicle vehicle;

/// Create a copy of DriverProfile
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$DriverProfileCopyWith<_DriverProfile> get copyWith => __$DriverProfileCopyWithImpl<_DriverProfile>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$DriverProfileToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _DriverProfile&&(identical(other.userId, userId) || other.userId == userId)&&(identical(other.displayName, displayName) || other.displayName == displayName)&&(identical(other.status, status) || other.status == status)&&(identical(other.isOnline, isOnline) || other.isOnline == isOnline)&&(identical(other.vehicle, vehicle) || other.vehicle == vehicle));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,userId,displayName,status,isOnline,vehicle);
}

@override
String toString() {
    return 'DriverProfile(userId: $userId, displayName: $displayName, status: $status, isOnline: $isOnline, vehicle: $vehicle)';
}


}

/// @nodoc
abstract mixin class _$DriverProfileCopyWith<$Res> implements $DriverProfileCopyWith<$Res> {
  factory _$DriverProfileCopyWith(_DriverProfile value, $Res Function(_DriverProfile) _then) = __$DriverProfileCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'user_id') String userId,@JsonKey(name: 'display_name') String? displayName, String status,@JsonKey(name: 'is_online') bool isOnline, DriverVehicle vehicle
});


@override $DriverVehicleCopyWith<$Res> get vehicle;

}
/// @nodoc
class __$DriverProfileCopyWithImpl<$Res>
    implements _$DriverProfileCopyWith<$Res> {
  __$DriverProfileCopyWithImpl(this._self, this._then);

  final _DriverProfile _self;
  final $Res Function(_DriverProfile) _then;

/// Create a copy of DriverProfile
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? userId = null,Object? displayName = freezed,Object? status = null,Object? isOnline = null,Object? vehicle = null,}) {
  return _then(_DriverProfile(
userId: null == userId ? _self.userId : userId // ignore: cast_nullable_to_non_nullable
as String,displayName: freezed == displayName ? _self.displayName : displayName // ignore: cast_nullable_to_non_nullable
as String?,status: null == status ? _self.status : status // ignore: cast_nullable_to_non_nullable
as String,isOnline: null == isOnline ? _self.isOnline : isOnline // ignore: cast_nullable_to_non_nullable
as bool,vehicle: null == vehicle ? _self.vehicle : vehicle // ignore: cast_nullable_to_non_nullable
as DriverVehicle,
  ));
}

/// Create a copy of DriverProfile
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$DriverVehicleCopyWith<$Res> get vehicle {
  
  return $DriverVehicleCopyWith<$Res>(_self.vehicle, (value) {
    return _then(_self.copyWith(vehicle: value));
  });
}
}


/// @nodoc
mixin _$PublishedDriverLocation {

 double get latitude; double get longitude;@JsonKey(name: 'updated_at') DateTime get updatedAt;
/// Create a copy of PublishedDriverLocation
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$PublishedDriverLocationCopyWith<PublishedDriverLocation> get copyWith => _$PublishedDriverLocationCopyWithImpl<PublishedDriverLocation>(this as PublishedDriverLocation, _$identity);

  /// Serializes this PublishedDriverLocation to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as PublishedDriverLocation;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is PublishedDriverLocation&&(identical(other.latitude, _this.latitude) || other.latitude == _this.latitude)&&(identical(other.longitude, _this.longitude) || other.longitude == _this.longitude)&&(identical(other.updatedAt, _this.updatedAt) || other.updatedAt == _this.updatedAt));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as PublishedDriverLocation;
  return Object.hash(runtimeType,_this.latitude,_this.longitude,_this.updatedAt);
}

@override
String toString() {
  final _this = this as PublishedDriverLocation;
  return 'PublishedDriverLocation(latitude: ${_this.latitude}, longitude: ${_this.longitude}, updatedAt: ${_this.updatedAt})';
}


}

/// @nodoc
abstract mixin class $PublishedDriverLocationCopyWith<$Res>  {
  factory $PublishedDriverLocationCopyWith(PublishedDriverLocation value, $Res Function(PublishedDriverLocation) _then) = _$PublishedDriverLocationCopyWithImpl;
@useResult
$Res call({
 double latitude, double longitude,@JsonKey(name: 'updated_at') DateTime updatedAt
});




}
/// @nodoc
class _$PublishedDriverLocationCopyWithImpl<$Res>
    implements $PublishedDriverLocationCopyWith<$Res> {
  _$PublishedDriverLocationCopyWithImpl(this._self, this._then);

  final PublishedDriverLocation _self;
  final $Res Function(PublishedDriverLocation) _then;

/// Create a copy of PublishedDriverLocation
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? latitude = null,Object? longitude = null,Object? updatedAt = null,}) {
  return _then(PublishedDriverLocation(
latitude: null == latitude ? _self.latitude : latitude // ignore: cast_nullable_to_non_nullable
as double,longitude: null == longitude ? _self.longitude : longitude // ignore: cast_nullable_to_non_nullable
as double,updatedAt: null == updatedAt ? _self.updatedAt : updatedAt // ignore: cast_nullable_to_non_nullable
as DateTime,
  ));
}

}


/// Adds pattern-matching-related methods to [PublishedDriverLocation].
extension PublishedDriverLocationPatterns on PublishedDriverLocation {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _PublishedDriverLocation value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _PublishedDriverLocation() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _PublishedDriverLocation value)  $default,){
final _that = this;
switch (_that) {
case _PublishedDriverLocation():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _PublishedDriverLocation value)?  $default,){
final _that = this;
switch (_that) {
case _PublishedDriverLocation() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( double latitude,  double longitude, @JsonKey(name: 'updated_at')  DateTime updatedAt)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _PublishedDriverLocation() when $default != null:
return $default(_that.latitude,_that.longitude,_that.updatedAt);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( double latitude,  double longitude, @JsonKey(name: 'updated_at')  DateTime updatedAt)  $default,) {final _that = this;
switch (_that) {
case _PublishedDriverLocation():
return $default(_that.latitude,_that.longitude,_that.updatedAt);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( double latitude,  double longitude, @JsonKey(name: 'updated_at')  DateTime updatedAt)?  $default,) {final _that = this;
switch (_that) {
case _PublishedDriverLocation() when $default != null:
return $default(_that.latitude,_that.longitude,_that.updatedAt);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _PublishedDriverLocation implements PublishedDriverLocation {
  const _PublishedDriverLocation({required this.latitude, required this.longitude, @JsonKey(name: 'updated_at') required this.updatedAt});
  factory _PublishedDriverLocation.fromJson(Map<String, dynamic> json) => _$PublishedDriverLocationFromJson(json);

@override final  double latitude;
@override final  double longitude;
@override@JsonKey(name: 'updated_at') final  DateTime updatedAt;

/// Create a copy of PublishedDriverLocation
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$PublishedDriverLocationCopyWith<_PublishedDriverLocation> get copyWith => __$PublishedDriverLocationCopyWithImpl<_PublishedDriverLocation>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$PublishedDriverLocationToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _PublishedDriverLocation&&(identical(other.latitude, latitude) || other.latitude == latitude)&&(identical(other.longitude, longitude) || other.longitude == longitude)&&(identical(other.updatedAt, updatedAt) || other.updatedAt == updatedAt));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,latitude,longitude,updatedAt);
}

@override
String toString() {
    return 'PublishedDriverLocation(latitude: $latitude, longitude: $longitude, updatedAt: $updatedAt)';
}


}

/// @nodoc
abstract mixin class _$PublishedDriverLocationCopyWith<$Res> implements $PublishedDriverLocationCopyWith<$Res> {
  factory _$PublishedDriverLocationCopyWith(_PublishedDriverLocation value, $Res Function(_PublishedDriverLocation) _then) = __$PublishedDriverLocationCopyWithImpl;
@override @useResult
$Res call({
 double latitude, double longitude,@JsonKey(name: 'updated_at') DateTime updatedAt
});




}
/// @nodoc
class __$PublishedDriverLocationCopyWithImpl<$Res>
    implements _$PublishedDriverLocationCopyWith<$Res> {
  __$PublishedDriverLocationCopyWithImpl(this._self, this._then);

  final _PublishedDriverLocation _self;
  final $Res Function(_PublishedDriverLocation) _then;

/// Create a copy of PublishedDriverLocation
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? latitude = null,Object? longitude = null,Object? updatedAt = null,}) {
  return _then(_PublishedDriverLocation(
latitude: null == latitude ? _self.latitude : latitude // ignore: cast_nullable_to_non_nullable
as double,longitude: null == longitude ? _self.longitude : longitude // ignore: cast_nullable_to_non_nullable
as double,updatedAt: null == updatedAt ? _self.updatedAt : updatedAt // ignore: cast_nullable_to_non_nullable
as DateTime,
  ));
}


}

// dart format on
