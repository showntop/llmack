package dag

import (
	wf "github.com/showntop/llmack/workflow"
)

type (
	// Graph ...
	// relations of Nodes
	// list of Graph Nodes with relations
	Graph struct {
		ID          string
		Metadata    map[string]interface{}
		StartNode   wf.Node // now only support one starter
		Nodes       wf.Nodes
		ChildNodes  map[string]wf.Nodes
		ParentNodes map[string]wf.Nodes
		Index       map[string]wf.Node   // index for quick get one node
		Incoming    map[string][]wf.Edge // index for incoming nodes
		Outgoing    map[string][]wf.Edge // index for outgoing nodes

		// Listeners map[string]wf.Listener
		PreNodes  []wf.Node
		PostNodes []wf.Node
	}
)

// NewGraph TODO
func NewGraph() *Graph {
	wf := &Graph{
		Nodes:       make(wf.Nodes, 0, 1024),
		ChildNodes:  make(map[string]wf.Nodes),
		ParentNodes: make(map[string]wf.Nodes),
		Index:       make(map[string]wf.Node),
		Incoming:    make(map[string][]wf.Edge),
		Outgoing:    make(map[string][]wf.Edge),
	}

	return wf
}

// Starter Start Node
func (g *Graph) Starter() wf.Node {
	return g.StartNode
}

// Len TODO
func (g *Graph) Len() int {
	return len(g.Nodes)
}

// NodeByID TODO
func (g *Graph) NodeByID(ID string) wf.Node {
	return g.Index[ID]
}

// AddNode TODO
func (g *Graph) AddNode(s wf.Node, cc ...wf.Node) {
	if g.NodeByID(s.ID).ID == "" {
		g.Nodes = append(g.Nodes, s)
	}

	if id := s.ID; id != "" {
		g.Index[id] = s
	}

	if len(cc) > 0 {
		for _, c := range cc {
			g.AddParent(c, s)
		}
	}
}

// AddStarter TODO
func (g *Graph) AddStarter(s wf.Node) {
	g.StartNode = s
}

// AddParent TODO
func (g *Graph) AddParent(c, p wf.Node) {
	g.ParentNodes[c.ID] = append(g.ParentNodes[c.ID], p)
	g.ChildNodes[p.ID] = append(g.ChildNodes[p.ID], c)
}

// Children TODO
func (g *Graph) Children(s wf.Node) wf.Nodes {
	return g.ChildNodes[s.ID]
}

// Parents TODO
func (g *Graph) Parents(s wf.Node) wf.Nodes {
	return g.ParentNodes[s.ID]
}
