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
      FakeDriverPresenceService(),
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
      FakeDriverPresenceService(),
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
      FakeDriverPresenceService(),
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
      FakeDriverPresenceService(),
    );
    await controller.login('rider@example.com', 'password');
    await controller.selectCapability(Capability.driver);
    expect(controller.state.capability, Capability.rider);
  });

  test('starts verification for an existing signed-out account', () async {
    final controller = SessionController(
      FakeAuthRepository(),
      MemoryCapabilityStore(),
      FakeDriverPresenceService(),
    );
    final challenge = await controller.startVerification('rider@example.com');
    expect(challenge, 'challenge');
    expect(controller.state.status, SessionStatus.signedOut);
    expect(controller.state.busy, isFalse);
  });

  test('logout stops Driver presence even when remote logout fails', () async {
    final presence = FakeDriverPresenceService()..runningFor = 'user-1';
    final controller = SessionController(
      FakeAuthRepository(
        account: bothCapabilities,
        logoutError: Exception('offline'),
      ),
      MemoryCapabilityStore(),
      presence,
    );
    await controller.restore();
    await controller.logout();
    expect(presence.runningFor, isNull);
    expect(controller.state.status, SessionStatus.signedOut);
  });

  test('new account login stops previous session publisher first', () async {
    final presence = FakeDriverPresenceService()..runningFor = 'user-1';
    final controller = SessionController(
      FakeAuthRepository(
        account: const Account(id: 'user-2', capabilities: [Capability.rider]),
      ),
      MemoryCapabilityStore(),
      presence,
    );
    await controller.login('other@example.com', 'password');
    expect(presence.runningFor, isNull);
    expect(controller.state.account?.id, 'user-2');
  });

  test('missing restored session stops a stale Driver service', () async {
    final presence = FakeDriverPresenceService()..runningFor = 'user-1';
    final controller = SessionController(
      FakeAuthRepository(),
      MemoryCapabilityStore(),
      presence,
    );
    await controller.restore();
    expect(presence.runningFor, isNull);
    expect(controller.state.status, SessionStatus.signedOut);
  });
}
