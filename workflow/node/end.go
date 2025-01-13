package node

import (
	"context"
	"strings"

	"github.com/showntop/flatmap"

	"github.com/showntop/llmack/log"
	"github.com/showntop/llmack/workflow"
	wf "github.com/showntop/llmack/workflow"
)

type endNode struct {
	Node *wf.Node

	executeable
	Identifier
}

func (nd endNode) Schema() string {
	return `{
		"kind": "end",
		"outputs": [
			{
				"name": "result",
				"type": "json"
			}
		]
	}`
}

// EndNode initializes new end node with end definition
func EndNode(node *workflow.Node) (*endNode, error) {
	return &endNode{Node: node}, nil
}

// Exec executes end node
func (nd endNode) Execute(ctx context.Context, r *ExecRequest) (ExecResponse, error) {
	result := make(map[string]any)

	// 处理输出
	log.InfoContextf(ctx, "end node %+v inputs: %+v", nd.Node.Metadata, r.Scope)
	//	extract args
	mp, err := flatmap.Flatten(r.Scope, flatmap.DefaultTokenizer)
	if err != nil {
		return result, err
	}
	log.InfoContextf(ctx, "end node %+v inputs: %+v", nd.Node.Metadata, mp)
	for _, p := range nd.Node.Outputs {
		pointer := strings.TrimPrefix(p.Value, "{{")
		pointer = strings.TrimSuffix(pointer, "}}")
		value := mp.Get(pointer)
		values, ok := value.(<-chan interface{})
		if !ok {
			values, ok = value.(chan interface{})
		}
		if ok { // TODO 并行
			for v := range values {
				r.Events <- &workflow.Event{
					Type: "TODO",
					Data: v,
				}
			}
		}
		result[p.Name] = value
	}

	return result, nil
}
