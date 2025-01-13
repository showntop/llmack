package llm

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"io"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Hook ...
type Hook interface {
	OnBeforeInvoke(context.Context) context.Context
	OnAfterInvoke(ctx context.Context, err error)
	OnFirstChunk(context.Context, error) context.Context
}

// OtelHook ...
type OtelHook struct {
	provider trace.TracerProvider
	tracer   trace.Tracer
}

// HookOption 	...
type HookOption func(*OtelHook)

// NewOtelHook ...
func NewOtelHook(opts ...HookOption) Hook {
	h := &OtelHook{}
	h.provider = otel.GetTracerProvider()
	h.tracer = h.provider.Tracer("github.com/showntop/llmack/opentelemetry")

	return h
}

// OnFirstChunk ...
func (h *OtelHook) OnFirstChunk(ctx context.Context, _ error) context.Context {
	return ctx
}

// OnBeforeInvoke ...
func (h *OtelHook) OnBeforeInvoke(ctx context.Context) context.Context {
	ctx, _ = h.tracer.Start(ctx, "llm/invoke", trace.WithSpanKind(trace.SpanKindInternal))
	return ctx
}

// OnAfterInvoke ...
func (h *OtelHook) OnAfterInvoke(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}
	defer span.End()

	attrs := make([]attribute.KeyValue, 0, 4)

	span.SetAttributes(attrs...)
	switch err {
	case nil,
		driver.ErrSkip,
		io.EOF, // end of rows iterator
		sql.ErrNoRows:
		// ignore
	default:
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}
