package logger

import (
	"log/slog"
	"os"
	"strings"
)

func New() *slog.Logger {
	level := parseLevel(os.Getenv("KEYFORGE_LOG"))

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     level,
		AddSource: false,
		ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
			if attr.Key == slog.TimeKey {
				return slog.String(
					slog.TimeKey,
					attr.Value.Time().Format("15:04:05"),
				)
			}

			return attr
		},
	})

	return slog.New(handler)
}

func parseLevel(value string) slog.Level {
	switch strings.ToLower(value) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	case "info", "":
		return slog.LevelInfo
	default:
		return slog.LevelInfo
	}
}
