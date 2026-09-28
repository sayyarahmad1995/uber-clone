import 'dart:async';

import 'package:flutter_test/flutter_test.dart';
import 'package:uber_clone/core/models/account.dart';
import 'package:uber_clone/core/network/api_exception.dart';
import 'package:uber_clone/features/authentication/application/session_controller.dart';
import 'package:uber_clone/features/ride_flow/domain/ride_execution.dart';
import 'package:uber_clone/features/ride_flow/domain/trip.dart';

import 'test_doubles.dart';

void main() {
  test('restores a valid saved capability', () async {
    final capabilities = MemoryCapabilityStore(Capability.driver);
    final controller = SessionController(
      FakeAuthRepository(account: bothCapabilities),
      capabilities,
      FakeDriverPresenceService(),
      FakeDriverRepository(),
      FakeRideFlowRepository(),
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
      FakeDriverRepository(),
      FakeRideFlowRepository(),
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
      FakeDriverRepository(),
      FakeRideFlowRepository(),
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
      FakeDriverRepository(),
      FakeRideFlowRepository(),
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
      FakeDriverRepository(),
      FakeRideFlowRepository(),
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
      FakeDriverRepository(),
      FakeRideFlowRepository(),
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
      FakeDriverRepository(),
      FakeRideFlowRepository(),
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
      FakeDriverRepository(),
      FakeRideFlowRepository(),
    );
    await controller.restore();
    expect(presence.runningFor, isNull);
    expect(controller.state.status, SessionStatus.signedOut);
  });

  test('online Driver cannot switch to Rider or stop presence', () async {
    final capabilities = MemoryCapabilityStore(Capability.driver);
    final presence = FakeDriverPresenceService()..runningFor = 'user-1';
    final drivers = FakeDriverRepository(
      profile: driverProfile.copyWith(isOnline: true),
    );
    final controller = SessionController(
      FakeAuthRepository(account: bothCapabilities),
      capabilities,
      presence,
      drivers,
      FakeRideFlowRepository(),
    );
    await controller.restore();

    final result = await controller.selectCapability(Capability.rider);

    expect(result.outcome, CapabilitySelectionOutcome.driverOnline);
    expect(result.message, 'Go offline before switching to Rider mode.');
    expect(controller.state.capability, Capability.driver);
    expect(capabilities.value, Capability.driver);
    expect(presence.runningFor, 'user-1');
    expect(presence.events, isEmpty);
    expect(drivers.calls, isEmpty);
  });

  test('offline Driver without an active trip switches to Rider', () async {
    final capabilities = MemoryCapabilityStore(Capability.driver);
    final presence = FakeDriverPresenceService()..runningFor = 'user-1';
    final controller = SessionController(
      FakeAuthRepository(account: bothCapabilities),
      capabilities,
      presence,
      FakeDriverRepository(profile: driverProfile),
      FakeRideFlowRepository(),
    );
    await controller.restore();

    final result = await controller.selectCapability(Capability.rider);

    expect(result.outcome, CapabilitySelectionOutcome.selected);
    expect(controller.state.capability, Capability.rider);
    expect(capabilities.value, Capability.rider);
    expect(presence.runningFor, isNull);
    expect(presence.events, ['service.stop']);
  });

  for (final status in ['assigned', 'in_progress']) {
    test('$status trip blocks switching to Rider while offline', () async {
      final capabilities = MemoryCapabilityStore(Capability.driver);
      final presence = FakeDriverPresenceService()..runningFor = 'user-1';
      final controller = SessionController(
        FakeAuthRepository(account: bothCapabilities),
        capabilities,
        presence,
        FakeDriverRepository(profile: driverProfile),
        FakeRideFlowRepository(currentTrip: activeTrip(status)),
      );
      await controller.restore();

      final result = await controller.selectCapability(Capability.rider);

      expect(result.outcome, CapabilitySelectionOutcome.activeTrip);
      expect(
        result.message,
        'Finish or cancel your active trip before switching to Rider mode.',
      );
      expect(controller.state.capability, Capability.driver);
      expect(capabilities.value, Capability.driver);
      expect(presence.runningFor, 'user-1');
      expect(presence.events, isEmpty);
    });
  }

  test('active-trip restriction takes precedence over online status', () async {
    final drivers = FakeDriverRepository(
      profile: driverProfile.copyWith(isOnline: true),
    );
    final controller = SessionController(
      FakeAuthRepository(account: bothCapabilities),
      MemoryCapabilityStore(Capability.driver),
      FakeDriverPresenceService()..runningFor = 'user-1',
      drivers,
      FakeRideFlowRepository(currentTrip: activeTrip('assigned')),
    );
    await controller.restore();

    final result = await controller.selectCapability(Capability.rider);

    expect(result.outcome, CapabilitySelectionOutcome.activeTrip);
    expect(drivers.getCalls, 0);
  });

  test('trip lookup failure blocks switching with actionable error', () async {
    final capabilities = MemoryCapabilityStore(Capability.driver);
    final drivers = FakeDriverRepository(profile: driverProfile);
    final controller = SessionController(
      FakeAuthRepository(account: bothCapabilities),
      capabilities,
      FakeDriverPresenceService(),
      drivers,
      FakeRideFlowRepository(currentTripFailure: Exception('network down')),
    );
    await controller.restore();

    final result = await controller.selectCapability(Capability.rider);

    expect(result.outcome, CapabilitySelectionOutcome.verificationFailed);
    expect(result.message, contains('Unable to verify Driver status'));
    expect(controller.state.capability, Capability.driver);
    expect(capabilities.value, Capability.driver);
    expect(drivers.getCalls, 0);
  });

  test('availability lookup failure blocks switching', () async {
    final capabilities = MemoryCapabilityStore(Capability.driver);
    final drivers = FakeDriverRepository(profile: driverProfile)
      ..getFailure = Exception('network down');
    final controller = SessionController(
      FakeAuthRepository(account: bothCapabilities),
      capabilities,
      FakeDriverPresenceService(),
      drivers,
      FakeRideFlowRepository(),
    );
    await controller.restore();

    final result = await controller.selectCapability(Capability.rider);

    expect(result.outcome, CapabilitySelectionOutcome.verificationFailed);
    expect(result.message, contains('Unable to verify Driver availability'));
    expect(controller.state.capability, Capability.driver);
    expect(capabilities.value, Capability.driver);
  });

  test('stale presence cleanup failure retains Driver mode', () async {
    final capabilities = MemoryCapabilityStore(Capability.driver);
    final presence = FakeDriverPresenceService()
      ..runningFor = 'user-1'
      ..failStop = true;
    final controller = SessionController(
      FakeAuthRepository(account: bothCapabilities),
      capabilities,
      presence,
      FakeDriverRepository(profile: driverProfile),
      FakeRideFlowRepository(),
    );
    await controller.restore();

    final result = await controller.selectCapability(Capability.rider);

    expect(result.outcome, CapabilitySelectionOutcome.cleanupFailed);
    expect(result.message, contains('Unable to stop Driver presence'));
    expect(controller.state.capability, Capability.driver);
    expect(capabilities.value, Capability.driver);
  });

  test('repeated Rider switches do not overlap server verification', () async {
    final capabilities = MemoryCapabilityStore(Capability.driver);
    final trips = FakeRideFlowRepository(
      currentTripCompleter: Completer<TripSnapshot?>(),
    );
    final controller = SessionController(
      FakeAuthRepository(account: bothCapabilities),
      capabilities,
      FakeDriverPresenceService(),
      FakeDriverRepository(profile: driverProfile),
      trips,
    );
    await controller.restore();

    final first = controller.selectCapability(Capability.rider);
    final second = await controller.selectCapability(Capability.rider);

    expect(second.outcome, CapabilitySelectionOutcome.busy);
    expect(trips.currentTripCalls, 1);
    trips.currentTripCompleter!.complete(null);
    expect((await first).outcome, CapabilitySelectionOutcome.selected);
    expect(controller.state.capability, Capability.rider);
  });
}

TripSnapshot activeTrip(String status) => TripSnapshot(
  rideRequestId: 'ride-1',
  status: status,
  assignedAt: DateTime.utc(2026, 9, 28),
  settlement: const SettlementSnapshot(status: 'unsettled'),
);
