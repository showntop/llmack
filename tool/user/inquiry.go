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
	t.Parameters = append(t.Parameters, tool.Parameter{
		Name:          "properties",
		LLMDescrition: "用户需要提供的具体信息",
		Type:          tool.String,
		Required:      true,
	})
	t.Invokex = func(ctx context.Context, args map[string]any) (string, error) {
		// 等待用户输入
		properties := args["properties"].(string)
		fmt.Println("请输入 " + properties + ":")
		var input string
		fmt.Scanf("%s", &input)
		// for {
		// fmt.Scanf("%s", &input)
		// fmt.Scanln(input)
		// if input != "" {
		// 	break
		// }
		// }
		return input, nil
	}
	tool.Register(t)
}
