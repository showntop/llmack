package user

import (
	"context"

	"github.com/showntop/llmack/tool"
)

const Inquery = "Inquery"

func init() {
	t := &tool.Tool{}
	t.Name = Inquery
	t.Kind = "code"
	t.Description = "询问用户的具体需求以提供个性化的服务"
	t.Parameters = append(t.Parameters, tool.Parameter{})
	t.Invokex = func(ctx context.Context, args map[string]any) (string, error) {
		return "小猫钓鱼", nil
	}
	tool.Register(t)
}
