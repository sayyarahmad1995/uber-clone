// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'ride_request.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$GeoPoint {

 double get latitude; double get longitude;
/// Create a copy of GeoPoint
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$GeoPointCopyWith<GeoPoint> get copyWith => _$GeoPointCopyWithImpl<GeoPoint>(this as GeoPoint, _$identity);

  /// Serializes this GeoPoint to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as GeoPoint;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is GeoPoint&&(identical(other.latitude, _this.latitude) || other.latitude == _this.latitude)&&(identical(other.longitude, _this.longitude) || other.longitude == _this.longitude));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as GeoPoint;
  return Object.hash(runtimeType,_this.latitude,_this.longitude);
}

@override
String toString() {
  final _this = this as GeoPoint;
  return 'GeoPoint(latitude: ${_this.latitude}, longitude: ${_this.longitude})';
}


}

/// @nodoc
abstract mixin class $GeoPointCopyWith<$Res>  {
  factory $GeoPointCopyWith(GeoPoint value, $Res Function(GeoPoint) _then) = _$GeoPointCopyWithImpl;
@useResult
$Res call({
 double latitude, double longitude
});




}
/// @nodoc
class _$GeoPointCopyWithImpl<$Res>
    implements $GeoPointCopyWith<$Res> {
  _$GeoPointCopyWithImpl(this._self, this._then);

  final GeoPoint _self;
  final $Res Function(GeoPoint) _then;

/// Create a copy of GeoPoint
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? latitude = null,Object? longitude = null,}) {
  return _then(GeoPoint(
latitude: null == latitude ? _self.latitude : latitude // ignore: cast_nullable_to_non_nullable
as double,longitude: null == longitude ? _self.longitude : longitude // ignore: cast_nullable_to_non_nullable
as double,
  ));
}

}


/// Adds pattern-matching-related methods to [GeoPoint].
extension GeoPointPatterns on GeoPoint {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _GeoPoint value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _GeoPoint() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _GeoPoint value)  $default,){
final _that = this;
switch (_that) {
case _GeoPoint():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _GeoPoint value)?  $default,){
final _that = this;
switch (_that) {
case _GeoPoint() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( double latitude,  double longitude)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _GeoPoint() when $default != null:
return $default(_that.latitude,_that.longitude);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( double latitude,  double longitude)  $default,) {final _that = this;
switch (_that) {
case _GeoPoint():
return $default(_that.latitude,_that.longitude);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( double latitude,  double longitude)?  $default,) {final _that = this;
switch (_that) {
case _GeoPoint() when $default != null:
return $default(_that.latitude,_that.longitude);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _GeoPoint implements GeoPoint {
  const _GeoPoint({required this.latitude, required this.longitude});
  factory _GeoPoint.fromJson(Map<String, dynamic> json) => _$GeoPointFromJson(json);

@override final  double latitude;
@override final  double longitude;

/// Create a copy of GeoPoint
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$GeoPointCopyWith<_GeoPoint> get copyWith => __$GeoPointCopyWithImpl<_GeoPoint>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$GeoPointToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _GeoPoint&&(identical(other.latitude, latitude) || other.latitude == latitude)&&(identical(other.longitude, longitude) || other.longitude == longitude));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,latitude,longitude);
}

@override
String toString() {
    return 'GeoPoint(latitude: $latitude, longitude: $longitude)';
}


}

/// @nodoc
abstract mixin class _$GeoPointCopyWith<$Res> implements $GeoPointCopyWith<$Res> {
  factory _$GeoPointCopyWith(_GeoPoint value, $Res Function(_GeoPoint) _then) = __$GeoPointCopyWithImpl;
@override @useResult
$Res call({
 double latitude, double longitude
});




}
/// @nodoc
class __$GeoPointCopyWithImpl<$Res>
    implements _$GeoPointCopyWith<$Res> {
  __$GeoPointCopyWithImpl(this._self, this._then);

  final _GeoPoint _self;
  final $Res Function(_GeoPoint) _then;

/// Create a copy of GeoPoint
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? latitude = null,Object? longitude = null,}) {
  return _then(_GeoPoint(
latitude: null == latitude ? _self.latitude : latitude // ignore: cast_nullable_to_non_nullable
as double,longitude: null == longitude ? _self.longitude : longitude // ignore: cast_nullable_to_non_nullable
as double,
  ));
}


}


