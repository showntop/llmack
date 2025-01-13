package weather

import (
	"context"

	"github.com/showntop/llmack/tool"
)

func init() {
	t := tool.CodeTool{}
	t.Meta.Name = "weather"
	t.Meta.Description = "weather"
	t.Meta.Parameters = append(t.Meta.Parameters, tool.Parameter{
		Name: "city", Type: tool.String, Required: true, LLMDescrition: "城市", Default: "北京",
	})
	t.Invokex = func(ctx context.Context, args map[string]any) (string, error) {
		return "晴朗", nil
	}
	tool.Register(t.Name(), &t)
}
