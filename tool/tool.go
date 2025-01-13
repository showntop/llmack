package tool

import (
	"context"
	"fmt"
)

// Tool 工具接口
type Tool interface {
	Name() string
	Description() string
	Parameters() map[string]any
	Kind() string
	Invoke(context.Context, map[string]any) (string, error)
	Stream(context.Context, map[string]any) (<-chan any, error)
}

// NilTool 无工具
type NilTool struct {
	ID int64
}

// Kind 工具类型
func (t *NilTool) Kind() string {
	return "nil"
}

// Name 工具名称
func (t *NilTool) Name() string {
	return "nil"
}

// Description 工具名称
func (t *NilTool) Description() string {
	return "a nil tool when exception"
}

// Parameters 工具名称
func (t *NilTool) Parameters() map[string]any {
	return nil
}

// Invoke 调用工具
func (t *NilTool) Invoke(ctx context.Context, params map[string]any) (string, error) {
	return "", fmt.Errorf("tool not found")
}

// Stream 流式调用工具
func (t *NilTool) Stream(ctx context.Context, params map[string]any) (<-chan any, error) {
	return nil, fmt.Errorf("tool not found")
}
