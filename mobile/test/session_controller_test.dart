import 'package:flutter_test/flutter_test.dart';
import 'package:uber_clone/core/models/account.dart';
import 'package:uber_clone/features/authentication/application/session_controller.dart';

import 'test_doubles.dart';

void main() {
  test('restores a valid saved capability', () async {
    final capabilities = MemoryCapabilityStore(Capability.driver);
    final controller = SessionController(
      FakeAuthRepository(account: bothCapabilities),
      capabilities,
    );
    await controller.restore();
    expect(controller.state.status, SessionStatus.signedIn);
    expect(controller.state.capability, Capability.driver);
  });

  test('login enters Rider even when Driver is available', () async {
    final capabilities = MemoryCapabilityStore(Capability.driver);
    final controller = SessionController(
      FakeAuthRepository(account: bothCapabilities),
      capabilities,
    );
    await controller.login('rider@example.com', 'password');
    expect(controller.state.status, SessionStatus.signedIn);
    expect(controller.state.capability, Capability.rider);
    expect(capabilities.value, Capability.rider);
  });

  test('cannot select a capability the account does not own', () async {
    final controller = SessionController(
      FakeAuthRepository(account: riderAccount),
      MemoryCapabilityStore(),
    );
    await controller.login('rider@example.com', 'password');
    await controller.selectCapability(Capability.driver);
    expect(controller.state.capability, Capability.rider);
  });
}
