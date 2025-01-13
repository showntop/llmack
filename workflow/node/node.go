package node

import (
	"context"

	"github.com/showntop/llmack/workflow"
	"github.com/showntop/llmack/workflow/dag"
)

// Identifier TODO
type Identifier struct{ id string }

// ID TODO
func (i *Identifier) ID() string { return i.id }

// SetID TODO
func (i *Identifier) SetID(id string) { i.id = id }

type (
	executeable struct{}

	// Nodes TODO
	Nodes []Node

	// Node TODO
	Node interface {
		ID() string
		SetID(string)
		Execute(context.Context, *ExecRequest) (ExecResponse, error)
	}

	execFn func(context.Context, *ExecRequest) (ExecResponse, error)
)

type ExecRequest struct {
	ProcessID  string
	WorkflowID string
	NodeID     string
	Sponsor    string
	Runner     string // 运行人
	Inputs     map[string]any
	Scope      map[string]any
	Graph      *dag.Graph
	Events     chan *workflow.Event
}

type ExecResponse any
