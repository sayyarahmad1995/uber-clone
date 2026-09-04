import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:shared_preferences/shared_preferences.dart';

import '../models/account.dart';

abstract interface class SessionStore {
  Future<String?> readValidToken();
  Future<void> save(String token, DateTime expiresAt);
  Future<void> clear();
}

class SecureSessionStore implements SessionStore {
  SecureSessionStore(this._storage);
  final FlutterSecureStorage _storage;
  static const _tokenKey = 'auth_access_token';
  static const _expiryKey = 'auth_expires_at';

  @override
  Future<String?> readValidToken() async {
    final token = await _storage.read(key: _tokenKey);
    final rawExpiry = await _storage.read(key: _expiryKey);
    final expiry = rawExpiry == null ? null : DateTime.tryParse(rawExpiry);
    if (token == null || expiry == null || !expiry.isAfter(DateTime.now())) {
      await clear();
      return null;
    }
    return token;
  }

  @override
  Future<void> save(String token, DateTime expiresAt) async {
    await _storage.write(key: _tokenKey, value: token);
    await _storage.write(
      key: _expiryKey,
      value: expiresAt.toUtc().toIso8601String(),
    );
  }

  @override
  Future<void> clear() async {
    await _storage.delete(key: _tokenKey);
    await _storage.delete(key: _expiryKey);
  }
}

abstract interface class CapabilityStore {
  Future<Capability?> read();
  Future<void> save(Capability capability);
  Future<void> clear();
}

class PreferencesCapabilityStore implements CapabilityStore {
  static const _key = 'current_capability';

  @override
  Future<Capability?> read() async {
    final value = (await SharedPreferences.getInstance()).getString(_key);
    for (final capability in Capability.values) {
      if (capability.name == value) return capability;
    }
    return null;
  }

  @override
  Future<void> save(Capability capability) async =>
      (await SharedPreferences.getInstance()).setString(_key, capability.name);

  @override
  Future<void> clear() async =>
      (await SharedPreferences.getInstance()).remove(_key);
}
