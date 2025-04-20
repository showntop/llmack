package node

import (
	"context"
	"strings"
	"sync"

	"github.com/showntop/flatmap"

	"github.com/showntop/llmack/llm"
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
	// log.InfoContextf(ctx, "end node %+v inputs: %+v", nd.Node.Metadata, r.Inputs)
	//	extract args
	mp, err := flatmap.Flatten(r.Inputs, flatmap.DefaultTokenizer)
	if err != nil {
		return result, err
	}
	// log.InfoContextf(ctx, "end node %+v inputs: %+v", nd.Node.Metadata, mp)

	wg := sync.WaitGroup{}
	wg.Add(len(nd.Node.Outputs))
	for name, p := range nd.Node.Outputs {
		pointer := strings.TrimPrefix(p.Value, "{{")
		pointer = strings.TrimSuffix(pointer, "}}")
		value := mp.Get(pointer)
		if response, ok := value.(*llm.Response); ok {
			go func() {
				stream := response.Stream()
				for chunk := stream.Take(); chunk != nil; chunk = stream.Take() {
					r.Events <- &workflow.Event{
						Name: name,
						Data: chunk,
						Type: "end",
					}
				}
				wg.Done()
			}()
		} else {
			r.Events <- &workflow.Event{
				Name: name,
				Data: value,
				Type: "end",
			}
			wg.Done()
		}
		result[name] = value
	}
	r.Events <- &workflow.Event{
		Name: "end",
		Data: result,
	}
	wg.Wait()
	close(r.Events)
	return result, nil
}
