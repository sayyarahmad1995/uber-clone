import 'package:uber_clone/core/models/account.dart';
import 'package:uber_clone/core/session/session_store.dart';
import 'package:uber_clone/features/authentication/data/auth_repository.dart';

const riderAccount = Account(id: 'user-1', capabilities: [Capability.rider]);
const bothCapabilities = Account(
  id: 'user-1',
  capabilities: [Capability.rider, Capability.driver],
);

class FakeAuthRepository implements AuthRepository {
  FakeAuthRepository({this.account});
  final Account? account;
  @override
  Future<Account?> restore() async => account;
  @override
  Future<Account> login(String identifier, String password) async => account!;
  @override
  Future<String> register(String identifier, String password) async =>
      'challenge';
  @override
  Future<void> completeVerification(String verificationId, String code) async {}
  @override
  Future<void> logout() async {}
}

class MemoryCapabilityStore implements CapabilityStore {
  MemoryCapabilityStore([this.value]);
  Capability? value;
  @override
  Future<Capability?> read() async => value;
  @override
  Future<void> save(Capability capability) async => value = capability;
  @override
  Future<void> clear() async => value = null;
}
