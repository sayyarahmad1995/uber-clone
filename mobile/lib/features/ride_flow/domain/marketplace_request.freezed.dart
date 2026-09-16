// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'marketplace_request.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$RideLocation {

 double get latitude; double get longitude;
/// Create a copy of RideLocation
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$RideLocationCopyWith<RideLocation> get copyWith => _$RideLocationCopyWithImpl<RideLocation>(this as RideLocation, _$identity);

  /// Serializes this RideLocation to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as RideLocation;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is RideLocation&&(identical(other.latitude, _this.latitude) || other.latitude == _this.latitude)&&(identical(other.longitude, _this.longitude) || other.longitude == _this.longitude));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as RideLocation;
  return Object.hash(runtimeType,_this.latitude,_this.longitude);
}

@override
String toString() {
  final _this = this as RideLocation;
  return 'RideLocation(latitude: ${_this.latitude}, longitude: ${_this.longitude})';
}


}

/// @nodoc
abstract mixin class $RideLocationCopyWith<$Res>  {
  factory $RideLocationCopyWith(RideLocation value, $Res Function(RideLocation) _then) = _$RideLocationCopyWithImpl;
@useResult
$Res call({
 double latitude, double longitude
});




}
/// @nodoc
class _$RideLocationCopyWithImpl<$Res>
    implements $RideLocationCopyWith<$Res> {
  _$RideLocationCopyWithImpl(this._self, this._then);

  final RideLocation _self;
  final $Res Function(RideLocation) _then;

/// Create a copy of RideLocation
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? latitude = null,Object? longitude = null,}) {
  return _then(RideLocation(
latitude: null == latitude ? _self.latitude : latitude // ignore: cast_nullable_to_non_nullable
as double,longitude: null == longitude ? _self.longitude : longitude // ignore: cast_nullable_to_non_nullable
as double,
  ));
}

}


/// Adds pattern-matching-related methods to [RideLocation].
extension RideLocationPatterns on RideLocation {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _RideLocation value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _RideLocation() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _RideLocation value)  $default,){
final _that = this;
switch (_that) {
case _RideLocation():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _RideLocation value)?  $default,){
final _that = this;
switch (_that) {
case _RideLocation() when $default != null:
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
case _RideLocation() when $default != null:
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
case _RideLocation():
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
case _RideLocation() when $default != null:
return $default(_that.latitude,_that.longitude);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _RideLocation implements RideLocation {
  const _RideLocation({required this.latitude, required this.longitude});
  factory _RideLocation.fromJson(Map<String, dynamic> json) => _$RideLocationFromJson(json);

@override final  double latitude;
@override final  double longitude;

/// Create a copy of RideLocation
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$RideLocationCopyWith<_RideLocation> get copyWith => __$RideLocationCopyWithImpl<_RideLocation>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$RideLocationToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _RideLocation&&(identical(other.latitude, latitude) || other.latitude == latitude)&&(identical(other.longitude, longitude) || other.longitude == longitude));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,latitude,longitude);
}

@override
String toString() {
    return 'RideLocation(latitude: $latitude, longitude: $longitude)';
}


}

/// @nodoc
abstract mixin class _$RideLocationCopyWith<$Res> implements $RideLocationCopyWith<$Res> {
  factory _$RideLocationCopyWith(_RideLocation value, $Res Function(_RideLocation) _then) = __$RideLocationCopyWithImpl;
@override @useResult
$Res call({
 double latitude, double longitude
});




}
/// @nodoc
class __$RideLocationCopyWithImpl<$Res>
    implements _$RideLocationCopyWith<$Res> {
  __$RideLocationCopyWithImpl(this._self, this._then);

  final _RideLocation _self;
  final $Res Function(_RideLocation) _then;

/// Create a copy of RideLocation
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? latitude = null,Object? longitude = null,}) {
  return _then(_RideLocation(
latitude: null == latitude ? _self.latitude : latitude // ignore: cast_nullable_to_non_nullable
as double,longitude: null == longitude ? _self.longitude : longitude // ignore: cast_nullable_to_non_nullable
as double,
  ));
}


}


