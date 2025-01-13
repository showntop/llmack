package workflow

import "fmt"

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
