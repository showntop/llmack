package search

import (
	"context"
	"encoding/json"

	"github.com/sap-nocops/duckduckgogo/client"
	"github.com/showntop/llmack/tool"
	"github.com/showntop/llmack/tool/search/engine"
)

const Serper = "SerperSearch"
const Searxng = "SearxngSearch"
const DuckDuckGo = "DuckDuckGoSearch"

func init() {
	registDuckDuckGo()
	registSerper()
	registSearxng()
}

func registDuckDuckGo() {
	t := &tool.Tool{}
	t.Name = DuckDuckGo
	t.Kind = "code"
	t.Description = "A tool for performing a DuckDuckGo search and extracting snippets and webpages. Input should be a search query."
	t.Parameters = append(t.Parameters, tool.Parameter{
		Name:          "query",
		Type:          tool.String,
		Required:      true,
		LLMDescrition: "The search query for duckduckgo search.",
		Default:       "",
	})
	t.Invokex = func(ctx context.Context, args map[string]any) (string, error) {
		query, _ := args["query"].(string)
		ddg := client.NewDuckDuckGoSearchClient()
		originResults, err := ddg.SearchLimited(query, 10)
		if err != nil {
			return "", err
		}

		// crawl detail
		urls := []string{}
		for i := 0; i < len(originResults); i++ {
			urls = append(urls, originResults[i].FormattedUrl)
		}
		details, err := CrawlWebpage(ctx, urls)
		if err != nil {
			return "", err
		}
		var results []engine.Result = make([]engine.Result, len(originResults))
		for i := 0; i < len(originResults); i++ {
			results[i].Content = details[i]
			results[i].Title = originResults[i].Title
			results[i].Snippet = originResults[i].Snippet
		}
		bytes, err := json.Marshal(results)
		if err != nil {
			return "", err
		}
		return string(bytes), nil
	}
	tool.Register(t)
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
