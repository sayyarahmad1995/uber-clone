// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'ride_snapshot.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$RiderRideSnapshot {

@JsonKey(name: 'service_code') String get serviceCode; String get id; RideLocation get pickup; RideLocation get destination;@JsonKey(name: 'proposed_fare') RideFare get proposedFare; String get status;@JsonKey(name: 'created_at') DateTime get createdAt;@JsonKey(name: 'expires_at') DateTime get expiresAt; TripSnapshot? get trip;
/// Create a copy of RiderRideSnapshot
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$RiderRideSnapshotCopyWith<RiderRideSnapshot> get copyWith => _$RiderRideSnapshotCopyWithImpl<RiderRideSnapshot>(this as RiderRideSnapshot, _$identity);

  /// Serializes this RiderRideSnapshot to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as RiderRideSnapshot;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is RiderRideSnapshot&&(identical(other.serviceCode, _this.serviceCode) || other.serviceCode == _this.serviceCode)&&(identical(other.id, _this.id) || other.id == _this.id)&&(identical(other.pickup, _this.pickup) || other.pickup == _this.pickup)&&(identical(other.destination, _this.destination) || other.destination == _this.destination)&&(identical(other.proposedFare, _this.proposedFare) || other.proposedFare == _this.proposedFare)&&(identical(other.status, _this.status) || other.status == _this.status)&&(identical(other.createdAt, _this.createdAt) || other.createdAt == _this.createdAt)&&(identical(other.expiresAt, _this.expiresAt) || other.expiresAt == _this.expiresAt)&&(identical(other.trip, _this.trip) || other.trip == _this.trip));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as RiderRideSnapshot;
  return Object.hash(runtimeType,_this.serviceCode,_this.id,_this.pickup,_this.destination,_this.proposedFare,_this.status,_this.createdAt,_this.expiresAt,_this.trip);
}

@override
String toString() {
  final _this = this as RiderRideSnapshot;
  return 'RiderRideSnapshot(serviceCode: ${_this.serviceCode}, id: ${_this.id}, pickup: ${_this.pickup}, destination: ${_this.destination}, proposedFare: ${_this.proposedFare}, status: ${_this.status}, createdAt: ${_this.createdAt}, expiresAt: ${_this.expiresAt}, trip: ${_this.trip})';
}


}

/// @nodoc
abstract mixin class $RiderRideSnapshotCopyWith<$Res>  {
  factory $RiderRideSnapshotCopyWith(RiderRideSnapshot value, $Res Function(RiderRideSnapshot) _then) = _$RiderRideSnapshotCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'service_code') String serviceCode, String id, RideLocation pickup, RideLocation destination,@JsonKey(name: 'proposed_fare') RideFare proposedFare, String status,@JsonKey(name: 'created_at') DateTime createdAt,@JsonKey(name: 'expires_at') DateTime expiresAt, TripSnapshot? trip
});


