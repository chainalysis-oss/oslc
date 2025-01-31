package main

import (
	"io"
	"log/slog"
	"strings"
)

// logLevelFromStr returns the log level from a string.
// If the string is not a valid log level, the level is set to info.
func logLevelFromStr(level string) slog.Level {
	switch level {
	case strings.ToLower("debug"):
		return slog.LevelDebug
	case strings.ToLower("info"):
		return slog.LevelInfo
	case strings.ToLower("warn"):
		return slog.LevelWarn
	case strings.ToLower("error"):
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// getLogger returns a logger based on the provided level and kind.
// If the kind is not a valid kind, the logger is set to nil.
// If the level is not a valid level, the level is set to info.
func getLogger(level, kind string, writer io.Writer) *slog.Logger {
	ho := &slog.HandlerOptions{
		Level: logLevelFromStr(level),
	}
	switch kind {
	case strings.ToLower("text"):
		return slog.New(slog.NewTextHandler(writer, ho))
	case strings.ToLower("json"):
		return slog.New(slog.NewJSONHandler(writer, ho))
	case strings.ToLower("discard"):
		return slog.New(slog.NewTextHandler(io.Discard, ho))
	default:
		return nil
	}
}
