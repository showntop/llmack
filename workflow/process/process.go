package process

import (
	"github.com/showntop/llmack/workflow/dag"
)

type processable interface {
	Start() error
}

// Process is a workflow process
type Process struct {
	G *dag.Graph // process define

	WorkflowID string // process instance id equal workflow id when root
	RunID      string // workflow id
	Runner     string
}
