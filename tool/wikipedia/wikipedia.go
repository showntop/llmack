package wikipedia

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/showntop/llmack/tool"
	gowiki "github.com/trietmn/go-wiki"
)

func init() {
	t := tool.New(
		tool.WithName("wikipedia_search"),
		tool.WithKind("code"),
		tool.WithDescription("A tool for performing a Wikipedia search and extracting snippets and webpages. Input should be a search query."),
		tool.WithParameters(
			tool.Parameter{
				Name: "query", Type: tool.String, Required: true, LLMDescrition: `key words for searching, this should be in the language of "language" parameter`,
			},
			tool.Parameter{
				Name: "language", Type: tool.String, Required: true, LLMDescrition: `
		language of the wikipedia to be searched,
		only "de" for German,
		"en" for English,
		"fr" for French,
		"hi" for Hindi,
		"ja" for Japanese,
		"ko" for Korean,
		"pl" for Polish,
		"pt" for Portuguese,
		"ro" for Romanian,
		"uk" for Ukrainian,
		"vi" for Vietnamese,
		and "zh" for Chinese are supported
		`,
				Options: []string{"de", "en", "fr", "hi", "ja", "ko", "pl", "pt", "ro", "uk", "vi", "zh"},
			},
		),
		tool.WithFunction(Invoke),
	)
	tool.Register(t)
}

// Invoke ...
func Invoke(ctx context.Context, args string) (string, error) {
	var params struct {
		Query    string `json:"query"`
		Language string `json:"language"`
	}
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", err
	}

	if params.Query == "" {
		return "Please input query", fmt.Errorf("query is empty")
	}

	// Search for the Wikipedia page title
	searchResult, _, err := gowiki.Search(params.Query, 3, false)
	if err != nil {
		return "", fmt.Errorf("search failed: %v", err)
	}

	if len(searchResult) == 0 {
		return "No search results found", nil
	}

	// Get the first page
	page, err := gowiki.GetPage(searchResult[0], -1, false, true)
	if err != nil {
		return "", fmt.Errorf("failed to get page: %v", err)
	}

	// Get the content of the page
	content, err := page.GetContent()
	if err != nil {
		return "", fmt.Errorf("failed to get content: %v", err)
	}

	return content, nil
}
