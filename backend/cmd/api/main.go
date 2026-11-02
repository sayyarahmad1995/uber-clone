package main

import (
 "log/slog"
 "os"
)

func main() {
 os.Exit(run())
}

// run ensures deferred cleanup executes even when startup or serving fails.
func run() int {
 cfg := loadConfig()
 logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
 app, cleanup, err := newApplication(cfg)
 if err != nil {
  logger.Error("application initialization failed", "error", err)
  return 1
 }
 defer cleanup()
 if err := runServer(cfg.Port, app.handler, logger); err != nil {
  logger.Error("API server stopped", "error", err)
  return 1
 }
 return 0
}
