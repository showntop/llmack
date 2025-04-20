package hooks

// import (
// 	"context"
// 	"database/sql"
// 	"database/sql/driver"
// 	"io"

// 	"github.com/showntop/llmack/engine"
// 	"go.opentelemetry.io/otel"
// 	"go.opentelemetry.io/otel/attribute"
// 	"go.opentelemetry.io/otel/codes"
// 	"go.opentelemetry.io/otel/trace"
// )

// // EngineOtelHook ...
// type EngineOtelHook struct {
// 	provider trace.TracerProvider
// 	tracer   trace.Tracer
// }

// // Option 	...
// type Option func(*EngineOtelHook)

// // NewEngineOtelHook ...
// func NewEngineOtelHook(opts ...Option) engine.Hook {
// 	h := &EngineOtelHook{}
// 	h.provider = otel.GetTracerProvider()
// 	h.tracer = h.provider.Tracer("github.com/showntop/llmack/opentelemetry")

// 	return h
// }

// // OnStart ...
// func (h *EngineOtelHook) OnStart(ctx context.Context) context.Context {
// 	ctx, _ = h.tracer.Start(ctx, "engine/start", trace.WithSpanKind(trace.SpanKindInternal))
// 	return ctx
// }

// // OnFinish ...
// func (h *EngineOtelHook) OnFinish(ctx context.Context, err error) {
// 	span := trace.SpanFromContext(ctx)
// 	if !span.IsRecording() {
// 		return
// 	}
// 	defer span.End()

// 	attrs := make([]attribute.KeyValue, 0, 4)

// 	span.SetAttributes(attrs...)
// 	switch err {
// 	case nil,
// 		driver.ErrSkip,
// 		io.EOF, // end of rows iterator
// 		sql.ErrNoRows:
// 		// ignore
// 	default:
// 		span.RecordError(err)
// 		span.SetStatus(codes.Error, err.Error())
// 	}
// }

// func (h *EngineOtelHook) BeforeRetrieveStart(ctx context.Context) {}

// func (h *EngineOtelHook) AfterRetrieveFinish(ctx context.Context) {}

// func (h *EngineOtelHook) BeforeToolStart(ctx context.Context, toolID int64, inputs map[string]any) {
// }

// // AfterToolFinish ...
// func (h *EngineOtelHook) AfterToolFinish(ctx context.Context, toolID int64, outputs map[string]any, err error) {
// }

// func (h *EngineOtelHook) OnReasonStart(ctx context.Context) {}

// func (h *EngineOtelHook) OnReasonFinish(ctx context.Context) {}

// // BeforeLLMStart ...
// func (h *EngineOtelHook) BeforeLLMStart(ctx context.Context) context.Context {
// 	ctx, _ = h.tracer.Start(ctx, "engine/llm", trace.WithSpanKind(trace.SpanKindInternal))
// 	return ctx
// }

// // AfterLLMFinish ...
// func (h *EngineOtelHook) AfterLLMFinish(ctx context.Context, err error) {
// 	span := trace.SpanFromContext(ctx)
// 	if !span.IsRecording() {
// 		return
// 	}
// 	defer span.End()

// 	attrs := make([]attribute.KeyValue, 0, 4)

// 	span.SetAttributes(attrs...)
// 	switch err {
// 	case nil,
// 		driver.ErrSkip,
// 		io.EOF, // end of rows iterator
// 		sql.ErrNoRows:
// 		// ignore
// 	default:
// 		span.RecordError(err)
// 		span.SetStatus(codes.Error, err.Error())
// 	}

// }