/// @nodoc
mixin _$Money {

@JsonKey(name: 'amount_minor') int get amountMinor; String get currency;
/// Create a copy of Money
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$MoneyCopyWith<Money> get copyWith => _$MoneyCopyWithImpl<Money>(this as Money, _$identity);

  /// Serializes this Money to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as Money;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is Money&&(identical(other.amountMinor, _this.amountMinor) || other.amountMinor == _this.amountMinor)&&(identical(other.currency, _this.currency) || other.currency == _this.currency));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as Money;
  return Object.hash(runtimeType,_this.amountMinor,_this.currency);
}

@override
String toString() {
  final _this = this as Money;
  return 'Money(amountMinor: ${_this.amountMinor}, currency: ${_this.currency})';
}


}

/// @nodoc
abstract mixin class $MoneyCopyWith<$Res>  {
  factory $MoneyCopyWith(Money value, $Res Function(Money) _then) = _$MoneyCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'amount_minor') int amountMinor, String currency
});




}
/// @nodoc
class _$MoneyCopyWithImpl<$Res>
    implements $MoneyCopyWith<$Res> {
  _$MoneyCopyWithImpl(this._self, this._then);

  final Money _self;
  final $Res Function(Money) _then;

/// Create a copy of Money
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? amountMinor = null,Object? currency = null,}) {
  return _then(Money(
amountMinor: null == amountMinor ? _self.amountMinor : amountMinor // ignore: cast_nullable_to_non_nullable
as int,currency: null == currency ? _self.currency : currency // ignore: cast_nullable_to_non_nullable
as String,
  ));
}

}


/// Adds pattern-matching-related methods to [Money].
extension MoneyPatterns on Money {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _Money value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _Money() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _Money value)  $default,){
final _that = this;
switch (_that) {
case _Money():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _Money value)?  $default,){
final _that = this;
switch (_that) {
case _Money() when $default != null:
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
case _Money() when $default != null:
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
case _Money():
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
case _Money() when $default != null:
return $default(_that.amountMinor,_that.currency);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _Money implements Money {
  const _Money({@JsonKey(name: 'amount_minor') required this.amountMinor, required this.currency});
  factory _Money.fromJson(Map<String, dynamic> json) => _$MoneyFromJson(json);

@override@JsonKey(name: 'amount_minor') final  int amountMinor;
@override final  String currency;

/// Create a copy of Money
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$MoneyCopyWith<_Money> get copyWith => __$MoneyCopyWithImpl<_Money>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$MoneyToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _Money&&(identical(other.amountMinor, amountMinor) || other.amountMinor == amountMinor)&&(identical(other.currency, currency) || other.currency == currency));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,amountMinor,currency);
}

@override
String toString() {
    return 'Money(amountMinor: $amountMinor, currency: $currency)';
}


}

/// @nodoc
abstract mixin class _$MoneyCopyWith<$Res> implements $MoneyCopyWith<$Res> {
  factory _$MoneyCopyWith(_Money value, $Res Function(_Money) _then) = __$MoneyCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'amount_minor') int amountMinor, String currency
});




}
/// @nodoc
class __$MoneyCopyWithImpl<$Res>
    implements _$MoneyCopyWith<$Res> {
  __$MoneyCopyWithImpl(this._self, this._then);

  final _Money _self;
  final $Res Function(_Money) _then;

/// Create a copy of Money
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? amountMinor = null,Object? currency = null,}) {
  return _then(_Money(
amountMinor: null == amountMinor ? _self.amountMinor : amountMinor // ignore: cast_nullable_to_non_nullable
as int,currency: null == currency ? _self.currency : currency // ignore: cast_nullable_to_non_nullable
as String,
  ));
}


}


/// @nodoc
mixin _$TripSnapshot {

 String get status;
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
  return identical(this, other) || (other.runtimeType == runtimeType&&other is TripSnapshot&&(identical(other.status, _this.status) || other.status == _this.status));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as TripSnapshot;
  return Object.hash(runtimeType,_this.status);
}

@override
String toString() {
  final _this = this as TripSnapshot;
  return 'TripSnapshot(status: ${_this.status})';
}


}

