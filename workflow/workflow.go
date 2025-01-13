package workflow

// Result ...
type Result struct {
	Outputs map[string]any
}

// Workflow ...
type Workflow struct {
	ID          int64
	Name        string
	Description string
	Key         string
	Version     int64
	Nodes       []Node
	Edges       []Edge
	// Metadata    metadata
}

// Parameters ...
type Parameters map[string]Parameter

// Parameter ...
type Parameter struct {
	Name  string `json:"name"`  // 参数名称
	Value string `json:"value"` // 参数值
	Type  string `json:"type"`  // 参数类型
}

// Nodes ...
type Nodes []Node

// Event ...
type Event struct {
	Name string `json:"name"` // 事件名称
	Data any    `json:"data"` // 事件值
	Type string `json:"type"` // 事件类型
}