/// @nodoc
mixin _$RideFare {

@JsonKey(name: 'amount_minor') int get amountMinor; String get currency;
/// Create a copy of RideFare
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$RideFareCopyWith<RideFare> get copyWith => _$RideFareCopyWithImpl<RideFare>(this as RideFare, _$identity);

  /// Serializes this RideFare to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as RideFare;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is RideFare&&(identical(other.amountMinor, _this.amountMinor) || other.amountMinor == _this.amountMinor)&&(identical(other.currency, _this.currency) || other.currency == _this.currency));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as RideFare;
  return Object.hash(runtimeType,_this.amountMinor,_this.currency);
}

@override
String toString() {
  final _this = this as RideFare;
  return 'RideFare(amountMinor: ${_this.amountMinor}, currency: ${_this.currency})';
}


}

/// @nodoc
abstract mixin class $RideFareCopyWith<$Res>  {
  factory $RideFareCopyWith(RideFare value, $Res Function(RideFare) _then) = _$RideFareCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'amount_minor') int amountMinor, String currency
});




}
/// @nodoc
class _$RideFareCopyWithImpl<$Res>
    implements $RideFareCopyWith<$Res> {
  _$RideFareCopyWithImpl(this._self, this._then);

  final RideFare _self;
  final $Res Function(RideFare) _then;

/// Create a copy of RideFare
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? amountMinor = null,Object? currency = null,}) {
  return _then(RideFare(
amountMinor: null == amountMinor ? _self.amountMinor : amountMinor // ignore: cast_nullable_to_non_nullable
as int,currency: null == currency ? _self.currency : currency // ignore: cast_nullable_to_non_nullable
as String,
  ));
}

}


/// Adds pattern-matching-related methods to [RideFare].
extension RideFarePatterns on RideFare {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _RideFare value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _RideFare() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _RideFare value)  $default,){
final _that = this;
switch (_that) {
case _RideFare():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _RideFare value)?  $default,){
final _that = this;
switch (_that) {
case _RideFare() when $default != null:
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
case _RideFare() when $default != null:
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
case _RideFare():
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
case _RideFare() when $default != null:
return $default(_that.amountMinor,_that.currency);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _RideFare implements RideFare {
  const _RideFare({@JsonKey(name: 'amount_minor') required this.amountMinor, required this.currency});
  factory _RideFare.fromJson(Map<String, dynamic> json) => _$RideFareFromJson(json);

@override@JsonKey(name: 'amount_minor') final  int amountMinor;
@override final  String currency;

/// Create a copy of RideFare
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$RideFareCopyWith<_RideFare> get copyWith => __$RideFareCopyWithImpl<_RideFare>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$RideFareToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _RideFare&&(identical(other.amountMinor, amountMinor) || other.amountMinor == amountMinor)&&(identical(other.currency, currency) || other.currency == currency));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,amountMinor,currency);
}

@override
String toString() {
    return 'RideFare(amountMinor: $amountMinor, currency: $currency)';
}


}

/// @nodoc
abstract mixin class _$RideFareCopyWith<$Res> implements $RideFareCopyWith<$Res> {
  factory _$RideFareCopyWith(_RideFare value, $Res Function(_RideFare) _then) = __$RideFareCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'amount_minor') int amountMinor, String currency
});




}
/// @nodoc
class __$RideFareCopyWithImpl<$Res>
    implements _$RideFareCopyWith<$Res> {
  __$RideFareCopyWithImpl(this._self, this._then);

  final _RideFare _self;
  final $Res Function(_RideFare) _then;

/// Create a copy of RideFare
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? amountMinor = null,Object? currency = null,}) {
  return _then(_RideFare(
amountMinor: null == amountMinor ? _self.amountMinor : amountMinor // ignore: cast_nullable_to_non_nullable
as int,currency: null == currency ? _self.currency : currency // ignore: cast_nullable_to_non_nullable
as String,
  ));
}


}