/// @nodoc
abstract mixin class $TripSnapshotCopyWith<$Res>  {
  factory $TripSnapshotCopyWith(TripSnapshot value, $Res Function(TripSnapshot) _then) = _$TripSnapshotCopyWithImpl;
@useResult
$Res call({
 String status
});




}
/// @nodoc
class _$TripSnapshotCopyWithImpl<$Res>
    implements $TripSnapshotCopyWith<$Res> {
  _$TripSnapshotCopyWithImpl(this._self, this._then);

  final TripSnapshot _self;
  final $Res Function(TripSnapshot) _then;

/// Create a copy of TripSnapshot
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? status = null,}) {
  return _then(TripSnapshot(
status: null == status ? _self.status : status // ignore: cast_nullable_to_non_nullable
as String,
  ));
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( String status)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _TripSnapshot() when $default != null:
return $default(_that.status);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( String status)  $default,) {final _that = this;
switch (_that) {
case _TripSnapshot():
return $default(_that.status);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( String status)?  $default,) {final _that = this;
switch (_that) {
case _TripSnapshot() when $default != null:
return $default(_that.status);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _TripSnapshot implements TripSnapshot {
  const _TripSnapshot({required this.status});
  factory _TripSnapshot.fromJson(Map<String, dynamic> json) => _$TripSnapshotFromJson(json);

@override final  String status;

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
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _TripSnapshot&&(identical(other.status, status) || other.status == status));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,status);
}

@override
String toString() {
    return 'TripSnapshot(status: $status)';
}


}

/// @nodoc
abstract mixin class _$TripSnapshotCopyWith<$Res> implements $TripSnapshotCopyWith<$Res> {
  factory _$TripSnapshotCopyWith(_TripSnapshot value, $Res Function(_TripSnapshot) _then) = __$TripSnapshotCopyWithImpl;
@override @useResult
$Res call({
 String status
});




}
/// @nodoc
class __$TripSnapshotCopyWithImpl<$Res>
    implements _$TripSnapshotCopyWith<$Res> {
  __$TripSnapshotCopyWithImpl(this._self, this._then);

  final _TripSnapshot _self;
  final $Res Function(_TripSnapshot) _then;

/// Create a copy of TripSnapshot
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? status = null,}) {
  return _then(_TripSnapshot(
status: null == status ? _self.status : status // ignore: cast_nullable_to_non_nullable
as String,
  ));
}


}


/// @nodoc
mixin _$RideRequest {

 String get id; GeoPoint get pickup; GeoPoint get destination;@JsonKey(name: 'proposed_fare') Money get proposedFare; String get status;@JsonKey(name: 'created_at') DateTime get createdAt; TripSnapshot? get trip;
/// Create a copy of RideRequest
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$RideRequestCopyWith<RideRequest> get copyWith => _$RideRequestCopyWithImpl<RideRequest>(this as RideRequest, _$identity);

  /// Serializes this RideRequest to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as RideRequest;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is RideRequest&&(identical(other.id, _this.id) || other.id == _this.id)&&(identical(other.pickup, _this.pickup) || other.pickup == _this.pickup)&&(identical(other.destination, _this.destination) || other.destination == _this.destination)&&(identical(other.proposedFare, _this.proposedFare) || other.proposedFare == _this.proposedFare)&&(identical(other.status, _this.status) || other.status == _this.status)&&(identical(other.createdAt, _this.createdAt) || other.createdAt == _this.createdAt)&&(identical(other.trip, _this.trip) || other.trip == _this.trip));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as RideRequest;
  return Object.hash(runtimeType,_this.id,_this.pickup,_this.destination,_this.proposedFare,_this.status,_this.createdAt,_this.trip);
}

@override
String toString() {
  final _this = this as RideRequest;
  return 'RideRequest(id: ${_this.id}, pickup: ${_this.pickup}, destination: ${_this.destination}, proposedFare: ${_this.proposedFare}, status: ${_this.status}, createdAt: ${_this.createdAt}, trip: ${_this.trip})';
}


}

