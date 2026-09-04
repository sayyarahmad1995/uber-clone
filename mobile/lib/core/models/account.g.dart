// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'account.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_Account _$AccountFromJson(Map<String, dynamic> json) => _Account(
  id: json['id'] as String,
  capabilities: (json['capabilities'] as List<dynamic>)
      .map((e) => $enumDecode(_$CapabilityEnumMap, e))
      .toList(),
);

Map<String, dynamic> _$AccountToJson(_Account instance) => <String, dynamic>{
  'id': instance.id,
  'capabilities': instance.capabilities
      .map((e) => _$CapabilityEnumMap[e]!)
      .toList(),
};

const _$CapabilityEnumMap = {
  Capability.rider: 'rider',
  Capability.driver: 'driver',
};

_AuthSession _$AuthSessionFromJson(Map<String, dynamic> json) => _AuthSession(
  accessToken: json['access_token'] as String,
  expiresIn: (json['expires_in'] as num).toInt(),
);

Map<String, dynamic> _$AuthSessionToJson(_AuthSession instance) =>
    <String, dynamic>{
      'access_token': instance.accessToken,
      'expires_in': instance.expiresIn,
    };

_VerificationChallenge _$VerificationChallengeFromJson(
  Map<String, dynamic> json,
) => _VerificationChallenge(verificationId: json['verification_id'] as String);

Map<String, dynamic> _$VerificationChallengeToJson(
  _VerificationChallenge instance,
) => <String, dynamic>{'verification_id': instance.verificationId};
