package llm

import "context"

// NoneLogger ...
type NoneLogger struct {
}

// ErrorContextf ...
func (w *NoneLogger) ErrorContextf(ctx context.Context, format string, args ...interface{}) {
	// 在这里实现你的逻辑
}

// InfoContextf ...
func (w *NoneLogger) InfoContextf(ctx context.Context, format string, args ...interface{}) {
	// 在这里实现你的逻辑
}

// WarnContextf ...
func (w *NoneLogger) WarnContextf(ctx context.Context, format string, args ...interface{}) {
	// 在这里实现你的逻辑
}

// DebugContextf ...
func (w *NoneLogger) DebugContextf(ctx context.Context, format string, args ...interface{}) {
	// 在这里实现你的逻辑
}
