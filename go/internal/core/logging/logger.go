package logging

import (
	"log/slog"
	"os"
)

type Logger struct {
	inner *slog.Logger
}

func New(component string) Logger {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	return Logger{inner: slog.New(handler).With("component", component)}
}

func (l Logger) With(args ...any) Logger {
	return Logger{inner: l.inner.With(args...)}
}

func (l Logger) Info(msg string, args ...any) {
	l.inner.Info(msg, args...)
}

func (l Logger) Error(msg string, args ...any) {
	l.inner.Error(msg, args...)
}

func (l Logger) Debug(msg string, args ...any) {
	l.inner.Debug(msg, args...)
}