$RideLocationCopyWith<$Res> get pickup;$RideLocationCopyWith<$Res> get destination;$RideFareCopyWith<$Res> get proposedFare;$TripSnapshotCopyWith<$Res>? get trip;

}
/// @nodoc
class _$RiderRideSnapshotCopyWithImpl<$Res>
    implements $RiderRideSnapshotCopyWith<$Res> {
  _$RiderRideSnapshotCopyWithImpl(this._self, this._then);

  final RiderRideSnapshot _self;
  final $Res Function(RiderRideSnapshot) _then;

/// Create a copy of RiderRideSnapshot
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? serviceCode = null,Object? id = null,Object? pickup = null,Object? destination = null,Object? proposedFare = null,Object? status = null,Object? createdAt = null,Object? expiresAt = null,Object? trip = freezed,}) {
  return _then(RiderRideSnapshot(
serviceCode: null == serviceCode ? _self.serviceCode : serviceCode // ignore: cast_nullable_to_non_nullable
as String,id: null == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as String,pickup: null == pickup ? _self.pickup : pickup // ignore: cast_nullable_to_non_nullable
as RideLocation,destination: null == destination ? _self.destination : destination // ignore: cast_nullable_to_non_nullable
as RideLocation,proposedFare: null == proposedFare ? _self.proposedFare : proposedFare // ignore: cast_nullable_to_non_nullable
as RideFare,status: null == status ? _self.status : status // ignore: cast_nullable_to_non_nullable
as String,createdAt: null == createdAt ? _self.createdAt : createdAt // ignore: cast_nullable_to_non_nullable
as DateTime,expiresAt: null == expiresAt ? _self.expiresAt : expiresAt // ignore: cast_nullable_to_non_nullable
as DateTime,trip: freezed == trip ? _self.trip : trip // ignore: cast_nullable_to_non_nullable
as TripSnapshot?,
  ));
}
/// Create a copy of RiderRideSnapshot
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$RideLocationCopyWith<$Res> get pickup {

  return $RideLocationCopyWith<$Res>(_self.pickup, (value) {
    return _then(_self.copyWith(pickup: value));
  });
}/// Create a copy of RiderRideSnapshot
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$RideLocationCopyWith<$Res> get destination {

  return $RideLocationCopyWith<$Res>(_self.destination, (value) {
    return _then(_self.copyWith(destination: value));
  });
}/// Create a copy of RiderRideSnapshot
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$RideFareCopyWith<$Res> get proposedFare {

  return $RideFareCopyWith<$Res>(_self.proposedFare, (value) {
    return _then(_self.copyWith(proposedFare: value));
  });
}/// Create a copy of RiderRideSnapshot
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$TripSnapshotCopyWith<$Res>? get trip {
    if (_self.trip == null) {
    return null;
  }

  return $TripSnapshotCopyWith<$Res>(_self.trip!, (value) {
    return _then(_self.copyWith(trip: value));
  });
}
}


/// Adds pattern-matching-related methods to [RiderRideSnapshot].
extension RiderRideSnapshotPatterns on RiderRideSnapshot {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _RiderRideSnapshot value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _RiderRideSnapshot() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _RiderRideSnapshot value)  $default,){
final _that = this;
switch (_that) {
case _RiderRideSnapshot():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _RiderRideSnapshot value)?  $default,){
final _that = this;
switch (_that) {
case _RiderRideSnapshot() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(name: 'service_code')  String serviceCode,  String id,  RideLocation pickup,  RideLocation destination, @JsonKey(name: 'proposed_fare')  RideFare proposedFare,  String status, @JsonKey(name: 'created_at')  DateTime createdAt, @JsonKey(name: 'expires_at')  DateTime expiresAt,  TripSnapshot? trip)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _RiderRideSnapshot() when $default != null:
return $default(_that.serviceCode,_that.id,_that.pickup,_that.destination,_that.proposedFare,_that.status,_that.createdAt,_that.expiresAt,_that.trip);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(name: 'service_code')  String serviceCode,  String id,  RideLocation pickup,  RideLocation destination, @JsonKey(name: 'proposed_fare')  RideFare proposedFare,  String status, @JsonKey(name: 'created_at')  DateTime createdAt, @JsonKey(name: 'expires_at')  DateTime expiresAt,  TripSnapshot? trip)  $default,) {final _that = this;
switch (_that) {
case _RiderRideSnapshot():
return $default(_that.serviceCode,_that.id,_that.pickup,_that.destination,_that.proposedFare,_that.status,_that.createdAt,_that.expiresAt,_that.trip);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(name: 'service_code')  String serviceCode,  String id,  RideLocation pickup,  RideLocation destination, @JsonKey(name: 'proposed_fare')  RideFare proposedFare,  String status, @JsonKey(name: 'created_at')  DateTime createdAt, @JsonKey(name: 'expires_at')  DateTime expiresAt,  TripSnapshot? trip)?  $default,) {final _that = this;
switch (_that) {
case _RiderRideSnapshot() when $default != null:
return $default(_that.serviceCode,_that.id,_that.pickup,_that.destination,_that.proposedFare,_that.status,_that.createdAt,_that.expiresAt,_that.trip);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _RiderRideSnapshot implements RiderRideSnapshot {
  const _RiderRideSnapshot({@JsonKey(name: 'service_code') required this.serviceCode, required this.id, required this.pickup, required this.destination, @JsonKey(name: 'proposed_fare') required this.proposedFare, required this.status, @JsonKey(name: 'created_at') required this.createdAt, @JsonKey(name: 'expires_at') required this.expiresAt, this.trip});
  factory _RiderRideSnapshot.fromJson(Map<String, dynamic> json) => _$RiderRideSnapshotFromJson(json);

@override@JsonKey(name: 'service_code') final  String serviceCode;
@override final  String id;
@override final  RideLocation pickup;
@override final  RideLocation destination;
@override@JsonKey(name: 'proposed_fare') final  RideFare proposedFare;
@override final  String status;
@override@JsonKey(name: 'created_at') final  DateTime createdAt;
@override@JsonKey(name: 'expires_at') final  DateTime expiresAt;
@override final  TripSnapshot? trip;

/// Create a copy of RiderRideSnapshot
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$RiderRideSnapshotCopyWith<_RiderRideSnapshot> get copyWith => __$RiderRideSnapshotCopyWithImpl<_RiderRideSnapshot>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$RiderRideSnapshotToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _RiderRideSnapshot&&(identical(other.serviceCode, serviceCode) || other.serviceCode == serviceCode)&&(identical(other.id, id) || other.id == id)&&(identical(other.pickup, pickup) || other.pickup == pickup)&&(identical(other.destination, destination) || other.destination == destination)&&(identical(other.proposedFare, proposedFare) || other.proposedFare == proposedFare)&&(identical(other.status, status) || other.status == status)&&(identical(other.createdAt, createdAt) || other.createdAt == createdAt)&&(identical(other.expiresAt, expiresAt) || other.expiresAt == expiresAt)&&(identical(other.trip, trip) || other.trip == trip));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,serviceCode,id,pickup,destination,proposedFare,status,createdAt,expiresAt,trip);
}

