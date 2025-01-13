package log

import (
	"context"

	"log"
)

// Logger ...
type Logger interface {
	ErrorContextf(context.Context, string, ...interface{})
	InfoContextf(context.Context, string, ...interface{})
	WarnContextf(context.Context, string, ...interface{})
	DebugContextf(context.Context, string, ...interface{})
}

// NoneLogger ...
type NoneLogger struct {
}

// ErrorContextf ...
func (w *NoneLogger) ErrorContextf(ctx context.Context, format string, args ...interface{}) {
}

// InfoContextf ...
func (w *NoneLogger) InfoContextf(ctx context.Context, format string, args ...interface{}) {
}

// WarnContextf ...
func (w *NoneLogger) WarnContextf(ctx context.Context, format string, args ...interface{}) {
}

// DebugContextf ...
func (w *NoneLogger) DebugContextf(ctx context.Context, format string, args ...interface{}) {
}

// WrapLogger ...
type WrapLogger struct {
}

// ErrorContextf ...
func (w *WrapLogger) ErrorContextf(ctx context.Context, format string, args ...interface{}) {
	log.Printf(format, args...)
}

// InfoContextf ...
func (w *WrapLogger) InfoContextf(ctx context.Context, format string, args ...interface{}) {
	log.Printf(format, args...)
}

// WarnContextf ...
func (w *WrapLogger) WarnContextf(ctx context.Context, format string, args ...interface{}) {
	log.Printf(format, args...)
}

// DebugContextf ...
func (w *WrapLogger) DebugContextf(ctx context.Context, format string, args ...interface{}) {
	log.Printf(format, args...)
}

var defaultLogger Logger = new(NoneLogger)

// DefaultLogger ...
func DefaultLogger() Logger {
	return defaultLogger
}

// SetLogger ...
func SetLogger(logger Logger) {
	defaultLogger = logger
}

// ErrorContextf ...
func ErrorContextf(ctx context.Context, format string, args ...interface{}) {
	defaultLogger.ErrorContextf(ctx, format, args...)
}

// InfoContextf ...
func InfoContextf(ctx context.Context, format string, args ...interface{}) {
	defaultLogger.InfoContextf(ctx, format, args...)
}

// WarnContextf ...
func WarnContextf(ctx context.Context, format string, args ...interface{}) {
	defaultLogger.WarnContextf(ctx, format, args...)
}

// DebugContextf ...
func DebugContextf(ctx context.Context, format string, args ...interface{}) {
	defaultLogger.DebugContextf(ctx, format, args...)
}

// Info ...
func Info(format string, args ...interface{}) {
	defaultLogger.DebugContextf(context.TODO(), format, args...)
}
