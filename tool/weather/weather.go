package weather

import (
	"context"

	"github.com/showntop/llmack/tool"
)

const QueryWeather = "QueryWeather"

func init() {
	t := tool.New(
		tool.WithName(QueryWeather),
		tool.WithKind("code"),
		tool.WithDescription("查询天气"),
		tool.WithParameters(tool.Parameter{
			Name: "city", Type: tool.String, Required: true, LLMDescrition: "城市", Default: "北京",
		}),
		tool.WithFunction(func(ctx context.Context, args string) (string, error) {
			return "北京晴朗，北风三级，2-15 摄氏度，空气质量优。", nil
		}),
	)
	tool.Register(t)
}