@override
String toString() {
    return 'RiderRideSnapshot(serviceCode: $serviceCode, id: $id, pickup: $pickup, destination: $destination, proposedFare: $proposedFare, status: $status, createdAt: $createdAt, expiresAt: $expiresAt, trip: $trip)';
}


}

/// @nodoc
abstract mixin class _$RiderRideSnapshotCopyWith<$Res> implements $RiderRideSnapshotCopyWith<$Res> {
  factory _$RiderRideSnapshotCopyWith(_RiderRideSnapshot value, $Res Function(_RiderRideSnapshot) _then) = __$RiderRideSnapshotCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'service_code') String serviceCode, String id, RideLocation pickup, RideLocation destination,@JsonKey(name: 'proposed_fare') RideFare proposedFare, String status,@JsonKey(name: 'created_at') DateTime createdAt,@JsonKey(name: 'expires_at') DateTime expiresAt, TripSnapshot? trip
});


@override $RideLocationCopyWith<$Res> get pickup;@override $RideLocationCopyWith<$Res> get destination;@override $RideFareCopyWith<$Res> get proposedFare;@override $TripSnapshotCopyWith<$Res>? get trip;

}
/// @nodoc
class __$RiderRideSnapshotCopyWithImpl<$Res>
    implements _$RiderRideSnapshotCopyWith<$Res> {
  __$RiderRideSnapshotCopyWithImpl(this._self, this._then);

  final _RiderRideSnapshot _self;
  final $Res Function(_RiderRideSnapshot) _then;

/// Create a copy of RiderRideSnapshot
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? serviceCode = null,Object? id = null,Object? pickup = null,Object? destination = null,Object? proposedFare = null,Object? status = null,Object? createdAt = null,Object? expiresAt = null,Object? trip = freezed,}) {
  return _then(_RiderRideSnapshot(
serviceCode: null == serviceCode ? _self.serviceCode : serviceCode // ignore: cast_nullable_to_non_nullable
as String,id: null == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as String,pickup: null == pickup ? _self.pickup : pickup // ignore: cast_nullable_to_non_nullable
as RideLocation,destination: null == destination ? _self.destination : destination // ignore: cast_nullable_to_non_nullable
as RideLocation,proposedFare: null == proposedFare ? _self.proposedFare : proposedFare // ignore: cast_nullable_to_non_nullable
as RideFare,status: null == status ? _self.status : status // ignore: cast_nullable_to_non_nullable
as String,createdAt: null == createdAt ? _self.createdAt : createdAt // ignore: cast_nullable_to_non_nullable
as DateTime,expiresAt: null == expiresAt ? _self.expiresAt : expiresAt // ignore: cast_nullable_to_non_nullable
as DateTime,trip: freezed == trip ? _self.trip : trip // ignore: cast_nullable_to_non_nullable
as TripSnapshot?,
  ));
}

