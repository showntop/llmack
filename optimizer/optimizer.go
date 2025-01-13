package optimizer

import (
	"context"

	"github.com/showntop/llmack/program"
)

// OptimizeResult contains the results of optimization
type OptimizeResult struct {
	BestProgram    program.Program
	BestScore      float64
	History        []*OptimizeStep
	BestParameters map[string]interface{}
}

// OptimizeStep represents a single optimization step
type OptimizeStep struct {
	Program    program.Program
	Score      float64
	Parameters map[string]interface{}
	Timestamp  int64
}

// Optimizer defines the interface for optimization
type Optimizer interface {
	// Optimize performs optimization and returns the best result
	Optimize(context.Context, any, ...OptimizeOption) (*OptimizeResult, error)
}
