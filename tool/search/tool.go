package search

import (
	"context"
	"encoding/json"

	"github.com/showntop/llmack/tool"
	"github.com/showntop/llmack/tool/search/engine"
)

const Serper = "SerperSearch"
const Searxng = "SearxngSearch"

func init() {
	registSerper()
	registSearxng()
}

func registSerper() {
	t := &tool.Tool{}
	t.Name = Serper
	t.Kind = "code"
	t.Description = "A tool for performing a Google SERP search and extracting snippets and webpages.Input should be a search query."
	t.Parameters = append(t.Parameters, tool.Parameter{
		Name:          "query",
		Type:          tool.String,
		Required:      true,
		LLMDescrition: "The search query for Google SERP.",
		Default:       "",
	})
	t.Invokex = func(ctx context.Context, args map[string]any) (string, error) {

		apiKey := tool.DefaultConfig.GetString("serper.api_key")
		query, _ := args["query"].(string)
		engine := engine.NewSerper(apiKey, "")
		results, err := engine.Search(ctx, query)
		if err != nil {
			return "", err
		}
		// crawl detail
		urls := []string{}
		for i := 0; i < len(results); i++ {
			urls = append(urls, results[i].Link)
		}
		details, err := CrawlWebpage(ctx, urls)
		if err != nil {
			return "", err
		}

		for i := 0; i < len(results); i++ {
			results[i].Content = details[i]
		}
		bytes, err := json.Marshal(results)
		if err != nil {
			return "", err
		}
		return string(bytes), nil
	}
	tool.Register(t)
}
func registSearxng() {

	t := &tool.Tool{}
	t.Name = Searxng
	t.Kind = "code"
	t.Description = "A tool for performing a Searx search and extracting snippets and webpages.Input should be a search query."
	t.Parameters = append(t.Parameters, tool.Parameter{
		Name:          "query",
		Type:          tool.String,
		Required:      true,
		LLMDescrition: "The search query for the Searx search engine.",
		Default:       "",
	})
	t.Invokex = func(ctx context.Context, args map[string]any) (string, error) {
		query, _ := args["query"].(string)
		baseUrl := tool.DefaultConfig.GetString("searxng.base_url")

		engine := engine.NewSearxng(baseUrl)
		results, err := engine.Search(ctx, query)
		if err != nil {
			return "", err
		}
		// crawl detail
		urls := []string{}
		for i := 0; i < len(results); i++ {
			urls = append(urls, results[i].Link)
		}
		details, err := CrawlWebpage(ctx, urls)
		if err != nil {
			return "", err
		}

		for i := 0; i < len(results); i++ {
			results[i].Content = details[i]
		}
		bytes, err := json.Marshal(results)
		if err != nil {
			return "", err
		}
		return string(bytes), nil

	}
	tool.Register(t)
}
