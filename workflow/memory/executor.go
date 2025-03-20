package memory

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/expr-lang/expr"
	"github.com/showntop/flatmap"
	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/log"
	"github.com/showntop/llmack/workflow"
	nodePkg "github.com/showntop/llmack/workflow/node"
	"golang.org/x/sync/errgroup"
)

// Executor 工作流执行器
type Executor struct {
	workflow *workflow.Workflow
	ctx      context.Context
	scope    map[string]any       // 用于存储节点间共享的数据
	outputs  map[string]any       // 用于最终的结果数据
	events   chan *workflow.Event // 用于输出中间结果
	graph    *Graph
}

// NewExecutor 创建工作流执行器
func NewExecutor(wfdefine *workflow.Workflow) *Executor {
	return &Executor{
		workflow: wfdefine,
		scope:    make(map[string]any),
		events:   make(chan *workflow.Event, 30),
	}
}

// Events ...
func (e *Executor) Events() <-chan *workflow.Event {
	return e.events
}

// Execute 执行工作流
func (e *Executor) Execute(ctx context.Context, inputs map[string]any) (*workflow.Result, error) {
	log.InfoContextf(ctx, "workflow execute workflow %v, inputs: %+v", e.workflow.ID, inputs)

	e.scope = inputs // 初始化scope TODO 增加系统变量

	// 构建DAG图
	graph := NewGraph(e.workflow.Nodes, e.workflow.Edges)
	e.graph = graph

	// 执行单层所有节点
	levelExecute := func(currents []*workflow.Node) ([]*workflow.Node, error) {
		var nexts []*workflow.Node
		var err error
		errGroup := errgroup.Group{}
		for _, node := range currents {
			errGroup.Go(func() error {

				nexts, err = e.executeNode(ctx, node, graph.edges[node.ID]...)
				if err != nil {
					log.ErrorContextf(ctx, "workflow execute node(id: %s, kind:%s) error, %v", node.ID, node.Kind, err)
					return nil
				}
				return nil
			})
		}
		// 等待本层所有节点执行完成
		if err := errGroup.Wait(); err != nil {
			log.ErrorContextf(ctx, "workflow execution failed: %v", err)
			return nil, fmt.Errorf("workflow execution failed: %w", err)
		}
		return nexts, nil
	} ////

	// 遍历所有节点并执行
	for queue := graph.StartNodes(); len(queue) > 0; {
		currents := queue
		queue = queue[:0]
		// 层次遍历，解决并行问题。并行执行所有分支
		nexts, err := levelExecute(currents)
		if err != nil {
			return nil, fmt.Errorf("workflow execution failed: %w", err)
		}
		queue = append(queue, nexts...) // 将下一层节点加入队列
	}

	// 关闭流 @TODO fix it on 恰当位置
	// close(e.events)
	log.InfoContextf(ctx, "workflow execute workflow %v, finished with outputs: %+v", e.workflow.ID, e.outputs)
	return &workflow.Result{
		Outputs: e.outputs,
	}, nil
}

// executeNode 执行单个节点
func (e *Executor) executeNode(ctx context.Context, node *workflow.Node, outgoings ...*workflow.Edge) ([]*workflow.Node, error) {
	// 检查节点的所有前置依赖是否已完成
	// if !graph.AreAllDependenciesCompleted(node.ID) {
	// 	return nil, nil
	// }

	nodeIns, err := nodePkg.Build(node, outgoings...)
	if err != nil {
		return nil, fmt.Errorf("failed to build node %s: %w", node.ID, err)
	}
	inputs := make(map[string]any)
	if len(node.Inputs) > 0 {
		env, err := flatmap.Flatten(e.scope, flatmap.DefaultTokenizer)
		if err != nil {
			return nil, errors.Join(err)
		}

		for name, input := range node.Inputs {
			if !strings.HasPrefix(input.Value, "{{") { // expression
				program, _ := expr.Compile(input.Value)
				value, _ := expr.Run(program, nil)
				inputs[name] = value
				continue
			}
			pointer := strings.TrimPrefix(input.Value, "{{")
			pointer = strings.TrimSuffix(pointer, "}}")
			value := env.Get(pointer)
			inputs[name] = value
		}
	} else {
		inputs = e.scope
	}
	jsonInputs, _ := json.Marshal(inputs)
	jsonScope, _ := json.Marshal(e.scope)
	log.InfoContextf(ctx, "workflow execute node(id: %s, kind:%s) <====> inputs: %s, scope: %s", node.ID, node.Kind, string(jsonInputs), string(jsonScope))
	result, err := nodeIns.Execute(ctx, &nodePkg.ExecRequest{
		Inputs: inputs,
		Scope:  e.scope,
		Events: e.events,
	})
	if err != nil {
		log.ErrorContextf(ctx, "workflow execute node(id: %s, kind:%s) failed: %s", node.ID, node.Kind, err.Error())
		return nil, fmt.Errorf("failed to execute node %s: %w", node.ID, err)
	}
	jsonResult, _ := json.Marshal(result)
	log.InfoContextf(ctx, "workflow execute node(id: %s, kind:%s) <====> outputs: %T %s", node.ID, node.Kind, result, string(jsonResult))

	nexts := []*workflow.Node{}
	switch result := result.(type) {
	case map[string]any:
		e.scope[node.ID] = result // 更新节点结果
		e.outputs = result        // 重置outputs，TODO 注意并行下的bug，需要改得更优雅一些
	case *llm.Response:
		e.scope[node.ID] = map[string]any{"response": result}
		// e.outputs = result
	case *llm.Result:
		e.scope[node.ID] = map[string]any{"response": result}
		// e.outputs = result
	case workflow.NodeID:
		nexts = []*workflow.Node{e.graph.nodes[result.ID]}
	case error:
		return nil, result
	}
	e.graph.MarkNodeCompleted(node.ID) // 标记节点已完成，持久化存储TODO

	if len(nexts) <= 0 {
		nexts = e.graph.NextNodes(node.ID)
	}
	return nexts, nil
}
