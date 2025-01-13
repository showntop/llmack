package spider

import (
	"context"
	"encoding/json"

	"github.com/showntop/llmack/log"
	"github.com/showntop/llmack/tool"
)

func init() {
	t := tool.CodeTool{}
	t.Meta.Name = "spider"
	t.Meta.Description = "爬虫，爬取网页内容"
	t.Meta.Parameters = append(t.Meta.Parameters, tool.Parameter{
		Name:          "urls",
		Type:          tool.Array,
		Required:      true,
		LLMDescrition: "需要爬取的网页列表",
		Default:       "",
	})
	t.Invokex = func(ctx context.Context, args map[string]any) (string, error) {
		log.InfoContextf(ctx, "tool spider invoke with args: %+v", args)
		urls, _ := args["urls"].([]string)
		if len(urls) == 0 {
			urlsx, _ := args["urls"].([]any)
			for _, url := range urlsx {
				urls = append(urls, url.(string))
			}
		}
		spd := NewSpider()
		result, err := spd.Crawl(ctx, urls)
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
