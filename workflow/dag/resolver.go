package dag

import (
	"fmt"

	wf "github.com/showntop/llmack/workflow"
)

// Resolve TODO
func Resolve(flow *wf.Workflow) (*Graph, error) {
	gf := NewGraph()
	gf.ID = fmt.Sprint(flow.ID)
	// gf.Metadata = flow.Metadata
	nodes := flow.Nodes
	edges := flow.Edges

	for gf.Len() < len(nodes) {
		for _, n := range nodes {
			if gf.NodeByID(n.ID).ID != "" { // has resolved
				continue
			}

			incoming := make([]wf.Edge, 0)
			outgoing := make([]wf.Edge, 0)
			for _, edge := range edges {
				if edge.Source == n.ID {
					outgoing = append(outgoing, edge)
				} else if edge.Target == n.ID {
					incoming = append(incoming, edge)
				}
			}
			gf.Incoming[n.ID] = incoming
			gf.Outgoing[n.ID] = outgoing
			// validate
			gf.AddNode(n)
			if n.Kind == wf.NodeKindStart { // or trigger
				gf.AddStarter(n)
			}
		}
	}

	for _, edge := range edges {
		if n := gf.NodeByID(edge.Target); n.ID == "" {
			// error
			continue
		}
		if n := gf.NodeByID(edge.Source); n.ID == "" {
			// error
			continue
		}
		gf.AddParent(
			gf.NodeByID(edge.Target),
			gf.NodeByID(edge.Source),
		)
	}

	return gf, nil
}
