/// One online Driver's renewable presence publisher, independent of UI lifetime.
abstract interface class DriverPresenceService {
  Future<bool> isRunningFor(String driverUserId);
  Future<void> start(String driverUserId);
  Future<void> stop();
}
