package user

import (
	"context"
	"fmt"

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
		// 等待用户输入
		var input string
		fmt.Println("请根据提示输入你的需求:")
		fmt.Scanf("%s", &input)
		return input, nil
	}
	tool.Register(t)
}