/// Create a copy of RiderRideSnapshot
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$RideLocationCopyWith<$Res> get pickup {

  return $RideLocationCopyWith<$Res>(_self.pickup, (value) {
    return _then(_self.copyWith(pickup: value));
  });
}/// Create a copy of RiderRideSnapshot
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$RideLocationCopyWith<$Res> get destination {

  return $RideLocationCopyWith<$Res>(_self.destination, (value) {
    return _then(_self.copyWith(destination: value));
  });
}/// Create a copy of RiderRideSnapshot
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$RideFareCopyWith<$Res> get proposedFare {

  return $RideFareCopyWith<$Res>(_self.proposedFare, (value) {
    return _then(_self.copyWith(proposedFare: value));
  });
}/// Create a copy of RiderRideSnapshot
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$TripSnapshotCopyWith<$Res>? get trip {
    if (_self.trip == null) {
    return null;
  }

  return $TripSnapshotCopyWith<$Res>(_self.trip!, (value) {
    return _then(_self.copyWith(trip: value));
  });
}
}


/// @nodoc
mixin _$DriverLocationSnapshot {

@JsonKey(name: 'ride_request_id') String get rideRequestId; double get latitude; double get longitude;@JsonKey(name: 'updated_at') DateTime get updatedAt;
/// Create a copy of DriverLocationSnapshot
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$DriverLocationSnapshotCopyWith<DriverLocationSnapshot> get copyWith => _$DriverLocationSnapshotCopyWithImpl<DriverLocationSnapshot>(this as DriverLocationSnapshot, _$identity);

  /// Serializes this DriverLocationSnapshot to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as DriverLocationSnapshot;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is DriverLocationSnapshot&&(identical(other.rideRequestId, _this.rideRequestId) || other.rideRequestId == _this.rideRequestId)&&(identical(other.latitude, _this.latitude) || other.latitude == _this.latitude)&&(identical(other.longitude, _this.longitude) || other.longitude == _this.longitude)&&(identical(other.updatedAt, _this.updatedAt) || other.updatedAt == _this.updatedAt));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as DriverLocationSnapshot;
  return Object.hash(runtimeType,_this.rideRequestId,_this.latitude,_this.longitude,_this.updatedAt);
}

@override
String toString() {
  final _this = this as DriverLocationSnapshot;
  return 'DriverLocationSnapshot(rideRequestId: ${_this.rideRequestId}, latitude: ${_this.latitude}, longitude: ${_this.longitude}, updatedAt: ${_this.updatedAt})';
}


}

/// @nodoc
abstract mixin class $DriverLocationSnapshotCopyWith<$Res>  {
  factory $DriverLocationSnapshotCopyWith(DriverLocationSnapshot value, $Res Function(DriverLocationSnapshot) _then) = _$DriverLocationSnapshotCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'ride_request_id') String rideRequestId, double latitude, double longitude,@JsonKey(name: 'updated_at') DateTime updatedAt
});




}
/// @nodoc
class _$DriverLocationSnapshotCopyWithImpl<$Res>
    implements $DriverLocationSnapshotCopyWith<$Res> {
  _$DriverLocationSnapshotCopyWithImpl(this._self, this._then);

  final DriverLocationSnapshot _self;
  final $Res Function(DriverLocationSnapshot) _then;

/// Create a copy of DriverLocationSnapshot
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? rideRequestId = null,Object? latitude = null,Object? longitude = null,Object? updatedAt = null,}) {
  return _then(DriverLocationSnapshot(
rideRequestId: null == rideRequestId ? _self.rideRequestId : rideRequestId // ignore: cast_nullable_to_non_nullable
as String,latitude: null == latitude ? _self.latitude : latitude // ignore: cast_nullable_to_non_nullable
as double,longitude: null == longitude ? _self.longitude : longitude // ignore: cast_nullable_to_non_nullable
as double,updatedAt: null == updatedAt ? _self.updatedAt : updatedAt // ignore: cast_nullable_to_non_nullable
as DateTime,
  ));
}

}


