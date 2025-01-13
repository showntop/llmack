package crawl

import (
	"context"
	"encoding/json"

	"github.com/showntop/llmack/tool"
)

func init() {
	t := tool.CodeTool{}
	t.Meta.Name = "jina"
	t.Meta.Description = "从jina爬取网页信息"
	t.Meta.Parameters = append(t.Meta.Parameters, tool.Parameter{
		Name:          "link",
		Type:          tool.String,
		Required:      true,
		LLMDescrition: "要爬取的网站链接",
		Default:       "",
	})
	t.Invokex = func(ctx context.Context, args map[string]any) (string, error) {
		link, _ := args["link"].(string)
		engine := NewJinaCrawler()
		result, err := engine.Crawl(ctx, link)
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
