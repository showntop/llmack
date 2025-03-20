package memory

import (
	"sync"

	"github.com/showntop/llmack/workflow"
)

// Graph 表示工作流的DAG图结构
type Graph struct {
	queue []string // 用于存储待执行的节点

	nodes     map[string]*workflow.Node
	edges     map[string][]*workflow.Edge // sourceID -> []targetID
	completed map[string]bool             // 记录节点完成状态
	inDegree  map[string]int              // 记录节点入度
	mu        sync.RWMutex                // 保护并发访问
}

// StartNodes 获取所有入度为0的起始节点
func (g *Graph) StartNodes() []*workflow.Node {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var startNodes []*workflow.Node
	for nodeID, inDegree := range g.inDegree {
		if inDegree == 0 {
			if node, exists := g.nodes[nodeID]; exists {
				startNodes = append(startNodes, node)
			}
		}
	}
	return startNodes
}

// Dequeue 获取指定节点
func (g *Graph) Dequeue() *workflow.Node {
	g.mu.Lock()
	defer g.mu.Unlock()

	if len(g.queue) == 0 {
		return nil
	}

	nodeID := g.queue[0]
	g.queue = g.queue[1:]
	return g.nodes[nodeID]
}

// Enqueue 添加节点到队列
func (g *Graph) Enqueue(nodeIDs ...string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.queue = append(g.queue, nodeIDs...)
}

// Finished 检查是否所有节点都已完成
func (g *Graph) Finished() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return len(g.queue) == 0
}

// AreAllDependenciesCompleted 检查节点的所有前置依赖是否已完成
func (g *Graph) AreAllDependenciesCompleted(nodeID string) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()

	// 遍历所有边，找到指向当前节点的边的源节点
	for source, edges := range g.edges {
		for _, target := range edges {
			if target.Target == nodeID {
				// 如果有任何一个前置节点未完成，返回false
				if !g.completed[source] {
					return false
				}
			}
		}
	}
	return true
}

// MarkNodeCompleted 标记节点已完成
func (g *Graph) MarkNodeCompleted(nodeID string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.completed[nodeID] = true
}

// NextNodes 获取指定节点的所有后续节点
func (g *Graph) NextNodes(nodeID string) []*workflow.Node {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var nextNodes []*workflow.Node
	if edges, exists := g.edges[nodeID]; exists {
		for _, edge := range edges {
			if node, exists := g.nodes[edge.Target]; exists {
				nextNodes = append(nextNodes, node)
			}
		}
	}
	return nextNodes
}

// NewGraph 创建一个新的DAG图
func NewGraph(nodes []workflow.Node, edges []workflow.Edge) *Graph {
	g := &Graph{
		nodes:     make(map[string]*workflow.Node),
		edges:     make(map[string][]*workflow.Edge),
		completed: make(map[string]bool),
		inDegree:  make(map[string]int),
	}

	// 初始化图结构
	for _, node := range nodes {
		g.nodes[node.ID] = &node
		g.inDegree[node.ID] = 0
	}

	for _, edge := range edges {
		g.edges[edge.Source] = append(g.edges[edge.Source], &edge)
		g.inDegree[edge.Target]++
	}

	// 初始化队列，把入度为0的节点加入队列
	for _, node := range nodes {
		if g.inDegree[node.ID] == 0 {
			g.queue = append(g.queue, node.ID)
		}
	}

	return g
}
