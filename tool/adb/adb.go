package adb

type AdbTool struct {
	controller *Controller
}

func NewTools() []any {
	return registry.AvailableTools(nil)
}

type ToolParams struct {
	Thought *AgentThought
	Actions []map[string]any
}

type AgentThought struct {
	EvaluationPreviousGoal string
	Memory                 string
	CurrentGoal            string
}
