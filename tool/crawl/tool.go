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
	t := tool.New(
		tool.WithName(name),
		tool.WithKind("code"),
		tool.WithDescription("Used to scrape website urls and extract text content"),
		tool.WithParameters(tool.Parameter{
			Name:          "link",
			Type:          tool.String,
			Required:      true,
			LLMDescrition: "Valid website url without any quotes.",
			Default:       "",
		}),
		tool.WithFunction(func(ctx context.Context, args string) (string, error) {
			var params struct {
				Link string `json:"link"`
			}
			if err := json.Unmarshal([]byte(args), &params); err != nil {
				return "", err
			}
			engine, ok := Crawlers[name]
			if !ok {
				return "", fmt.Errorf("crawler %s not found", name)
			}
			result, err := engine.Crawl(ctx, params.Link)
			if err != nil {
				return "", err
			}
			bytes, err := json.Marshal(result)
			if err != nil {
				return "", err
			}
			return string(bytes), nil
		}),
	)
	tool.Register(t)
}