/// @nodoc
abstract mixin class $RideRequestCopyWith<$Res>  {
  factory $RideRequestCopyWith(RideRequest value, $Res Function(RideRequest) _then) = _$RideRequestCopyWithImpl;
@useResult
$Res call({
 String id, GeoPoint pickup, GeoPoint destination,@JsonKey(name: 'proposed_fare') Money proposedFare, String status,@JsonKey(name: 'created_at') DateTime createdAt, TripSnapshot? trip
});


$GeoPointCopyWith<$Res> get pickup;$GeoPointCopyWith<$Res> get destination;$MoneyCopyWith<$Res> get proposedFare;$TripSnapshotCopyWith<$Res>? get trip;

}
/// @nodoc
class _$RideRequestCopyWithImpl<$Res>
    implements $RideRequestCopyWith<$Res> {
  _$RideRequestCopyWithImpl(this._self, this._then);

  final RideRequest _self;
  final $Res Function(RideRequest) _then;

/// Create a copy of RideRequest
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? id = null,Object? pickup = null,Object? destination = null,Object? proposedFare = null,Object? status = null,Object? createdAt = null,Object? trip = freezed,}) {
  return _then(RideRequest(
id: null == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as String,pickup: null == pickup ? _self.pickup : pickup // ignore: cast_nullable_to_non_nullable
as GeoPoint,destination: null == destination ? _self.destination : destination // ignore: cast_nullable_to_non_nullable
as GeoPoint,proposedFare: null == proposedFare ? _self.proposedFare : proposedFare // ignore: cast_nullable_to_non_nullable
as Money,status: null == status ? _self.status : status // ignore: cast_nullable_to_non_nullable
as String,createdAt: null == createdAt ? _self.createdAt : createdAt // ignore: cast_nullable_to_non_nullable
as DateTime,trip: freezed == trip ? _self.trip : trip // ignore: cast_nullable_to_non_nullable
as TripSnapshot?,
  ));
}
/// Create a copy of RideRequest
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$GeoPointCopyWith<$Res> get pickup {

  return $GeoPointCopyWith<$Res>(_self.pickup, (value) {
    return _then(_self.copyWith(pickup: value));
  });
}/// Create a copy of RideRequest
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$GeoPointCopyWith<$Res> get destination {

  return $GeoPointCopyWith<$Res>(_self.destination, (value) {
    return _then(_self.copyWith(destination: value));
  });
}/// Create a copy of RideRequest
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$MoneyCopyWith<$Res> get proposedFare {

  return $MoneyCopyWith<$Res>(_self.proposedFare, (value) {
    return _then(_self.copyWith(proposedFare: value));
  });
}/// Create a copy of RideRequest
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


