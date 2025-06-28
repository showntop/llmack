package search

import (
	"context"
	"encoding/json"
	"errors"

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
	t := tool.New(
		tool.WithName(DuckDuckGo),
		tool.WithDescription("A tool for performing a DuckDuckGo search and extracting snippets and webpages. Input should be a search query."),
		tool.WithParameters(
			tool.Parameter{
				Name:          "query",
				Type:          tool.String,
				Required:      true,
				LLMDescrition: "The search query for duckduckgo search.",
				Default:       "",
			},
		),
		tool.WithFunction(func(ctx context.Context, params string) (string, error) {
			var args map[string]any
			err := json.Unmarshal([]byte(params), &args)
			if err != nil {
				return "", err
			}
			query, _ := args["query"].(string)
			ddg := client.NewDuckDuckGoSearchClient()
			originResults, err := ddg.SearchLimited(query, 100)
			if err != nil {
				return "", err
			}
			// crawl detail
			urls := []string{}
			for i := range originResults {
				urls = append(urls, originResults[i].FormattedUrl)
			}
			details, err := CrawlWebpage(ctx, urls)
			if err != nil {
				return "", err
			}
			var results []engine.Result = make([]engine.Result, len(originResults))
			for i := range originResults {
				results[i].Content = details[i]
				results[i].Title = originResults[i].Title
				results[i].Snippet = originResults[i].Snippet
			}
			bytes, err := json.Marshal(results)
			if err != nil {
				return "", err
			}
			return string(bytes), nil
		}),
	)
	tool.Register(t)
}

func registSerper() {
	t := tool.New(
		tool.WithName(Serper),
		tool.WithDescription("A tool for performing a Google SERP search and extracting snippets and webpages.Input should be a search query."),
		tool.WithParameters(
			tool.Parameter{
				Name:          "query",
				Type:          tool.String,
				Required:      true,
				LLMDescrition: "The search query for Google SERP.",
			},
		),
		tool.WithFunction(func(ctx context.Context, params string) (string, error) {
			var args map[string]any
			err := json.Unmarshal([]byte(params), &args)
			if err != nil {
				return "", err
			}

			apiKey := tool.DefaultConfig.GetString("serper.api_key")
			query, ok := args["query"].(string)
			if !ok {
				return "", errors.New("query is not a string")
			}

			engine := engine.NewSerper(apiKey, "search")
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
		}),
	)
	tool.Register(t)
}
func registSearxng() {

	t := tool.New(
		tool.WithName(Searxng),
		tool.WithDescription("A tool for performing a Searx search and extracting snippets and webpages.Input should be a search query."),
		tool.WithParameters(
			tool.Parameter{
				Name:          "query",
				Type:          tool.String,
				Required:      true,
				LLMDescrition: "The search query for the Searx search engine.",
				Default:       "",
			},
		),
		tool.WithFunction(func(ctx context.Context, params string) (string, error) {
			var args map[string]any
			err := json.Unmarshal([]byte(params), &args)
			if err != nil {
				return "", err
			}
			query, ok := args["query"].(string)
			if !ok {
				return "", errors.New("query is not a string")
			}
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
		}),
	)
	tool.Register(t)
}
