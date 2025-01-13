package node

import (
	"context"
	"errors"
	"strings"

	"github.com/expr-lang/expr"
	"github.com/showntop/flatmap"

	"github.com/showntop/llmack/workflow"
	wf "github.com/showntop/llmack/workflow"
)

// exprNode TODO
type exprNode struct {
	executeable
	Identifier

	Node *wf.Node
}

// ExprNode 创建expr node
func ExprNode(n *workflow.Node) *exprNode {
	return &exprNode{
		Node: n,
	}
}

// Execute 执行JSON节点，单次执行
func (n *exprNode) Execute(ctx context.Context, r *ExecRequest) (ExecResponse, error) {
	env, err := flatmap.Flatten(r.Scope, flatmap.DefaultTokenizer)
	if err != nil {
		return nil, errors.Join(err)
	}

	inputs := make(map[string]any)
	for _, input := range n.Node.Inputs {
		pointer := strings.TrimPrefix(input.Value, "{{")
		pointer = strings.TrimSuffix(pointer, "}}")
		value := env.Get(pointer)
		inputs[input.Name] = value
	}
	// inputs["input"] = `{"a": 1}`

	expression, _ := n.Node.Metadata["expr"].(string)
	if expression == "" {
		return nil, nil
	}

	program, err := expr.Compile(expression, expr.Env(inputs))
	if err != nil {
		return nil, errors.Join(err)
	}
	result, err := expr.Run(program, inputs)
	if err != nil {
		return nil, errors.Join(err)
	}

	// outputs := make(map[string]any)
	// for _, output := range n.Node.Outputs {
	// 	pointer := strings.TrimPrefix(output.Value, "{{")
	// 	pointer = strings.TrimSuffix(pointer, "}}")
	// 	value := env.Get(pointer)
	// 	outputs[output.Name] = value
	// }

	return map[string]any{"result": result}, nil
}
