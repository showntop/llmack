package serper

import (
	"context"
	"encoding/json"

	"github.com/showntop/llmack/tool"
)

func init() {
	t := tool.CodeTool{}
	t.Meta.Name = "multi_search"
	t.Meta.Description = "智能搜索引擎，从多个搜索引擎全网抓取信息，合并提取"
	t.Meta.Parameters = append(t.Meta.Parameters, tool.Parameter{
		Name:          "query",
		Type:          tool.String,
		Required:      true,
		LLMDescrition: "查询的关键词",
		Default:       "",
	})
	t.Invokex = func(ctx context.Context, args map[string]any) (string, error) {
		query, _ := args["query"].(string)
		engine := NewSerper()
		result, err := engine.Search(ctx, query)
		if err != nil {
			return "", err
		}
		bytes, err := json.Marshal(result)
		if err != nil {
			return "", err
		}
		return string(bytes), nil
	}
	tool.Register(t.Name(), &t)
}
