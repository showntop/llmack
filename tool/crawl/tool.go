package crawl

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/showntop/llmack/tool"
)

var (
	Jina  = "Jina"
	Colly = "Colly"
)

func init() {
	register(Jina)
	register(Colly)
}

func register(name string) {
	t := &tool.Tool{}
	t.Name = name
	t.Kind = "code"
	t.Description = "Used to scrape website urls and extract text content"
	t.Parameters = append(t.Parameters, tool.Parameter{
		Name:          "link",
		Type:          tool.String,
		Required:      true,
		LLMDescrition: "Valid website url without any quotes.",
		Default:       "",
	})
	t.Invokex = func(ctx context.Context, args map[string]any) (string, error) {
		link, _ := args["link"].(string)
		engine, ok := Crawlers[name]
		if !ok {
			return "", fmt.Errorf("crawler %s not found", name)
		}
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
	tool.Register(t)
}
