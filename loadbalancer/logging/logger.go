package logging

import (
	"log/slog"
	"os"
)

func setupLogger() *slog.Logger {
	var log *slog.Logger

	log = slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
	)
	return log
}

var logger = setupLogger()

func GetLogger() *slog.Logger {
	return logger
}
