import 'dart:async';

class PollingLoop {
  PollingLoop({required this.interval});

  final Duration interval;

  Timer? _timer;
  Future<void> Function()? _refresh;
  Completer<void>? _idle;
  bool _busy = false;
  bool _disposed = false;
  bool _foreground = true;

  bool get busy => _busy;

  Future<void> runRefresh(Future<void> Function() task) async {
    await _run(task, refreshTask: task);
  }

  Future<void> runCommand(
    Future<void> Function() task, {
    required Future<void> Function() reload,
  }) async {
    await _run(() async {
      try {
        await task();
      } catch (_) {
        try {
          await reload();
        } catch (_) {}
        rethrow;
      }
      await reload();
    }, waitForBusy: true);
  }

  void setForeground(bool value, Future<void> Function() refresh) {
    _foreground = value;
    _timer?.cancel();
    _timer = null;
    if (value) unawaited(runRefresh(refresh));
  }

  void dispose() {
    _disposed = true;
    _timer?.cancel();
    _timer = null;
  }

  Future<void> _run(
    Future<void> Function() task, {
    bool waitForBusy = false,
    Future<void> Function()? refreshTask,
  }) async {
    if (_disposed || !_foreground) return;

    if (waitForBusy) {
      while (_busy && !_disposed && _foreground) {
        final idle = _idle;
        if (idle == null) break;
        await idle.future;
      }
    } else if (_busy) {
      return;
    }

    if (_disposed || !_foreground) return;

    if (refreshTask != null) _refresh = refreshTask;
    _timer?.cancel();
    _timer = null;
    final idle = Completer<void>();
    _idle = idle;
    _busy = true;

    try {
      await task();
    } finally {
      _busy = false;
      if (!idle.isCompleted) idle.complete();
      if (identical(_idle, idle)) _idle = null;
      _schedule();
    }
  }

  void _schedule() {
    if (!_disposed && _foreground && _refresh != null) {
      _timer = Timer(interval, () {
        final refresh = _refresh;
        if (refresh != null) unawaited(runRefresh(refresh));
      });
    }
  }
}
