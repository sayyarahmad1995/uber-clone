import 'package:flutter_test/flutter_test.dart';
import 'package:uber_clone/core/models/account.dart';
import 'package:uber_clone/core/network/api_exception.dart';
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

  test('unverified login exposes verification recovery', () async {
    final controller = SessionController(
      FakeAuthRepository(
        loginError: const ApiException(
          'verification_required',
          'Account verification is required.',
          statusCode: 403,
        ),
      ),
      MemoryCapabilityStore(),
    );
    await controller.login('rider@example.com', 'password');
    expect(controller.state.status, SessionStatus.signedOut);
    expect(controller.state.verificationRequired, isTrue);

    controller.clearError();
    expect(controller.state.verificationRequired, isFalse);
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

  test('starts verification for an existing signed-out account', () async {
    final controller = SessionController(
      FakeAuthRepository(),
      MemoryCapabilityStore(),
    );
    final challenge = await controller.startVerification('rider@example.com');
    expect(challenge, 'challenge');
    expect(controller.state.status, SessionStatus.signedOut);
    expect(controller.state.busy, isFalse);
  });
}
