package tool

import (
	"context"

	"github.com/showntop/llmack/log"
)

// CodeTool 结构体，用于表示基于API的工具
type CodeTool struct {
	Meta
	Invokex func(context.Context, map[string]any) (string, error)
	Streamx func(context.Context, map[string]any) (chan any, error)
}

// Kind 返回工具类型
func (t *CodeTool) Kind() string {
	return "code"
}

// Name 返回工具名称
func (t *CodeTool) Name() string {
	return t.Meta.Name
}

// Description 返回工具名称
func (t *CodeTool) Description() string {
	return t.Meta.Description
}

// Parameters 返回工具参数
func (t *CodeTool) Parameters() map[string]any {
	xxx := make(map[string]any, len(t.Meta.Parameters))
	for i := 0; i < len(t.Meta.Parameters); i++ {
		xxx[t.Meta.Parameters[i].Name] = t.Meta.Parameters[i]
	}
	return xxx
}

// Invoke 调用工具
func (t *CodeTool) Invoke(ctx context.Context, args map[string]any) (string, error) {
	log.InfoContextf(ctx, "code tool invoke with args: %v", args)
	return t.Invokex(ctx, args)
}

// Stream 调用工具
func (t *CodeTool) Stream(ctx context.Context, args map[string]any) (<-chan any, error) {
	return t.Streamx(ctx, args)
}

// NewCodeTool TODO
func NewCodeTool(name string) Tool {
	x, ok := CodeTools[name]
	if !ok {
		return &NilTool{Target: name}
	}
	return x
}

// CodeTools 工具注册表
var CodeTools map[string]*CodeTool = make(map[string]*CodeTool)

// Register 注册工具
func Register(name string, tool *CodeTool) {
	CodeTools[name] = tool
}

// GetCodeTool 获取工具元信息
func GetCodeTool(name string) *CodeTool {
	return CodeTools[name]
}
