package workflow

// NodeKind TODO
type NodeKind string

const (
	// NodeKindTool TODO
	NodeKindTool NodeKind = "tool" // ref = <tool ref>
	// NodeKindExpr TODO
	NodeKindExpr NodeKind = "expr" // ref = <expr>
	// NodeKindLLM TODO
	NodeKindLLM NodeKind = "llm" // ref = <llm ref>
	// NodeKindGateway TODO
	NodeKindGateway NodeKind = "gateway" // ref = join|fork|excl|incl
	// NodeKindIterator TODO
	NodeKindIterator NodeKind = "iterator" // ref = <iterator function ref>
	// NodeKindError TODO
	NodeKindError NodeKind = "error" // no ref
	// NodeKindTermination TODO
	NodeKindTermination NodeKind = "termination" // no ref
	// NodeKindPrompt TODO
	NodeKindPrompt NodeKind = "prompt" // ref = <client function>
	// NodeKindDelay TODO
	NodeKindDelay NodeKind = "delay" // no ref
	// NodeKindWait TODO
	NodeKindWait NodeKind = "wait" // no ref
	// NodeKindVisual TODO
	NodeKindVisual NodeKind = "visual" // ref = <*>
	// NodeKindDebug TODO
	NodeKindDebug NodeKind = "debug" // ref = <*>
	// NodeKindBreak TODO
	NodeKindBreak NodeKind = "break" // ref = <*>
	// NodeKindContinue TODO
	NodeKindContinue NodeKind = "continue" // ref = <*>
	// NodeKindStart TODO
	NodeKindStart NodeKind = "start" // start event
	// NodeKindEnd TODO
	NodeKindEnd NodeKind = "end" // end event
	// NodeKindHuman TODO
	NodeKindHuman NodeKind = "human" // user involved in node
)

// Node ...
type Node struct {
	ID          string   `json:"id"`          // 节点ID
	Name        string   `json:"name"`        // 节点名称 support expr @TODO
	Description string   `json:"description"` // 节点描述
	Kind        NodeKind `json:"kind"`        // 节点类型
	Subref      string   `json:"subref"`      // reference to function or subprocess (gateway)
	// set of expressions to evaluate, test or pass to function
	Inputs  Parameters `json:"inputs"`
	Outputs Parameters `json:"outputs"`

	// Events    []Event   `json:"events2"` // 边界事件
	// Callbacks []*Action `json:"events"`  // 普通事件
	// for business configure
	// its free now, need tobe constraint in the future
	Metadata map[string]any `json:"metadata"`
}
