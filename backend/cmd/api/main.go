package main

import (
	"log/slog"
	"os"
)

func main() {
	cfg := loadConfig()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	app, cleanup, err := newApplication(cfg)
	if err != nil {
		logger.Error("application initialization failed", "error", err)
		os.Exit(1)
	}
	defer cleanup()

	if err := runServer(cfg.Port, app.handler, logger); err != nil {
		logger.Error("API server stopped", "error", err)
		os.Exit(1)
	}
}