/// Adds pattern-matching-related methods to [DriverLocationSnapshot].
extension DriverLocationSnapshotPatterns on DriverLocationSnapshot {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _DriverLocationSnapshot value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _DriverLocationSnapshot() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _DriverLocationSnapshot value)  $default,){
final _that = this;
switch (_that) {
case _DriverLocationSnapshot():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _DriverLocationSnapshot value)?  $default,){
final _that = this;
switch (_that) {
case _DriverLocationSnapshot() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(name: 'ride_request_id')  String rideRequestId,  double latitude,  double longitude, @JsonKey(name: 'updated_at')  DateTime updatedAt)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _DriverLocationSnapshot() when $default != null:
return $default(_that.rideRequestId,_that.latitude,_that.longitude,_that.updatedAt);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(name: 'ride_request_id')  String rideRequestId,  double latitude,  double longitude, @JsonKey(name: 'updated_at')  DateTime updatedAt)  $default,) {final _that = this;
switch (_that) {
case _DriverLocationSnapshot():
return $default(_that.rideRequestId,_that.latitude,_that.longitude,_that.updatedAt);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(name: 'ride_request_id')  String rideRequestId,  double latitude,  double longitude, @JsonKey(name: 'updated_at')  DateTime updatedAt)?  $default,) {final _that = this;
switch (_that) {
case _DriverLocationSnapshot() when $default != null:
return $default(_that.rideRequestId,_that.latitude,_that.longitude,_that.updatedAt);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _DriverLocationSnapshot implements DriverLocationSnapshot {
  const _DriverLocationSnapshot({@JsonKey(name: 'ride_request_id') required this.rideRequestId, required this.latitude, required this.longitude, @JsonKey(name: 'updated_at') required this.updatedAt});
  factory _DriverLocationSnapshot.fromJson(Map<String, dynamic> json) => _$DriverLocationSnapshotFromJson(json);

@override@JsonKey(name: 'ride_request_id') final  String rideRequestId;
@override final  double latitude;
@override final  double longitude;
@override@JsonKey(name: 'updated_at') final  DateTime updatedAt;

/// Create a copy of DriverLocationSnapshot
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$DriverLocationSnapshotCopyWith<_DriverLocationSnapshot> get copyWith => __$DriverLocationSnapshotCopyWithImpl<_DriverLocationSnapshot>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$DriverLocationSnapshotToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _DriverLocationSnapshot&&(identical(other.rideRequestId, rideRequestId) || other.rideRequestId == rideRequestId)&&(identical(other.latitude, latitude) || other.latitude == latitude)&&(identical(other.longitude, longitude) || other.longitude == longitude)&&(identical(other.updatedAt, updatedAt) || other.updatedAt == updatedAt));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,rideRequestId,latitude,longitude,updatedAt);
}

@override
String toString() {
    return 'DriverLocationSnapshot(rideRequestId: $rideRequestId, latitude: $latitude, longitude: $longitude, updatedAt: $updatedAt)';
}


}

/// @nodoc
abstract mixin class _$DriverLocationSnapshotCopyWith<$Res> implements $DriverLocationSnapshotCopyWith<$Res> {
  factory _$DriverLocationSnapshotCopyWith(_DriverLocationSnapshot value, $Res Function(_DriverLocationSnapshot) _then) = __$DriverLocationSnapshotCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'ride_request_id') String rideRequestId, double latitude, double longitude,@JsonKey(name: 'updated_at') DateTime updatedAt
});




}
/// @nodoc
class __$DriverLocationSnapshotCopyWithImpl<$Res>
    implements _$DriverLocationSnapshotCopyWith<$Res> {
  __$DriverLocationSnapshotCopyWithImpl(this._self, this._then);

  final _DriverLocationSnapshot _self;
  final $Res Function(_DriverLocationSnapshot) _then;

/// Create a copy of DriverLocationSnapshot
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? rideRequestId = null,Object? latitude = null,Object? longitude = null,Object? updatedAt = null,}) {
  return _then(_DriverLocationSnapshot(
rideRequestId: null == rideRequestId ? _self.rideRequestId : rideRequestId // ignore: cast_nullable_to_non_nullable
as String,latitude: null == latitude ? _self.latitude : latitude // ignore: cast_nullable_to_non_nullable
as double,longitude: null == longitude ? _self.longitude : longitude // ignore: cast_nullable_to_non_nullable
as double,updatedAt: null == updatedAt ? _self.updatedAt : updatedAt // ignore: cast_nullable_to_non_nullable
as DateTime,
  ));
}


}

// dart format on