/// @nodoc
mixin _$MarketplaceRequest {

 String get id; RideLocation get pickup; RideLocation get destination;@JsonKey(name: 'proposed_fare') RideFare get proposedFare;@JsonKey(name: 'created_at') DateTime get createdAt;@JsonKey(name: 'expires_at') DateTime get expiresAt;@JsonKey(name: 'response_deadline') DateTime get responseDeadline;@JsonKey(name: 'own_offer') RideOffer? get ownOffer;@JsonKey(name: 'pickup_distance_meters') int get pickupDistanceMeters;
/// Create a copy of MarketplaceRequest
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$MarketplaceRequestCopyWith<MarketplaceRequest> get copyWith => _$MarketplaceRequestCopyWithImpl<MarketplaceRequest>(this as MarketplaceRequest, _$identity);

  /// Serializes this MarketplaceRequest to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as MarketplaceRequest;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is MarketplaceRequest&&(identical(other.id, _this.id) || other.id == _this.id)&&(identical(other.pickup, _this.pickup) || other.pickup == _this.pickup)&&(identical(other.destination, _this.destination) || other.destination == _this.destination)&&(identical(other.proposedFare, _this.proposedFare) || other.proposedFare == _this.proposedFare)&&(identical(other.createdAt, _this.createdAt) || other.createdAt == _this.createdAt)&&(identical(other.expiresAt, _this.expiresAt) || other.expiresAt == _this.expiresAt)&&(identical(other.responseDeadline, _this.responseDeadline) || other.responseDeadline == _this.responseDeadline)&&(identical(other.ownOffer, _this.ownOffer) || other.ownOffer == _this.ownOffer)&&(identical(other.pickupDistanceMeters, _this.pickupDistanceMeters) || other.pickupDistanceMeters == _this.pickupDistanceMeters));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as MarketplaceRequest;
  return Object.hash(runtimeType,_this.id,_this.pickup,_this.destination,_this.proposedFare,_this.createdAt,_this.expiresAt,_this.responseDeadline,_this.ownOffer,_this.pickupDistanceMeters);
}

@override
String toString() {
  final _this = this as MarketplaceRequest;
  return 'MarketplaceRequest(id: ${_this.id}, pickup: ${_this.pickup}, destination: ${_this.destination}, proposedFare: ${_this.proposedFare}, createdAt: ${_this.createdAt}, expiresAt: ${_this.expiresAt}, responseDeadline: ${_this.responseDeadline}, ownOffer: ${_this.ownOffer}, pickupDistanceMeters: ${_this.pickupDistanceMeters})';
}


}

