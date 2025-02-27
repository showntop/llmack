package weather

import (
	"context"

	"github.com/showntop/llmack/tool"
)

const QueryWeather = "QueryWeather"

func init() {
	t := &tool.Tool{}
	t.Name = QueryWeather
	t.Kind = "code"
	t.Description = "查询天气"
	t.Parameters = append(t.Parameters, tool.Parameter{
		Name: "city", Type: tool.String, Required: true, LLMDescrition: "城市", Default: "北京",
	})
	t.Invokex = func(ctx context.Context, args map[string]any) (string, error) {
		return "晴朗，北风三级，2-15 摄氏度，空气质量优。", nil
	}
	tool.Register(t)
}
