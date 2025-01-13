package engine

import "context"

// Hook ...
type Hook interface {
	OnStart(context.Context) context.Context
	OnFinish(context.Context, error)

	BeforeRetrieveStart(context.Context)
	AfterRetrieveFinish(context.Context)

	BeforeToolStart(context.Context, int64, map[string]any)
	AfterToolFinish(context.Context, int64, map[string]any, error)

	OnReasonStart(context.Context)
	OnReasonFinish(context.Context)

	BeforeLLMStart(context.Context) context.Context
	AfterLLMFinish(context.Context, error)
}