/// @nodoc
abstract mixin class $MarketplaceRequestCopyWith<$Res>  {
  factory $MarketplaceRequestCopyWith(MarketplaceRequest value, $Res Function(MarketplaceRequest) _then) = _$MarketplaceRequestCopyWithImpl;
@useResult
$Res call({
 String id, RideLocation pickup, RideLocation destination,@JsonKey(name: 'proposed_fare') RideFare proposedFare,@JsonKey(name: 'created_at') DateTime createdAt,@JsonKey(name: 'expires_at') DateTime expiresAt,@JsonKey(name: 'response_deadline') DateTime responseDeadline,@JsonKey(name: 'own_offer') RideOffer? ownOffer,@JsonKey(name: 'pickup_distance_meters') int pickupDistanceMeters
});


$RideLocationCopyWith<$Res> get pickup;$RideLocationCopyWith<$Res> get destination;$RideFareCopyWith<$Res> get proposedFare;$RideOfferCopyWith<$Res>? get ownOffer;

}
/// @nodoc
class _$MarketplaceRequestCopyWithImpl<$Res>
    implements $MarketplaceRequestCopyWith<$Res> {
  _$MarketplaceRequestCopyWithImpl(this._self, this._then);

  final MarketplaceRequest _self;
  final $Res Function(MarketplaceRequest) _then;

/// Create a copy of MarketplaceRequest
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? id = null,Object? pickup = null,Object? destination = null,Object? proposedFare = null,Object? createdAt = null,Object? expiresAt = null,Object? responseDeadline = null,Object? ownOffer = freezed,Object? pickupDistanceMeters = null,}) {
  return _then(MarketplaceRequest(
id: null == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as String,pickup: null == pickup ? _self.pickup : pickup // ignore: cast_nullable_to_non_nullable
as RideLocation,destination: null == destination ? _self.destination : destination // ignore: cast_nullable_to_non_nullable
as RideLocation,proposedFare: null == proposedFare ? _self.proposedFare : proposedFare // ignore: cast_nullable_to_non_nullable
as RideFare,createdAt: null == createdAt ? _self.createdAt : createdAt // ignore: cast_nullable_to_non_nullable
as DateTime,expiresAt: null == expiresAt ? _self.expiresAt : expiresAt // ignore: cast_nullable_to_non_nullable
as DateTime,responseDeadline: null == responseDeadline ? _self.responseDeadline : responseDeadline // ignore: cast_nullable_to_non_nullable
as DateTime,ownOffer: freezed == ownOffer ? _self.ownOffer : ownOffer // ignore: cast_nullable_to_non_nullable
as RideOffer?,pickupDistanceMeters: null == pickupDistanceMeters ? _self.pickupDistanceMeters : pickupDistanceMeters // ignore: cast_nullable_to_non_nullable
as int,
  ));
}
/// Create a copy of MarketplaceRequest
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$RideLocationCopyWith<$Res> get pickup {

  return $RideLocationCopyWith<$Res>(_self.pickup, (value) {
    return _then(_self.copyWith(pickup: value));
  });
}/// Create a copy of MarketplaceRequest
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$RideLocationCopyWith<$Res> get destination {

  return $RideLocationCopyWith<$Res>(_self.destination, (value) {
    return _then(_self.copyWith(destination: value));
  });
}/// Create a copy of MarketplaceRequest
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$RideFareCopyWith<$Res> get proposedFare {

  return $RideFareCopyWith<$Res>(_self.proposedFare, (value) {
    return _then(_self.copyWith(proposedFare: value));
  });
}/// Create a copy of MarketplaceRequest
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$RideOfferCopyWith<$Res>? get ownOffer {
    if (_self.ownOffer == null) {
    return null;
  }

  return $RideOfferCopyWith<$Res>(_self.ownOffer!, (value) {
    return _then(_self.copyWith(ownOffer: value));
  });
}
}


