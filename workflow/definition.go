package workflow

import (
	"fmt"
)

// Link 链接节点
func (w *Workflow) LinkWithCondition(node1, expr string, node2 any, nodes ...Node) *Workflow {
	var targetID string
	if node, ok := node2.(Node); ok {
		targetID = node.ID
		w.AddNode(node)
	} else {
		targetID = node2.(string)
	}
	edges := make([]Edge, 0)
	edges = append(edges, Edge{
		ID:      fmt.Sprintf("%s-%s", node1, targetID),
		Source:  node1,
		Target:  targetID,
		Express: expr,
	})
	w.Edges = append(w.Edges, edges...)

	if node, ok := node2.(Node); ok && len(nodes) > 0 {
		nodes = append([]Node{node}, nodes...)
		w.Link(nodes...)
	}
	return w
}

// Link 链接节点
func (w *Workflow) Link(nodes ...Node) *Workflow {
	w.AddNode(nodes...)

	edges := make([]Edge, 0)
	for i := 0; i < len(nodes)-1; i++ {
		edges = append(edges, Edge{
			ID:     fmt.Sprintf("%s-%s", nodes[i].ID, nodes[i+1].ID),
			Source: nodes[i].ID,
			Target: nodes[i+1].ID,
		})
	}
	w.Edges = append(w.Edges, edges...)
	return w
}

// AddNode 添加节点
func (w *Workflow) AddNode(nodes ...Node) *Workflow {
	w.Nodes = append(w.Nodes, nodes...)
	return w
}

// NewWorkflow 初始化workflow
func NewWorkflow(id int64, name string) *Workflow {
	return &Workflow{
		ID:   id,
		Name: name,
	}
}
