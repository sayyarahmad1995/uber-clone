import 'dart:async';

import 'package:fake_async/fake_async.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:uber_clone/features/ride_flow/application/polling_loop.dart';

void main() {
  test(
    'schedules the next refresh only after the current refresh completes',
    () {
      fakeAsync((async) {
        final firstRefresh = Completer<void>();
        var refreshes = 0;
        final loop = PollingLoop(interval: const Duration(milliseconds: 5));

        loop.runRefresh(() async {
          refreshes++;
          await firstRefresh.future;
        });

        async.elapse(const Duration(milliseconds: 20));
        expect(refreshes, 1);

        firstRefresh.complete();
        async.flushMicrotasks();
        expect(refreshes, 1);

        async.elapse(const Duration(milliseconds: 4));
        expect(refreshes, 1);
        async.elapse(const Duration(milliseconds: 1));
        async.flushMicrotasks();
        expect(refreshes, 2);

        loop.dispose();
      });
    },
  );

  test('background suspension cancels a pending timer', () {
    fakeAsync((async) {
      var refreshes = 0;
      final loop = PollingLoop(interval: const Duration(milliseconds: 5));

      loop.runRefresh(() async => refreshes++);
      async.flushMicrotasks();
      loop.setForeground(false, () async => refreshes++);

      async.elapse(const Duration(milliseconds: 20));

      expect(refreshes, 1);
      loop.dispose();
    });
  });

  test('resume triggers one immediate refresh', () {
    fakeAsync((async) {
      var refreshes = 0;
      final loop = PollingLoop(interval: const Duration(milliseconds: 5));
      Future<void> refresh() async => refreshes++;

      loop.runRefresh(refresh);
      async.flushMicrotasks();
      loop.setForeground(false, refresh);
      loop.setForeground(true, refresh);
      async.flushMicrotasks();

      expect(refreshes, 2);
      loop.dispose();
    });
  });

  test('ignores a concurrent passive refresh', () {
    fakeAsync((async) {
      final firstRefresh = Completer<void>();
      var refreshes = 0;
      final loop = PollingLoop(interval: const Duration(milliseconds: 5));

      loop.runRefresh(() async {
        refreshes++;
        await firstRefresh.future;
      });
      loop.runRefresh(() async => refreshes++);
      async.flushMicrotasks();

      expect(refreshes, 1);
      expect(loop.busy, isTrue);

      firstRefresh.complete();
      async.flushMicrotasks();
      expect(loop.busy, isFalse);
      loop.dispose();
    });
  });

  test('command waits for an in-flight refresh before executing', () {
    fakeAsync((async) {
      final firstRefresh = Completer<void>();
      var commands = 0;
      final loop = PollingLoop(interval: const Duration(milliseconds: 5));

      loop.runRefresh(() => firstRefresh.future);
      loop.runCommand(() async {
        commands++;
      }, reload: () async {});
      async.flushMicrotasks();

      expect(commands, 0);

      firstRefresh.complete();
      async.flushMicrotasks();

      expect(commands, 1);
      loop.dispose();
    });
  });

  test('reloads authoritatively before surfacing a failed command', () {
    fakeAsync((async) {
      final events = <String>[];
      Object? surfacedError;
      var reloadSawSurfacedError = false;
      final loop = PollingLoop(interval: const Duration(milliseconds: 5));

      loop
          .runCommand(
            () async {
              events.add('command');
              throw StateError('response lost');
            },
            reload: () async {
              events.add('reload');
              reloadSawSurfacedError = surfacedError != null;
            },
          )
          .then<void>(
            (_) {},
            onError: (Object error, StackTrace stackTrace) {
              surfacedError = error;
            },
          );
      async.flushMicrotasks();

      expect(events, ['command', 'reload']);
      expect(reloadSawSurfacedError, isFalse);
      expect(surfacedError, isA<StateError>());
      loop.dispose();
    });
  });

  test(
    'dispose prevents future scheduling and completes an in-flight refresh',
    () {
      fakeAsync((async) {
        final refresh = Completer<void>();
        var refreshes = 0;
        final loop = PollingLoop(interval: const Duration(milliseconds: 5));

        loop.runRefresh(() async {
          refreshes++;
          await refresh.future;
        });
        loop.dispose();
        refresh.complete();
        async.flushMicrotasks();
        async.elapse(const Duration(milliseconds: 20));

        expect(refreshes, 1);
        expect(loop.busy, isFalse);
      });
    },
  );
}
