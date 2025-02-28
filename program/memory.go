package program

import (
	"context"

	"github.com/showntop/llmack/llm"
)

// Extra ...
type Extra struct {
	Responder   string
	ResponderID string
}

// Memory ...
type Memory interface {
	FetchMemories(context.Context, int64) ([]llm.Message, error)
	SaveMemory(context.Context, int64, string, *Extra) error
}
