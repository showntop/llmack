package node

import (
	"fmt"

	"github.com/showntop/llmack/workflow"
)

// Build 根据节点类型构建节点
func Build(n *workflow.Node, outgoing ...workflow.Edge) (Node, error) {

	switch n.Kind {
	case workflow.NodeKindStart:
		return StartNode(n), nil
	case workflow.NodeKindLLM:
		return LLMNode(n), nil
	case workflow.NodeKindGateway:
		switch n.Subref {
		case "exclusive":
			return ExclGateway(n, outgoing...)
		case "inclusive":
			return ForkGateway(n, outgoing...)
		case "parallel":
			return ForkGateway(n, outgoing...)
		}
	case workflow.NodeKindWait:
		// return WaitNode(n)
	case workflow.NodeKindHuman:
		switch n.Subref {
		case "approval":
		case "human":
		}
	case workflow.NodeKindEnd:
		return EndNode(n)
	case workflow.NodeKindTool:
		return ToolNode(n)
	case workflow.NodeKindExpr:
		return ExprNode(n), nil
	}
	return nil, fmt.Errorf("unknown workflow node kind %q", n.Kind)
}