/// Adds pattern-matching-related methods to [MarketplaceRequest].
extension MarketplaceRequestPatterns on MarketplaceRequest {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _MarketplaceRequest value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _MarketplaceRequest() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _MarketplaceRequest value)  $default,){
final _that = this;
switch (_that) {
case _MarketplaceRequest():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _MarketplaceRequest value)?  $default,){
final _that = this;
switch (_that) {
case _MarketplaceRequest() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( String id,  RideLocation pickup,  RideLocation destination, @JsonKey(name: 'proposed_fare')  RideFare proposedFare, @JsonKey(name: 'created_at')  DateTime createdAt, @JsonKey(name: 'expires_at')  DateTime expiresAt, @JsonKey(name: 'response_deadline')  DateTime responseDeadline, @JsonKey(name: 'own_offer')  RideOffer? ownOffer, @JsonKey(name: 'pickup_distance_meters')  int pickupDistanceMeters)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _MarketplaceRequest() when $default != null:
return $default(_that.id,_that.pickup,_that.destination,_that.proposedFare,_that.createdAt,_that.expiresAt,_that.responseDeadline,_that.ownOffer,_that.pickupDistanceMeters);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( String id,  RideLocation pickup,  RideLocation destination, @JsonKey(name: 'proposed_fare')  RideFare proposedFare, @JsonKey(name: 'created_at')  DateTime createdAt, @JsonKey(name: 'expires_at')  DateTime expiresAt, @JsonKey(name: 'response_deadline')  DateTime responseDeadline, @JsonKey(name: 'own_offer')  RideOffer? ownOffer, @JsonKey(name: 'pickup_distance_meters')  int pickupDistanceMeters)  $default,) {final _that = this;
switch (_that) {
case _MarketplaceRequest():
return $default(_that.id,_that.pickup,_that.destination,_that.proposedFare,_that.createdAt,_that.expiresAt,_that.responseDeadline,_that.ownOffer,_that.pickupDistanceMeters);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( String id,  RideLocation pickup,  RideLocation destination, @JsonKey(name: 'proposed_fare')  RideFare proposedFare, @JsonKey(name: 'created_at')  DateTime createdAt, @JsonKey(name: 'expires_at')  DateTime expiresAt, @JsonKey(name: 'response_deadline')  DateTime responseDeadline, @JsonKey(name: 'own_offer')  RideOffer? ownOffer, @JsonKey(name: 'pickup_distance_meters')  int pickupDistanceMeters)?  $default,) {final _that = this;
switch (_that) {
case _MarketplaceRequest() when $default != null:
return $default(_that.id,_that.pickup,_that.destination,_that.proposedFare,_that.createdAt,_that.expiresAt,_that.responseDeadline,_that.ownOffer,_that.pickupDistanceMeters);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _MarketplaceRequest implements MarketplaceRequest {
  const _MarketplaceRequest({required this.id, required this.pickup, required this.destination, @JsonKey(name: 'proposed_fare') required this.proposedFare, @JsonKey(name: 'created_at') required this.createdAt, @JsonKey(name: 'expires_at') required this.expiresAt, @JsonKey(name: 'response_deadline') required this.responseDeadline, @JsonKey(name: 'own_offer') this.ownOffer, @JsonKey(name: 'pickup_distance_meters') required this.pickupDistanceMeters});
  factory _MarketplaceRequest.fromJson(Map<String, dynamic> json) => _$MarketplaceRequestFromJson(json);

@override final  String id;
@override final  RideLocation pickup;
@override final  RideLocation destination;
@override@JsonKey(name: 'proposed_fare') final  RideFare proposedFare;
@override@JsonKey(name: 'created_at') final  DateTime createdAt;
@override@JsonKey(name: 'expires_at') final  DateTime expiresAt;
@override@JsonKey(name: 'response_deadline') final  DateTime responseDeadline;
@override@JsonKey(name: 'own_offer') final  RideOffer? ownOffer;
@override@JsonKey(name: 'pickup_distance_meters') final  int pickupDistanceMeters;

/// Create a copy of MarketplaceRequest
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$MarketplaceRequestCopyWith<_MarketplaceRequest> get copyWith => __$MarketplaceRequestCopyWithImpl<_MarketplaceRequest>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$MarketplaceRequestToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _MarketplaceRequest&&(identical(other.id, id) || other.id == id)&&(identical(other.pickup, pickup) || other.pickup == pickup)&&(identical(other.destination, destination) || other.destination == destination)&&(identical(other.proposedFare, proposedFare) || other.proposedFare == proposedFare)&&(identical(other.createdAt, createdAt) || other.createdAt == createdAt)&&(identical(other.expiresAt, expiresAt) || other.expiresAt == expiresAt)&&(identical(other.responseDeadline, responseDeadline) || other.responseDeadline == responseDeadline)&&(identical(other.ownOffer, ownOffer) || other.ownOffer == ownOffer)&&(identical(other.pickupDistanceMeters, pickupDistanceMeters) || other.pickupDistanceMeters == pickupDistanceMeters));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,id,pickup,destination,proposedFare,createdAt,expiresAt,responseDeadline,ownOffer,pickupDistanceMeters);
}

@override
String toString() {
    return 'MarketplaceRequest(id: $id, pickup: $pickup, destination: $destination, proposedFare: $proposedFare, createdAt: $createdAt, expiresAt: $expiresAt, responseDeadline: $responseDeadline, ownOffer: $ownOffer, pickupDistanceMeters: $pickupDistanceMeters)';
}


}

/// @nodoc
abstract mixin class _$MarketplaceRequestCopyWith<$Res> implements $MarketplaceRequestCopyWith<$Res> {
  factory _$MarketplaceRequestCopyWith(_MarketplaceRequest value, $Res Function(_MarketplaceRequest) _then) = __$MarketplaceRequestCopyWithImpl;
@override @useResult
$Res call({
 String id, RideLocation pickup, RideLocation destination,@JsonKey(name: 'proposed_fare') RideFare proposedFare,@JsonKey(name: 'created_at') DateTime createdAt,@JsonKey(name: 'expires_at') DateTime expiresAt,@JsonKey(name: 'response_deadline') DateTime responseDeadline,@JsonKey(name: 'own_offer') RideOffer? ownOffer,@JsonKey(name: 'pickup_distance_meters') int pickupDistanceMeters
});


@override $RideLocationCopyWith<$Res> get pickup;@override $RideLocationCopyWith<$Res> get destination;@override $RideFareCopyWith<$Res> get proposedFare;@override $RideOfferCopyWith<$Res>? get ownOffer;

}
/// @nodoc
class __$MarketplaceRequestCopyWithImpl<$Res>
    implements _$MarketplaceRequestCopyWith<$Res> {
  __$MarketplaceRequestCopyWithImpl(this._self, this._then);

  final _MarketplaceRequest _self;
  final $Res Function(_MarketplaceRequest) _then;

/// Create a copy of MarketplaceRequest
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? id = null,Object? pickup = null,Object? destination = null,Object? proposedFare = null,Object? createdAt = null,Object? expiresAt = null,Object? responseDeadline = null,Object? ownOffer = freezed,Object? pickupDistanceMeters = null,}) {
  return _then(_MarketplaceRequest(
id: null == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as String,pickup: null == pickup ? _self.pickup : pickup // ignore: cast_nullable_to_non_nullable
as RideLocation,destination: null == destination ? _self.destination : destination // ignore: cast_nullable_to_non_nullable
as RideLocation,proposedFare: null == proposedFare ? _self.proposedFare : proposedFare // ignore: cast_nullable_to_non_nullable
as RideFare,createdAt: null == createdAt ? _self.createdAt : createdAt // ignore: cast_nullable_to_non_nullable
as DateTime,expiresAt: null == expiresAt ? _self.expiresAt : expiresAt // ignore: cast_nullable_to_non_nullable
as DateTime,responseDeadline: null == responseDeadline ? _self.responseDeadline : responseDeadline // ignore: cast_nullable_to_non_nullable
as DateTime,ownOffer: freezed == ownOffer ? _self.ownOffer : ownOffer // ignore: cast_nullable_to_non_nullable
as RideOffer?,pickupDistanceMeters: null == pickupDistanceMeters ? _self.pickupDistanceMeters : pickupDistanceMeters // ignore: cast_nullable_to_non_nullable
as int,
  ));
}

/// Create a copy of MarketplaceRequest
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$RideLocationCopyWith<$Res> get pickup {

  return $RideLocationCopyWith<$Res>(_self.pickup, (value) {
    return _then(_self.copyWith(pickup: value));
  });
}/// Create a copy of MarketplaceRequest
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$RideLocationCopyWith<$Res> get destination {

  return $RideLocationCopyWith<$Res>(_self.destination, (value) {
    return _then(_self.copyWith(destination: value));
  });
}/// Create a copy of MarketplaceRequest
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$RideFareCopyWith<$Res> get proposedFare {

  return $RideFareCopyWith<$Res>(_self.proposedFare, (value) {
    return _then(_self.copyWith(proposedFare: value));
  });
}/// Create a copy of MarketplaceRequest
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$RideOfferCopyWith<$Res>? get ownOffer {
    if (_self.ownOffer == null) {
    return null;
  }

  return $RideOfferCopyWith<$Res>(_self.ownOffer!, (value) {
    return _then(_self.copyWith(ownOffer: value));
  });
}
}

// dart format on
