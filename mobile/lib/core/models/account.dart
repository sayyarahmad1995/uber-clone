import 'package:freezed_annotation/freezed_annotation.dart';

part 'account.freezed.dart';
part 'account.g.dart';

enum Capability { rider, driver }

@freezed
abstract class Account with _$Account {
  const factory Account({
    required String id,
    required List<Capability> capabilities,
  }) = _Account;
  factory Account.fromJson(Map<String, dynamic> json) =>
      _$AccountFromJson(json);
}

@freezed
abstract class AuthSession with _$AuthSession {
  const factory AuthSession({
    @JsonKey(name: 'access_token') required String accessToken,
    @JsonKey(name: 'expires_in') required int expiresIn,
  }) = _AuthSession;
  factory AuthSession.fromJson(Map<String, dynamic> json) =>
      _$AuthSessionFromJson(json);
}

@freezed
abstract class VerificationChallenge with _$VerificationChallenge {
  const factory VerificationChallenge({
    @JsonKey(name: 'verification_id') required String verificationId,
  }) = _VerificationChallenge;
  factory VerificationChallenge.fromJson(Map<String, dynamic> json) =>
      _$VerificationChallengeFromJson(json);
}
