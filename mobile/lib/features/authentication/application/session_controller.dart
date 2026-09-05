import 'package:flutter/foundation.dart';

import '../../../core/models/account.dart';
import '../../../core/session/session_store.dart';
import '../data/auth_repository.dart';

enum SessionStatus { bootstrapping, signedOut, signedIn }

@immutable
class SessionState {
  const SessionState({
    required this.status,
    this.account,
    this.capability = Capability.rider,
    this.busy = false,
    this.error,
  });
  const SessionState.bootstrapping()
    : this(status: SessionStatus.bootstrapping);
  final SessionStatus status;
  final Account? account;
  final Capability capability;
  final bool busy;
  final String? error;

  SessionState copyWith({
    SessionStatus? status,
    Account? account,
    Capability? capability,
    bool? busy,
    String? error,
    bool clearError = false,
  }) => SessionState(
    status: status ?? this.status,
    account: account ?? this.account,
    capability: capability ?? this.capability,
    busy: busy ?? this.busy,
    error: clearError ? null : error ?? this.error,
  );
}

class SessionController extends ChangeNotifier {
  SessionController(this._auth, this._capabilities) {
    restore();
  }
  final AuthRepository _auth;
  final CapabilityStore _capabilities;
  SessionState _state = const SessionState.bootstrapping();
  SessionState get state => _state;

  Future<void> restore() async {
    try {
      final account = await _auth.restore();
      if (account == null) {
        _set(const SessionState(status: SessionStatus.signedOut));
        return;
      }
      final stored = await _capabilities.read();
      final selected = stored != null && account.capabilities.contains(stored)
          ? stored
          : Capability.rider;
      _set(
        SessionState(
          status: SessionStatus.signedIn,
          account: account,
          capability: selected,
        ),
      );
    } catch (_) {
      _set(
        const SessionState(
          status: SessionStatus.signedOut,
          error: 'Unable to restore your session. Please sign in again.',
        ),
      );
    }
  }

  Future<bool> login(String identifier, String password) async {
    _set(_state.copyWith(busy: true, clearError: true));
    try {
      final account = await _auth.login(identifier, password);
      await _capabilities.save(Capability.rider);
      _set(SessionState(status: SessionStatus.signedIn, account: account));
      return true;
    } catch (error) {
      _set(SessionState(status: SessionStatus.signedOut, error: '$error'));
      return false;
    }
  }

  Future<String?> register(String identifier, String password) async {
    _set(_state.copyWith(busy: true, clearError: true));
    try {
      final id = await _auth.register(identifier, password);
      _set(const SessionState(status: SessionStatus.signedOut));
      return id;
    } catch (error) {
      _set(SessionState(status: SessionStatus.signedOut, error: '$error'));
      return null;
    }
  }

  Future<bool> verify(String verificationId, String code) async {
    _set(_state.copyWith(busy: true, clearError: true));
    try {
      await _auth.completeVerification(verificationId, code);
      _set(const SessionState(status: SessionStatus.signedOut));
      return true;
    } catch (error) {
      _set(SessionState(status: SessionStatus.signedOut, error: '$error'));
      return false;
    }
  }

  Future<void> selectCapability(Capability capability) async {
    if (!(_state.account?.capabilities.contains(capability) ?? false)) return;
    await _capabilities.save(capability);
    _set(_state.copyWith(capability: capability, clearError: true));
  }

  Future<bool> enableDriver() async {
    if (_state.busy || _state.status != SessionStatus.signedIn) return false;
    _set(_state.copyWith(busy: true, clearError: true));
    try {
      final account = await _auth.enableDriver();
      _set(_state.copyWith(account: account, busy: false));
      return account.capabilities.contains(Capability.driver);
    } catch (error) {
      _set(_state.copyWith(busy: false, error: '$error'));
      return false;
    }
  }

  Future<void> logout() async {
    _set(_state.copyWith(busy: true, clearError: true));
    try {
      await _auth.logout();
    } catch (_) {
      // Local sign-out remains authoritative when the remote session is gone.
    } finally {
      await _capabilities.clear();
      _set(const SessionState(status: SessionStatus.signedOut));
    }
  }

  void clearError() => _set(_state.copyWith(clearError: true));
  void _set(SessionState value) {
    _state = value;
    notifyListeners();
  }
}
