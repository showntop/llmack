package ai

import (
	"context"

	"github.com/showntop/llmack/tool"
)

func init() {
	t := tool.CodeTool{}
	t.Meta.Name = "ai_search"
	t.Meta.Description = "智能搜索引擎，从多个搜索引擎全网抓取信息，合并提取"
	t.Meta.Parameters = append(t.Meta.Parameters, tool.Parameter{
		Name:          "query",
		Type:          tool.String,
		Required:      true,
		LLMDescrition: "查询的关键词",
		Default:       "",
	})
	t.Streamx = func(ctx context.Context, args map[string]any) (chan any, error) {
		query, _ := args["query"].(string)

		agent := NewAgent(nil)
		stream, err := agent.Stream(ctx, query)
		if err != nil {
			return nil, err
		}
		return stream, nil
	}
	tool.Register(t.Name(), &t)
}