/// Adds pattern-matching-related methods to [RideRequest].
extension RideRequestPatterns on RideRequest {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _RideRequest value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _RideRequest() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _RideRequest value)  $default,){
final _that = this;
switch (_that) {
case _RideRequest():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _RideRequest value)?  $default,){
final _that = this;
switch (_that) {
case _RideRequest() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( String id,  GeoPoint pickup,  GeoPoint destination, @JsonKey(name: 'proposed_fare')  Money proposedFare,  String status, @JsonKey(name: 'created_at')  DateTime createdAt,  TripSnapshot? trip)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _RideRequest() when $default != null:
return $default(_that.id,_that.pickup,_that.destination,_that.proposedFare,_that.status,_that.createdAt,_that.trip);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( String id,  GeoPoint pickup,  GeoPoint destination, @JsonKey(name: 'proposed_fare')  Money proposedFare,  String status, @JsonKey(name: 'created_at')  DateTime createdAt,  TripSnapshot? trip)  $default,) {final _that = this;
switch (_that) {
case _RideRequest():
return $default(_that.id,_that.pickup,_that.destination,_that.proposedFare,_that.status,_that.createdAt,_that.trip);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( String id,  GeoPoint pickup,  GeoPoint destination, @JsonKey(name: 'proposed_fare')  Money proposedFare,  String status, @JsonKey(name: 'created_at')  DateTime createdAt,  TripSnapshot? trip)?  $default,) {final _that = this;
switch (_that) {
case _RideRequest() when $default != null:
return $default(_that.id,_that.pickup,_that.destination,_that.proposedFare,_that.status,_that.createdAt,_that.trip);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _RideRequest implements RideRequest {
  const _RideRequest({required this.id, required this.pickup, required this.destination, @JsonKey(name: 'proposed_fare') required this.proposedFare, required this.status, @JsonKey(name: 'created_at') required this.createdAt, this.trip});
  factory _RideRequest.fromJson(Map<String, dynamic> json) => _$RideRequestFromJson(json);

@override final  String id;
@override final  GeoPoint pickup;
@override final  GeoPoint destination;
@override@JsonKey(name: 'proposed_fare') final  Money proposedFare;
@override final  String status;
@override@JsonKey(name: 'created_at') final  DateTime createdAt;
@override final  TripSnapshot? trip;

/// Create a copy of RideRequest
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$RideRequestCopyWith<_RideRequest> get copyWith => __$RideRequestCopyWithImpl<_RideRequest>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$RideRequestToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _RideRequest&&(identical(other.id, id) || other.id == id)&&(identical(other.pickup, pickup) || other.pickup == pickup)&&(identical(other.destination, destination) || other.destination == destination)&&(identical(other.proposedFare, proposedFare) || other.proposedFare == proposedFare)&&(identical(other.status, status) || other.status == status)&&(identical(other.createdAt, createdAt) || other.createdAt == createdAt)&&(identical(other.trip, trip) || other.trip == trip));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,id,pickup,destination,proposedFare,status,createdAt,trip);
}

@override
String toString() {
    return 'RideRequest(id: $id, pickup: $pickup, destination: $destination, proposedFare: $proposedFare, status: $status, createdAt: $createdAt, trip: $trip)';
}


}

/// @nodoc
abstract mixin class _$RideRequestCopyWith<$Res> implements $RideRequestCopyWith<$Res> {
  factory _$RideRequestCopyWith(_RideRequest value, $Res Function(_RideRequest) _then) = __$RideRequestCopyWithImpl;
@override @useResult
$Res call({
 String id, GeoPoint pickup, GeoPoint destination,@JsonKey(name: 'proposed_fare') Money proposedFare, String status,@JsonKey(name: 'created_at') DateTime createdAt, TripSnapshot? trip
});


@override $GeoPointCopyWith<$Res> get pickup;@override $GeoPointCopyWith<$Res> get destination;@override $MoneyCopyWith<$Res> get proposedFare;@override $TripSnapshotCopyWith<$Res>? get trip;

}
/// @nodoc
class __$RideRequestCopyWithImpl<$Res>
    implements _$RideRequestCopyWith<$Res> {
  __$RideRequestCopyWithImpl(this._self, this._then);

  final _RideRequest _self;
  final $Res Function(_RideRequest) _then;

/// Create a copy of RideRequest
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? id = null,Object? pickup = null,Object? destination = null,Object? proposedFare = null,Object? status = null,Object? createdAt = null,Object? trip = freezed,}) {
  return _then(_RideRequest(
id: null == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as String,pickup: null == pickup ? _self.pickup : pickup // ignore: cast_nullable_to_non_nullable
as GeoPoint,destination: null == destination ? _self.destination : destination // ignore: cast_nullable_to_non_nullable
as GeoPoint,proposedFare: null == proposedFare ? _self.proposedFare : proposedFare // ignore: cast_nullable_to_non_nullable
as Money,status: null == status ? _self.status : status // ignore: cast_nullable_to_non_nullable
as String,createdAt: null == createdAt ? _self.createdAt : createdAt // ignore: cast_nullable_to_non_nullable
as DateTime,trip: freezed == trip ? _self.trip : trip // ignore: cast_nullable_to_non_nullable
as TripSnapshot?,
  ));
}

/// Create a copy of RideRequest
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$GeoPointCopyWith<$Res> get pickup {

  return $GeoPointCopyWith<$Res>(_self.pickup, (value) {
    return _then(_self.copyWith(pickup: value));
  });
}/// Create a copy of RideRequest
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$GeoPointCopyWith<$Res> get destination {

  return $GeoPointCopyWith<$Res>(_self.destination, (value) {
    return _then(_self.copyWith(destination: value));
  });
}/// Create a copy of RideRequest
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$MoneyCopyWith<$Res> get proposedFare {

  return $MoneyCopyWith<$Res>(_self.proposedFare, (value) {
    return _then(_self.copyWith(proposedFare: value));
  });
}/// Create a copy of RideRequest
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
mixin _$RideRequestList {

@JsonKey(name: 'ride_requests') List<RideRequest> get rideRequests;
/// Create a copy of RideRequestList
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$RideRequestListCopyWith<RideRequestList> get copyWith => _$RideRequestListCopyWithImpl<RideRequestList>(this as RideRequestList, _$identity);

  /// Serializes this RideRequestList to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as RideRequestList;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is RideRequestList&&const DeepCollectionEquality().equals(other.rideRequests, _this.rideRequests));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as RideRequestList;
  return Object.hash(runtimeType,const DeepCollectionEquality().hash(_this.rideRequests));
}

