package search

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/showntop/llmack/log"
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
		url := baseUrl + "/search"

		payload := strings.NewReader(fmt.Sprintf(`{"q":"%s","gl":"cn"}`, query))

		req, err := http.NewRequest(http.MethodGet, url, payload)

		if err != nil {
			return "", err
		}
		req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux i686; rv:109.0) Gecko/20100101 Firefox/114.0")
		req.Header.Add("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		log.InfoContextf(ctx, "searxng search result: %s", string(body))
		return string(body), nil
	}
	tool.Register(t)
}
