package user

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/showntop/llmack/tool"
)

const Inquery = "Inquery"

func init() {
	t := tool.New(
		tool.WithName(Inquery),
		tool.WithKind("code"),
		tool.WithDescription("询问用户的具体需求以提供个性化的服务"),
		tool.WithParameters(tool.Parameter{
			Name:          "inquiry",
			LLMDescrition: "询问用户的具体需求以提供个性化的服务",
			Required:      true,
			Type:          tool.String,
		}),
		tool.WithFunction(func(ctx context.Context, args string) (string, error) {
			var params struct {
				Inquiry string `json:"inquiry"`
			}
			if err := json.Unmarshal([]byte(args), &params); err != nil {
				return "", err
			}
			// 等待用户输入
			var input string
			fmt.Printf("请根据提示输入你的需求 (%s): ", params.Inquiry)
			fmt.Scanf("%s", &input)
			return input, nil
		}),
	)
	tool.Register(t)
}