@override
String toString() {
  final _this = this as RideRequestList;
  return 'RideRequestList(rideRequests: ${_this.rideRequests})';
}


}

/// @nodoc
abstract mixin class $RideRequestListCopyWith<$Res>  {
  factory $RideRequestListCopyWith(RideRequestList value, $Res Function(RideRequestList) _then) = _$RideRequestListCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'ride_requests') List<RideRequest> rideRequests
});




}
/// @nodoc
class _$RideRequestListCopyWithImpl<$Res>
    implements $RideRequestListCopyWith<$Res> {
  _$RideRequestListCopyWithImpl(this._self, this._then);

  final RideRequestList _self;
  final $Res Function(RideRequestList) _then;

/// Create a copy of RideRequestList
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? rideRequests = null,}) {
  return _then(RideRequestList(
rideRequests: null == rideRequests ? _self.rideRequests : rideRequests // ignore: cast_nullable_to_non_nullable
as List<RideRequest>,
  ));
}

}


/// Adds pattern-matching-related methods to [RideRequestList].
extension RideRequestListPatterns on RideRequestList {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _RideRequestList value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _RideRequestList() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _RideRequestList value)  $default,){
final _that = this;
switch (_that) {
case _RideRequestList():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _RideRequestList value)?  $default,){
final _that = this;
switch (_that) {
case _RideRequestList() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(name: 'ride_requests')  List<RideRequest> rideRequests)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _RideRequestList() when $default != null:
return $default(_that.rideRequests);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(name: 'ride_requests')  List<RideRequest> rideRequests)  $default,) {final _that = this;
switch (_that) {
case _RideRequestList():
return $default(_that.rideRequests);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(name: 'ride_requests')  List<RideRequest> rideRequests)?  $default,) {final _that = this;
switch (_that) {
case _RideRequestList() when $default != null:
return $default(_that.rideRequests);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _RideRequestList implements RideRequestList {
  const _RideRequestList({@JsonKey(name: 'ride_requests') required  List<RideRequest> rideRequests}): _rideRequests = rideRequests;
  factory _RideRequestList.fromJson(Map<String, dynamic> json) => _$RideRequestListFromJson(json);

 final  List<RideRequest> _rideRequests;
@override@JsonKey(name: 'ride_requests') List<RideRequest> get rideRequests {
  if (_rideRequests is EqualUnmodifiableListView) return _rideRequests;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_rideRequests);
}


/// Create a copy of RideRequestList
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$RideRequestListCopyWith<_RideRequestList> get copyWith => __$RideRequestListCopyWithImpl<_RideRequestList>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$RideRequestListToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _RideRequestList&&const DeepCollectionEquality().equals(other.rideRequests, _rideRequests));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,const DeepCollectionEquality().hash(_rideRequests));
}

@override
String toString() {
    return 'RideRequestList(rideRequests: $rideRequests)';
}


}

/// @nodoc
abstract mixin class _$RideRequestListCopyWith<$Res> implements $RideRequestListCopyWith<$Res> {
  factory _$RideRequestListCopyWith(_RideRequestList value, $Res Function(_RideRequestList) _then) = __$RideRequestListCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'ride_requests') List<RideRequest> rideRequests
});




}
/// @nodoc
class __$RideRequestListCopyWithImpl<$Res>
    implements _$RideRequestListCopyWith<$Res> {
  __$RideRequestListCopyWithImpl(this._self, this._then);

  final _RideRequestList _self;
  final $Res Function(_RideRequestList) _then;

/// Create a copy of RideRequestList
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? rideRequests = null,}) {
  return _then(_RideRequestList(
rideRequests: null == rideRequests ? _self._rideRequests : rideRequests // ignore: cast_nullable_to_non_nullable
as List<RideRequest>,
  ));
}


}

// dart format on
